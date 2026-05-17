# Commit 08 — Add GORM Gen Setup

## GORM Gen

GORM Gen generates type-safe query objects from the GORM models.

The generator lives at:

```text
apps/api-go/cmd/gormgen/main.go
```

Generated code is written to:

```text
apps/api-go/internal/query
```

## Go Generate

Run generation from `apps/api-go`:

```sh
go generate ./...
```

The directive lives in `apps/api-go/generate.go`:

```go
//go:generate go run ./cmd/gormgen
```

## Why Generated Queries Help

Generated query fields reduce stringly typed database code. For example, using generated `DeliveryRule.Status` is safer than typing `"status"` manually in many places.
