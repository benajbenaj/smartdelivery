# Commit 15 - Buffered Audit Channel

The audit worker now uses a configurable buffered channel:

```env
AUDIT_BUFFER_SIZE=100
AUDIT_OVERFLOW_BEHAVIOR=block
```

`block` keeps audit delivery reliable. If the buffer is full, `Publish` waits
until the worker catches up and increments the blocked metric.

`drop` keeps requests moving during overload. If the buffer is full, `Publish`
drops the audit event, logs it, and increments the dropped metric.

The trade-off:

- `block`: better reliability, lower throughput under audit pressure.
- `drop`: better request throughput, possible missing audit logs.
