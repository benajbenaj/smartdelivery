# Commit 21 - Generated API Docs

API docs are generated with:

```bash
cd apps/api-go
go generate ./...
```

The generator reads the registered Gin routes and writes:

```text
docs/generated/openapi.json
```

The generated file should be committed because it is useful to read without
running Go tooling. The generator code is the source of truth for how the file is
produced.

Good generated-file rule:

- commit small generated artifacts that help humans or CI
- do not hand edit generated files
- update them by rerunning `go generate`
