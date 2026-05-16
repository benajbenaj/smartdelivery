# Application 1 — Go + GORM + Shopify Delivery Customization Function

## Business idea

**Smart Delivery Rules Admin** is a merchant-facing app for Shopify stores.  
The merchant defines delivery rules in a Go backend, stores them in PostgreSQL, and deploys a Rust Shopify **Delivery Customization Function** that hides, renames, or sorts checkout delivery options.

Example merchant rules:

- Hide express shipping for hazardous products.
- Rename local pickup to “Pickup today before 18:00”.
- Sort low-carbon delivery above express delivery for eco-tagged customers.
- Hide pickup when the delivery address is outside a supported postal-code range.

The app is intentionally built as a learning project. The commits are small, and each commit introduces one or two technical concepts.

## Main stack

- Go backend
- PostgreSQL
- GORM
- GORM Gen
- GORM CLI / migration workflow
- Shopify Admin API integration layer
- Rust Shopify Function extension
- Docker Compose
- Makefile
- Unit, integration, and function tests

## Shopify Function used

**Delivery Customization Function API**

Reason: this API teaches how Shopify Functions can affect checkout delivery options by renaming, sorting, or hiding delivery methods. The Rust function should stay deterministic and fast. It should receive rule-like configuration through Shopify metafields or function configuration, not call the Go backend at runtime.

## Domain model

Core entities:

- `Shop`
- `DeliveryRule`
- `RuleCondition`
- `RuleAction`
- `AuditLog`
- `FunctionDeployment`

Recommended rule shape:

```text
DeliveryRule
- id
- shop_domain
- name
- priority
- status: draft | active | disabled
- condition_type: country | postal_code | cart_total | product_tag | customer_tag
- condition_value
- action_type: hide | rename | sort
- action_value
- created_at
- updated_at
```

## User stories and acceptance criteria

### US-01 — Register a Shopify shop

As a merchant, I want to connect my Shopify shop so that I can manage delivery rules.

Acceptance criteria:

- Given a new shop domain, when the app receives a valid installation callback, then a `Shop` record is created.
- Given an existing shop domain, when the callback is received again, then the shop record is updated, not duplicated.
- Given invalid input, the API returns a typed validation error.

### US-02 — Create delivery rule

As a merchant, I want to create a delivery rule so that checkout delivery options can change automatically.

Acceptance criteria:

- Rule name, condition, and action are required.
- Invalid condition/action pairs are rejected.
- Successful creation returns the created rule.
- Audit log entry is created.

### US-03 — List and filter delivery rules

As a merchant, I want to list delivery rules by status so that I can review active checkout behavior.

Acceptance criteria:

- Rules can be filtered by status.
- Rules are sorted by priority.
- API supports request cancellation through `context.Context`.

### US-04 — Activate rule

As a merchant, I want to activate a tested rule so that it becomes available to the Shopify Function.

Acceptance criteria:

- Only valid rules can be activated.
- Activation stores a deployable config snapshot.
- Concurrent activation requests do not corrupt priority ordering.

### US-05 — Generate function config

As a merchant, I want the app to generate a compact Shopify Function configuration so that the Rust function can run without network calls.

Acceptance criteria:

- Active rules are serialized into a compact JSON configuration.
- Config output is deterministic.
- Unsupported rules are skipped with an audit log entry.

### US-06 — Rust delivery customization execution

As a buyer, I want delivery options to reflect merchant rules during checkout.

Acceptance criteria:

- The Rust function hides matching delivery options.
- The Rust function can rename matching delivery options.
- The Rust function can sort matching delivery options.
- Function tests cover matching and non-matching carts.

## Commit-by-commit plan

### Commit 01 — Initialize repository

Message:

```text
chore: initialize smart delivery rules app
```

Do:

- Create monorepo structure:

```text
/apps/api-go
/extensions/delivery-customization-rust
/docs
/docker
```

- Add `README.md`, `.gitignore`, `.editorconfig`, `Makefile`.
- Add Docker Compose with PostgreSQL.

Learn:

- How to structure a business app as a monorepo.
- Why the Shopify Function is separated from the backend.

---

### Commit 02 — Add Go HTTP skeleton

Message:

```text
feat(api): add minimal Go HTTP server
```

Do:

- Create `cmd/api/main.go`.
- Add health endpoint: `GET /healthz`.
- Add graceful shutdown.
- Pass `context.Context` from signal handling to server shutdown.

Learn:

- Go project layout.
- `context.Context` as a cancellation and timeout carrier.

---

### Commit 03 — Add configuration package

Message:

```text
feat(api): load configuration from environment
```

Do:

- Add config package.
- Read database URL, port, app environment, Shopify app credentials.
- Return custom config errors.

Learn:

- `errors.New` for static sentinel errors.
- `fmt.Errorf("...: %w", err)` for error wrapping.

Example concepts:

```go
var ErrMissingConfig = errors.New("missing required config")
return fmt.Errorf("%w: DATABASE_URL", ErrMissingConfig)
```

---

### Commit 04 — Add error taxonomy

Message:

```text
feat(api): add typed application errors
```

Do:

- Create `internal/apperr`.
- Define sentinel errors: `ErrNotFound`, `ErrValidation`, `ErrConflict`.
- Define a custom struct error:

```go
type ValidationError struct {
    Field string
    Rule  string
}
```

- Add examples/tests for:

```go
errors.Is(err, ErrValidation)
errors.As(err, &validationErr)
```

Learn:

- Difference between sentinel errors and typed errors.
- `errors.Is`.
- `errors.As`.
- Important correction: Go has `errors.As`, but not `errors.Astype`. A type assertion is a separate Go language feature: `v, ok := err.(*MyError)`.

---

### Commit 05 — Add database connection with GORM

Message:

```text
feat(db): connect PostgreSQL with GORM
```

Do:

- Add GORM and PostgreSQL driver.
- Create `internal/db`.
- Add connection health check.
- Ensure DB operations receive `context.Context`.

Learn:

- GORM setup.
- `db.WithContext(ctx)`.
- Why database calls should be cancellable.

---

### Commit 06 — Add first GORM models

Message:

```text
feat(db): add shop and delivery rule models
```

Do:

- Add `Shop`, `DeliveryRule`, `AuditLog`.
- Use GORM tags.
- Add model unit tests where useful.

Learn:

- GORM model tags.
- ID, timestamps, indexes, and unique constraints.

---

### Commit 07 — Add migration workflow

Message:

```text
feat(db): add migration commands
```

Do:

- Add migration tool command through Makefile.
- Add initial SQL migration.
- Add commands:

```text
make migrate-up
make migrate-down
make migrate-status
```

- Document how GORM AutoMigrate differs from explicit migrations.

Learn:

- Why production apps should prefer explicit migrations.
- How to keep schema changes reviewable.

---

### Commit 08 — Add GORM Gen setup

Message:

```text
feat(db): generate type-safe query layer with gorm gen
```

Do:

- Add `cmd/gormgen/main.go`.
- Generate query code into `internal/query`.
- Add `//go:generate go run ./cmd/gormgen`.

Learn:

- GORM Gen.
- `go generate`.
- Why generated query code can reduce runtime mistakes.

---

### Commit 09 — Add repository layer

Message:

```text
feat(api): add delivery rule repository
```

Do:

- Add repository methods:
  - `CreateRule`
  - `GetRule`
  - `ListRules`
  - `UpdateRuleStatus`
- Use generated GORM Gen query objects.

Learn:

- Repository pattern.
- How to map database errors to app errors with `errors.Is` and `errors.As`.

---

### Commit 10 — Add service layer validation

Message:

```text
feat(api): validate delivery rule commands
```

Do:

- Add command structs.
- Validate rule/action combinations.
- Return `ValidationError`.

Learn:

- Domain validation.
- Custom error struct.
- How to keep HTTP handlers thin.

---

### Commit 11 — Add REST endpoints for rules

Message:

```text
feat(api): expose delivery rule CRUD endpoints
```

Do:

- Add:
  - `POST /shops/{shop}/delivery-rules`
  - `GET /shops/{shop}/delivery-rules`
  - `PATCH /shops/{shop}/delivery-rules/{id}`
- Convert app errors into HTTP status codes.

Learn:

- REST handler design.
- Centralized error-to-response mapping.

---

### Commit 12 — Add joined errors for batch validation

Message:

```text
feat(api): return joined validation errors
```

Do:

- When multiple fields are invalid, collect errors and return:

```go
return errors.Join(errs...)
```

- Add test that checks individual errors using `errors.Is` / `errors.As`.

Learn:

- `errors.Join`.
- Why joined errors are useful for form validation.

---

### Commit 13 — Add context deadlines

Message:

```text
feat(api): enforce request deadlines
```

Do:

- Add middleware that creates a timeout context.
- Ensure service and repository methods accept `ctx`.
- Add a test for cancelled context.

Learn:

- `context.WithTimeout`.
- `context.Canceled`.
- `context.DeadlineExceeded`.

---

### Commit 14 — Add asynchronous audit pipeline with unbuffered channel

Message:

```text
feat(audit): write audit logs through unbuffered channel
```

Do:

- Create audit worker.
- Send audit events through an unbuffered channel.
- Shut down worker using context.

Learn:

- Unbuffered channels.
- Backpressure.
- Worker lifecycle with context.

---

### Commit 15 — Switch audit pipeline to buffered channel

Message:

```text
perf(audit): use buffered audit event channel
```

Do:

- Add configurable buffer size.
- Add overflow behavior.
- Add metrics/logging for dropped events or blocking behavior.

Learn:

- Buffered channels.
- Trade-off between throughput and reliability.

---

### Commit 16 — Add select-based worker shutdown

Message:

```text
feat(audit): handle worker shutdown with select
```

Do:

- Use:

```go
select {
case event := <-events:
case <-ctx.Done():
}
```

- Add graceful drain behavior.

Learn:

- `select`.
- Coordinating cancellation and channel reads.

---

### Commit 17 — Add fan-out validation workers

Message:

```text
feat(validation): validate rule imports with fan-out workers
```

Do:

- Add bulk rule import endpoint.
- Split validation jobs across worker goroutines.
- Collect results.

Learn:

- Fan-out pattern.
- Worker pools.

---

### Commit 18 — Add fan-in result aggregation

Message:

```text
feat(validation): aggregate import results with fan-in
```

Do:

- Merge validation result channels into one output channel.
- Preserve original row number in each result.

Learn:

- Fan-in pattern.
- Closing channels safely.

---

### Commit 19 — Add reflection-based patch helper

Message:

```text
feat(api): add reflection-based patch mapper
```

Do:

- Implement a small reflection helper that copies only non-nil patch fields to a target struct.
- Keep it constrained and tested.

Learn:

- `reflect.Value`, `reflect.Type`, `Kind`.
- Why reflection is powerful but should be limited.

---

### Commit 20 — Add build tags

Message:

```text
chore(build): add integration build tags
```

Do:

- Mark integration tests:

```go
//go:build integration
```

- Add Makefile commands:

```text
make test
make test-integration
```

Learn:

- Go build tags.
- Separating fast unit tests from slower integration tests.

---

### Commit 21 — Add generated API docs

Message:

```text
chore(api): generate route documentation
```

Do:

- Add `go:generate` command for docs generation.
- Generate route list or OpenAPI skeleton.

Learn:

- `//go:generate`.
- What belongs in generated files and what should be committed.

---

### Commit 22 — Add Shopify OAuth skeleton

Message:

```text
feat(shopify): add shop installation flow skeleton
```

Do:

- Add install endpoint.
- Add callback endpoint.
- Validate required callback params.
- Store shop record.

Learn:

- OAuth flow shape.
- Safe error handling around external callbacks.

---

### Commit 23 — Add function config generator

Message:

```text
feat(function-config): generate delivery customization config
```

Do:

- Convert active rules to compact JSON.
- Store config snapshot in DB.
- Add deterministic snapshot tests.

Learn:

- Keeping backend config generation separate from function runtime.
- Deterministic output tests.

---

### Commit 24 — Initialize Rust Shopify Function extension

Message:

```text
feat(function): initialize delivery customization function
```

Do:

- Create Shopify Function extension for Delivery Customization.
- Add Rust crate structure.
- Add function input query.
- Add minimal no-op output.

Learn:

- Shopify Function extension anatomy.
- Rust function entrypoint.
- GraphQL input query.

---

### Commit 25 — Implement hide delivery option logic

Message:

```text
feat(function): hide delivery options by rule
```

Do:

- Parse config.
- Match delivery option title/code/cost/address.
- Return hide operation.

Learn:

- Rust pattern matching.
- Deterministic checkout logic.

---

### Commit 26 — Implement rename and sort delivery options

Message:

```text
feat(function): rename and sort delivery options
```

Do:

- Add rename operation.
- Add sort operation.
- Add fixture tests.

Learn:

- Multiple operation types in a Shopify Function.
- Test-first function development.

---

### Commit 27 — Add end-to-end rule activation test

Message:

```text
test(e2e): cover rule creation to function config
```

Do:

- Create shop.
- Create rule.
- Activate rule.
- Generate config.
- Validate expected function config.

Learn:

- Integration testing across layers.
- Keeping Function runtime separate from backend tests.

---

### Commit 28 — Add observability

Message:

```text
feat(observability): add structured logging and request ids
```

Do:

- Add request ID middleware.
- Add structured logs.
- Include shop domain and rule ID where safe.

Learn:

- Operational basics.
- Why logs must not expose secrets.

---

### Commit 29 — Add hardening and security checks

Message:

```text
chore(security): validate inputs and protect secrets
```

Do:

- Validate shop domain format.
- Avoid logging tokens.
- Add secure config checks.
- Add dependency scanning command.

Learn:

- Security basics for Shopify apps.
- Defensive validation.

---

### Commit 30 — Final learning documentation

Message:

```text
docs: add learning guide for Go and Shopify Function concepts
```

Do:

- Add documentation for:
  - Go errors
  - GORM / GORM Gen
  - migrations
  - channels
  - context
  - reflection
  - build tags
  - go:generate
  - Shopify Delivery Customization Function

Learn:

- How to explain architecture and trade-offs.

## Suggested milestone order

1. Backend skeleton and error handling.
2. Database and generated query layer.
3. Rule CRUD.
4. Go concurrency learning features.
5. Shopify installation and config generation.
6. Rust Delivery Customization Function.
7. Tests, observability, hardening.

## Definition of done

- All endpoints have tests.
- Migration up/down works.
- GORM Gen code is reproducible.
- Unit tests run without Docker.
- Integration tests run with Docker.
- Rust function has fixture tests.
- README explains each learning topic.
