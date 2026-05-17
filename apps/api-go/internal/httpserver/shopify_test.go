package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/shopify"

	"github.com/gin-gonic/gin"
)

func TestShopifyInstallRedirectsToOAuthURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	installations := &fakeShopifyInstallationService{
		installURL: "https://example.myshopify.com/admin/oauth/authorize?state=test",
	}

	response := performRequest(NewHandler(Dependencies{ShopifyInstallation: installations}), http.MethodGet, "/shopify/install?shop=example.myshopify.com", "")

	if response.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusFound)
	}
	if response.Header().Get("Location") != installations.installURL {
		t.Fatalf("Location = %q, want %q", response.Header().Get("Location"), installations.installURL)
	}
	if installations.installCmd.Shop != "example.myshopify.com" {
		t.Fatalf("Shop = %q, want example.myshopify.com", installations.installCmd.Shop)
	}
	if installations.installCmd.State == "" {
		t.Fatalf("State is empty, want generated state")
	}
	if !strings.HasPrefix(installations.installCmd.AppBaseURL, "http://") {
		t.Fatalf("AppBaseURL = %q, want http URL", installations.installCmd.AppBaseURL)
	}
	if response.Header().Get("Set-Cookie") == "" {
		t.Fatalf("Set-Cookie is empty, want OAuth state cookie")
	}
}

func TestShopifyCallbackStoresShop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	installations := &fakeShopifyInstallationService{
		callbackShop: &model.Shop{ID: 7, Domain: "example.myshopify.com"},
	}

	request := httptestRequest(http.MethodGet, "/shopify/callback?shop=example.myshopify.com&code=code-123&state=state-123", "")
	request.AddCookie(&http.Cookie{Name: shopifyOAuthStateCookie, Value: "state-123"})
	response := serveRequest(NewHandler(Dependencies{ShopifyInstallation: installations}), request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if installations.callbackCmd.Shop != "example.myshopify.com" {
		t.Fatalf("Shop = %q, want example.myshopify.com", installations.callbackCmd.Shop)
	}
	if installations.callbackCmd.Code != "code-123" {
		t.Fatalf("Code = %q, want code-123", installations.callbackCmd.Code)
	}

	var payload shopifyInstallResponse
	decodeResponse(t, response, &payload)
	if !payload.Installed {
		t.Fatalf("Installed = false, want true")
	}
	if payload.ShopID != 7 {
		t.Fatalf("ShopID = %d, want 7", payload.ShopID)
	}
}

func TestShopifyCallbackRejectsMissingState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	response := performRequest(NewHandler(Dependencies{ShopifyInstallation: &fakeShopifyInstallationService{}}), http.MethodGet, "/shopify/callback?shop=example.myshopify.com&code=code-123", "")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestShopifyCallbackRejectsStateCookieMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	request := httptestRequest(http.MethodGet, "/shopify/callback?shop=example.myshopify.com&code=code-123&state=state-123", "")
	request.AddCookie(&http.Cookie{Name: shopifyOAuthStateCookie, Value: "other-state"})
	response := serveRequest(NewHandler(Dependencies{ShopifyInstallation: &fakeShopifyInstallationService{}}), request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestShopifyCallbackMapsValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	installations := &fakeShopifyInstallationService{
		callbackErr: &apperr.ValidationError{Field: "code", Rule: "required"},
	}

	request := httptestRequest(http.MethodGet, "/shopify/callback?shop=example.myshopify.com&state=state-123", "")
	request.AddCookie(&http.Cookie{Name: shopifyOAuthStateCookie, Value: "state-123"})
	response := serveRequest(NewHandler(Dependencies{ShopifyInstallation: installations}), request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func httptestRequest(method string, path string, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}

func serveRequest(handler http.Handler, request *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

type fakeShopifyInstallationService struct {
	installCmd   shopify.BuildInstallURLCommand
	installURL   string
	installErr   error
	callbackCmd  shopify.CompleteInstallCommand
	callbackShop *model.Shop
	callbackErr  error
}

func (svc *fakeShopifyInstallationService) BuildInstallURL(_ context.Context, cmd shopify.BuildInstallURLCommand) (string, error) {
	svc.installCmd = cmd
	return svc.installURL, svc.installErr
}

func (svc *fakeShopifyInstallationService) CompleteInstall(_ context.Context, cmd shopify.CompleteInstallCommand) (*model.Shop, error) {
	svc.callbackCmd = cmd
	return svc.callbackShop, svc.callbackErr
}
