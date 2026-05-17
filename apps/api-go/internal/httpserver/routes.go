package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	DeliveryRules       DeliveryRuleService
	ShopifyInstallation ShopifyInstallationService
	RequestTimeout      time.Duration
}

func NewHandler(deps Dependencies) http.Handler {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if deps.RequestTimeout > 0 {
		router.Use(requestTimeoutMiddleware(deps.RequestTimeout))
	}
	router.GET("/healthz", healthzHandler)

	if deps.DeliveryRules != nil {
		registerDeliveryRuleRoutes(router, deps.DeliveryRules)
	}
	if deps.ShopifyInstallation != nil {
		registerShopifyRoutes(router, deps.ShopifyInstallation)
	}

	return router
}

func healthzHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
