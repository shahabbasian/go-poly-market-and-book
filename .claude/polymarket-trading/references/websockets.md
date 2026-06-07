# Polymarket WebSockets — Complete Reference

Four channels. Full raw docs in `api-pages/market-data_websocket_*.md` and `api-pages/api-reference_wss_*.md`.
Machine-readable AsyncAPI specs in `specs/asyncapi-market.json`, `specs/asyncapi-user.json`, `specs/asyncapi-sports.json`.

| Channel | Endpoint | Auth |
| --- | --- | --- |
| Market | `wss://ws-subscriptions-clob.polymarket.com/ws/market` | No |
| User | `wss://ws-subscriptions-clob.polymarket.com/ws/user` | Yes (API creds) |
| Sports | `wss://sports-api.polymarket.com/ws` | No |
| RTDS (Real-Time Data Socket) | `wss://ws-live-data.polymarket.com` | Optional |

**Rust:** prefer the SDK `ws` feature (`polymarket_client_sdk_v2::clob::ws::Client`) which wraps the
market+user channels with typed streams (`subscribe_orderbook`, `subscribe_prices`, `subscribe_midpoints`,
`subscribe_orders`, `subscribe_trades`) and handles heartbeats. The `rtds` feature wraps the RTDS socket.
Raw protocol below is for when you need a hand-rolled client (e.g. `tokio-tungstenite`).

---

## Heartbeats (critical — connections drop without these)
- **Market & User channels:** client sends `PING` every **10 seconds**; server replies `PONG`.
- **Sports channel:** server sends `ping` every **5 seconds**; client must reply `pong` within **10 seconds** or the connection closes.
- A connection that drops ~10s after opening almost always means missing heartbeats.
- A connection that closes immediately means you didn't send a valid subscription message right after connecting.

---

## Market Channel (public, level-2 data)
Subscribe by **asset/token IDs**.

Subscribe message:
```json
{
  "assets_ids": ["<token_id_1>", "<token_id_2>"],
  "type": "market",
  "custom_feature_enabled": true
}
```
`custom_feature_enabled: true` is required to receive `best_bid_ask`, `new_market`, and `market_resolved`.

Dynamic (sub)scription without reconnecting:
```json
{ "assets_ids": ["new_id_1","new_id_2"], "operation": "subscribe", "custom_feature_enabled": true }
{ "assets_ids": ["id_to_remove"], "operation": "unsubscribe" }
```

Every message carries an `event_type` field:

| event_type | When emitted | Custom feature? |
| --- | --- | --- |
| `book` | On first subscribe and on any trade that affects the book — full snapshot | No |
| `price_change` | New order placed or order cancelled (level update). `size:"0"` = level removed | No |
| `tick_size_change` | Min tick changes (price > 0.96 or < 0.04) | No |
| `last_trade_price` | A maker+taker match created a trade | No |
| `best_bid_ask` | Best bid/ask changed | Yes |
| `new_market` | New market created | Yes |
| `market_resolved` | Market resolved | Yes |

`book` payload: `{ event_type, asset_id, market, bids:[{price,size}], asks:[{price,size}], timestamp, hash }`
`price_change` payload: `{ market, price_changes:[{asset_id,price,size,side,hash,best_bid,best_ask}], timestamp, event_type }`
`tick_size_change`: `{ event_type, asset_id, market, old_tick_size, new_tick_size, timestamp }`
`last_trade_price`: `{ asset_id, event_type, fee_rate_bps, market, price, side, size, timestamp }`
`best_bid_ask`: `{ event_type, market, asset_id, best_bid, best_ask, spread, timestamp }`
`new_market`: rich payload incl. `id, question, market, slug, description, assets_ids, outcomes, event_message{...}, timestamp, tags, condition_id, active, clob_token_ids, sports_market_type, line, game_start_time, order_price_min_tick_size, group_item_title, taker_base_fee, fees_enabled, fee_schedule{exponent,rate,taker_only,rebate_rate}`
`market_resolved`: like `new_market` plus `winning_asset_id`, `winning_outcome`.

---

## User Channel (authenticated, your orders & trades)
Subscribe by **condition IDs** (market identifiers), NOT asset IDs. Each market = 1 condition ID, 2 asset IDs (Yes/No).
Server-side only — never expose API creds client-side.

Subscribe message:
```json
{
  "auth": { "apiKey": "...", "secret": "...", "passphrase": "..." },
  "markets": ["0x...condition_id"],
  "type": "user"
}
```
Dynamic subscription uses `markets` instead of `assets_ids`:
```json
{ "markets": ["0x...condition_id"], "operation": "subscribe" }
```

Message types (field `type` / `event_type`):

### `trade`
Emitted on a match (`MATCHED`) and subsequent status changes (`MINED`, `CONFIRMED`, `RETRYING`, `FAILED`).
Payload: `{ asset_id, event_type:"trade", id, last_update, maker_orders:[{asset_id,matched_amount,order_id,outcome,owner,price}], market, matchtime, outcome, owner, price, side, size, status, taker_order_id, timestamp, trade_owner, type:"TRADE" }`

Trade status machine:
```
MATCHED → MINED → CONFIRMED
    ↓        ↑
RETRYING ───┘
    ↓
  FAILED
```
| Status | Terminal | Meaning |
| --- | --- | --- |
| `MATCHED` | No | Matched and sent to executor |
| `MINED` | No | Mined into chain, no finality threshold yet |
| `CONFIRMED` | Yes | Strong probabilistic finality, success |
| `RETRYING` | No | Tx failed (revert/reorg), being resubmitted |
| `FAILED` | Yes | Failed, not retried |

### `order`
Emitted on `PLACEMENT`, `UPDATE` (partial match), `CANCELLATION`.
Payload: `{ asset_id, associate_trades, event_type:"order", id, market, order_owner, original_size, outcome, owner, price, side, size_matched, timestamp, type }`

---

## Sports Channel (public)
Endpoint `wss://sports-api.polymarket.com/ws`. No subscription message — connect and receive all active events.
Message type `sport_result`: live game scores, periods, status. Remember the pong-within-10s rule.
Full schema: `api-pages/market-data_websocket_sports.md`, `specs/asyncapi-sports.json`.

---

## RTDS — Real-Time Data Socket (optional auth)
Endpoint `wss://ws-live-data.polymarket.com`. Streams comments and crypto prices (Binance, Chainlink feeds).
Rust: `rtds` feature. Full schema: `api-pages/market-data_websocket_rtds.md`.

---

## Reconnection playbook
1. On connect, immediately send the subscription message (market: `assets_ids`; user: `auth`+`markets`).
2. Start the heartbeat loop appropriate to the channel.
3. On `book` snapshot, replace local book; apply `price_change` deltas (remove level when `size == "0"`).
4. On disconnect: reconnect, re-subscribe, re-request a fresh `book` snapshot (don't trust stale deltas).
5. Track `hash`/`timestamp` to detect gaps; if out of sync, resnapshot via REST `GET /book` or resubscribe.
