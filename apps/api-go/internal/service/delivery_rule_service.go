package service

import (
	"context"
	"errors"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/repository"
)

const defaultRulePriority = 100

type DeliveryRuleRepository interface {
	CreateRule(ctx context.Context, rule *model.DeliveryRule) (*model.DeliveryRule, error)
	GetRule(ctx context.Context, id uint) (*model.DeliveryRule, error)
	ListRules(ctx context.Context, filter repository.ListRulesFilter) ([]*model.DeliveryRule, error)
	UpdateRuleStatus(ctx context.Context, shopID uint, id uint, status model.DeliveryRuleStatus) (*model.DeliveryRule, error)
}

type DeliveryRuleService struct {
	rules DeliveryRuleRepository
}

type CreateRuleCommand struct {
	ShopID         uint
	Name           string
	Priority       int
	Status         model.DeliveryRuleStatus
	ConditionType  model.RuleConditionType
	ConditionValue string
	ActionType     model.RuleActionType
	ActionValue    string
}

type GetRuleCommand struct {
	ID uint
}

type ListRulesCommand struct {
	ShopID uint
	Status *model.DeliveryRuleStatus
}

type UpdateRuleStatusCommand struct {
	ShopID uint
	ID     uint
	Status model.DeliveryRuleStatus
}

func NewDeliveryRuleService(rules DeliveryRuleRepository) *DeliveryRuleService {
	return &DeliveryRuleService{rules: rules}
}

func (svc *DeliveryRuleService) CreateRule(ctx context.Context, cmd CreateRuleCommand) (*model.DeliveryRule, error) {
	if err := validateCreateRuleCommand(cmd); err != nil {
		return nil, err
	}

	status := cmd.Status
	if status == "" {
		status = model.DeliveryRuleStatusDraft
	}

	priority := cmd.Priority
	if priority == 0 {
		priority = defaultRulePriority
	}

	return svc.rules.CreateRule(ctx, &model.DeliveryRule{
		ShopID:         cmd.ShopID,
		Name:           cmd.Name,
		Priority:       priority,
		Status:         status,
		ConditionType:  cmd.ConditionType,
		ConditionValue: cmd.ConditionValue,
		ActionType:     cmd.ActionType,
		ActionValue:    cmd.ActionValue,
	})
}

func (svc *DeliveryRuleService) GetRule(ctx context.Context, cmd GetRuleCommand) (*model.DeliveryRule, error) {
	if cmd.ID == 0 {
		return nil, &apperr.ValidationError{Field: "id", Rule: "required"}
	}
	return svc.rules.GetRule(ctx, cmd.ID)
}

func (svc *DeliveryRuleService) ListRules(ctx context.Context, cmd ListRulesCommand) ([]*model.DeliveryRule, error) {
	if err := validateListRulesCommand(cmd); err != nil {
		return nil, err
	}

	return svc.rules.ListRules(ctx, repository.ListRulesFilter{
		ShopID: cmd.ShopID,
		Status: cmd.Status,
	})
}

func (svc *DeliveryRuleService) UpdateRuleStatus(ctx context.Context, cmd UpdateRuleStatusCommand) (*model.DeliveryRule, error) {
	if err := validateUpdateRuleStatusCommand(cmd); err != nil {
		return nil, err
	}

	return svc.rules.UpdateRuleStatus(ctx, cmd.ShopID, cmd.ID, cmd.Status)
}

func validateCreateRuleCommand(cmd CreateRuleCommand) error {
	var errs []error

	if cmd.ShopID == 0 {
		errs = append(errs, &apperr.ValidationError{Field: "shop_id", Rule: "required"})
	}
	if cmd.Name == "" {
		errs = append(errs, &apperr.ValidationError{Field: "name", Rule: "required"})
	}
	if cmd.Priority < 0 {
		errs = append(errs, &apperr.ValidationError{Field: "priority", Rule: "non_negative"})
	}
	if cmd.Status != "" && !isValidRuleStatus(cmd.Status) {
		errs = append(errs, &apperr.ValidationError{Field: "status", Rule: "supported_value"})
	}

	validCondition := isValidConditionType(cmd.ConditionType)
	if !validCondition {
		errs = append(errs, &apperr.ValidationError{Field: "condition_type", Rule: "supported_value"})
	}
	if cmd.ConditionValue == "" {
		errs = append(errs, &apperr.ValidationError{Field: "condition_value", Rule: "required"})
	}

	validAction := isValidActionType(cmd.ActionType)
	if !validAction {
		errs = append(errs, &apperr.ValidationError{Field: "action_type", Rule: "supported_value"})
	}
	if cmd.ActionValue == "" {
		errs = append(errs, &apperr.ValidationError{Field: "action_value", Rule: "required"})
	}
	if validCondition && validAction && !isValidRuleCombination(cmd.ConditionType, cmd.ActionType) {
		errs = append(errs, &apperr.ValidationError{Field: "action_type", Rule: "compatible_condition"})
	}

	return errors.Join(errs...)
}

func validateListRulesCommand(cmd ListRulesCommand) error {
	var errs []error

	if cmd.ShopID == 0 {
		errs = append(errs, &apperr.ValidationError{Field: "shop_id", Rule: "required"})
	}
	if cmd.Status != nil && !isValidRuleStatus(*cmd.Status) {
		errs = append(errs, &apperr.ValidationError{Field: "status", Rule: "supported_value"})
	}

	return errors.Join(errs...)
}

func validateUpdateRuleStatusCommand(cmd UpdateRuleStatusCommand) error {
	var errs []error

	if cmd.ShopID == 0 {
		errs = append(errs, &apperr.ValidationError{Field: "shop_id", Rule: "required"})
	}
	if cmd.ID == 0 {
		errs = append(errs, &apperr.ValidationError{Field: "id", Rule: "required"})
	}
	if !isValidRuleStatus(cmd.Status) {
		errs = append(errs, &apperr.ValidationError{Field: "status", Rule: "supported_value"})
	}

	return errors.Join(errs...)
}

func isValidRuleStatus(status model.DeliveryRuleStatus) bool {
	switch status {
	case model.DeliveryRuleStatusDraft, model.DeliveryRuleStatusActive, model.DeliveryRuleStatusDisabled:
		return true
	default:
		return false
	}
}

func isValidConditionType(conditionType model.RuleConditionType) bool {
	switch conditionType {
	case model.RuleConditionCountry,
		model.RuleConditionPostalCode,
		model.RuleConditionCartTotal,
		model.RuleConditionProductTag,
		model.RuleConditionCustomerTag:
		return true
	default:
		return false
	}
}

func isValidActionType(actionType model.RuleActionType) bool {
	switch actionType {
	case model.RuleActionHide, model.RuleActionRename, model.RuleActionSort:
		return true
	default:
		return false
	}
}

func isValidRuleCombination(conditionType model.RuleConditionType, actionType model.RuleActionType) bool {
	switch actionType {
	case model.RuleActionHide, model.RuleActionRename:
		return true
	case model.RuleActionSort:
		return conditionType == model.RuleConditionCustomerTag
	default:
		return false
	}
}
