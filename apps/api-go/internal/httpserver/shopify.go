package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"smartdelivery/apps/api-go/internal/apperr"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/shopify"

	"github.com/gin-gonic/gin"
)

const shopifyOAuthStateCookie = "shopify_oauth_state"

type ShopifyInstallationService interface {
	BuildInstallURL(ctx context.Context, cmd shopify.BuildInstallURLCommand) (string, error)
	CompleteInstall(ctx context.Context, cmd shopify.CompleteInstallCommand) (*model.Shop, error)
}

type shopifyHandler struct {
	installations ShopifyInstallationService
}

type shopifyInstallResponse struct {
	Installed bool   `json:"installed"`
	ShopID    uint   `json:"shop_id"`
	Shop      string `json:"shop"`
}

func registerShopifyRoutes(router *gin.Engine, installations ShopifyInstallationService) {
	handler := shopifyHandler{installations: installations}

	router.GET("/shopify/install", handler.install)
	router.GET("/shopify/callback", handler.callback)
}

func (handler shopifyHandler) install(ctx *gin.Context) {
	state, err := newOAuthState()
	if err != nil {
		writeError(ctx, err)
		return
	}

	installURL, err := handler.installations.BuildInstallURL(ctx.Request.Context(), shopify.BuildInstallURLCommand{
		Shop:       ctx.Query("shop"),
		AppBaseURL: requestBaseURL(ctx.Request),
		State:      state,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	http.SetCookie(ctx.Writer, oauthStateCookie(ctx.Request, state, 300))
	ctx.Redirect(http.StatusFound, installURL)
}

func (handler shopifyHandler) callback(ctx *gin.Context) {
	state := ctx.Query("state")
	if state == "" {
		writeError(ctx, &apperr.ValidationError{Field: "state", Rule: "required"})
		return
	}
	if !oauthStateMatches(ctx.Request, state) {
		writeError(ctx, &apperr.ValidationError{Field: "state", Rule: "matches_cookie"})
		return
	}

	shop, err := handler.installations.CompleteInstall(ctx.Request.Context(), shopify.CompleteInstallCommand{
		Shop:  ctx.Query("shop"),
		Code:  ctx.Query("code"),
		State: state,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}
	if shop == nil {
		writeError(ctx, errors.New("shopify install returned nil shop"))
		return
	}

	http.SetCookie(ctx.Writer, oauthStateCookie(ctx.Request, "", -1))
	ctx.JSON(http.StatusOK, shopifyInstallResponse{
		Installed: true,
		ShopID:    shop.ID,
		Shop:      shop.Domain,
	})
}

func newOAuthState() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

func oauthStateCookie(request *http.Request, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     shopifyOAuthStateCookie,
		Value:    value,
		Path:     "/shopify/callback",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   request.TLS != nil,
	}
}

func oauthStateMatches(request *http.Request, state string) bool {
	cookie, err := request.Cookie(shopifyOAuthStateCookie)
	return err == nil && cookie.Value != "" && cookie.Value == state
}

func requestBaseURL(request *http.Request) string {
	scheme := "http"
	if request.TLS != nil {
		scheme = "https"
	}
	if forwardedProto := request.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}

	return scheme + "://" + request.Host
}
