# Commit 18 - Fan-In Result Aggregation

Import validation now gives each worker its own result channel, then merges
those channels into one output channel.

Fan-in shape:

```text
worker 1 results -
worker 2 results -- fan-in -> merged results
worker N results -
```

Each validation result keeps:

- `index`: zero-based input position, useful for sorting
- `row_number`: one-based original row number, useful for users

The fan-in helper closes the merged output only after every worker result
channel has closed. That avoids sending on closed channels and lets callers
range over one final result stream safely.
