//! Read-only market data — no signer, no credentials.
//! Run: `cargo run --bin read`
use polymarket_client_sdk_v2::clob::Client;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Public CLOB client. Use https://clob-v2.polymarket.com for V2 (pUSD),
    // or https://clob.polymarket.com for V1 (USDC.e). Protocol auto-detected.
    let client = Client::default();

    // Health check
    let ok = client.ok().await?;
    println!("CLOB ok: {ok}");

    // Server time (handy for syncing local clock before signing)
    // let t = client.server_time().await?;

    // Fetch markets (paginated). Inspect `condition_id` and the two token IDs
    // (Yes/No outcomes) you will trade on.
    let markets = client.markets(None).await?;
    println!("fetched a page of markets: {markets:?}");

    // Order book for a specific token id:
    // let book = client.order_book("<token-id>".parse()?).await?;
    // println!("best bid {:?} / best ask {:?}", book.bids.first(), book.asks.first());

    Ok(())
}
