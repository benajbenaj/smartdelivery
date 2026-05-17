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

type DeliveryRuleRepository struct {
	query *query.Query
}

type ListRulesFilter struct {
	ShopID uint
	Status *model.DeliveryRuleStatus
}

func NewDeliveryRuleRepository(gormDB *gorm.DB) *DeliveryRuleRepository {
	return &DeliveryRuleRepository{
		query: query.Use(gormDB),
	}
}

func (repo *DeliveryRuleRepository) CreateRule(ctx context.Context, rule *model.DeliveryRule) (*model.DeliveryRule, error) {
	if rule == nil {
		return nil, &apperr.ValidationError{Field: "rule", Rule: "required"}
	}

	if err := repo.query.DeliveryRule.WithContext(ctx).Create(rule); err != nil {
		return nil, mapDatabaseError(err)
	}

	return rule, nil
}

func (repo *DeliveryRuleRepository) GetRule(ctx context.Context, id uint) (*model.DeliveryRule, error) {
	if id == 0 {
		return nil, &apperr.ValidationError{Field: "id", Rule: "required"}
	}

	rule := repo.query.DeliveryRule
	result, err := rule.WithContext(ctx).
		Where(rule.ID.Eq(id)).
		First()
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return result, nil
}

func (repo *DeliveryRuleRepository) ListRules(ctx context.Context, filter ListRulesFilter) ([]*model.DeliveryRule, error) {
	if filter.ShopID == 0 {
		return nil, &apperr.ValidationError{Field: "shop_id", Rule: "required"}
	}

	rule := repo.query.DeliveryRule
	stmt := rule.WithContext(ctx).
		Where(rule.ShopID.Eq(filter.ShopID)).
		Order(rule.Priority.Asc(), rule.ID.Asc())

	if filter.Status != nil {
		stmt = stmt.Where(rule.Status.Eq(string(*filter.Status)))
	}

	rules, err := stmt.Find()
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return rules, nil
}

func (repo *DeliveryRuleRepository) UpdateRuleStatus(ctx context.Context, shopID uint, id uint, status model.DeliveryRuleStatus) (*model.DeliveryRule, error) {
	if shopID == 0 {
		return nil, &apperr.ValidationError{Field: "shop_id", Rule: "required"}
	}
	if id == 0 {
		return nil, &apperr.ValidationError{Field: "id", Rule: "required"}
	}
	if status == "" {
		return nil, &apperr.ValidationError{Field: "status", Rule: "required"}
	}

	rule := repo.query.DeliveryRule
	info, err := rule.WithContext(ctx).
		Where(rule.ShopID.Eq(shopID), rule.ID.Eq(id)).
		Update(rule.Status, string(status))
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if info.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: delivery rule %d", apperr.ErrNotFound, id)
	}

	return repo.GetRule(ctx, id)
}

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: delivery rule", apperr.ErrNotFound)
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: %s", apperr.ErrConflict, pgErr.ConstraintName)
		case "23503", "23514":
			return fmt.Errorf("%w: %s", apperr.ErrValidation, pgErr.Message)
		}
	}

	return err
}
