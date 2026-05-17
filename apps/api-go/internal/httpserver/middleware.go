package httpserver

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func requestTimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestCtx, cancel := context.WithTimeout(ctx.Request.Context(), timeout)
		defer cancel()

		ctx.Request = ctx.Request.WithContext(requestCtx)
		ctx.Next()
	}
}
