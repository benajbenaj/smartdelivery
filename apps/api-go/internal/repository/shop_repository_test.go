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

func TestFindOrCreateShopRequiresDomain(t *testing.T) {
	repo := &ShopRepository{}

	_, err := repo.FindOrCreateShop(context.Background(), "")
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("FindOrCreateShop() error = %v, want ErrValidation", err)
	}
}

func TestMapShopDatabaseErrorMapsGormNotFound(t *testing.T) {
	err := fmt.Errorf("load shop: %w", gorm.ErrRecordNotFound)

	mapped := mapShopDatabaseError(err)
	if !errors.Is(mapped, apperr.ErrNotFound) {
		t.Fatalf("mapShopDatabaseError() = %v, want ErrNotFound", mapped)
	}
}

func TestMapShopDatabaseErrorMapsPostgresConflict(t *testing.T) {
	err := fmt.Errorf("insert shop: %w", &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "idx_shops_domain",
	})

	mapped := mapShopDatabaseError(err)
	if !errors.Is(mapped, apperr.ErrConflict) {
		t.Fatalf("mapShopDatabaseError() = %v, want ErrConflict", mapped)
	}
}
