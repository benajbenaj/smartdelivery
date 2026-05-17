# Commit 12 — Add Joined Errors For Batch Validation

## Joined Validation Errors

Service validation now collects all invalid fields and returns:

```go
return errors.Join(errs...)
```

This lets callers receive every form problem at once instead of fixing one field per request.

## Inspection

Joined errors still work with:

```go
errors.Is(err, apperr.ErrValidation)
errors.As(err, &validationErr)
```

HTTP responses return a single validation error in the old shape, or multiple validation errors in an `errors` array.
