package repository

import (
	"context"
	"errors"
	"testing"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
)

func TestCreateSnapshotRequiresSnapshot(t *testing.T) {
	repo := &FunctionConfigSnapshotRepository{}

	_, err := repo.CreateSnapshot(context.Background(), nil)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("CreateSnapshot() error = %v, want ErrValidation", err)
	}
}

func TestCreateSnapshotRequiresShopID(t *testing.T) {
	repo := &FunctionConfigSnapshotRepository{}

	_, err := repo.CreateSnapshot(context.Background(), &model.FunctionConfigSnapshot{ConfigJSON: "{}"})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("CreateSnapshot() error = %v, want ErrValidation", err)
	}
}

func TestCreateSnapshotRequiresConfigJSON(t *testing.T) {
	repo := &FunctionConfigSnapshotRepository{}

	_, err := repo.CreateSnapshot(context.Background(), &model.FunctionConfigSnapshot{ShopID: 1})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("CreateSnapshot() error = %v, want ErrValidation", err)
	}
}
