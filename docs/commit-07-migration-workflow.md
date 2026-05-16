# Commit 07 — Add Migration Workflow

## Commands

Run migrations from the repository root:

```sh
make migrate-up
make migrate-status
make migrate-down
```

These commands run `apps/api-go/cmd/migrate`, which uses GORM migration APIs.

## GORM AutoMigrate

`make migrate-up` calls:

```go
db.AutoMigrate(&Shop{}, &DeliveryRule{}, &AuditLog{})
```

GORM inspects the model structs and creates or updates tables.

`make migrate-down` uses GORM's `Migrator().DropTable(...)` to remove the current tables in dependency order.

`make migrate-status` uses `Migrator().HasTable(...)` to show whether each model table exists.

## AutoMigrate vs Explicit SQL

`AutoMigrate` is fast for learning projects because the Go model is the source of truth.

Production apps often prefer explicit SQL migrations because schema changes are easier to review and roll back precisely. For this learning app, using GORM migration capabilities keeps the workflow simple while the domain model is still changing.
