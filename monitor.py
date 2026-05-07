#!/usr/bin/env python3
"""
DNS Monitor — автоматический мониторинг DNS-серверов из Telegram-канала.
Парсит IPv4 из постов, проверяет DNS через asyncio, сохраняет статистику,
генерирует README.md с топ-10.
"""

import asyncio
import json
import logging
import os
import re
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

import dns.asyncresolver
import dns.exception
from telethon import TelegramClient
from telethon.errors import SessionPasswordNeededError, AuthKeyError
from telethon.sessions import StringSession

# ─────────────────────────── Logging ────────────────────────────────────────
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%H:%M:%S",
)
log = logging.getLogger("dns-monitor")

# ─────────────────────────── Config ─────────────────────────────────────────
TG_API_ID      = int(os.environ["TG_API_ID"])
TG_API_HASH    = os.environ["TG_API_HASH"]
TG_SESSION_STR = os.environ.get("TG_SESSION_STRING", "")

TG_CHANNEL     = "https://t.me/dns_lists"          # целевой канал
MESSAGE_LIMIT  = 50                                  # последних постов
STATS_FILE     = Path("dns_stats.json")
README_FILE    = Path("README.md")

PROBE_DOMAINS  = ["google.com", "cloudflare.com"]
DNS_TIMEOUT    = 3.0     # секунды
DNS_PORT       = 53
PROBES_PER_DNS = 3       # повторений для усреднения latency
CONCURRENCY    = 10      # семафор параллельных проверок
TOP_N          = 10      # строк в таблице README

IPv4_RE = re.compile(
    r"\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}"
    r"(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\b"
)

# Исключаем очевидно «не-DNS» адреса (loopback, link-local, broadcast …)
EXCLUDE_PREFIXES = ("0.", "127.", "169.254.", "224.", "240.", "255.")


# ═══════════════════════════════════════════════════════════════════════════
#  Telegram — парсинг IPv4
# ═══════════════════════════════════════════════════════════════════════════

async def fetch_dns_from_telegram() -> list[str]:
    """Возвращает дедуплицированный список IPv4 из последних постов канала."""
    log.info("📡  Подключаемся к Telegram …")

    session = StringSession(TG_SESSION_STR) if TG_SESSION_STR else StringSession()

    try:
        async with TelegramClient(session, TG_API_ID, TG_API_HASH) as client:
            log.info("✅  Telegram-сессия установлена")
            found: set[str] = set()

            async for msg in client.iter_messages(TG_CHANNEL, limit=MESSAGE_LIMIT):
                if not msg.text:
                    continue
                for ip in IPv4_RE.findall(msg.text):
                    if not any(ip.startswith(p) for p in EXCLUDE_PREFIXES):
                        found.add(ip)

            log.info("🔍  Найдено уникальных IP: %d", len(found))
            return sorted(found)

    except AuthKeyError as exc:
        log.error("❌  Ошибка авторизации Telegram: %s", exc)
        raise
    except SessionPasswordNeededError:
        log.error("❌  Требуется двухфакторный пароль — используйте SESSION_STRING")
        raise
    except Exception as exc:
        log.error("❌  Ошибка Telegram: %s", exc)
        raise


# ═══════════════════════════════════════════════════════════════════════════
#  DNS — проверка одного сервера
# ═══════════════════════════════════════════════════════════════════════════

async def check_dns(ip: str, semaphore: asyncio.Semaphore) -> dict:
    """
    Проверяет DNS-сервер: выполняет PROBES_PER_DNS A-запросов для каждого
    домена в PROBE_DOMAINS, замеряет latency и packet loss.
    """
    async with semaphore:
        latencies: list[float] = []
        total_queries = 0
        failed_queries = 0

        resolver = dns.asyncresolver.Resolver(configure=False)
        resolver.nameservers = [ip]
        resolver.port = DNS_PORT
        resolver.timeout = DNS_TIMEOUT
        resolver.lifetime = DNS_TIMEOUT

        for domain in PROBE_DOMAINS:
            for _ in range(PROBES_PER_DNS):
                total_queries += 1
                t0 = time.perf_counter()
                try:
                    await resolver.resolve(domain, "A")
                    latencies.append((time.perf_counter() - t0) * 1000)
                except (dns.exception.Timeout, dns.exception.DNSException) as exc:
                    failed_queries += 1
                    log.debug("  ⚠  %s → %s: %s", ip, domain, exc)
                except Exception as exc:
                    failed_queries += 1
                    log.debug("  ⚠  %s → %s: %s", ip, domain, exc)

        avg_latency   = sum(latencies) / len(latencies) if latencies else None
        packet_loss   = (failed_queries / total_queries * 100) if total_queries else 100.0
        reachable     = avg_latency is not None

        status = "✅" if reachable and packet_loss < 20 else (
                 "⚠️" if reachable else "❌")

        log.info(
            "  %s  %-16s  latency=%s ms  loss=%.0f%%",
            status, ip,
            f"{avg_latency:.1f}" if avg_latency else "N/A",
            packet_loss,
        )

        return {
            "ip": ip,
            "reachable": reachable,
            "avg_latency_ms": round(avg_latency, 2) if avg_latency else None,
            "packet_loss_pct": round(packet_loss, 1),
        }


# ═══════════════════════════════════════════════════════════════════════════
#  Статистика — load / merge / save
# ═══════════════════════════════════════════════════════════════════════════

def load_stats() -> dict:
    if STATS_FILE.exists():
        try:
            return json.loads(STATS_FILE.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            log.warning("⚠  dns_stats.json повреждён — начинаем заново")
    return {}


def merge_stats(existing: dict, results: list[dict]) -> dict:
    """
    Обновляет накопленную статистику новыми замерами.
    Инкрементирует счётчик проверок, пересчитывает скользящее среднее
    latency и uptime_percentage.
    """
    updated = dict(existing)

    for r in results:
        ip  = r["ip"]
        old = updated.get(ip, {
            "ip": ip,
            "checks": 0,
            "successful_checks": 0,
            "avg_latency_ms": None,
            "packet_loss_pct": 100.0,
            "uptime_percentage": 0.0,
            "first_seen": datetime.now(timezone.utc).isoformat(),
        })

        checks     = old["checks"] + 1
        successes  = old["successful_checks"] + (1 if r["reachable"] else 0)
        uptime     = round(successes / checks * 100, 1)

        # Скользящее среднее latency (только по успешным)
        if r["avg_latency_ms"] is not None:
            prev_lat = old.get("avg_latency_ms")
            if prev_lat is None:
                new_lat = r["avg_latency_ms"]
            else:
                # Взвешенное среднее: старые данные имеют вес (checks-1)
                new_lat = round(
                    (prev_lat * (checks - 1) + r["avg_latency_ms"]) / checks, 2
                )
        else:
            new_lat = old.get("avg_latency_ms")

        updated[ip] = {
            "ip": ip,
            "checks": checks,
            "successful_checks": successes,
            "avg_latency_ms": new_lat,
            "packet_loss_pct": r["packet_loss_pct"],
            "uptime_percentage": uptime,
            "last_checked": datetime.now(timezone.utc).isoformat(),
            "first_seen": old.get("first_seen", datetime.now(timezone.utc).isoformat()),
        }

    return updated


def sort_and_save(stats: dict) -> list[dict]:
    """Сортирует по (uptime DESC, latency ASC), сохраняет JSON, возвращает список."""
    ranked = sorted(
        stats.values(),
        key=lambda x: (
            -(x.get("uptime_percentage") or 0),
             (x.get("avg_latency_ms") or 9999),
        ),
    )

    payload = {
        "updated_at": datetime.now(timezone.utc).isoformat(),
        "total_dns": len(ranked),
        "dns_servers": ranked,
    }
    STATS_FILE.write_text(
        json.dumps(payload, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    log.info("💾  Статистика сохранена → %s  (%d серверов)", STATS_FILE, len(ranked))
    return ranked


# ═══════════════════════════════════════════════════════════════════════════
#  README — генерация таблицы
# ═══════════════════════════════════════════════════════════════════════════

_README_START = "<!-- DNS_TABLE_START -->"
_README_END   = "<!-- DNS_TABLE_END -->"

def _status_badge(uptime: float, latency: Optional[float]) -> str:
    if latency is None:
        return "🔴 Offline"
    if uptime >= 95 and latency < 50:
        return "🟢 Excellent"
    if uptime >= 80:
        return "🟡 Good"
    return "🟠 Unstable"


def build_table(ranked: list[dict]) -> str:
    top = [s for s in ranked if s.get("avg_latency_ms") is not None][:TOP_N]
    now = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")

    lines = [
        _README_START,
        "",
        f"## 🌐 DNS Monitor — Top {TOP_N} Fastest & Most Stable Servers",
        "",
        f"> 🕐 Last updated: **{now}** &nbsp;|&nbsp; "
        f"Проверяется каждые 3 часа через GitHub Actions",
        "",
        "| # | IP Address | Status | Avg Latency | Packet Loss | Uptime | Checks |",
        "|:-:|:----------:|:------:|:-----------:|:-----------:|:------:|:------:|",
    ]

    for i, s in enumerate(top, 1):
        lat  = f"{s['avg_latency_ms']:.1f} ms" if s["avg_latency_ms"] else "—"
        loss = f"{s['packet_loss_pct']:.0f}%"
        up   = f"{s['uptime_percentage']:.1f}%"
        badge = _status_badge(s["uptime_percentage"], s["avg_latency_ms"])
        lines.append(
            f"| {i} | `{s['ip']}` | {badge} | {lat} | {loss} | {up} | {s['checks']} |"
        )

    lines += [
        "",
        "<details>",
        "<summary>ℹ️ How it works</summary>",
        "",
        "1. **Source** — парсинг IPv4 из постов Telegram-канала (Telethon)",
        "2. **Validation** — A-запросы к `google.com` + `cloudflare.com` через dnspython",
        "3. **Metrics** — latency (мс) и packet loss усредняются по нескольким пробам",
        "4. **Persistence** — накопленная статистика в `dns_stats.json`",
        "5. **Ranking** — сортировка по uptime↓ → latency↑",
        "",
        "</details>",
        "",
        _README_END,
    ]
    return "\n".join(lines)


def update_readme(table: str) -> None:
    if README_FILE.exists():
        content = README_FILE.read_text(encoding="utf-8")
        if _README_START in content and _README_END in content:
            before = content[: content.index(_README_START)]
            after  = content[content.index(_README_END) + len(_README_END) :]
            new_content = before + table + after
        else:
            new_content = content.rstrip() + "\n\n" + table + "\n"
    else:
        new_content = table + "\n"

    README_FILE.write_text(new_content, encoding="utf-8")
    log.info("📝  README.md обновлён")


# ═══════════════════════════════════════════════════════════════════════════
#  Main
# ═══════════════════════════════════════════════════════════════════════════

async def main() -> None:
    log.info("═" * 60)
    log.info("🚀  DNS Monitor запущен")
    log.info("═" * 60)

    # 1. Парсим Telegram
    try:
        ips = await fetch_dns_from_telegram()
    except Exception:
        log.error("🛑  Не удалось получить IP из Telegram — выходим")
        return

    if not ips:
        log.warning("⚠  IP-адресов не найдено — нечего проверять")
        return

    log.info("─" * 60)
    log.info("🔬  Проверяем %d DNS-серверов (concurrency=%d) …", len(ips), CONCURRENCY)

    # 2. Параллельная валидация
    semaphore = asyncio.Semaphore(CONCURRENCY)
    results   = await asyncio.gather(*[check_dns(ip, semaphore) for ip in ips])

    reachable_count = sum(1 for r in results if r["reachable"])
    log.info("─" * 60)
    log.info("✔  Доступно: %d / %d", reachable_count, len(results))

    # 3. Слияние и сохранение статистики
    existing = load_stats()
    # load_stats() может вернуть обёртку с ключом dns_servers
    if "dns_servers" in existing:
        existing = {s["ip"]: s for s in existing["dns_servers"]}

    merged = merge_stats(existing, results)
    ranked = sort_and_save(merged)

    # 4. Генерация README
    table = build_table(ranked)
    update_readme(table)

    log.info("═" * 60)
    log.info("🏁  Готово! Топ-5 лучших DNS:")
    top5 = [s for s in ranked if s.get("avg_latency_ms") is not None][:5]
    for i, s in enumerate(top5, 1):
        log.info(
            "  %d. %-16s  %.1f ms  uptime %.1f%%",
            i, s["ip"], s["avg_latency_ms"], s["uptime_percentage"],
        )
    log.info("═" * 60)


if __name__ == "__main__":
    asyncio.run(main())
