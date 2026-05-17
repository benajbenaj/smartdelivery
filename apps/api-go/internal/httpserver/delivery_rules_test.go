package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/service"

	"github.com/gin-gonic/gin"
)

func TestCreateDeliveryRuleEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := &fakeDeliveryRuleService{
		createResult: &model.DeliveryRule{
			ID:             7,
			ShopID:         42,
			Name:           "Hide express",
			Priority:       10,
			Status:         model.DeliveryRuleStatusDraft,
			ConditionType:  model.RuleConditionProductTag,
			ConditionValue: "hazardous",
			ActionType:     model.RuleActionHide,
			ActionValue:    "express",
		},
	}

	body := `{"name":"Hide express","priority":10,"condition_type":"product_tag","condition_value":"hazardous","action_type":"hide","action_value":"express"}`
	response := performRequest(NewHandler(Dependencies{DeliveryRules: rules}), http.MethodPost, "/shops/42/delivery-rules", body)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusCreated, response.Body.String())
	}
	if rules.createCmd.ShopID != 42 {
		t.Fatalf("ShopID = %d, want 42", rules.createCmd.ShopID)
	}
	if rules.createCmd.Name != "Hide express" {
		t.Fatalf("Name = %q, want Hide express", rules.createCmd.Name)
	}

	var payload deliveryRuleResponse
	decodeResponse(t, response, &payload)
	if payload.ID != 7 {
		t.Fatalf("payload.ID = %d, want 7", payload.ID)
	}
}

func TestListDeliveryRulesEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := &fakeDeliveryRuleService{
		listResult: []*model.DeliveryRule{
			{ID: 1, ShopID: 42, Name: "First", Status: model.DeliveryRuleStatusActive},
			{ID: 2, ShopID: 42, Name: "Second", Status: model.DeliveryRuleStatusActive},
		},
	}

	response := performRequest(NewHandler(Dependencies{DeliveryRules: rules}), http.MethodGet, "/shops/42/delivery-rules?status=active", "")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if rules.listCmd.ShopID != 42 {
		t.Fatalf("ShopID = %d, want 42", rules.listCmd.ShopID)
	}
	if rules.listCmd.Status == nil || *rules.listCmd.Status != model.DeliveryRuleStatusActive {
		t.Fatalf("Status = %v, want active", rules.listCmd.Status)
	}

	var payload []deliveryRuleResponse
	decodeResponse(t, response, &payload)
	if len(payload) != 2 {
		t.Fatalf("len(payload) = %d, want 2", len(payload))
	}
}

func TestPatchDeliveryRuleStatusEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := &fakeDeliveryRuleService{
		updateResult: &model.DeliveryRule{
			ID:     9,
			ShopID: 42,
			Name:   "Hide express",
			Status: model.DeliveryRuleStatusActive,
		},
	}

	response := performRequest(NewHandler(Dependencies{DeliveryRules: rules}), http.MethodPatch, "/shops/42/delivery-rules/9", `{"status":"active"}`)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if rules.updateCmd.ShopID != 42 {
		t.Fatalf("ShopID = %d, want 42", rules.updateCmd.ShopID)
	}
	if rules.updateCmd.ID != 9 {
		t.Fatalf("ID = %d, want 9", rules.updateCmd.ID)
	}
	if rules.updateCmd.Status != model.DeliveryRuleStatusActive {
		t.Fatalf("Status = %q, want active", rules.updateCmd.Status)
	}
}

func TestDeliveryRuleEndpointMapsValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := &fakeDeliveryRuleService{
		createErr: &apperr.ValidationError{Field: "name", Rule: "required"},
	}

	response := performRequest(NewHandler(Dependencies{DeliveryRules: rules}), http.MethodPost, "/shops/42/delivery-rules", `{}`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	var payload errorResponse
	decodeResponse(t, response, &payload)
	if payload.Field != "name" {
		t.Fatalf("Field = %q, want name", payload.Field)
	}
}

func TestDeliveryRuleEndpointMapsJoinedValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := &fakeDeliveryRuleService{
		createErr: errors.Join(
			&apperr.ValidationError{Field: "name", Rule: "required"},
			&apperr.ValidationError{Field: "action_type", Rule: "supported_value"},
		),
	}

	response := performRequest(NewHandler(Dependencies{DeliveryRules: rules}), http.MethodPost, "/shops/42/delivery-rules", `{}`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	var payload errorResponse
	decodeResponse(t, response, &payload)
	if len(payload.Errors) != 2 {
		t.Fatalf("len(payload.Errors) = %d, want 2", len(payload.Errors))
	}
	if payload.Errors[0].Field != "name" {
		t.Fatalf("first field = %q, want name", payload.Errors[0].Field)
	}
	if payload.Errors[1].Field != "action_type" {
		t.Fatalf("second field = %q, want action_type", payload.Errors[1].Field)
	}
}

func TestDeliveryRuleEndpointRejectsInvalidPathParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	response := performRequest(NewHandler(Dependencies{DeliveryRules: &fakeDeliveryRuleService{}}), http.MethodGet, "/shops/not-a-number/delivery-rules", "")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestRequestTimeoutCancelsServiceContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := &fakeDeliveryRuleService{waitForListContextCancel: true}

	response := performRequest(
		NewHandler(Dependencies{DeliveryRules: rules, RequestTimeout: time.Nanosecond}),
		http.MethodGet,
		"/shops/42/delivery-rules",
		"",
	)

	if response.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusGatewayTimeout, response.Body.String())
	}
	if !errors.Is(rules.listContextErr, context.DeadlineExceeded) {
		t.Fatalf("listContextErr = %v, want DeadlineExceeded", rules.listContextErr)
	}
}

func performRequest(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v; body = %s", err, response.Body.String())
	}
}

type fakeDeliveryRuleService struct {
	createCmd                service.CreateRuleCommand
	createResult             *model.DeliveryRule
	createErr                error
	listCmd                  service.ListRulesCommand
	listResult               []*model.DeliveryRule
	listErr                  error
	listContextErr           error
	waitForListContextCancel bool
	updateCmd                service.UpdateRuleStatusCommand
	updateResult             *model.DeliveryRule
	updateErr                error
}

func (svc *fakeDeliveryRuleService) CreateRule(_ context.Context, cmd service.CreateRuleCommand) (*model.DeliveryRule, error) {
	svc.createCmd = cmd
	return svc.createResult, svc.createErr
}

func (svc *fakeDeliveryRuleService) ListRules(ctx context.Context, cmd service.ListRulesCommand) ([]*model.DeliveryRule, error) {
	svc.listCmd = cmd
	if svc.waitForListContextCancel {
		<-ctx.Done()
		svc.listContextErr = ctx.Err()
		return nil, svc.listContextErr
	}
	return svc.listResult, svc.listErr
}

func (svc *fakeDeliveryRuleService) UpdateRuleStatus(_ context.Context, cmd service.UpdateRuleStatusCommand) (*model.DeliveryRule, error) {
	svc.updateCmd = cmd
	return svc.updateResult, svc.updateErr
}
