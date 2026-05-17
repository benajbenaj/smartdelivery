package httpserver

import (
	"context"
	"errors"
	"net/http"

	"smartdelivery/apps/api-go/internal/apperr"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Error  string                    `json:"error"`
	Field  string                    `json:"field,omitempty"`
	Rule   string                    `json:"rule,omitempty"`
	Errors []validationErrorResponse `json:"errors,omitempty"`
}

type validationErrorResponse struct {
	Error string `json:"error"`
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

func writeError(ctx *gin.Context, err error) {
	validationErrors := collectValidationErrors(err)
	if len(validationErrors) == 1 {
		validationErr := validationErrors[0]
		ctx.JSON(http.StatusBadRequest, errorResponse{
			Error: validationErr.Error(),
			Field: validationErr.Field,
			Rule:  validationErr.Rule,
		})
		return
	}
	if len(validationErrors) > 1 {
		response := errorResponse{
			Error:  apperr.ErrValidation.Error(),
			Errors: make([]validationErrorResponse, 0, len(validationErrors)),
		}
		for _, validationErr := range validationErrors {
			response.Errors = append(response.Errors, validationErrorResponse{
				Error: validationErr.Error(),
				Field: validationErr.Field,
				Rule:  validationErr.Rule,
			})
		}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		ctx.JSON(http.StatusGatewayTimeout, errorResponse{Error: "request deadline exceeded"})
	case errors.Is(err, context.Canceled):
		ctx.JSON(http.StatusRequestTimeout, errorResponse{Error: "request canceled"})
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

func collectValidationErrors(err error) []*apperr.ValidationError {
	var result []*apperr.ValidationError
	for _, candidate := range flattenErrors(err) {
		var validationErr *apperr.ValidationError
		if errors.As(candidate, &validationErr) {
			result = append(result, validationErr)
		}
	}
	return result
}

func flattenErrors(err error) []error {
	if err == nil {
		return nil
	}

	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var result []error
		for _, child := range joined.Unwrap() {
			result = append(result, flattenErrors(child)...)
		}
		return result
	}

	return []error{err}
}
