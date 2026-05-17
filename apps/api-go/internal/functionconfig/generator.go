package functionconfig

import (
	"context"
	"encoding/json"
	"sort"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/repository"
)

const version = 1

type RuleLister interface {
	ListRules(ctx context.Context, filter repository.ListRulesFilter) ([]*model.DeliveryRule, error)
}

type SnapshotStore interface {
	CreateSnapshot(ctx context.Context, snapshot *model.FunctionConfigSnapshot) (*model.FunctionConfigSnapshot, error)
}

type Generator struct {
	rules     RuleLister
	snapshots SnapshotStore
}

type configDocument struct {
	Version int          `json:"version"`
	Rules   []configRule `json:"rules"`
}

type configRule struct {
	ID        uint            `json:"id"`
	Priority  int             `json:"priority"`
	Condition configPredicate `json:"condition"`
	Action    configPredicate `json:"action"`
}

type configPredicate struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func NewGenerator(rules RuleLister, snapshots SnapshotStore) *Generator {
	return &Generator{rules: rules, snapshots: snapshots}
}

func (generator *Generator) GenerateSnapshot(ctx context.Context, shopID uint) (*model.FunctionConfigSnapshot, error) {
	if shopID == 0 {
		return nil, &apperr.ValidationError{Field: "shop_id", Rule: "required"}
	}

	status := model.DeliveryRuleStatusActive
	rules, err := generator.rules.ListRules(ctx, repository.ListRulesFilter{
		ShopID: shopID,
		Status: &status,
	})
	if err != nil {
		return nil, err
	}

	configJSON, err := BuildJSON(rules)
	if err != nil {
		return nil, err
	}

	return generator.snapshots.CreateSnapshot(ctx, &model.FunctionConfigSnapshot{
		ShopID:     shopID,
		ConfigJSON: configJSON,
	})
}

func BuildJSON(rules []*model.DeliveryRule) (string, error) {
	activeRules := make([]*model.DeliveryRule, 0, len(rules))
	for _, rule := range rules {
		if rule == nil || rule.Status != model.DeliveryRuleStatusActive {
			continue
		}
		activeRules = append(activeRules, rule)
	}

	sort.Slice(activeRules, func(i, j int) bool {
		if activeRules[i].Priority == activeRules[j].Priority {
			return activeRules[i].ID < activeRules[j].ID
		}
		return activeRules[i].Priority < activeRules[j].Priority
	})

	document := configDocument{
		Version: version,
		Rules:   make([]configRule, 0, len(activeRules)),
	}
	for _, rule := range activeRules {
		document.Rules = append(document.Rules, configRule{
			ID:       rule.ID,
			Priority: rule.Priority,
			Condition: configPredicate{
				Type:  string(rule.ConditionType),
				Value: rule.ConditionValue,
			},
			Action: configPredicate{
				Type:  string(rule.ActionType),
				Value: rule.ActionValue,
			},
		})
	}

	bytes, err := json.Marshal(document)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
