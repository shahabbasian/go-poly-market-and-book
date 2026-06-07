# Polymarket Rust Client — polymarket_client_sdk_v2 (official README, archived)

Repo: https://github.com/Polymarket/rs-clob-client-v2
Crate: https://crates.io/crates/polymarket_client_sdk_v2
MSRV: Rust 1.88

An ergonomic Rust client for Polymarket services, primarily the Central Limit Order Book (CLOB).
Strongly typed request builders, authenticated endpoints, `alloy` support.

## Overview / design
- Typed CLOB requests (orders, trades, markets, balances, ...)
- Dual auth flows: normal authenticated flow + Builder authentication flow
- Type-level state machine: prevents using authenticated endpoints before authenticating; compile-time enforcement of correct transitions
- Signer support via `alloy::signers::Signer` (including remote signers, e.g. AWS KMS)
- Zero-cost abstractions — no dynamic dispatch in hot paths
- Order builders for easy construction & signing
- Full `serde` support
- Async-first design with `reqwest`

## Getting started
```toml
[dependencies]
polymarket_client_sdk_v2 = "0.5"
```
`cargo add polymarket_client_sdk_v2`
Run examples: `cargo run --example unauthenticated`

## Feature Flags
| Feature | Description |
| --- | --- |
| `clob` | Core CLOB client for order placement, market data, and authentication |
| `tracing` | Structured logging via `tracing` for HTTP requests, auth flows, and caching |
| `ws` | WebSocket client for real-time orderbook, price, and user event streaming |
| `rtds` | Real-time data streams for crypto prices (Binance, Chainlink) and comments |
| `data` | Data API client for positions, trades, leaderboards, and analytics |
| `gamma` | Gamma API client for market/event discovery, search, and metadata |
| `bridge` | Bridge API client for cross-chain deposits (EVM, Solana, Bitcoin) |
| `rfq` | RFQ API (within CLOB) for submitting and querying quotes |
| `heartbeats` | CLOB feature that automatically sends heartbeat messages; if the client disconnects all open orders are cancelled |
| `ctf` | CTF API client to perform split/merge/redeem on binary and neg risk markets |

```toml
polymarket_client_sdk_v2 = { version = "0.5", features = ["ws", "data"] }
```

## Re-exported types
From `types` module:
```rust
use polymarket_client_sdk_v2::types::{
    Address, ChainId, Signature, address,  // from alloy::primitives
    DateTime, NaiveDate, Utc,              // from chrono
    Decimal, dec,                          // from rust_decimal + rust_decimal_macros
};
```
From `auth` module:
```rust
use polymarket_client_sdk_v2::auth::{
    LocalSigner, Signer,          // from alloy::signers (LocalSigner + trait)
    Uuid, ApiKey,                 // from uuid (ApiKey = Uuid)
    SecretString, ExposeSecret,   // from secrecy
};
```
From `error` module:
```rust
use polymarket_client_sdk_v2::error::{ StatusCode, Method }; // from reqwest
```

## CLOB client examples

### Unauthenticated (read-only)
```rust
use polymarket_client_sdk_v2::clob::Client;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let client = Client::default();
    let ok = client.ok().await?;
    println!("Ok: {ok}");
    Ok(())
}
```

### Authenticated (EOA)
EOA/MetaMask/hardware users must set token allowances first (see allowances).
```rust
use std::str::FromStr as _;
use alloy::signers::Signer as _;
use alloy::signers::local::LocalSigner;
use polymarket_client_sdk_v2::{POLYGON, PRIVATE_KEY_VAR};
use polymarket_client_sdk_v2::clob::{Client, Config};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let private_key = std::env::var(PRIVATE_KEY_VAR).expect("Need a private key");
    let signer = LocalSigner::from_str(&private_key)?.with_chain_id(Some(POLYGON));
    let client = Client::new("https://clob-v2.polymarket.com", Config::default())?
        .authentication_builder(&signer)
        .authenticate()
        .await?;
    let ok = client.ok().await?;
    let api_keys = client.api_keys().await?;
    Ok(())
}
```

### Proxy/Safe wallets — funder auto-derived via CREATE2
```rust
let client = Client::new("https://clob-v2.polymarket.com", Config::default())?
    .authentication_builder(&signer)
    .signature_type(SignatureType::GnosisSafe)  // Funder auto-derived via CREATE2
    .authenticate()
    .await?;

// Override derived address if needed:
let client = Client::new("https://clob-v2.polymarket.com", Config::default())?
    .authentication_builder(&signer)
    .funder(address!("<your-polymarket-wallet-address>"))
    .signature_type(SignatureType::GnosisSafe)
    .authenticate()
    .await?;

// Derive addresses manually:
use polymarket_client_sdk_v2::{derive_safe_wallet, derive_proxy_wallet, POLYGON};
let safe_address = derive_safe_wallet(signer.address(), POLYGON);   // browser wallet (GnosisSafe)
let proxy_address = derive_proxy_wallet(signer.address(), POLYGON); // Magic/email wallet (Proxy)
```

### Signature types
- `signature_type=0` (default): EOA — MetaMask, hardware wallets, direct private key control
- `signature_type=1`: Email/Magic wallet (delegated signing) — Proxy
- `signature_type=2`: Browser wallet proxy signatures (proxy contract) — GnosisSafe
- `signature_type=3`: EIP-1271 smart contract wallet signatures (V2 orders only) — Poly1271
The funder is the actual address holding funds. With proxy wallets the signing key differs from the funder; SDK derives it via CREATE2 for `SignatureType::Proxy`/`GnosisSafe`. Override with `.funder(address)`.

### Place a market order
```rust
use polymarket_client_sdk_v2::clob::types::{Amount, OrderType, Side};
use polymarket_client_sdk_v2::types::Decimal;

let order = client
    .market_order()
    .token_id("<token-id>")
    .amount(Amount::usdc(Decimal::ONE_HUNDRED)?)
    .side(Side::Buy)
    .order_type(OrderType::FOK)
    .build()
    .await?;
let signed_order = client.sign(&signer, order).await?;
let response = client.post_order(signed_order).await?;
```

### Place a limit order
```rust
use polymarket_client_sdk_v2::clob::types::Side;
use polymarket_client_sdk_v2::types::Decimal;
use rust_decimal_macros::dec;

let order = client
    .limit_order()
    .token_id("<token-id>")
    .size(Decimal::ONE_HUNDRED)
    .price(dec!(0.1))
    .side(Side::Buy)
    .build()
    .await?;
let signed_order = client.sign(&signer, order).await?;
let response = client.post_order(signed_order).await?;
```

### V1 and V2 protocols
Protocol auto-detected on first order build via `GET /version`, cached for the lifetime of the `Client`.
Pick protocol by host:
| Protocol | Host | Collateral | EIP-712 domain version |
| --- | --- | --- | --- |
| V2 | `https://clob-v2.polymarket.com` | pUSD | `"2"` |
| V1 | `https://clob.polymarket.com` | USDC.e | `"1"` |

V2 orders add `timestamp`, `metadata`, `builder` fields. V1 orders use `taker`, `nonce`, `feeRateBps`.
The order builder exposes both sets — non-applicable fields are silently ignored at build time, so one code-path works against either server.

V2-specific builder fields (ignored on V1):
```rust
use polymarket_client_sdk_v2::types::B256;
let order = client.limit_order()
    .token_id("<token-id>").size(Decimal::ONE_HUNDRED).price(dec!(0.5)).side(Side::Buy)
    .metadata(B256::ZERO)      // 32 bytes of custom metadata
    .builder_code(B256::ZERO)  // builder fee attribution
    .defer_exec(false)         // defer execution
    .build().await?;
```
V1-specific builder fields (ignored on V2):
```rust
use polymarket_client_sdk_v2::types::Address;
let order = client.limit_order()
    .token_id("<token-id>").size(Decimal::ONE_HUNDRED).price(dec!(0.5)).side(Side::Buy)
    .taker(Address::ZERO)  // explicit taker; default zero = public order
    .nonce(0)              // on-chain cancel nonce; default 0
    .fee_rate_bps(0)       // must match the market rate when set
    .build().await?;
```

### Builder-attributed trading
In V2 builder attribution is carried on the order's `builder_code` field (and as a query param on `builder_trades`).
Set a default `builder_code` on `Config` and every order inherits it unless overridden.
```rust
use polymarket_client_sdk_v2::clob::types::request::TradesRequest;
use polymarket_client_sdk_v2::types::B256;

let builder_code = B256::from_str(&std::env::var("POLYMARKET_BUILDER_CODE")?)?;
let config = Config::builder().builder_code(builder_code).build();
let client = Client::new("https://clob-v2.polymarket.com", config)?
    .authentication_builder(&signer)
    .authenticate().await?;
let builder_trades = client
    .builder_trades(builder_code, &TradesRequest::default(), None)
    .await?;
```

## WebSocket streaming (feature `ws`)
```toml
polymarket_client_sdk_v2 = { version = "0.5", features = ["ws"] }
```
```rust
use std::str::FromStr as _;
use futures::StreamExt as _;
use polymarket_client_sdk_v2::clob::ws::Client;
use polymarket_client_sdk_v2::types::U256;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let client = Client::default();
    let asset_ids = vec![U256::from_str("<asset-id>")?];
    let stream = client.subscribe_orderbook(asset_ids)?;
    let mut stream = Box::pin(stream);
    while let Some(book_result) = stream.next().await {
        let book = book_result?;
        println!("Orderbook update for {}: {} bids, {} asks",
            book.asset_id, book.bids.len(), book.asks.len());
    }
    Ok(())
}
```
Available streams:
- `subscribe_orderbook()` — bid/ask levels for assets
- `subscribe_prices()` — price change events
- `subscribe_midpoints()` — calculated midpoint prices
- `subscribe_orders()` — user order updates (authenticated)
- `subscribe_trades()` — user trade executions (authenticated)
See `examples/clob/ws/` for authenticated user streams.

## Optional APIs

### Data API (feature `data`)
```rust
use polymarket_client_sdk_v2::data::Client;
use polymarket_client_sdk_v2::data::types::request::PositionsRequest;
use polymarket_client_sdk_v2::types::address;

let client = Client::default();
let user = address!("0x0000000000000000000000000000000000000000");
let request = PositionsRequest::builder().user(user).limit(10)?.build();
let positions = client.positions(&request).await?;
```

### Gamma API (feature `gamma`)
```rust
use polymarket_client_sdk_v2::gamma::Client;
use polymarket_client_sdk_v2::gamma::types::request::{EventsRequest, SearchRequest};

let client = Client::default();
let request = EventsRequest::builder().active(true).limit(5).build();
let events = client.events(&request).await?;
let search = SearchRequest::builder().q("bitcoin").build();
let results = client.search(&search).await?;
```

### Bridge API (feature `bridge`)
```rust
use polymarket_client_sdk_v2::bridge::Client;
use polymarket_client_sdk_v2::bridge::types::DepositRequest;
use polymarket_client_sdk_v2::types::address;

let client = Client::default();
let request = DepositRequest::builder()
    .address(address!("0x0000000000000000000000000000000000000000")).build();
let response = client.deposit(&request).await?;
println!("EVM: {}  Solana: {}  Bitcoin: {}",
    response.address.evm, response.address.svm, response.address.btc);
```

## Additional CLOB capabilities
- Rewards & Earnings — query maker rewards, daily earnings, reward percentages
- Streaming pagination — `stream_data()` for iterating large result sets
- Batch operations — `post_orders()` and `cancel_orders()` for multiple orders
- Order scoring — check if orders qualify for maker rewards
- Notifications — manage trading notifications
- Balance management — query and refresh balance/allowance caches
- Geoblock detection — check if trading is available in your region
See `examples/clob/authenticated.rs`.

## Token allowances
EOA/MetaMask users MUST set allowances; proxy/Safe wallets do NOT.
Approve two token types: USDC (deposits/trading) and Conditional Tokens (outcome tokens).
Use `examples/approvals.rs` — run once for USDC, then change `TOKEN_TO_APPROVE` and run per conditional token. Set once per wallet.

## Example files in repo (examples/)
- `unauthenticated.rs`
- `clob/authenticated.rs` (comprehensive)
- `clob/ws/` (WebSocket incl. authenticated user streams)
- `approvals.rs` (token allowances)
- `data.rs`
- `gamma/client.rs`
- `bridge.rs`

---

## SDK module map (`polymarket_client_sdk_v2::*`)
| Module | Feature | Purpose |
| --- | --- | --- |
| `clob` | `clob` | CLOB `Client`, `Config`, order builders, `sign`, `post_order(s)`, `cancel*`, market data, rewards |
| `clob::ws` | `ws` | WebSocket `Client` with typed streams |
| `clob::types` | `clob` | `Side`, `OrderType`, `Amount`, `SignatureType`, `request::*` (e.g. `TradesRequest`) |
| `data` | `data` | Data API `Client`: positions, trades, activity, holders, leaderboard, value, OI |
| `gamma` | `gamma` | Gamma API `Client`: events, markets, search, tags, series, sports, comments, profiles |
| `bridge` | `bridge` | Bridge `Client`: deposit/withdraw addresses, quote, status, supported assets |
| `ctf` | `ctf` | Split / merge / redeem on binary and neg-risk markets |
| `rtds` | `rtds` | Real-Time Data Socket: crypto prices + comments |
| `auth` | — | `LocalSigner`, `Signer`, `Uuid`/`ApiKey`, `SecretString`, `ExposeSecret` |
| `types` | — | `Address`, `ChainId`, `Signature`, `U256`, `B256`, `Decimal`, `dec`, `address!`, chrono types |
| `error` | — | error types + re-exported `StatusCode`, `Method` |

Constants: `POLYGON` (chain id 137), `PRIVATE_KEY_VAR` (env var name `POLYMARKET_PRIVATE_KEY`).
Wallet derivation: `derive_safe_wallet(addr, chain)`, `derive_proxy_wallet(addr, chain)`.

## CLOB method catalogue (cross-language — Rust uses builders, names map closely)
Public (no auth): `ok`, `markets`/`getMarkets`, `getMarket`, simplified/sampling markets, `getOrderBook(s)`,
`getPrice(s)`, `getMidpoint(s)`, `getSpread(s)`, `getPricesHistory`, `getLastTradePrice(s)`,
`getMarketTradesEvents`, `getClobMarketInfo`, `getFeeRateBps`, `getFeeExponent`, `getTickSize`,
`getNegRisk`, `getServerTime`, `calculateMarketPrice`.
L1 (signer, no creds): `createApiKey(nonce?)`, `deriveApiKey(nonce?)`, `createOrDeriveApiKey(nonce?)`,
`createOrder`, `createMarketOrder` (sign locally, post later).
L2 (creds): `createAndPostOrder`, `createAndPostMarketOrder`, `postOrder`, `postOrders` (≤15),
`cancelOrder`, `cancelOrders`, `cancelAll`, `cancelMarketOrders`, `getOrder`, `getOpenOrders`,
`getTrades`, `getTradesPaginated`, `getBalanceAllowance`, `updateBalanceAllowance`,
`getApiKeys`, `deleteApiKey`, `getNotifications`, `dropNotifications`.
Builder: `builder_trades`/`getBuilderTrades`, plus `getOrder`/`getOpenOrders` scoped to builder code.

In Rust the order flow is always: `client.limit_order()/market_order().…build().await?`
→ `client.sign(&signer, order).await?` → `client.post_order(signed).await?`
(or `post_orders(vec)` for batches). Cancels: `client.cancel_orders(...)`, `cancel_all()`, etc.

## Order types & tick sizes
- Order types: `GTC` (resting limit, default for limit), `GTD` (good-till-date, needs `expiration`),
  `FOK` (fill-or-kill, default for market), `FAK` (fill-and-kill / IOC).
- Tick sizes: one of `0.1`, `0.01`, `0.001`, `0.0001`. Tick can change when price crosses 0.96 / 0.04.
- Market `BUY` market-orders are denominated in **USDC dollar amount**; `SELL` in **shares**.
- Limit orders: `price` (0–1) and `size` (shares).
- `post_only` flag posts a maker-only order; rejected if it would cross.

## Production notes
- **Rate limits**: see `references/rest-endpoints.md` notes + `api-pages/api-reference_rate-limits.md`.
  Cloudflare throttles (queues) rather than hard-rejects; `POST /order` & `DELETE /order` allow
  5,000 req/10s burst, 48,000 req/10min sustained. Back off on `429`.
- **Heartbeats**: with the `heartbeats` feature the CLOB client auto-pings; if the client disconnects,
  all open orders are cancelled (a safety feature for automated systems). The REST `POST /heartbeats`
  endpoint does the same for hand-rolled clients.
- **Matching engine restarts**: scheduled maintenance windows exist; after a restart the book may enter
  post-only mode briefly. See `api-pages/trading_matching-engine.md`.
- **Geoblock**: check trading availability per region before placing orders
  (`api-pages/api-reference_geoblock.md`). The SDK exposes geoblock detection.
- **Gasless**: deposit-wallet (`SignatureType::Poly1271`, type 3) flow uses the relayer for gasless
  onchain ops; see `api-pages/trading_gasless.md`.
- **Errors**: CLOB returns `{"error": "<message>"}`. Full catalogue in `api-pages/resources_error-codes.md`.
  Common: `INVALID_SIGNATURE`, `NONCE_ALREADY_USED`, `401 Unauthorized/Invalid api key`,
  `503 Trading is currently disabled`, `429 Too Many Requests`.
