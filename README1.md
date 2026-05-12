# 🌐 DNS Monitor

Автоматический мониторинг публичных DNS-серверов, извлечённых из Telegram.
Проверяется каждые **3 часа** через GitHub Actions.

<!-- DNS_TABLE_START -->

## 🌐 DNS Monitor — Top 10 Fastest & Most Stable Servers

> 🕐 Last updated: **pending first run** &nbsp;|&nbsp; Проверяется каждые 3 часа через GitHub Actions

| # | IP Address | Status | Avg Latency | Packet Loss | Uptime | Checks |
|:-:|:----------:|:------:|:-----------:|:-----------:|:------:|:------:|
| — | — | ⏳ Waiting for first run | — | — | — | — |

<details>
<summary>ℹ️ How it works</summary>

1. **Source** — парсинг IPv4 из постов Telegram-канала (Telethon)
2. **Validation** — A-запросы к `google.com` + `cloudflare.com` через dnspython
3. **Metrics** — latency (мс) и packet loss усредняются по нескольким пробам
4. **Persistence** — накопленная статистика в `dns_stats.json`
5. **Ranking** — сортировка по uptime↓ → latency↑

</details>

<!-- DNS_TABLE_END -->

## ⚙️ Setup

### 1. Secrets (GitHub → Settings → Secrets)

| Secret | Описание |
|--------|----------|
| `TG_API_ID` | API ID из [my.telegram.org](https://my.telegram.org) |
| `TG_API_HASH` | API Hash оттуда же |
| `TG_SESSION_STRING` | Строка сессии (см. ниже) |

### 2. Генерация SESSION_STRING

```python
from telethon.sync import TelegramClient
from telethon.sessions import StringSession

api_id   = int(input("API ID: "))
api_hash = input("API Hash: ")

with TelegramClient(StringSession(), api_id, api_hash) as client:
    print("SESSION_STRING:", client.session.save())
```

Запустите локально, войдите по номеру телефона — скопируйте вывод в секрет `TG_SESSION_STRING`.

### 3. Ручной запуск

**Actions → 🌐 DNS Monitor → Run workflow**

## 📁 Files

| Файл | Описание |
|------|----------|
| `monitor.py` | Основной скрипт мониторинга |
| `dns_stats.json` | Накопленная статистика (обновляется автоматически) |
| `requirements.txt` | Python-зависимости |
| `.github/workflows/dns_update.yml` | Расписание GitHub Actions |
