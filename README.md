# Polymarket Crypto Market Fetcher

Go application that continuously discovers and tracks Polymarket crypto up/down markets across 5m, 15m, 1h, and 4h timeframes. Stores market data, token IDs, and order book info in PostgreSQL.

## Prerequisites

- External **PostgreSQL** database (already running in another Docker container or server)
- Docker & Docker Compose

## Quick Start

```bash
# 1. Clone and enter
cd /path/to/project

# 2. Copy env file
cp .env.example .env

# 3. Edit .env — set your DATABASE_URL pointing to existing Postgres
# DATABASE_URL=postgres://user:password@your-postgres-host:5432/polymarket?sslmode=disable

# 4. Run
docker compose up -d
```

## Deploy on Coolify

### 1. Push to Git

```bash
git init
git add .
git commit -m "init: polymarket market fetcher"
git remote add origin <your-gitea-repo-url>
git push -u origin main
```

### 2. Create Resource in Coolify

- **Type**: `Application`
- **Source**: Your Git repository
- **Build Pack**: `Docker Compose`
- **Compose File**: `docker-compose.yml`

### 3. Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | **Yes** | — | Postgres connection string to your existing DB |
| `LOG_LEVEL` | No | `info` | debug / info / warn / error |
| `LOG_FORMAT` | No | `json` | json / text |
| `RATE_LIMIT_DELAY_MS` | No | `250` | Min delay between Gamma API calls |
| `DISCOVERY_INTERVAL_S` | No | `30` | Discovery cycle interval (seconds) |
| `REFRESH_INTERVAL_S` | No | `60` | Refresh cycle interval (seconds) |
| `FUTURE_SLUG_COUNT` | No | `10` | Slugs to look ahead (when DB has data) |
| `INITIAL_SLUG_COUNT` | No | `20` | Slugs to look ahead (when DB is empty) |
| `CLOB_BATCH_SIZE` | No | `50` | Max tokens per CLOB batch request |
| `REFRESH_MAX_AGE_M` | No | `5` | Minutes before a market is stale |
| `REFRESH_LIMIT` | No | `100` | Max markets to refresh per cycle |
| `HTTP_TIMEOUT_S` | No | `30` | HTTP request timeout |
| `MAX_RETRIES` | No | `5` | Max retry attempts for transient failures |

### 4. Important Notes

- **No ports exposed** — this is a background worker, not a web service
- The existing PostgreSQL must be reachable from this container (same Docker network or public network)
- If Postgres is in another Coolify project, use the internal hostname: `postgres` or the service name
- Table `new_markets` must already exist in your database (the app auto-creates it via migration if not)

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Discovery  │────▶│ Gamma API   │     │  Existing   │
│  Worker     │     │ (slug→data) │     │  PostgreSQL │
│  (30s)      │     └─────────────┘     └─────────────┘
└─────────────┘              │                  ▲
      │                      ▼                  │
      │               ┌─────────────┐           │
      │               │  CLOB API   │           │
      └──────────────▶│ (batch books)───────────┘
                      └─────────────┘
┌─────────────┐
│  Refresh    │────▶ Re-fetch upcoming + stale markets
│  Worker     │
│  (60s)      │
└─────────────┘
```

## Supported Markets

| Coin | 5m Slug | 15m Slug | 1h Slug | 4h Slug |
|------|---------|----------|---------|---------|
| BTC | `btc-updown-5m-{ts}` | `btc-updown-15m-{ts}` | `bitcoin-up-or-down-{date}` | `btc-updown-4h-{ts}` |
| ETH | `eth-updown-5m-{ts}` | `eth-updown-15m-{ts}` | `ethereum-up-or-down-{date}` | `eth-updown-4h-{ts}` |
| SOL | `sol-updown-5m-{ts}` | `sol-updown-15m-{ts}` | `solana-up-or-down-{date}` | `sol-updown-4h-{ts}` |
| XRP | `xrp-updown-5m-{ts}` | `xrp-updown-15m-{ts}` | `xrp-up-or-down-{date}` | `xrp-updown-4h-{ts}` |
| DOGE | `doge-updown-5m-{ts}` | `doge-updown-15m-{ts}` | `dogecoin-up-or-down-{date}` | `doge-updown-4h-{ts}` |
| HYPE | `hype-updown-5m-{ts}` | `hype-updown-15m-{ts}` | `hype-up-or-down-{date}` | `hype-updown-4h-{ts}` |
| BNB | `bnb-updown-5m-{ts}` | `bnb-updown-15m-{ts}` | `bnb-up-or-down-{date}` | `bnb-updown-4h-{ts}` |

## Database

The app auto-creates the `new_markets` table and indexes on startup (idempotent). Only `DATABASE_URL` is required — no other DB config needed.
