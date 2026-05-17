# Commit 14 - Audit Worker

The API now writes audit logs through an asynchronous worker.

Flow:

1. Service methods create or update a delivery rule.
2. The service publishes an audit event.
3. The audit worker receives the event on an unbuffered channel.
4. The worker writes `audit_logs` with GORM Gen.
5. When the app context is canceled, the worker exits.

The channel is intentionally unbuffered:

```go
events: make(chan Event)
```

That means `Publish` waits until the worker receives the event. The request does
not wait for the audit database insert itself, but it does wait until the worker
accepts the event. If the worker is busy, this creates backpressure instead of
letting audit events pile up in memory.
