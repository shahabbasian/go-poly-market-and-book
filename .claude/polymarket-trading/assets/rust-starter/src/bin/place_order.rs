//! Authenticated trading — derives API creds (L1) then places orders (L2).
//! Set POLYMARKET_PRIVATE_KEY in the environment. Run: `cargo run --bin trade`
//!
//! EOA/MetaMask wallets MUST set token allowances once before trading
//! (see the SDK `examples/approvals.rs`). Proxy/Safe wallets do not.
use std::str::FromStr as _;

use alloy::signers::Signer as _;
use alloy::signers::local::LocalSigner;
use polymarket_client_sdk_v2::{POLYGON, PRIVATE_KEY_VAR};
use polymarket_client_sdk_v2::clob::{Client, Config};
use polymarket_client_sdk_v2::clob::types::{Amount, OrderType, Side};
use polymarket_client_sdk_v2::types::Decimal;
use rust_decimal_macros::dec;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let private_key = std::env::var(PRIVATE_KEY_VAR).expect("set POLYMARKET_PRIVATE_KEY");
    let signer = LocalSigner::from_str(&private_key)?.with_chain_id(Some(POLYGON));

    // authenticate() does L1 -> derive/create API creds -> ready for L2 calls.
    // For deposit-wallet (new API users): add
    //   .signature_type(SignatureType::Poly1271).funder(deposit_wallet_addr)
    // For browser/email wallets: .signature_type(SignatureType::GnosisSafe) (funder auto-derived).
    let client = Client::new("https://clob-v2.polymarket.com", Config::default())?
        .authentication_builder(&signer)
        .authenticate()
        .await?;

    let token_id = "<token-id>"; // a Yes/No outcome token id from a market

    // --- Limit order (resting, GTC by default) ---
    let limit = client
        .limit_order()
        .token_id(token_id)
        .size(Decimal::from(10)) // 10 shares
        .price(dec!(0.45))       // price in [0,1]
        .side(Side::Buy)
        .build()
        .await?;
    let signed = client.sign(&signer, limit).await?;
    let resp = client.post_order(signed).await?;
    println!("limit order resp: {resp:?}");

    // --- Market order (FOK by default) ---
    let market = client
        .market_order()
        .token_id(token_id)
        .amount(Amount::usdc(dec!(5))?) // spend $5 (BUY uses USDC amount)
        .side(Side::Buy)
        .order_type(OrderType::FOK)
        .build()
        .await?;
    let signed = client.sign(&signer, market).await?;
    let resp = client.post_order(signed).await?;
    println!("market order resp: {resp:?}");

    // Manage:
    // let open = client.open_orders(Default::default()).await?;
    // client.cancel_all().await?;

    Ok(())
}
