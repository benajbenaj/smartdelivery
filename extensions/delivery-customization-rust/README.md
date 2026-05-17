# Delivery Customization Function

This extension is the Shopify Function runtime for Smart Delivery Rules.

Important files:

1. `shopify.extension.toml` declares the Shopify Function target and build command.
2. `src/run.graphql` declares the exact checkout data Shopify passes to the function.
3. `src/lib.rs` wires Shopify Rust type generation.
4. `src/run.rs` is the exported Rust function entrypoint.

The current function reads the compact config snapshot and returns hide
operations for matching delivery rules. Rename and sort rules can be added later.

`schema.graphql` is a minimal local schema for this skeleton. Refresh it from
Shopify CLI before implementing real checkout behavior.

Build locally after installing Rust:

```sh
rustup target add wasm32-wasip1
cargo build --target=wasm32-wasip1 --release
```
