package db

import "testing"

func TestMigrationStatusModelNames(t *testing.T) {
	status := MigrationStatus{
		Tables: []TableStatus{
			{Name: "shops", Exists: true},
			{Name: "delivery_rules", Exists: true},
			{Name: "audit_logs", Exists: true},
		},
	}

	if len(status.Tables) != 3 {
		t.Fatalf("len(status.Tables) = %d, want 3", len(status.Tables))
	}
}
