# Commit 06 — Add First GORM Models

## Models

The first persistence models are:

- `Shop`
- `DeliveryRule`
- `AuditLog`

They live in `apps/api-go/internal/db` because they describe the database shape.

## GORM Tags

The models use tags for:

- primary keys: `gorm:"primaryKey"`
- required fields: `gorm:"not null"`
- indexes: `gorm:"index"`
- unique constraints: `gorm:"uniqueIndex"`
- string sizes and text columns

## IDs And Timestamps

`uint` IDs are used as simple auto-incrementing primary keys.

`CreatedAt` and `UpdatedAt` are recognized by GORM automatically. `AuditLog` only has `CreatedAt` because audit entries should be append-only.
