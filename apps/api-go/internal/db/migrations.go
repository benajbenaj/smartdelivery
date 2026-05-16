package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type MigrationStatus struct {
	Tables []TableStatus
}

type TableStatus struct {
	Name   string
	Exists bool
}

func MigrateUp(ctx context.Context, gormDB *gorm.DB) error {
	if err := gormDB.WithContext(ctx).AutoMigrate(&Shop{}, &DeliveryRule{}, &AuditLog{}); err != nil {
		return fmt.Errorf("gorm auto migrate: %w", err)
	}
	return nil
}

func MigrateDown(ctx context.Context, gormDB *gorm.DB) error {
	if err := gormDB.WithContext(ctx).Migrator().DropTable(&AuditLog{}, &DeliveryRule{}, &Shop{}); err != nil {
		return fmt.Errorf("gorm drop tables: %w", err)
	}
	return nil
}

func CheckMigrationStatus(ctx context.Context, gormDB *gorm.DB) (MigrationStatus, error) {
	conn := gormDB.WithContext(ctx)
	tables := []TableStatus{
		{Name: "shops", Exists: conn.Migrator().HasTable(&Shop{})},
		{Name: "delivery_rules", Exists: conn.Migrator().HasTable(&DeliveryRule{})},
		{Name: "audit_logs", Exists: conn.Migrator().HasTable(&AuditLog{})},
	}

	return MigrationStatus{Tables: tables}, nil
}
