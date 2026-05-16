# Commit 04 — Add Error Taxonomy

## Sentinel Errors

Sentinel errors are stable package-level values. Use `errors.Is` to check the error category.

Examples:

- `apperr.ErrNotFound`
- `apperr.ErrValidation`
- `apperr.ErrConflict`

## Typed Errors

Typed errors carry structured details.

`apperr.ValidationError` stores:

- `Field`
- `Rule`

It unwraps to `apperr.ErrValidation`, so callers can use both:

```go
errors.Is(err, apperr.ErrValidation)
errors.As(err, &validationErr)
```

`errors.As` finds a typed error in an error chain. Go does not have `errors.Astype`.

A type assertion is different:

```go
v, ok := err.(*apperr.ValidationError)
```
