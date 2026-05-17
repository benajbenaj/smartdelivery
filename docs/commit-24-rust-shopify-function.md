# Commit 24 - Rust Shopify Function Extension

The delivery customization extension now has a Rust Shopify Function skeleton.

The extension includes:

1. `shopify.extension.toml` for the function target, input query, export, and build command.
2. `src/run.graphql` for the checkout data Shopify sends into the function.
3. `src/lib.rs` for Shopify Rust type generation.
4. `src/run.rs` for the exported function entrypoint.
5. A minimal no-op output that returns no delivery option operations.

Shopify Function anatomy:

1. Shopify runs the GraphQL input query.
2. Shopify passes that JSON input into the Wasm function.
3. The Rust entrypoint returns operations, such as hide, rename, or move.

This first version returns an empty operations list so checkout behavior stays
unchanged until rule execution is implemented.
