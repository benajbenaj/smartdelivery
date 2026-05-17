# Commit 22 - Shopify OAuth Skeleton

The API now has Shopify installation routes:

```http
GET /shopify/install?shop={shop}.myshopify.com
GET /shopify/callback?shop={shop}.myshopify.com&code=...&state=...
```

Install builds the Shopify OAuth authorize URL and sets a short-lived
HTTP-only state cookie.

Callback validates the required external callback params and checks the state
against the cookie before storing the shop record.

This is still a skeleton. A production OAuth flow also needs token exchange,
HMAC verification, nonce/session persistence, and encrypted token storage.
