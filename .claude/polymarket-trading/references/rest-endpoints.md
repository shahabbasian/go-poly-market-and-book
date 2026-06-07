# Polymarket REST Endpoints — Complete Cheat-Sheet

All HTTP endpoints across the four base APIs, extracted from the official OpenAPI specs in `specs/`.
For full params/schemas of any endpoint, open the matching page in `api-pages/` or the spec file in `specs/`.
Gamma & Data are fully public (no auth). CLOB read endpoints are public; CLOB trading/account endpoints need L2 auth.

## CLOB API
Base URL: `https://clob.polymarket.com` — spec: `specs/clob-openapi.yaml`

| Method | Path | Summary |
| --- | --- | --- |
| POST | `/auth/api-key` | Create API key |
| DELETE | `/auth/api-key` | Delete API key |
| GET | `/auth/api-keys` | Get API keys |
| GET | `/auth/ban-status/closed-only` | Get closed-only mode status |
| GET | `/auth/builder-api-key` | Get builder API keys |
| POST | `/auth/builder-api-key` | Create builder API key |
| DELETE | `/auth/builder-api-key` | Revoke builder API key |
| GET | `/auth/derive-api-key` | Derive API key |
| GET | `/balance-allowance` | Get balance and allowance |
| PUT | `/balance-allowance` | Update balance and allowance |
| GET | `/balance-allowance/update` | Update balance and allowance |
| POST | `/batch-prices-history` | Get batch prices history |
| GET | `/book` | Get order book |
| GET | `/books` | Get order books (query parameters) |
| POST | `/books` | Get order books (request body) |
| GET | `/builder/trades` | Get builder trades |
| DELETE | `/cancel-all` | Cancel all orders |
| DELETE | `/cancel-market-orders` | Cancel orders for a market |
| GET | `/clob-markets/{condition_id}` | Get CLOB market info |
| GET | `/data/order/{orderID}` | Get single order by ID |
| GET | `/data/orders` | Get user orders |
| GET | `/data/trades` | Get trades |
| GET | `/fee-rate` | Get fee rate |
| GET | `/fee-rate/{token_id}` | Get fee rate by path parameter |
| POST | `/heartbeats` | Send heartbeat |
| GET | `/last-trade-price` | Get last trade price |
| GET | `/last-trades-prices` | Get last trade prices (query parameters) |
| POST | `/last-trades-prices` | Get last trade prices (request body) |
| GET | `/markets-by-token/{token_id}` | Get market by token |
| POST | `/markets/live-activity` | Get live activity markets by condition IDs |
| GET | `/markets/live-activity/{condition_id}` | Get live activity market by condition ID |
| GET | `/midpoint` | Get midpoint price |
| GET | `/midpoints` | Get midpoint prices (query parameters) |
| POST | `/midpoints` | Get midpoint prices (request body) |
| GET | `/neg-risk` | Get negative risk flag |
| GET | `/neg-risk/{token_id}` | Get negative risk flag by path parameter |
| GET | `/notifications` | Get notifications |
| DELETE | `/notifications` | Mark notifications as read |
| POST | `/order` | Post a new order |
| DELETE | `/order` | Cancel single order |
| GET | `/order-scoring` | Get order scoring status |
| POST | `/orders` | Post multiple orders |
| DELETE | `/orders` | Cancel multiple orders |
| GET | `/orders-scoring` | Get scoring status for multiple orders |
| POST | `/orders-scoring` | Get scoring status for multiple orders (POST) |
| GET | `/price` | Get market price |
| GET | `/prices` | Get market prices (query parameters) |
| POST | `/prices` | Get market prices (request body) |
| GET | `/prices-history` | Get prices history |
| GET | `/rebates/current` | Get current rebated fees for a maker |
| GET | `/rewards/markets/current` | Get current active rewards configurations |
| GET | `/rewards/markets/multi` | Get multiple markets with rewards |
| GET | `/rewards/markets/{condition_id}` | Get raw rewards for a specific market |
| GET | `/rewards/user` | Get earnings for user by date |
| GET | `/rewards/user/markets` | Get user earnings and markets configuration |
| GET | `/rewards/user/percentages` | Get reward percentages for user |
| GET | `/rewards/user/total` | Get total earnings for user by date |
| GET | `/sampling-markets` | Get sampling markets |
| GET | `/sampling-simplified-markets` | Get sampling simplified markets |
| GET | `/simplified-markets` | Get simplified markets |
| GET | `/spread` | Get spread |
| POST | `/spreads` | Get spreads |
| GET | `/tick-size` | Get tick size |
| GET | `/tick-size/{token_id}` | Get tick size by path parameter |
| GET | `/time` | Get server time |
| POST | `/v1/heartbeats` | Send heartbeat (v1) |

## Gamma API
Base URL: `https://gamma-api.polymarket.com` — spec: `specs/gamma-openapi.yaml`

| Method | Path | Summary |
| --- | --- | --- |
| GET | `/comments` | List comments |
| GET | `/comments/user_address/{user_address}` | Get comments by user address |
| GET | `/comments/{id}` | Get comments by comment id |
| GET | `/events` | List events |
| GET | `/events/creators` | List event creators |
| GET | `/events/creators/{id}` | Get event creator by id |
| GET | `/events/keyset` | List events (keyset pagination) |
| GET | `/events/pagination` | List events (paginated) |
| GET | `/events/results` | List sport events results |
| GET | `/events/slug/{slug}` | Get event by slug |
| GET | `/events/{id}` | Get event by id |
| GET | `/events/{id}/comments/count` | Get event comment count |
| GET | `/events/{id}/tags` | Get event tags |
| GET | `/events/{id}/tweet-count` | Get event tweet count |
| GET | `/markets` | List markets |
| POST | `/markets/abridged` | Query abridged markets by information filters |
| POST | `/markets/information` | Query markets by information filters |
| GET | `/markets/keyset` | List markets (keyset pagination) |
| GET | `/markets/slug/{slug}` | Get market by slug |
| GET | `/markets/{id}` | Get market by id |
| GET | `/markets/{id}/description` | Get market description by id |
| GET | `/markets/{id}/tags` | Get market tags by id |
| GET | `/profiles/user_address/{user_address}` | Get public profile by user address |
| GET | `/public-profile` | Get public profile by wallet address |
| GET | `/public-search` | Search markets, events, and profiles |
| GET | `/series` | List series |
| GET | `/series-summary/slug/{slug}` | Get series summary by slug |
| GET | `/series-summary/{id}` | Get series summary by id |
| GET | `/series/{id}` | Get series by id |
| GET | `/series/{id}/comments/count` | Get series comment count |
| GET | `/sports` | Get sports metadata information |
| GET | `/sports/market-types` | Get valid sports market types |
| GET | `/status` | Gamma API Health check |
| GET | `/tags` | List tags |
| GET | `/tags/slug/{slug}` | Get tag by slug |
| GET | `/tags/slug/{slug}/related-tags` | Get related tags (relationships) by tag slug |
| GET | `/tags/slug/{slug}/related-tags/tags` | Get tags related to a tag slug |
| GET | `/tags/{id}` | Get tag by id |
| GET | `/tags/{id}/related-tags` | Get related tags (relationships) by tag id |
| GET | `/tags/{id}/related-tags/tags` | Get tags related to a tag id |
| GET | `/teams` | List teams |
| GET | `/teams/{id}` | Get team by id |

## Data API
Base URL: `https://data-api.polymarket.com` — spec: `specs/data-openapi.yaml`

| Method | Path | Summary |
| --- | --- | --- |
| GET | `/` | Data API Health check |
| GET | `/activity` | Get user activity |
| GET | `/closed-positions` | Get closed positions for a user |
| GET | `/holders` | Get top holders for markets |
| GET | `/live-volume` | Get live volume for an event |
| GET | `/oi` | Get open interest |
| GET | `/other` | Get "Other" size for an augmented neg risk event and user |
| GET | `/positions` | Get current positions for a user |
| GET | `/revisions` | Get moderated revisions for a question |
| GET | `/traded` | Get total markets a user has traded |
| GET | `/trades` | Get trades for a user or markets |
| GET | `/v1/accounting/snapshot` | Download an accounting snapshot (ZIP of CSVs) |
| GET | `/v1/builders/leaderboard` | Get aggregated builder leaderboard |
| GET | `/v1/builders/volume` | Get daily builder volume time-series |
| GET | `/v1/leaderboard` | Get trader leaderboard rankings |
| GET | `/v1/market-positions` | Get positions for a market |
| GET | `/value` | Get total value of a user's positions |

## Bridge API
Base URL: `https://bridge.polymarket.com` — spec: `specs/bridge-openapi.yaml`

| Method | Path | Summary |
| --- | --- | --- |
| POST | `/deposit` | Create bridge addresses |
| POST | `/quote` | Get a quote |
| GET | `/status/{address}` | Get transaction status |
| GET | `/supported-assets` | Get supported assets |
| POST | `/withdraw` | Create withdrawal addresses |
