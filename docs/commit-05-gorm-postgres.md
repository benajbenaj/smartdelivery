# Commit 05 — Add Database Connection With GORM

## GORM Setup

The `internal/db` package opens PostgreSQL through GORM:

```go
gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
```

It also configures the underlying `database/sql` pool.

## Health Check

Startup runs:

```sql
SELECT 1
```

through:

```go
gormDB.WithContext(ctx)
```

## Why Context Matters

Database calls should receive `context.Context` so shutdowns, request cancellations, and timeouts can stop slow database work instead of waiting forever.
