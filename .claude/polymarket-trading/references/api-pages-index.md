# Polymarket Docs — Raw Page Index

All official documentation pages archived from docs.polymarket.com (.md form), grouped by area.
Filenames use the URL slug with '/' replaced by '_'. Open any file for the FULL, unabridged page.

## Getting Started & Concepts

- `concepts_markets-events.md` — **Markets & Events**: Understanding the fundamental building blocks of Polymarket
- `concepts_order-lifecycle.md` — **Order Lifecycle**: Understanding how orders flow from creation to settlement
- `concepts_positions-tokens.md` — **Positions & Tokens**: Understanding outcome tokens and how positions work on Polymarket
- `concepts_prices-orderbook.md` — **Prices & Orderbook**: How prices work and how the order book enables peer-to-peer trading
- `concepts_resolution.md` — **Resolution**: How markets are resolved and winning positions redeemed
- `index.md` — **Overview**: Build on the world's largest prediction market. Trade, integrate, and access real-time market data with the Polymarket API.
- `polymarket-101.md` — **Polymarket 101**: An intro to Polymarket - the world's largest prediction market
- `quickstart.md` — **Quickstart**: Fetch a market and place your first order

## API Reference — Core/Auth

- `api-reference_authentication.md` — **Authentication**: How to authenticate requests to the CLOB API
- `api-reference_clients-sdks.md` — **Clients & SDKs**: Official open-source libraries for interacting with Polymarket
- `api-reference_geoblock.md` — **Geographic Restrictions**: Check geographic restrictions before placing orders on the Polymarket API
- `api-reference_introduction.md` — **Introduction**: Overview of the Polymarket APIs
- `api-reference_rate-limits.md` — **Rate Limits**: API rate limits for all Polymarket endpoints

## Trading — SDK Client Methods

- `trading_clients_builder.md` — **Builder Methods**: Methods for querying orders and trades attributed to your builder code.
- `trading_clients_l1.md` — **L1 Methods**: These methods require a wallet signer (private key) but do not require user API credentials. Use these for initial setup.
- `trading_clients_l2.md` — **L2 Methods**: These methods require user API credentials (L2 headers). Use these for placing trades and managing your positions.
- `trading_clients_public.md` — **Public Methods**: These methods can be called without a signer or user credentials. Use these for reading market data, prices, and order books.

## Trading — Orders & Operations

- `trading_fees.md` — **Fees**: Understanding trading fees on Polymarket
- `trading_gasless.md` — **Gasless Transactions**: Execute onchain operations without paying gas fees
- `trading_matching-engine.md` — **Matching Engine Restarts**: Maintenance windows, restart handling, and post-restart post-only mode
- `trading_orderbook.md` — **Orderbook**: Reading the orderbook, prices, spreads, and midpoints
- `trading_orders_attribution.md` — **Order Attribution**: Attribute orders to your builder code for volume credit and fee rewards
- `trading_orders_cancel.md` — **Cancel Order**: Cancel single, multiple, or all open orders
- `trading_orders_create.md` — **Create Order**: Build, sign, and submit orders
- `trading_orders_overview.md` — **Overview**: Order types, tick sizes, and querying orders
- `trading_overview.md` — **Overview**: Trading on the Polymarket CLOB
- `trading_quickstart.md` — **Quickstart**: Place your first order on Polymarket

## Trading — CTF (Split/Merge/Redeem)

- `trading_ctf_merge.md` — **Merge Tokens**: Convert outcome token pairs back to pUSD
- `trading_ctf_overview.md` — **Conditional Token Framework**: Onchain token mechanics powering Polymarket positions
- `trading_ctf_redeem.md` — **Redeem Tokens**: Exchange winning tokens for pUSD after market resolution
- `trading_ctf_split.md` — **Split Tokens**: Convert pUSD into outcome token pairs

## Trading — Bridge (Deposit/Withdraw)

- `trading_bridge_deposit.md` — **Deposit**: Bridge assets from any supported chain to fund your Polymarket account
- `trading_bridge_quote.md` — **Quote**: Preview fees and estimated output for deposits and withdrawals
- `trading_bridge_status.md` — **Deposit Status**: Track the progress of your bridge deposits
- `trading_bridge_supported-assets.md` — **Supported Assets**: Chains and tokens supported for deposits to Polymarket
- `trading_bridge_withdraw.md` — **Withdraw**: Bridge pUSD from Polymarket to any supported chain

## WebSockets

- `api-reference_wss_market.md` — **Market Channel**: Public WebSocket for real-time orderbook, price, and market lifecycle updates.
- `api-reference_wss_sports.md` — **Sports Channel**: Public WebSocket for real-time sports match results.
- `api-reference_wss_user.md` — **User Channel**: Authenticated WebSocket for real-time order and trade updates.
- `market-data_websocket_market-channel.md` — **Market Channel**: Real-time orderbook, price, and trade data
- `market-data_websocket_overview.md` — **Overview**: Real-time market data and trading updates via WebSocket
- `market-data_websocket_rtds.md` — **Real-Time Data Socket**: Stream comments, crypto prices, and equity prices via WebSocket
- `market-data_websocket_sports.md` — **Sports WebSocket**: Live sports scores and game state
- `market-data_websocket_user-channel.md` — **User Channel**: Authenticated order and trade updates

## Market Data

- `market-data_fetching-markets.md` — **Fetching Markets**: Three strategies for discovering and querying markets
- `market-data_overview.md` — **Overview**: Fetch market data with no authentication required

## Market Makers

- `market-makers_getting-started.md` — **Getting Started**: One-time setup for market making on Polymarket
- `market-makers_inventory.md` — **Inventory Management**: Managing outcome token inventory for market making
- `market-makers_liquidity-rewards.md` — **Liquidity Rewards**: Earn rewards for providing liquidity on Polymarket
- `market-makers_maker-rebates.md` — **Maker Rebates Program**: Earn daily USDC rebates by providing liquidity on Polymarket
- `market-makers_overview.md` — **Overview**: Market making on Polymarket
- `market-makers_trading.md` — **Trading**: Order entry, management, and best practices for market makers

## Builders Program

- `builders_api-keys.md` — **Builder Code**: Your builder code for order attribution
- `builders_overview.md` — **Builder Program**: Build applications that route orders through Polymarket
- `builders_tiers.md` — **Tiers**: Rate limits, rewards, and how to upgrade

## Advanced & Resources

- `advanced_neg-risk.md` — **Negative Risk Markets**: Capital-efficient trading for multi-outcome events
- `resources_contract-addresses.md` — **Contracts**: All Polymarket smart contract addresses, audits, and security resources
- `resources_blockchain-data.md` — **Data Resources**: Access Polymarket on-chain activity for data & analytics
- `resources_error-codes.md` — **Error Codes**: Complete reference for CLOB API error responses

## Gamma API — Markets

- `api-reference_markets_get-market-by-id.md` — **Get market by id**
- `api-reference_markets_get-market-by-slug.md` — **Get market by slug**
- `api-reference_markets_get-market-tags-by-id.md` — **Get market tags by id**
- `api-reference_markets_get-prices-history.md` — **Get prices history**: Retrieve historical price data for a market.
- `api-reference_markets_get-sampling-markets.md` — **Get sampling markets**
- `api-reference_markets_get-sampling-simplified-markets.md` — **Get sampling simplified markets**
- `api-reference_markets_get-simplified-markets.md` — **Get simplified markets**
- `api-reference_markets_list-markets.md` — **List markets**

## Gamma API — Events

- `api-reference_events_get-event-by-id.md` — **Get event by id**
- `api-reference_events_get-event-by-slug.md` — **Get event by slug**
- `api-reference_events_get-event-tags.md` — **Get event tags**
- `api-reference_events_list-events.md` — **List events**

## Gamma API — Series

- `api-reference_series_get-series-by-id.md` — **Get series by id**
- `api-reference_series_list-series.md` — **List series**

## Gamma API — Tags

- `api-reference_tags_get-related-tags-relationships-by-tag-id.md` — **Get related tags (relationships) by tag id**
- `api-reference_tags_get-related-tags-relationships-by-tag-slug.md` — **Get related tags (relationships) by tag slug**
- `api-reference_tags_get-tag-by-id.md` — **Get tag by id**
- `api-reference_tags_get-tag-by-slug.md` — **Get tag by slug**
- `api-reference_tags_get-tags-related-to-a-tag-id.md` — **Get tags related to a tag id**
- `api-reference_tags_get-tags-related-to-a-tag-slug.md` — **Get tags related to a tag slug**
- `api-reference_tags_list-tags.md` — **List tags**

## Gamma API — Search / Sports / Comments / Profiles

- `api-reference_comments_get-comments-by-comment-id.md` — **Get comments by comment id**
- `api-reference_comments_get-comments-by-user-address.md` — **Get comments by user address**
- `api-reference_comments_list-comments.md` — **List comments**
- `api-reference_profiles_get-public-profile-by-wallet-address.md` — **Get public profile by wallet address**
- `api-reference_search_search-markets-events-and-profiles.md` — **Search markets, events, and profiles**
- `api-reference_sports_get-sports-metadata-information.md` — **Get sports metadata information**
- `api-reference_sports_get-valid-sports-market-types.md` — **Get valid sports market types**
- `api-reference_sports_list-teams.md` — **List teams**

## Data API — Core (positions/trades/activity/holders/leaderboard)

- `api-reference_core_get-closed-positions-for-a-user.md` — **Get closed positions for a user**
- `api-reference_core_get-current-positions-for-a-user.md` — **Get current positions for a user**
- `api-reference_core_get-positions-for-a-market.md` — **Get positions for a market**
- `api-reference_core_get-top-holders-for-markets.md` — **Get top holders for markets**
- `api-reference_core_get-total-value-of-a-users-positions.md` — **Get total value of a user's positions**
- `api-reference_core_get-trader-leaderboard-rankings.md` — **Get trader leaderboard rankings**
- `api-reference_core_get-trades-for-a-user-or-markets.md` — **Get trades for a user or markets**
- `api-reference_core_get-user-activity.md` — **Get user activity**

## CLOB — Market Data Endpoints

- `api-reference_data_get-midpoint-price.md` — **Get midpoint price**: Retrieves the midpoint price for a specific token ID.
- `api-reference_data_get-server-time.md` — **Get server time**: Returns the current Unix timestamp of the server.
- `api-reference_market-data_get-fee-rate-by-path-parameter.md` — **Get fee rate by path parameter**: Retrieves the base fee rate for a specific token ID using the token ID as a path parameter.
- `api-reference_market-data_get-fee-rate.md` — **Get fee rate**: Retrieves the base fee rate for a specific token ID.
- `api-reference_market-data_get-last-trade-price.md` — **Get last trade price**: Retrieves the last trade price and side for a specific token ID.
- `api-reference_market-data_get-last-trade-prices-query-parameters.md` — **Get last trade prices (query parameters)**: Retrieves last trade prices for multiple token IDs using query parameters.
- `api-reference_market-data_get-last-trade-prices-request-body.md` — **Get last trade prices (request body)**: Retrieves last trade prices for multiple token IDs using a request body.
- `api-reference_market-data_get-market-price.md` — **Get market price**: Retrieves the best market price for a specific token ID and side (bid or ask).
- `api-reference_market-data_get-market-prices-query-parameters.md` — **Get market prices (query parameters)**: Retrieves market prices for multiple token IDs and sides using query parameters.
- `api-reference_market-data_get-market-prices-request-body.md` — **Get market prices (request body)**: Retrieves market prices for multiple token IDs and sides using a request body.
- `api-reference_market-data_get-midpoint-prices-query-parameters.md` — **Get midpoint prices (query parameters)**: Retrieves midpoint prices for multiple token IDs using query parameters.
- `api-reference_market-data_get-midpoint-prices-request-body.md` — **Get midpoint prices (request body)**: Retrieves midpoint prices for multiple token IDs using a request body.
- `api-reference_market-data_get-order-book.md` — **Get order book**: Retrieves the order book summary for a specific token ID.
- `api-reference_market-data_get-order-books-request-body.md` — **Get order books (request body)**: Retrieves order book summaries for multiple token IDs using a request body.
- `api-reference_market-data_get-spread.md` — **Get spread**: Retrieves the spread for a specific token ID.
- `api-reference_market-data_get-spreads.md` — **Get spreads**: Retrieves spreads for multiple token IDs.
- `api-reference_market-data_get-tick-size-by-path-parameter.md` — **Get tick size by path parameter**: Retrieves the minimum tick size (price increment) for a specific token ID using the token ID as a path parameter.
- `api-reference_market-data_get-tick-size.md` — **Get tick size**: Retrieves the minimum tick size (price increment) for a specific token ID.

## CLOB — Trade Endpoints

- `api-reference_trade_cancel-all-orders.md` — **Cancel all orders**: Cancels all open orders for the authenticated user. Works even in cancel-only mode.
- `api-reference_trade_cancel-multiple-orders.md` — **Cancel multiple orders**: Cancels multiple orders by their IDs. Maximum 3000 orders per request.
- `api-reference_trade_cancel-orders-for-a-market.md` — **Cancel orders for a market**: Cancels all open orders for the authenticated user in a specific market (condition) and asset.
- `api-reference_trade_cancel-single-order.md` — **Cancel single order**: Cancels a single order by its ID. Works even in cancel-only mode.
- `api-reference_trade_get-builder-trades.md` — **Get builder trades**: Retrieves trades attributed to a builder code.
- `api-reference_trade_get-order-scoring-status.md` — **Get order scoring status**: Checks if a specific order is currently scoring for rewards.
- `api-reference_trade_get-single-order-by-id.md` — **Get single order by ID**: Retrieves a specific order by its ID (order hash) for the authenticated user.
- `api-reference_trade_get-trades.md` — **Get trades**: Retrieves trades for the authenticated user. Returns paginated results.
- `api-reference_trade_get-user-orders.md` — **Get user orders**: Retrieves open orders for the authenticated user. Returns paginated results.
- `api-reference_trade_post-a-new-order.md` — **Post a new order**: Creates a new order in the order book
- `api-reference_trade_post-multiple-orders.md` — **Post multiple orders**: Creates multiple new orders in the order book. Orders are processed in parallel.
- `api-reference_trade_send-heartbeat.md` — **Send heartbeat**: Sends a heartbeat signal to maintain active session status.

## Misc / Rebates / Builder Analytics / Bridge API

- `api-reference_bridge_create-withdrawal-addresses.md` — **Create withdrawal addresses**
- `api-reference_bridge_get-a-quote.md` — **Get a quote**
- `api-reference_bridge_get-supported-assets.md` — **Get supported assets**
- `api-reference_bridge_get-transaction-status.md` — **Get transaction status**
- `api-reference_builders_get-aggregated-builder-leaderboard.md` — **Get aggregated builder leaderboard**
- `api-reference_builders_get-daily-builder-volume-time-series.md` — **Get daily builder volume time-series**
- `api-reference_misc_download-an-accounting-snapshot-zip-of-csvs.md` — **Download an accounting snapshot (ZIP of CSVs)**
- `api-reference_misc_get-live-volume-for-an-event.md` — **Get live volume for an event**
- `api-reference_misc_get-open-interest.md` — **Get open interest**
- `api-reference_misc_get-total-markets-a-user-has-traded.md` — **Get total markets a user has traded**
- `api-reference_rebates_get-current-rebated-fees-for-a-maker.md` — **Get current rebated fees for a maker**: Returns the current rebated fees for a maker address on a given date.

