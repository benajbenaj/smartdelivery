package repository

import (
	"context"
	"errors"
	"fmt"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/query"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type ShopRepository struct {
	query *query.Query
}

func NewShopRepository(gormDB *gorm.DB) *ShopRepository {
	return &ShopRepository{query: query.Use(gormDB)}
}

func (repo *ShopRepository) FindOrCreateShop(ctx context.Context, domain string) (*model.Shop, error) {
	if domain == "" {
		return nil, &apperr.ValidationError{Field: "shop", Rule: "required"}
	}

	shopQuery := repo.query.Shop
	shop, err := shopQuery.WithContext(ctx).
		Where(shopQuery.Domain.Eq(domain)).
		First()
	if err == nil {
		return shop, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mapShopDatabaseError(err)
	}

	shop = &model.Shop{Domain: domain}
	if err := shopQuery.WithContext(ctx).Create(shop); err != nil {
		mapped := mapShopDatabaseError(err)
		if errors.Is(mapped, apperr.ErrConflict) {
			return shopQuery.WithContext(ctx).
				Where(shopQuery.Domain.Eq(domain)).
				First()
		}
		return nil, mapped
	}

	return shop, nil
}

func mapShopDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: shop", apperr.ErrNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: %s", apperr.ErrConflict, pgErr.ConstraintName)
		case "23503", "23514":
			return fmt.Errorf("%w: %s", apperr.ErrValidation, pgErr.Message)
		}
	}

	return err
}
