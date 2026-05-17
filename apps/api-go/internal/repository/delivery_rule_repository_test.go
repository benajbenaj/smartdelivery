package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"smartdelivery/apps/api-go/internal/apperr"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestCreateRuleReturnsTypedValidationError(t *testing.T) {
	repo := &DeliveryRuleRepository{}

	_, err := repo.CreateRule(context.Background(), nil)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("CreateRule() error = %v, want ErrValidation", err)
	}

	var validationErr *apperr.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("errors.As(err, &validationErr) = false, want true")
	}
	if validationErr.Field != "rule" {
		t.Fatalf("validationErr.Field = %q, want rule", validationErr.Field)
	}
}

func TestMapDatabaseErrorMapsGormNotFound(t *testing.T) {
	err := fmt.Errorf("load delivery rule: %w", gorm.ErrRecordNotFound)

	mapped := mapDatabaseError(err)
	if !errors.Is(mapped, apperr.ErrNotFound) {
		t.Fatalf("mapDatabaseError() = %v, want ErrNotFound", mapped)
	}
}

func TestMapDatabaseErrorMapsPostgresConflict(t *testing.T) {
	err := fmt.Errorf("insert delivery rule: %w", &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "idx_delivery_rules_unique",
	})

	mapped := mapDatabaseError(err)
	if !errors.Is(mapped, apperr.ErrConflict) {
		t.Fatalf("mapDatabaseError() = %v, want ErrConflict", mapped)
	}
}

func TestMapDatabaseErrorMapsPostgresValidation(t *testing.T) {
	err := fmt.Errorf("insert delivery rule: %w", &pgconn.PgError{
		Code:    "23503",
		Message: "foreign key violation",
	})

	mapped := mapDatabaseError(err)
	if !errors.Is(mapped, apperr.ErrValidation) {
		t.Fatalf("mapDatabaseError() = %v, want ErrValidation", mapped)
	}
}
