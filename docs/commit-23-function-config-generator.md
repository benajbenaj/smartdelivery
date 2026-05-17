# Commit 23 - Function Config Generator

The backend can now build a compact JSON config for the delivery customization
function from active delivery rules.

The generator:

1. Loads active rules for a shop.
2. Sorts them by priority, then ID.
3. Converts them to compact JSON.
4. Stores the JSON as a `function_config_snapshots` row.

The Shopify Function runtime stays separate from this backend code. The backend
prepares versioned config snapshots; the function can later consume the compact
JSON without needing database access or GORM models.

Tests assert exact JSON strings so output order stays deterministic.
