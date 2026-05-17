package functionconfig

import (
	"context"
	"errors"
	"strings"
	"testing"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/repository"
)

func TestBuildJSONReturnsDeterministicCompactConfig(t *testing.T) {
	configJSON, err := BuildJSON([]*model.DeliveryRule{
		{
			ID:             3,
			Priority:       30,
			Status:         model.DeliveryRuleStatusDisabled,
			ConditionType:  model.RuleConditionCustomerTag,
			ConditionValue: "eco",
			ActionType:     model.RuleActionSort,
			ActionValue:    "low-carbon",
		},
		{
			ID:             2,
			Priority:       10,
			Status:         model.DeliveryRuleStatusActive,
			ConditionType:  model.RuleConditionCustomerTag,
			ConditionValue: "eco",
			ActionType:     model.RuleActionSort,
			ActionValue:    "low-carbon",
		},
		{
			ID:             1,
			Priority:       10,
			Status:         model.DeliveryRuleStatusActive,
			ConditionType:  model.RuleConditionPostalCode,
			ConditionValue: "10000-19999",
			ActionType:     model.RuleActionHide,
			ActionValue:    "pickup",
		},
	})
	if err != nil {
		t.Fatalf("BuildJSON() error = %v", err)
	}

	want := `{"version":1,"rules":[{"id":1,"priority":10,"condition":{"type":"postal_code","value":"10000-19999"},"action":{"type":"hide","value":"pickup"}},{"id":2,"priority":10,"condition":{"type":"customer_tag","value":"eco"},"action":{"type":"sort","value":"low-carbon"}}]}`
	if configJSON != want {
		t.Fatalf("BuildJSON() = %s, want %s", configJSON, want)
	}
	if strings.Contains(configJSON, "\n") || strings.Contains(configJSON, "  ") {
		t.Fatalf("BuildJSON() = %q, want compact JSON", configJSON)
	}
}

func TestBuildJSONEmptyRulesIsDeterministic(t *testing.T) {
	configJSON, err := BuildJSON(nil)
	if err != nil {
		t.Fatalf("BuildJSON() error = %v", err)
	}

	want := `{"version":1,"rules":[]}`
	if configJSON != want {
		t.Fatalf("BuildJSON() = %s, want %s", configJSON, want)
	}
}

func TestGenerateSnapshotStoresConfigJSON(t *testing.T) {
	rules := &fakeRuleLister{
		rules: []*model.DeliveryRule{
			{
				ID:             5,
				ShopID:         9,
				Priority:       20,
				Status:         model.DeliveryRuleStatusActive,
				ConditionType:  model.RuleConditionProductTag,
				ConditionValue: "fragile",
				ActionType:     model.RuleActionRename,
				ActionValue:    "Fragile delivery",
			},
		},
	}
	snapshots := &fakeSnapshotStore{}
	generator := NewGenerator(rules, snapshots)

	snapshot, err := generator.GenerateSnapshot(context.Background(), 9)
	if err != nil {
		t.Fatalf("GenerateSnapshot() error = %v", err)
	}

	want := `{"version":1,"rules":[{"id":5,"priority":20,"condition":{"type":"product_tag","value":"fragile"},"action":{"type":"rename","value":"Fragile delivery"}}]}`
	if snapshot.ConfigJSON != want {
		t.Fatalf("ConfigJSON = %s, want %s", snapshot.ConfigJSON, want)
	}
	if snapshot.ShopID != 9 {
		t.Fatalf("ShopID = %d, want 9", snapshot.ShopID)
	}
	if rules.filter.Status == nil || *rules.filter.Status != model.DeliveryRuleStatusActive {
		t.Fatalf("Status filter = %v, want active", rules.filter.Status)
	}
}

func TestGenerateSnapshotRequiresShopID(t *testing.T) {
	generator := NewGenerator(&fakeRuleLister{}, &fakeSnapshotStore{})

	_, err := generator.GenerateSnapshot(context.Background(), 0)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("GenerateSnapshot() error = %v, want ErrValidation", err)
	}
}

type fakeRuleLister struct {
	filter repository.ListRulesFilter
	rules  []*model.DeliveryRule
	err    error
}

func (lister *fakeRuleLister) ListRules(_ context.Context, filter repository.ListRulesFilter) ([]*model.DeliveryRule, error) {
	lister.filter = filter
	return lister.rules, lister.err
}

type fakeSnapshotStore struct {
	snapshot *model.FunctionConfigSnapshot
	err      error
}

func (store *fakeSnapshotStore) CreateSnapshot(_ context.Context, snapshot *model.FunctionConfigSnapshot) (*model.FunctionConfigSnapshot, error) {
	store.snapshot = snapshot
	return snapshot, store.err
}
