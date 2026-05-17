package db

import "testing"

func TestMigrationStatusModelNames(t *testing.T) {
	status := MigrationStatus{
		Tables: []TableStatus{
			{Name: "shops", Exists: true},
			{Name: "delivery_rules", Exists: true},
			{Name: "audit_logs", Exists: true},
			{Name: "function_config_snapshots", Exists: true},
		},
	}

	if len(status.Tables) != 4 {
		t.Fatalf("len(status.Tables) = %d, want 4", len(status.Tables))
	}
}
