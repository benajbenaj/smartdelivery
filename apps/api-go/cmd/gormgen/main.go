package main

import (
	"smartdelivery/apps/api-go/internal/model"

	"gorm.io/gen"
)

func main() {
	generator := gen.NewGenerator(gen.Config{
		OutPath: "internal/query",
		Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	generator.ApplyBasic(
		model.Shop{},
		model.DeliveryRule{},
		model.AuditLog{},
	)

	generator.Execute()
}
