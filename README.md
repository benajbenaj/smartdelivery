# Smart Delivery Rules

Smart Delivery Rules is a learning monorepo for a Shopify delivery customization app.

The backend in `apps/api-go` stores merchant delivery rules in PostgreSQL. The Shopify Function in `extensions/delivery-customization-rust` applies a compact rule configuration at checkout without calling the backend.

## Repository Layout

```text
apps/
  api-go/                         Go backend API
extensions/
  delivery-customization-rust/    Shopify Delivery Customization Function
docs/                             Architecture and learning notes
docker/                           Local development services
```

## Local Development

Start PostgreSQL:

```sh
make db-up
```

Stop PostgreSQL:

```sh
make db-down
```

The local database URL is:

```text
postgres://smart_delivery:smart_delivery@localhost:5432/smart_delivery?sslmode=disable
```

## Why The Function Is Separate

Shopify Functions run inside Shopify checkout and must be deterministic, fast, and network-free. The Go backend owns merchant-facing workflows, persistence, and rule management. The Rust function receives a prepared configuration snapshot and applies it during checkout.
