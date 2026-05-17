package shopify

import (
	"context"
	"net/url"
	"strings"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
)

const DefaultRedirectPath = "/shopify/callback"

var DefaultScopes = []string{"read_shipping", "write_shipping"}

type ShopStore interface {
	FindOrCreateShop(ctx context.Context, domain string) (*model.Shop, error)
}

type OAuthConfig struct {
	APIKey       string
	Scopes       []string
	RedirectPath string
}

type OAuthService struct {
	shops ShopStore
	cfg   OAuthConfig
}

type BuildInstallURLCommand struct {
	Shop       string
	AppBaseURL string
	State      string
}

type CompleteInstallCommand struct {
	Shop  string
	Code  string
	State string
}

func NewOAuthService(shops ShopStore, cfg OAuthConfig) *OAuthService {
	return &OAuthService{
		shops: shops,
		cfg:   cfg.withDefaults(),
	}
}

func (svc *OAuthService) BuildInstallURL(_ context.Context, cmd BuildInstallURLCommand) (string, error) {
	if err := validateShopDomain(cmd.Shop); err != nil {
		return "", err
	}
	if cmd.AppBaseURL == "" {
		return "", &apperr.ValidationError{Field: "app_base_url", Rule: "required"}
	}
	if cmd.State == "" {
		return "", &apperr.ValidationError{Field: "state", Rule: "required"}
	}
	if svc.cfg.APIKey == "" {
		return "", &apperr.ValidationError{Field: "shopify_api_key", Rule: "required"}
	}

	values := url.Values{}
	values.Set("client_id", svc.cfg.APIKey)
	values.Set("scope", strings.Join(svc.cfg.Scopes, ","))
	values.Set("redirect_uri", strings.TrimRight(cmd.AppBaseURL, "/")+svc.cfg.RedirectPath)
	values.Set("state", cmd.State)

	installURL := url.URL{
		Scheme:   "https",
		Host:     cmd.Shop,
		Path:     "/admin/oauth/authorize",
		RawQuery: values.Encode(),
	}

	return installURL.String(), nil
}

func (svc *OAuthService) CompleteInstall(ctx context.Context, cmd CompleteInstallCommand) (*model.Shop, error) {
	if err := validateShopDomain(cmd.Shop); err != nil {
		return nil, err
	}
	if cmd.Code == "" {
		return nil, &apperr.ValidationError{Field: "code", Rule: "required"}
	}
	if cmd.State == "" {
		return nil, &apperr.ValidationError{Field: "state", Rule: "required"}
	}

	return svc.shops.FindOrCreateShop(ctx, cmd.Shop)
}

func validateShopDomain(shop string) error {
	if shop == "" {
		return &apperr.ValidationError{Field: "shop", Rule: "required"}
	}
	if strings.Contains(shop, "://") || strings.Contains(shop, "/") || !strings.HasSuffix(shop, ".myshopify.com") {
		return &apperr.ValidationError{Field: "shop", Rule: "myshopify_domain"}
	}
	return nil
}

func (cfg OAuthConfig) withDefaults() OAuthConfig {
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = DefaultScopes
	}
	if cfg.RedirectPath == "" {
		cfg.RedirectPath = DefaultRedirectPath
	}
	return cfg
}
