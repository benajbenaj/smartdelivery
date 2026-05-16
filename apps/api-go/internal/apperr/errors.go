package apperr

import (
	"errors"
	"fmt"
)

var (
	ErrMissingConfig = errors.New("missing required config")
	ErrNotFound      = errors.New("not found")
	ErrValidation    = errors.New("validation failed")
	ErrConflict      = errors.New("conflict")
)

type ValidationError struct {
	Field string
	Rule  string
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("%s: field %q failed rule %q", ErrValidation, err.Field, err.Rule)
}

func (err *ValidationError) Unwrap() error {
	return ErrValidation
}
