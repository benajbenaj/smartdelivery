package service

import (
	"context"
	"errors"
	"testing"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/repository"
)

func TestCreateRuleValidatesRequiredFields(t *testing.T) {
	svc := NewDeliveryRuleService(&fakeRuleRepository{})

	_, err := svc.CreateRule(context.Background(), CreateRuleCommand{})
	assertValidationError(t, err, "shop_id", "required")
}

func TestCreateRuleValidatesActionConditionCombination(t *testing.T) {
	svc := NewDeliveryRuleService(&fakeRuleRepository{})

	_, err := svc.CreateRule(context.Background(), validCreateRuleCommand(func(cmd *CreateRuleCommand) {
		cmd.ConditionType = model.RuleConditionProductTag
		cmd.ActionType = model.RuleActionSort
	}))

	assertValidationError(t, err, "action_type", "compatible_condition")
}

func TestCreateRuleBuildsModelAndAppliesDefaults(t *testing.T) {
	repo := &fakeRuleRepository{}
	svc := NewDeliveryRuleService(repo)

	rule, err := svc.CreateRule(context.Background(), validCreateRuleCommand())
	if err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}

	if rule.Status != model.DeliveryRuleStatusDraft {
		t.Fatalf("Status = %q, want %q", rule.Status, model.DeliveryRuleStatusDraft)
	}
	if rule.Priority != defaultRulePriority {
		t.Fatalf("Priority = %d, want %d", rule.Priority, defaultRulePriority)
	}
	if repo.created == nil {
		t.Fatalf("repo.created = nil, want created rule")
	}
}

func TestUpdateRuleStatusValidatesStatus(t *testing.T) {
	svc := NewDeliveryRuleService(&fakeRuleRepository{})

	_, err := svc.UpdateRuleStatus(context.Background(), UpdateRuleStatusCommand{
		ShopID: 1,
		ID:     1,
		Status: model.DeliveryRuleStatus("archived"),
	})

	assertValidationError(t, err, "status", "supported_value")
}

func TestListRulesValidatesStatusFilter(t *testing.T) {
	svc := NewDeliveryRuleService(&fakeRuleRepository{})
	status := model.DeliveryRuleStatus("archived")

	_, err := svc.ListRules(context.Background(), ListRulesCommand{
		ShopID: 1,
		Status: &status,
	})

	assertValidationError(t, err, "status", "supported_value")
}

func validCreateRuleCommand(mutators ...func(*CreateRuleCommand)) CreateRuleCommand {
	cmd := CreateRuleCommand{
		ShopID:         1,
		Name:           "Hide express for hazardous products",
		ConditionType:  model.RuleConditionProductTag,
		ConditionValue: "hazardous",
		ActionType:     model.RuleActionHide,
		ActionValue:    "express",
	}

	for _, mutate := range mutators {
		mutate(&cmd)
	}

	return cmd
}

func assertValidationError(t *testing.T, err error, field string, rule string) {
	t.Helper()

	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}

	var validationErr *apperr.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("errors.As(err, &validationErr) = false, want true")
	}
	if validationErr.Field != field {
		t.Fatalf("Field = %q, want %q", validationErr.Field, field)
	}
	if validationErr.Rule != rule {
		t.Fatalf("Rule = %q, want %q", validationErr.Rule, rule)
	}
}

type fakeRuleRepository struct {
	created *model.DeliveryRule
}

func (repo *fakeRuleRepository) CreateRule(_ context.Context, rule *model.DeliveryRule) (*model.DeliveryRule, error) {
	repo.created = rule
	rule.ID = 1
	return rule, nil
}

func (repo *fakeRuleRepository) GetRule(_ context.Context, id uint) (*model.DeliveryRule, error) {
	return &model.DeliveryRule{ID: id}, nil
}

func (repo *fakeRuleRepository) ListRules(_ context.Context, _ repository.ListRulesFilter) ([]*model.DeliveryRule, error) {
	return []*model.DeliveryRule{}, nil
}

func (repo *fakeRuleRepository) UpdateRuleStatus(_ context.Context, shopID uint, id uint, status model.DeliveryRuleStatus) (*model.DeliveryRule, error) {
	return &model.DeliveryRule{ID: id, ShopID: shopID, Status: status}, nil
}
