name: 🌐 DNS Monitor

on:
  schedule:
    - cron: "0 */3 * * *"   # каждые 3 часа
  workflow_dispatch:          # ручной запуск через UI

permissions:
  contents: write             # нужно для коммита изменений

jobs:
  monitor:
    name: Fetch & Validate DNS Servers
    runs-on: ubuntu-latest
    timeout-minutes: 15

    steps:
      # ── 1. Checkout ──────────────────────────────────────────────────────
      - name: ⬇️  Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0      # полная история для корректного коммита

      # ── 2. Python ────────────────────────────────────────────────────────
      - name: 🐍  Set up Python 3.12
        uses: actions/setup-python@v5
        with:
          python-version: "3.12"
          cache: pip

      # ── 3. Dependencies ───────────────────────────────────────────────────
      - name: 📦  Install dependencies
        run: pip install -r requirements.txt

      # ── 4. Run monitor ───────────────────────────────────────────────────
      - name: 🔬  Run DNS monitor
        env:
          TG_API_ID:         ${{ secrets.TG_API_ID }}
          TG_API_HASH:       ${{ secrets.TG_API_HASH }}
          TG_SESSION_STRING: ${{ secrets.TG_SESSION_STRING }}
        run: python monitor.py

      # ── 5. Commit results ─────────────────────────────────────────────────
      - name: 💾  Commit updated stats & README
        uses: stefanzweifel/git-auto-commit-action@v5
        with:
          commit_message: "chore(dns): auto-update stats & README [skip ci]"
          file_pattern: "dns_stats.json README.md"
          commit_user_name:  "dns-monitor-bot"
          commit_user_email: "dns-monitor-bot@users.noreply.github.com"
          commit_author:     "dns-monitor-bot <dns-monitor-bot@users.noreply.github.com>"
          skip_dirty_check:  false
          status_options:    "--untracked-files=no"
