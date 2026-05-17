# Commit 16 - Select Worker Shutdown

The audit worker now uses `select` for both normal event reads and shutdown:

```go
select {
case event := <-worker.events:
case <-ctx.Done():
}
```

When shutdown starts, the worker drains events already accepted into the buffered
channel before exiting. It uses a separate drain timeout so canceled app context
does not immediately cancel the final audit writes.

This means:

- new events can still follow normal overflow rules while the app is running
- accepted buffered events get a final chance to be written on shutdown
- the worker still exits cleanly when context is canceled
