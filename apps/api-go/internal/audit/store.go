package audit

import (
	"context"

	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/query"

	"gorm.io/gorm"
)

type GormStore struct {
	query *query.Query
}

func NewStore(gormDB *gorm.DB) *GormStore {
	return &GormStore{query: query.Use(gormDB)}
}

func (store *GormStore) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	return store.query.AuditLog.WithContext(ctx).Create(log)
}
