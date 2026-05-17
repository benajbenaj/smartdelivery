package shopify

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
)

func TestBuildInstallURL(t *testing.T) {
	svc := NewOAuthService(&fakeShopStore{}, OAuthConfig{APIKey: "test-key"})

	installURL, err := svc.BuildInstallURL(context.Background(), BuildInstallURLCommand{
		Shop:       "example.myshopify.com",
		AppBaseURL: "https://app.example.com",
		State:      "state-123",
	})
	if err != nil {
		t.Fatalf("BuildInstallURL() error = %v", err)
	}

	parsed, err := url.Parse(installURL)
	if err != nil {
		t.Fatalf("parse install URL: %v", err)
	}
	if parsed.Host != "example.myshopify.com" {
		t.Fatalf("Host = %q, want example.myshopify.com", parsed.Host)
	}
	if parsed.Path != "/admin/oauth/authorize" {
		t.Fatalf("Path = %q, want /admin/oauth/authorize", parsed.Path)
	}
	if parsed.Query().Get("client_id") != "test-key" {
		t.Fatalf("client_id = %q, want test-key", parsed.Query().Get("client_id"))
	}
	if parsed.Query().Get("redirect_uri") != "https://app.example.com/shopify/callback" {
		t.Fatalf("redirect_uri = %q", parsed.Query().Get("redirect_uri"))
	}
	if parsed.Query().Get("state") != "state-123" {
		t.Fatalf("state = %q, want state-123", parsed.Query().Get("state"))
	}
}

func TestCompleteInstallStoresShop(t *testing.T) {
	store := &fakeShopStore{}
	svc := NewOAuthService(store, OAuthConfig{APIKey: "test-key"})

	shop, err := svc.CompleteInstall(context.Background(), CompleteInstallCommand{
		Shop:  "example.myshopify.com",
		Code:  "code-123",
		State: "state-123",
	})
	if err != nil {
		t.Fatalf("CompleteInstall() error = %v", err)
	}

	if shop.Domain != "example.myshopify.com" {
		t.Fatalf("Domain = %q, want example.myshopify.com", shop.Domain)
	}
	if store.domain != "example.myshopify.com" {
		t.Fatalf("stored domain = %q, want example.myshopify.com", store.domain)
	}
}

func TestCompleteInstallValidatesRequiredParams(t *testing.T) {
	svc := NewOAuthService(&fakeShopStore{}, OAuthConfig{APIKey: "test-key"})

	_, err := svc.CompleteInstall(context.Background(), CompleteInstallCommand{
		Shop:  "example.myshopify.com",
		State: "state-123",
	})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("CompleteInstall() error = %v, want ErrValidation", err)
	}
}

func TestBuildInstallURLRejectsInvalidShopDomain(t *testing.T) {
	svc := NewOAuthService(&fakeShopStore{}, OAuthConfig{APIKey: "test-key"})

	_, err := svc.BuildInstallURL(context.Background(), BuildInstallURLCommand{
		Shop:       "https://example.myshopify.com",
		AppBaseURL: "https://app.example.com",
		State:      "state-123",
	})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("BuildInstallURL() error = %v, want ErrValidation", err)
	}
}

type fakeShopStore struct {
	domain string
}

func (store *fakeShopStore) FindOrCreateShop(_ context.Context, domain string) (*model.Shop, error) {
	store.domain = domain
	return &model.Shop{ID: 1, Domain: domain}, nil
}
