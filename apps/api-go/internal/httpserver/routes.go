package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewHandler() http.Handler {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/healthz", healthzHandler)
	return router
}

func healthzHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
