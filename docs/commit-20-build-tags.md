# Commit 20 - Integration Build Tags

Integration tests use a build tag:

```go
//go:build integration
```

Normal unit tests stay fast:

```bash
make test
```

Integration tests opt in to slower external dependencies:

```bash
make test-integration
```

`make test-integration` starts PostgreSQL with Docker Compose and runs:

```bash
go test -tags=integration ./...
```

Without the tag, Go ignores files marked with `//go:build integration`.
