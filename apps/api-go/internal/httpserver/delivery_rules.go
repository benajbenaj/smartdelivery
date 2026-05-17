package httpserver

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/service"

	"github.com/gin-gonic/gin"
)

type DeliveryRuleService interface {
	CreateRule(ctx context.Context, cmd service.CreateRuleCommand) (*model.DeliveryRule, error)
	ListRules(ctx context.Context, cmd service.ListRulesCommand) ([]*model.DeliveryRule, error)
	UpdateRuleStatus(ctx context.Context, cmd service.UpdateRuleStatusCommand) (*model.DeliveryRule, error)
	ValidateRuleImport(ctx context.Context, cmd service.ValidateRuleImportCommand) (service.RuleImportValidationSummary, error)
}

type deliveryRuleHandler struct {
	rules DeliveryRuleService
}

type createDeliveryRuleRequest struct {
	Name           string                   `json:"name"`
	Priority       int                      `json:"priority"`
	Status         model.DeliveryRuleStatus `json:"status"`
	ConditionType  model.RuleConditionType  `json:"condition_type"`
	ConditionValue string                   `json:"condition_value"`
	ActionType     model.RuleActionType     `json:"action_type"`
	ActionValue    string                   `json:"action_value"`
}

type updateDeliveryRuleStatusRequest struct {
	Status model.DeliveryRuleStatus `json:"status"`
}

type importDeliveryRulesRequest struct {
	WorkerCount int                         `json:"worker_count"`
	Rules       []createDeliveryRuleRequest `json:"rules"`
}

type deliveryRuleResponse struct {
	ID             uint                     `json:"id"`
	ShopID         uint                     `json:"shop_id"`
	Name           string                   `json:"name"`
	Priority       int                      `json:"priority"`
	Status         model.DeliveryRuleStatus `json:"status"`
	ConditionType  model.RuleConditionType  `json:"condition_type"`
	ConditionValue string                   `json:"condition_value"`
	ActionType     model.RuleActionType     `json:"action_type"`
	ActionValue    string                   `json:"action_value"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

type importDeliveryRulesResponse struct {
	Total   int                            `json:"total"`
	Valid   int                            `json:"valid"`
	Invalid int                            `json:"invalid"`
	Results []importDeliveryRuleValidation `json:"results"`
}

type importDeliveryRuleValidation struct {
	Index     int                       `json:"index"`
	RowNumber int                       `json:"row_number"`
	Valid     bool                      `json:"valid"`
	Errors    []validationErrorResponse `json:"errors,omitempty"`
}

func registerDeliveryRuleRoutes(router *gin.Engine, rules DeliveryRuleService) {
	handler := deliveryRuleHandler{rules: rules}

	router.POST("/shops/:shop/delivery-rules", handler.createRule)
	router.GET("/shops/:shop/delivery-rules", handler.listRules)
	router.PATCH("/shops/:shop/delivery-rules/:id", handler.updateRuleStatus)
	router.POST("/shops/:shop/delivery-rules/imports", handler.importRules)
}

func (handler deliveryRuleHandler) createRule(ctx *gin.Context) {
	shopID, ok := parseUintPathParam(ctx, "shop", "shop")
	if !ok {
		return
	}

	var request createDeliveryRuleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeBadRequest(ctx, "invalid request body")
		return
	}

	rule, err := handler.rules.CreateRule(ctx.Request.Context(), service.CreateRuleCommand{
		ShopID:         shopID,
		Name:           request.Name,
		Priority:       request.Priority,
		Status:         request.Status,
		ConditionType:  request.ConditionType,
		ConditionValue: request.ConditionValue,
		ActionType:     request.ActionType,
		ActionValue:    request.ActionValue,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toDeliveryRuleResponse(rule))
}

func (handler deliveryRuleHandler) listRules(ctx *gin.Context) {
	shopID, ok := parseUintPathParam(ctx, "shop", "shop")
	if !ok {
		return
	}

	var status *model.DeliveryRuleStatus
	if value := ctx.Query("status"); value != "" {
		parsed := model.DeliveryRuleStatus(value)
		status = &parsed
	}

	rules, err := handler.rules.ListRules(ctx.Request.Context(), service.ListRulesCommand{
		ShopID: shopID,
		Status: status,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := make([]deliveryRuleResponse, 0, len(rules))
	for _, rule := range rules {
		response = append(response, toDeliveryRuleResponse(rule))
	}

	ctx.JSON(http.StatusOK, response)
}

func (handler deliveryRuleHandler) updateRuleStatus(ctx *gin.Context) {
	shopID, ok := parseUintPathParam(ctx, "shop", "shop")
	if !ok {
		return
	}
	id, ok := parseUintPathParam(ctx, "id", "id")
	if !ok {
		return
	}

	var request updateDeliveryRuleStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeBadRequest(ctx, "invalid request body")
		return
	}

	rule, err := handler.rules.UpdateRuleStatus(ctx.Request.Context(), service.UpdateRuleStatusCommand{
		ShopID: shopID,
		ID:     id,
		Status: request.Status,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toDeliveryRuleResponse(rule))
}

func (handler deliveryRuleHandler) importRules(ctx *gin.Context) {
	shopID, ok := parseUintPathParam(ctx, "shop", "shop")
	if !ok {
		return
	}

	var request importDeliveryRulesRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeBadRequest(ctx, "invalid request body")
		return
	}

	rules := make([]service.ImportRuleCommand, 0, len(request.Rules))
	for _, rule := range request.Rules {
		rules = append(rules, service.ImportRuleCommand{
			Name:           rule.Name,
			Priority:       rule.Priority,
			Status:         rule.Status,
			ConditionType:  rule.ConditionType,
			ConditionValue: rule.ConditionValue,
			ActionType:     rule.ActionType,
			ActionValue:    rule.ActionValue,
		})
	}

	summary, err := handler.rules.ValidateRuleImport(ctx.Request.Context(), service.ValidateRuleImportCommand{
		ShopID:      shopID,
		Rules:       rules,
		WorkerCount: request.WorkerCount,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toImportDeliveryRulesResponse(summary))
}

func parseUintPathParam(ctx *gin.Context, param string, field string) (uint, bool) {
	value, err := strconv.ParseUint(ctx.Param(param), 10, 64)
	if err != nil || value == 0 {
		writeError(ctx, &apperr.ValidationError{Field: field, Rule: "uint"})
		return 0, false
	}

	return uint(value), true
}

func toImportDeliveryRulesResponse(summary service.RuleImportValidationSummary) importDeliveryRulesResponse {
	response := importDeliveryRulesResponse{
		Total:   summary.Total,
		Valid:   summary.Valid,
		Invalid: summary.Invalid,
		Results: make([]importDeliveryRuleValidation, 0, len(summary.Results)),
	}
	for _, result := range summary.Results {
		item := importDeliveryRuleValidation{
			Index:     result.Index,
			RowNumber: result.RowNumber,
			Valid:     result.Valid,
		}
		for _, validationErr := range result.Errors {
			item.Errors = append(item.Errors, validationErrorResponse{
				Error: validationErr.Error(),
				Field: validationErr.Field,
				Rule:  validationErr.Rule,
			})
		}
		response.Results = append(response.Results, item)
	}
	return response
}

func toDeliveryRuleResponse(rule *model.DeliveryRule) deliveryRuleResponse {
	return deliveryRuleResponse{
		ID:             rule.ID,
		ShopID:         rule.ShopID,
		Name:           rule.Name,
		Priority:       rule.Priority,
		Status:         rule.Status,
		ConditionType:  rule.ConditionType,
		ConditionValue: rule.ConditionValue,
		ActionType:     rule.ActionType,
		ActionValue:    rule.ActionValue,
		CreatedAt:      rule.CreatedAt,
		UpdatedAt:      rule.UpdatedAt,
	}
}
