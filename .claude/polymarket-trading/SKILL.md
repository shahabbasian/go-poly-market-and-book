---
name: polymarket-trading
description: >-
  Complete mastery of the Polymarket APIs, WebSockets, and the official Rust SDK
  (polymarket_client_sdk_v2). Use WHENEVER the user builds, debugs, or integrates
  anything touching Polymarket: placing/cancelling CLOB orders, fetching
  markets/events/prices/orderbooks, reading positions/trades/leaderboards/holders
  from the Data API, discovering markets via the Gamma API, streaming real-time
  data over WebSockets (market, user, sports, RTDS channels), authenticating (L1
  private-key / L2 API-key / builder / deposit-wallet flows), bridging deposits,
  splitting/merging/redeeming CTF tokens, market making, or builder attribution —
  in Rust, TypeScript, or Python. Trigger for any mention of Polymarket,
  prediction-market trading bots, the CLOB, condition IDs, outcome/Yes-No tokens,
  pUSD/USDC.e collateral, or polymarket_client_sdk_v2, even when the user does not
  say "use the docs". Prefer this skill over guessing endpoint shapes or SDK
  method names from memory; the bundled references are the ground truth.
---

# Polymarket Trading & Data Integration

This skill gives you authoritative, archived knowledge of the entire Polymarket
developer surface: every REST endpoint across four APIs, all four WebSocket
channels, the full authentication model, and the official **Rust** SDK
(`polymarket_client_sdk_v2`), plus TypeScript/Python equivalents.

**Ground rule:** do not invent endpoint paths, parameters, message shapes, or SDK
method names. They are all captured in the bundled references below. Open the
relevant reference before writing code, and cite exact fields from it.

## How to use this skill

1. Identify which surface the task touches (Rust SDK? a specific REST endpoint?
   a WebSocket channel? auth?).
2. Open the matching reference file(s) in `references/`.
3. For an exact endpoint contract (params, request/response schema, examples),
   open the specific raw page in `references/api-pages/` (use
   `references/api-pages-index.md` to locate it) or the OpenAPI/AsyncAPI spec in
   `references/specs/`.
4. Write code against the verified shapes. Default to the official SDK over
   hand-rolled HTTP unless the user needs a custom client.

### Reference map
| File | Use it for |
| --- | --- |
| `references/rust-sdk.md` | **Primary for Rust.** Crate setup, feature flags, auth flows, order builders, signing, WS streams, module map, full CLOB method catalogue, production notes. |
| `references/websockets.md` | All four channels: endpoints, subscribe messages, every event type + payload schema, heartbeats, reconnection playbook. |
| `references/rest-endpoints.md` | Complete cheat-sheet of all 118 REST endpoints across CLOB / Gamma / Data / Bridge with base URLs. |
| `references/api-pages-index.md` | Index of all 137 archived doc pages, grouped by area, so you can jump to the exact page. |
| `references/api-pages/*.md` | The full, unabridged official doc pages (params, schemas, examples) — the source of truth for any single endpoint or concept. |
| `references/specs/*` | Machine-readable OpenAPI (`clob`/`gamma`/`data`/`bridge`) + AsyncAPI (`market`/`user`/`sports` WS) specs. Parse these for exhaustive schemas. |
| `assets/rust-starter/` | A ready Cargo project: read market data, place orders, stream the orderbook. Copy and adapt. |

## The four APIs (mental model)

| API | Base URL | Auth | Domain |
| --- | --- | --- | --- |
| **Gamma** | `https://gamma-api.polymarket.com` | none | Discovery: markets, events, series, tags, search, sports, comments, profiles |
| **Data** | `https://data-api.polymarket.com` | none | Analytics: positions, trades, activity, holders, open interest, leaderboards, value |
| **CLOB** | `https://clob.polymarket.com` (V1, USDC.e) / `https://clob-v2.polymarket.com` (V2, pUSD) | read = none; trade = L2 | Orderbook, prices, midpoints, spreads, history **and** order placement/cancellation/heartbeat |
| **Bridge** | `https://bridge.polymarket.com` | none | Cross-chain deposits/withdrawals (proxy of fun.xyz) |

WebSockets (see `references/websockets.md`):
`wss://ws-subscriptions-clob.polymarket.com/ws/market` (public),
`/ws/user` (auth), `wss://sports-api.polymarket.com/ws` (public),
`wss://ws-live-data.polymarket.com` (RTDS, optional auth).

## Core concepts (don't get these wrong)

- **Market**: one binary Yes/No question. Identified by a **condition ID**
  (CTF), a **question ID**, and **two token IDs** (one Yes ERC1155 token, one No).
  Tradable on the CLOB only when `enableOrderBook` is true.
- **Event**: a container grouping one or more related markets (multi-outcome).
- **Outcome tokens**: ERC1155 on Polygon via the Gnosis Conditional Token
  Framework (CTF). Each Yes/No pair is backed by exactly $1 of collateral.
  **Split** $1 → 1 Yes + 1 No. **Merge** 1 Yes + 1 No → $1. **Redeem** winning
  tokens → $1 after resolution. (CTF ops: `references/api-pages/trading_ctf_*.md`.)
- **Position**: your balance of outcome tokens for a market.
- **Price**: 0–1, equals the market's implied probability. Yes + No prices ≈ $1.
- **Collateral**: V2 markets settle in **pUSD**; V1 in **USDC.e**.
- **Subscription keys**: market WS channel uses **asset/token IDs**; user WS
  channel uses **condition IDs**.

## Authentication (CLOB only; Gamma/Data/CLOB-reads are public)

Two levels — full details in `references/api-pages/api-reference_authentication.md`.
- **L1 (private key)**: EIP-712 signature. Used to create/derive API credentials
  and to sign orders locally. Non-custodial — the key never leaves the user.
- **L2 (API key)**: `apiKey` + `secret` + `passphrase`, generated from L1. Sent as
  five `POLY_*` HMAC-SHA256 headers on trading/account requests. Even with L2, the
  user must still sign each order payload.
- **Signature types / funder**: `0` EOA, `1` email/Magic proxy, `2` browser-wallet
  Gnosis Safe, `3` EIP-1271 deposit wallet (recommended for new API users, V2 only).
  The **funder** is the address holding funds; for proxy/Safe it is auto-derived
  via CREATE2. EOA users must set token allowances once before trading.
- In the Rust SDK the whole flow is one builder chain:
  `Client::new(host, cfg)?.authentication_builder(&signer).authenticate().await?`.

## Orders

- **Types**: `GTC` (resting limit, default for limit), `GTD` (good-till-date,
  needs `expiration`), `FOK` (fill-or-kill, default for market), `FAK`
  (fill-and-kill / IOC). `post_only` posts maker-only.
- **Tick sizes**: `0.1` / `0.01` / `0.001` / `0.0001`; can change at price >0.96 or <0.04.
- **Market BUY** amount is in **USDC dollars**; **market SELL** in **shares**.
  Limit orders use `price` (0–1) + `size` (shares).
- **Batch**: post up to **15** orders, cancel up to **3000** per request.
- Rust flow: `client.limit_order()/market_order()...build().await?`
  → `client.sign(&signer, order).await?` → `client.post_order(signed).await?`.
- Full contracts: `references/api-pages/trading_orders_*.md` and the
  `trade/` pages; SDK methods in `references/rust-sdk.md`.

## Rust SDK quick facts (read `references/rust-sdk.md` for the rest)

- Crate `polymarket_client_sdk_v2` (repo `Polymarket/rs-clob-client-v2`), MSRV 1.88,
  async/`tokio`, `alloy` signers, type-state machine that forbids authed calls
  before `.authenticate()`.
- Feature flags: `clob`, `ws`, `data`, `gamma`, `bridge`, `ctf`, `rtds`, `rfq`,
  `heartbeats`, `tracing`. Enable only what you need.
- Re-exports under `::types` (`Address`, `U256`, `B256`, `Decimal`, `dec`,
  `address!`, chrono) and `::auth` (`LocalSigner`, `Signer`, `ApiKey`, `SecretString`).
- Constants: `POLYGON` (137), `PRIVATE_KEY_VAR` (`POLYMARKET_PRIVATE_KEY`).
- WS via `clob::ws::Client`: `subscribe_orderbook`, `subscribe_prices`,
  `subscribe_midpoints`, `subscribe_orders` (auth), `subscribe_trades` (auth).
- Verify the latest version with `cargo search polymarket_client_sdk_v2` before pinning.

## Operational guardrails

- **Rate limits** (`references/api-pages/api-reference_rate-limits.md`): Cloudflare
  throttles (queues) rather than rejects; back off on `429`. `POST/DELETE /order`
  allow 5,000 req/10s burst, 48,000 req/10min sustained.
- **Heartbeats**: WS market/user → client sends `PING` every 10s; sports → reply
  `pong` to server ping within 10s. The CLOB `heartbeats` feature / `POST /heartbeats`
  auto-cancels open orders if the client goes silent (a safety feature).
- **Matching-engine restarts** (`trading_matching-engine.md`): expect maintenance
  windows and brief post-only mode after restart.
- **Geoblock** (`api-reference_geoblock.md`): check regional availability before ordering.
- **Errors** (`resources_error-codes.md`): CLOB returns `{"error":"..."}`. Watch for
  `INVALID_SIGNATURE`, `NONCE_ALREADY_USED`, `401`, `429`, `503 Trading disabled`.
- **Security**: never hardcode/commit private keys or API secrets; run the user WS
  channel and any L2 requests server-side only.

## Output expectations

- Default to the official SDK for the user's language (Rust unless stated otherwise).
- Produce compilable, idiomatic code with correct imports from the verified module map.
- When using a REST endpoint directly, state the exact base URL, method, path, and
  required headers, taken from the references — not from memory.
- If the user asks for something the references don't cover, say so and check the
  live docs index at `https://docs.polymarket.com/llms.txt` rather than guessing.
