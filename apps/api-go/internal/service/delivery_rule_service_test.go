package service

import (
	"context"
	"errors"
	"testing"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/audit"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/repository"
)

func TestCreateRuleValidatesRequiredFields(t *testing.T) {
	svc := NewDeliveryRuleService(&fakeRuleRepository{})

	_, err := svc.CreateRule(context.Background(), CreateRuleCommand{})
	assertValidationError(t, err, "shop_id", "required")
}

func TestCreateRuleReturnsJoinedValidationErrors(t *testing.T) {
	svc := NewDeliveryRuleService(&fakeRuleRepository{})

	_, err := svc.CreateRule(context.Background(), CreateRuleCommand{
		Priority: -1,
		Status:   model.DeliveryRuleStatus("archived"),
	})

	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("error = %v, want ErrValidation", err)
	}

	assertJoinedValidationError(t, err, "shop_id", "required")
	assertJoinedValidationError(t, err, "name", "required")
	assertJoinedValidationError(t, err, "priority", "non_negative")
	assertJoinedValidationError(t, err, "status", "supported_value")
	assertJoinedValidationError(t, err, "condition_type", "supported_value")
	assertJoinedValidationError(t, err, "condition_value", "required")
	assertJoinedValidationError(t, err, "action_type", "supported_value")
	assertJoinedValidationError(t, err, "action_value", "required")
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

func TestCreateRulePublishesAuditEvent(t *testing.T) {
	repo := &fakeRuleRepository{}
	publisher := &fakeAuditPublisher{}
	svc := NewDeliveryRuleServiceWithAudit(repo, publisher)

	rule, err := svc.CreateRule(context.Background(), validCreateRuleCommand())
	if err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}

	if len(publisher.events) != 1 {
		t.Fatalf("published events = %d, want 1", len(publisher.events))
	}
	event := publisher.events[0]
	if event.ShopID != rule.ShopID {
		t.Fatalf("ShopID = %d, want %d", event.ShopID, rule.ShopID)
	}
	if event.DeliveryRuleID == nil || *event.DeliveryRuleID != rule.ID {
		t.Fatalf("DeliveryRuleID = %v, want %d", event.DeliveryRuleID, rule.ID)
	}
	if event.Type != audit.EventDeliveryRuleCreated {
		t.Fatalf("Type = %q, want %q", event.Type, audit.EventDeliveryRuleCreated)
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

func TestUpdateRuleStatusPublishesAuditEvent(t *testing.T) {
	publisher := &fakeAuditPublisher{}
	svc := NewDeliveryRuleServiceWithAudit(&fakeRuleRepository{}, publisher)

	rule, err := svc.UpdateRuleStatus(context.Background(), UpdateRuleStatusCommand{
		ShopID: 1,
		ID:     2,
		Status: model.DeliveryRuleStatusActive,
	})
	if err != nil {
		t.Fatalf("UpdateRuleStatus() error = %v", err)
	}

	if len(publisher.events) != 1 {
		t.Fatalf("published events = %d, want 1", len(publisher.events))
	}
	event := publisher.events[0]
	if event.ShopID != rule.ShopID {
		t.Fatalf("ShopID = %d, want %d", event.ShopID, rule.ShopID)
	}
	if event.DeliveryRuleID == nil || *event.DeliveryRuleID != rule.ID {
		t.Fatalf("DeliveryRuleID = %v, want %d", event.DeliveryRuleID, rule.ID)
	}
	if event.Type != audit.EventDeliveryRuleStatusUpdated {
		t.Fatalf("Type = %q, want %q", event.Type, audit.EventDeliveryRuleStatusUpdated)
	}
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

func assertJoinedValidationError(t *testing.T, err error, field string, rule string) {
	t.Helper()

	for _, candidate := range flattenErrors(err) {
		var validationErr *apperr.ValidationError
		if errors.As(candidate, &validationErr) && validationErr.Field == field && validationErr.Rule == rule {
			return
		}
	}

	t.Fatalf("joined error %v does not contain ValidationError{%q, %q}", err, field, rule)
}

func flattenErrors(err error) []error {
	if err == nil {
		return nil
	}

	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var result []error
		for _, child := range joined.Unwrap() {
			result = append(result, flattenErrors(child)...)
		}
		return result
	}

	return []error{err}
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

type fakeAuditPublisher struct {
	events []audit.Event
}

func (publisher *fakeAuditPublisher) Publish(_ context.Context, event audit.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}
