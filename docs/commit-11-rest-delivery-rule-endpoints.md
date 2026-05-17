# Commit 11 — Add REST Endpoints For Rules

## Endpoints

- `POST /shops/{shop}/delivery-rules`
- `GET /shops/{shop}/delivery-rules`
- `PATCH /shops/{shop}/delivery-rules/{id}`

The `{shop}` and `{id}` path values are numeric IDs.

## Handler Design

Handlers parse HTTP input, build service command structs, call the service, and write HTTP responses.

Domain validation stays in the service layer. Database access stays in the repository layer.

## Error Mapping

HTTP errors are centralized in `internal/httpserver/errors.go`:

- validation errors -> `400`
- not found -> `404`
- conflicts -> `409`
- unexpected errors -> `500`
