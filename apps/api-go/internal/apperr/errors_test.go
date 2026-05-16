package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidationErrorMatchesSentinel(t *testing.T) {
	err := &ValidationError{
		Field: "name",
		Rule:  "required",
	}

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("errors.Is(err, ErrValidation) = false, want true")
	}
}

func TestWrappedValidationErrorMatchesSentinel(t *testing.T) {
	err := fmt.Errorf("create delivery rule: %w", &ValidationError{
		Field: "condition_type",
		Rule:  "supported_value",
	})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("errors.Is(err, ErrValidation) = false, want true")
	}
}

func TestErrorsAsFindsValidationError(t *testing.T) {
	err := fmt.Errorf("create delivery rule: %w", &ValidationError{
		Field: "action_type",
		Rule:  "required",
	})

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("errors.As(err, &validationErr) = false, want true")
	}

	if validationErr.Field != "action_type" {
		t.Fatalf("validationErr.Field = %q, want action_type", validationErr.Field)
	}
	if validationErr.Rule != "required" {
		t.Fatalf("validationErr.Rule = %q, want required", validationErr.Rule)
	}
}
