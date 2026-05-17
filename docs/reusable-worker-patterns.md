# Reusable Worker Patterns

Shared concurrency helpers live in `apps/api-go/internal/worker`.

`QueueWorker` is for one background worker that consumes queued items. Audit logs
use this pattern because requests publish audit events and one audit worker
writes them.

`FanOut` is for splitting jobs across multiple workers. Import validation uses
this pattern because many rule rows can be validated in parallel.

`FanIn` is for merging many result channels into one result stream. Import
validation uses this after fan-out so callers can read from one output channel.

Shape:

```text
audit:       publishers -> QueueWorker -> store
validation: jobs -> FanOut workers -> FanIn results
```
