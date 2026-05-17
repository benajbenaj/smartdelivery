# Commit 25 - Hide Delivery Options By Rule

The Rust Shopify Function now reads the compact config snapshot and returns
hide operations for matching rules.

Supported hide matching:

1. Rule conditions: `country`, `postal_code`, and `cart_total`.
2. Delivery option targets by handle, title, code, or cost.
3. Postal-code lists and ranges, such as `10001,10002` or `10000-19999`.
4. Cart or delivery amounts with `>=`, `<=`, `>`, `<`, exact, or range checks.

The function stays deterministic by using only Shopify input data and the
precomputed config metafield. It does not call the Go API or database during
checkout.
