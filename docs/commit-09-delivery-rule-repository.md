# Commit 09 — Add Repository Layer

## Repository Pattern

`DeliveryRuleRepository` wraps database access behind app-level methods:

- `CreateRule`
- `GetRule`
- `ListRules`
- `UpdateRuleStatus`

This keeps handlers and services from building GORM queries directly.

## GORM Gen

The repository uses generated query objects from `internal/query`, such as:

```go
rule := repo.query.DeliveryRule
rule.WithContext(ctx).Where(rule.ID.Eq(id)).First()
```

Generated fields like `rule.ID` and `rule.Status` reduce string-based query mistakes.

## Error Mapping

Database errors are mapped to app errors:

- `gorm.ErrRecordNotFound` -> `apperr.ErrNotFound` with `errors.Is`
- PostgreSQL unique violations -> `apperr.ErrConflict` with `errors.As`
- PostgreSQL foreign key/check violations -> `apperr.ErrValidation` with `errors.As`
