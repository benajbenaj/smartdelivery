package repository

import (
	"context"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/query"

	"gorm.io/gorm"
)

type FunctionConfigSnapshotRepository struct {
	query *query.Query
}

func NewFunctionConfigSnapshotRepository(gormDB *gorm.DB) *FunctionConfigSnapshotRepository {
	return &FunctionConfigSnapshotRepository{query: query.Use(gormDB)}
}

func (repo *FunctionConfigSnapshotRepository) CreateSnapshot(ctx context.Context, snapshot *model.FunctionConfigSnapshot) (*model.FunctionConfigSnapshot, error) {
	if snapshot == nil {
		return nil, &apperr.ValidationError{Field: "snapshot", Rule: "required"}
	}
	if snapshot.ShopID == 0 {
		return nil, &apperr.ValidationError{Field: "shop_id", Rule: "required"}
	}
	if snapshot.ConfigJSON == "" {
		return nil, &apperr.ValidationError{Field: "config_json", Rule: "required"}
	}

	if err := repo.query.FunctionConfigSnapshot.WithContext(ctx).Create(snapshot); err != nil {
		return nil, mapFunctionConfigSnapshotDatabaseError(err)
	}

	return snapshot, nil
}

func mapFunctionConfigSnapshotDatabaseError(err error) error {
	return mapDatabaseError(err)
}
