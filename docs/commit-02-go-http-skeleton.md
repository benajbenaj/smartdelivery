# Commit 02 — Add Go HTTP Skeleton

## Go Project Layout

The API executable starts in `apps/api-go/cmd/api/main.go`. The `cmd` directory is the usual place for deployable Go binaries.

HTTP routing and server lifecycle code live in `apps/api-go/internal/httpserver`. Routes use Gin, while graceful shutdown still uses Go's standard `http.Server`.

## Context For Shutdown

The server creates a root `context.Context` from operating system signals. When the process receives `SIGINT` or `SIGTERM`, that context is cancelled.

Shutdown then uses a timeout context so in-flight requests have a short window to finish before the process exits.
