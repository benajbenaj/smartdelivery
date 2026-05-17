# Commit 19 - Reflection Patch Helper

`internal/patch.Apply` copies only non-nil pointer fields from a patch struct to
a target struct with matching field names.

Example:

```go
type request struct {
	Status *model.DeliveryRuleStatus
}

cmd := service.UpdateRuleStatusCommand{ShopID: shopID, ID: id}
patch.Apply(&cmd, request)
```

Why pointer fields:

- `nil` means the field was omitted
- non-nil means the caller wants to update that field

The helper is intentionally constrained:

- target must be a pointer to a struct
- patch must be a struct or pointer to struct
- patch fields must be pointers
- missing or incompatible target fields return an error

Reflection is useful here because the mapper can work across small patch structs,
but keeping the rules strict avoids hiding mistakes.
