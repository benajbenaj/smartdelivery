# Commit 03 — Add Configuration Package

## Environment Configuration

The API loads configuration from environment variables in `apps/api-go/internal/config`.

Required values:

- `DATABASE_URL`
- `SHOPIFY_API_KEY`
- `SHOPIFY_API_SECRET`

Defaulted values:

- `APP_ENV=development`
- `PORT=8080`

## Custom Errors

`apperr.ErrMissingConfig` is a sentinel error created with `errors.New`.

Missing values wrap that sentinel with `fmt.Errorf`, for example:

```go
return Config{}, fmt.Errorf("%w: DATABASE_URL", apperr.ErrMissingConfig)
```

Callers can check the error type with `errors.Is(err, apperr.ErrMissingConfig)` while still seeing which variable is missing.
