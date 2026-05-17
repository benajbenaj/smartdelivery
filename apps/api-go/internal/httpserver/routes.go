package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	DeliveryRules DeliveryRuleService
}

func NewHandler(deps Dependencies) http.Handler {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/healthz", healthzHandler)

	if deps.DeliveryRules != nil {
		registerDeliveryRuleRoutes(router, deps.DeliveryRules)
	}

	return router
}

func healthzHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
