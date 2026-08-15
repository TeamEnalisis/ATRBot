# D.P.Sharma Trading Bot

Automated ATR-trailing-stop ("UT Bot") trading system for Delta Exchange,
matching the Pine Script strategy you provided. Two parts:

- **`backend/`** — Go server. Talks to Delta Exchange, runs the strategy,
  places bracket (entry + TP + SL) orders, exposes a small REST API.
  Your Delta API key/secret live **only** here.
- **`android/`** — Kotlin/Jetpack Compose app (dashboard, positions,
  configuration, logs) that talks to your backend over the REST API. Built
  automatically into an APK by GitHub Actions.

---

## ⚠️ Read before running with real money

- This places live orders with **market execution** as soon as a signal
  fires. Test on Delta's **testnet** (`https://cdn-ind.testnet.deltaex.org`)
  with a small size first, and paper-watch the `/api/logs` output for at
  least a few sessions before pointing it at your live account.
- The strategy math (ATR trailing stop, TP/SL) is re-implemented from your
  Pine Script as closely as possible, but Pine's exact internals and Delta's
  order/margin rules can differ subtly (e.g. rounding, tick size, contract
  value). Verify every field against your account/product before trusting it
  with size.
- Not financial advice, and I can't guarantee this code is bug-free or that
  it will be profitable — you are responsible for testing and for anything
  it trades.

---

## 1. Backend — deploy to a free VPS

The backend is a single static Go binary with **zero external dependencies**
(standard library only), so it builds fast and runs on the smallest VPS tier.

```bash
cd backend
go build -o tradingbot .
APP_TOKEN="pick-a-long-random-string" ./tradingbot
```

Or with Docker:

```bash
cd backend
docker build -t dpsharma-tradingbot .
docker run -d --name tradingbot -p 8080:8080 \
  -e APP_TOKEN="pick-a-long-random-string" \
  -v $(pwd)/data:/app/data \
  dpsharma-tradingbot
```

Then open `http://<vps-ip>:8080/api/config` once to confirm it's up, and
set your real values from the app's **Configuration** tab (API key/secret,
product id/symbol, strategy inputs, fees, beta, etc.) — everything is saved
to `data/config.json` on the VPS.

`APP_TOKEN` is a shared secret between the app and backend so a random
visitor to your VPS IP can't start/stop the bot or read your config. Put the
same value into the app's Configuration → "App Token" field. Run behind
HTTPS (e.g. Caddy/nginx + Let's Encrypt) once you're comfortable — the app
defaults to allowing plain HTTP so you can get started immediately.

## 2. Android app — build the APK via GitHub Actions

1. Push this whole folder to a new GitHub repo.
2. GitHub Actions (`.github/workflows/android-build.yml`) builds automatically
   on every push to `main`/`master` — or trigger it manually from the
   **Actions** tab ("Run workflow").
3. Open the finished run → **Artifacts** → download
   `dpsharma-tradingbot-debug-apk` → install `app-debug.apk` on your phone
   (enable "Install unknown apps" for your file manager/browser).
4. On first launch, go to **Configuration** → set **Backend URL** to
   `http://<vps-ip>:8080` and **App Token** to the same value you set on the
   server → Save connection. Then fill in the rest of the fields and Save.
5. Back on **Dashboard**, tap **START BOT**.

The workflow ships a **debug** APK (unsigned, installs fine for personal
use). If you want a signed release build later, add a keystore as GitHub
Secrets and extend the workflow's `assembleRelease` step — ask me and I'll
add it.

## 3. How the strategy → order flow works

1. Every `poll_interval_seconds`, the backend fetches the latest candles for
   `product_symbol`/`resolution` and re-runs the ATR trailing-stop logic
   from your Pine script on the last **closed** candle.
2. On a fresh **BUY** signal: if a short is open it's closed first
   (reduce-only market order), then a market **buy** is sent with a bracket:
   - `take_profit = entry_price + (fee + funding + brokerage) × beta`
   - `stop_loss = configured/auto-detected support level`
3. On a fresh **SELL** signal: symmetric, using resistance as the stop and
   `take_profit = entry_price − expenses × beta`.
4. Leave `support_level` / `resistance_level` at `0` in Configuration to let
   the backend auto-detect them as the swing low/high over the last
   `sr_lookback` candles; set them explicitly to pin your own levels.

## 4. REST API (used by the app, useful for debugging)

All routes require header `X-App-Token: <APP_TOKEN>` if you set one.

| Method | Path            | Purpose                                   |
|--------|-----------------|--------------------------------------------|
| GET    | `/api/status`   | running state, current position            |
| GET    | `/api/dashboard`| balances, positions, ticker, bot state     |
| GET    | `/api/config`   | current configuration (secret hidden)      |
| POST   | `/api/config`   | update configuration                       |
| POST   | `/api/bot/start`| start the strategy loop                    |
| POST   | `/api/bot/stop` | stop the strategy loop                     |
| GET    | `/api/logs`     | recent decision log                        |

## 5. Repo layout

```
tradingbot/
├── backend/            Go REST API + strategy engine + Delta Exchange client
├── android/             Kotlin/Compose app (Dashboard, Positions, Configuration, Logs)
└── .github/workflows/   CI that builds the APK on every push
```
