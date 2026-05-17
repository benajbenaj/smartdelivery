# Commit 17 - Fan-Out Validation Workers

Bulk rule imports now have a validation endpoint:

```http
POST /shops/{shop}/delivery-rules/imports
```

The request body accepts `rules` plus optional `worker_count`. The service sends
each rule into a jobs channel and starts multiple worker goroutines to validate
rules in parallel. Results are collected, sorted back into input order, and
returned per row.

Fan-out shape:

```text
import request
-> jobs channel
-> validation worker 1
-> validation worker 2
-> validation worker N
-> results channel
-> ordered response
```

This is useful for larger imports because validation work can be split across a
small worker pool instead of running one row at a time.
