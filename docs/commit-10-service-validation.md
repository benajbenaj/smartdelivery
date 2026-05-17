# Commit 10 — Add Service Layer Validation

## Service Layer

`DeliveryRuleService` accepts command structs and validates them before calling the repository.

Commands:

- `CreateRuleCommand`
- `GetRuleCommand`
- `ListRulesCommand`
- `UpdateRuleStatusCommand`

## Domain Validation

The service validates required fields, supported enum values, and rule/action combinations.

For now, `sort` rules are only valid for `customer_tag` conditions. `hide` and `rename` can work with any supported condition.

## Thin HTTP Handlers

Future handlers should parse HTTP input into commands, call the service, and translate returned errors into HTTP responses. They should not contain domain validation rules.
