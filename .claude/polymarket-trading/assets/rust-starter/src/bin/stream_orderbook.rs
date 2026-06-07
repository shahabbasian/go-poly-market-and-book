//! Real-time orderbook streaming over WebSocket (public market channel).
//! Requires the `ws` feature. Run: `cargo run --bin stream`
use std::str::FromStr as _;

use futures::StreamExt as _;
use polymarket_client_sdk_v2::clob::ws::Client;
use polymarket_client_sdk_v2::types::U256;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let client = Client::default();

    // Subscribe to one or more outcome token ids. The SDK handles the
    // PING/PONG heartbeat (every 10s) and reconnection plumbing for you.
    let asset_ids = vec![U256::from_str("<asset-id>")?];
    let stream = client.subscribe_orderbook(asset_ids)?;
    let mut stream = Box::pin(stream);

    while let Some(book) = stream.next().await {
        let book = book?;
        println!(
            "book {} -> {} bids / {} asks",
            book.asset_id,
            book.bids.len(),
            book.asks.len()
        );
    }

    // Other typed streams:
    //   client.subscribe_prices(asset_ids)     -> price_change events
    //   client.subscribe_midpoints(asset_ids)  -> midpoint updates
    //   client.subscribe_orders(...)           -> user order updates (auth)
    //   client.subscribe_trades(...)           -> user trade executions (auth)
    Ok(())
}
