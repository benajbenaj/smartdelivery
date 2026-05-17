package service

import (
	"context"
	"errors"
	"sort"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/audit"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/repository"
	workerpkg "smartdelivery/apps/api-go/internal/worker"
)

const defaultRulePriority = 100
const defaultImportValidationWorkers = 4

type DeliveryRuleRepository interface {
	CreateRule(ctx context.Context, rule *model.DeliveryRule) (*model.DeliveryRule, error)
	GetRule(ctx context.Context, id uint) (*model.DeliveryRule, error)
	ListRules(ctx context.Context, filter repository.ListRulesFilter) ([]*model.DeliveryRule, error)
	UpdateRuleStatus(ctx context.Context, shopID uint, id uint, status model.DeliveryRuleStatus) (*model.DeliveryRule, error)
}

type DeliveryRuleService struct {
	rules DeliveryRuleRepository
	audit audit.Publisher
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

type ImportRuleCommand struct {
	Name           string
	Priority       int
	Status         model.DeliveryRuleStatus
	ConditionType  model.RuleConditionType
	ConditionValue string
	ActionType     model.RuleActionType
	ActionValue    string
}

type ValidateRuleImportCommand struct {
	ShopID      uint
	Rules       []ImportRuleCommand
	WorkerCount int
}

type RuleImportValidationSummary struct {
	Total   int
	Valid   int
	Invalid int
	Results []RuleImportValidationResult
}

type RuleImportValidationResult struct {
	Index     int
	RowNumber int
	Valid     bool
	Errors    []apperr.ValidationError
}

func NewDeliveryRuleService(rules DeliveryRuleRepository) *DeliveryRuleService {
	return &DeliveryRuleService{rules: rules}
}

func NewDeliveryRuleServiceWithAudit(rules DeliveryRuleRepository, auditPublisher audit.Publisher) *DeliveryRuleService {
	return &DeliveryRuleService{rules: rules, audit: auditPublisher}
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

	rule, err := svc.rules.CreateRule(ctx, &model.DeliveryRule{
		ShopID:         cmd.ShopID,
		Name:           cmd.Name,
		Priority:       priority,
		Status:         status,
		ConditionType:  cmd.ConditionType,
		ConditionValue: cmd.ConditionValue,
		ActionType:     cmd.ActionType,
		ActionValue:    cmd.ActionValue,
	})
	if err != nil {
		return nil, err
	}

	if err := svc.publishAudit(ctx, audit.Event{
		ShopID:         rule.ShopID,
		DeliveryRuleID: uintPtr(rule.ID),
		Type:           audit.EventDeliveryRuleCreated,
		Message:        "delivery rule created",
	}); err != nil {
		return nil, err
	}

	return rule, nil
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

	rule, err := svc.rules.UpdateRuleStatus(ctx, cmd.ShopID, cmd.ID, cmd.Status)
	if err != nil {
		return nil, err
	}

	if err := svc.publishAudit(ctx, audit.Event{
		ShopID:         rule.ShopID,
		DeliveryRuleID: uintPtr(rule.ID),
		Type:           audit.EventDeliveryRuleStatusUpdated,
		Message:        "delivery rule status updated",
	}); err != nil {
		return nil, err
	}

	return rule, nil
}

func (svc *DeliveryRuleService) ValidateRuleImport(ctx context.Context, cmd ValidateRuleImportCommand) (RuleImportValidationSummary, error) {
	if cmd.ShopID == 0 {
		return RuleImportValidationSummary{}, &apperr.ValidationError{Field: "shop_id", Rule: "required"}
	}
	if len(cmd.Rules) == 0 {
		return RuleImportValidationSummary{}, &apperr.ValidationError{Field: "rules", Rule: "min_items"}
	}

	jobs := make(chan ruleImportValidationJob)
	workerCount := importValidationWorkerCount(cmd.WorkerCount, len(cmd.Rules))

	workerResults := workerpkg.FanOut(ctx, workerCount, jobs, func(_ context.Context, job ruleImportValidationJob) (RuleImportValidationResult, bool) {
		return validateRuleImportJob(cmd.ShopID, job), true
	})
	results := workerpkg.FanIn(ctx, len(cmd.Rules), workerResults...)

	for index, rule := range cmd.Rules {
		select {
		case <-ctx.Done():
			close(jobs)
			for range results {
			}
			return RuleImportValidationSummary{}, ctx.Err()
		case jobs <- ruleImportValidationJob{index: index, rule: rule}:
		}
	}

	close(jobs)

	summary := RuleImportValidationSummary{
		Total:   len(cmd.Rules),
		Results: make([]RuleImportValidationResult, 0, len(cmd.Rules)),
	}
	for result := range results {
		if result.Valid {
			summary.Valid++
		} else {
			summary.Invalid++
		}
		summary.Results = append(summary.Results, result)
	}

	if err := ctx.Err(); err != nil {
		return RuleImportValidationSummary{}, err
	}

	sort.Slice(summary.Results, func(i, j int) bool {
		return summary.Results[i].Index < summary.Results[j].Index
	})

	return summary, nil
}

func (svc *DeliveryRuleService) publishAudit(ctx context.Context, event audit.Event) error {
	if svc.audit == nil {
		return nil
	}
	return svc.audit.Publish(ctx, event)
}

func uintPtr(value uint) *uint {
	return &value
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

type ruleImportValidationJob struct {
	index int
	rule  ImportRuleCommand
}

func validateRuleImportJob(shopID uint, job ruleImportValidationJob) RuleImportValidationResult {
	err := validateCreateRuleCommand(CreateRuleCommand{
		ShopID:         shopID,
		Name:           job.rule.Name,
		Priority:       job.rule.Priority,
		Status:         job.rule.Status,
		ConditionType:  job.rule.ConditionType,
		ConditionValue: job.rule.ConditionValue,
		ActionType:     job.rule.ActionType,
		ActionValue:    job.rule.ActionValue,
	})

	validationErrors := collectValidationErrors(err)
	return RuleImportValidationResult{
		Index:     job.index,
		RowNumber: job.index + 1,
		Valid:     len(validationErrors) == 0,
		Errors:    validationErrors,
	}
}

func collectValidationErrors(err error) []apperr.ValidationError {
	if err == nil {
		return nil
	}

	var result []apperr.ValidationError
	for _, candidate := range flattenValidationErrors(err) {
		var validationErr *apperr.ValidationError
		if errors.As(candidate, &validationErr) {
			result = append(result, *validationErr)
		}
	}
	return result
}

func flattenValidationErrors(err error) []error {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var result []error
		for _, child := range joined.Unwrap() {
			result = append(result, flattenValidationErrors(child)...)
		}
		return result
	}
	return []error{err}
}

func importValidationWorkerCount(requested int, ruleCount int) int {
	if ruleCount <= 1 {
		return 1
	}
	if requested <= 0 {
		requested = defaultImportValidationWorkers
	}
	if requested > ruleCount {
		return ruleCount
	}
	return requested
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
