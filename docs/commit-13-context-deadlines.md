# Commit 13 — Add Context Deadlines

## Request Deadline Middleware

HTTP requests now pass through middleware that wraps the request context with:

```go
context.WithTimeout(...)
```

The default request timeout is three seconds.

## Cancellation Flow

Handlers pass `ctx.Request.Context()` into the service layer. Services pass it into repositories. Repositories use GORM Gen methods with context-aware query execution.

If the deadline expires, downstream work sees:

```go
context.DeadlineExceeded
```

If the request is cancelled, downstream work can see:

```go
context.Canceled
```

## HTTP Mapping

Central error mapping now returns:

- `context.DeadlineExceeded` -> `504`
- `context.Canceled` -> `408`
