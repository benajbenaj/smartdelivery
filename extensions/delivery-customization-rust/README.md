# Delivery Customization Function

This extension is the Shopify Function runtime for Smart Delivery Rules.

Important files:

1. `shopify.extension.toml` declares the Shopify Function target and build command.
2. `src/run.graphql` declares the exact checkout data Shopify passes to the function.
3. `src/lib.rs` wires Shopify Rust type generation.
4. `src/run.rs` is the exported Rust function entrypoint.

The current function is intentionally no-op: it returns an empty operations list.
Future commits can read the compact config snapshot and turn active rules into
hide, rename, or sort operations.

`schema.graphql` is a minimal local schema for this skeleton. Refresh it from
Shopify CLI before implementing real checkout behavior.

Build locally after installing Rust:

```sh
rustup target add wasm32-wasip1
cargo build --target=wasm32-wasip1 --release
```
