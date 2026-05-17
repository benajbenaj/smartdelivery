package httpserver

import (
	"errors"
	"net/http"

	"smartdelivery/apps/api-go/internal/apperr"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Error string `json:"error"`
	Field string `json:"field,omitempty"`
	Rule  string `json:"rule,omitempty"`
}

func writeError(ctx *gin.Context, err error) {
	if validationErr, ok := errors.AsType[*apperr.ValidationError](err); ok {
		ctx.JSON(http.StatusBadRequest, errorResponse{
			Error: validationErr.Error(),
			Field: validationErr.Field,
			Rule:  validationErr.Rule,
		})
		return
	}

	switch {
	case errors.Is(err, apperr.ErrNotFound):
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "not found"})
	case errors.Is(err, apperr.ErrConflict):
		ctx.JSON(http.StatusConflict, errorResponse{Error: "conflict"})
	default:
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func writeBadRequest(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusBadRequest, errorResponse{Error: message})
}
