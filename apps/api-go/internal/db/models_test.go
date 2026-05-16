package db

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestShopModelSchema(t *testing.T) {
	model := parseModel(t, &Shop{})

	if model.Table != "shops" {
		t.Fatalf("table = %q, want shops", model.Table)
	}

	assertPrimaryKey(t, model, "ID")
	assertTagSetting(t, model, "Domain", "UNIQUEINDEX")
	assertTagSetting(t, model, "Domain", "NOT NULL")
}

func TestDeliveryRuleModelSchema(t *testing.T) {
	model := parseModel(t, &DeliveryRule{})

	if model.Table != "delivery_rules" {
		t.Fatalf("table = %q, want delivery_rules", model.Table)
	}

	assertPrimaryKey(t, model, "ID")
	assertTagSetting(t, model, "ShopID", "INDEX")
	assertTagSetting(t, model, "Status", "INDEX")
	assertTagSetting(t, model, "Priority", "DEFAULT")
}

func TestAuditLogModelSchema(t *testing.T) {
	model := parseModel(t, &AuditLog{})

	if model.Table != "audit_logs" {
		t.Fatalf("table = %q, want audit_logs", model.Table)
	}

	assertPrimaryKey(t, model, "ID")
	assertTagSetting(t, model, "ShopID", "INDEX")
	assertTagSetting(t, model, "DeliveryRuleID", "INDEX")
	assertTagSetting(t, model, "CreatedAt", "INDEX")
}

func parseModel(t *testing.T, value any) *schema.Schema {
	t.Helper()

	model, err := schema.Parse(value, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse model: %v", err)
	}
	return model
}

func assertPrimaryKey(t *testing.T, model *schema.Schema, fieldName string) {
	t.Helper()

	field := model.LookUpField(fieldName)
	if field == nil {
		t.Fatalf("field %q not found", fieldName)
	}
	if !field.PrimaryKey {
		t.Fatalf("field %q PrimaryKey = false, want true", fieldName)
	}
}

func assertTagSetting(t *testing.T, model *schema.Schema, fieldName string, setting string) {
	t.Helper()

	field := model.LookUpField(fieldName)
	if field == nil {
		t.Fatalf("field %q not found", fieldName)
	}
	if _, ok := field.TagSettings[setting]; !ok {
		t.Fatalf("field %q missing GORM tag setting %q", fieldName, setting)
	}
}
