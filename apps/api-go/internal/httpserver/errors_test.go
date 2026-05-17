package httpserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"smartdelivery/apps/api-go/internal/apperr"

	"github.com/gin-gonic/gin"
)

func TestWriteErrorMapsSingleValidationError(t *testing.T) {
	response := writeErrorForTest(&apperr.ValidationError{Field: "name", Rule: "required"})

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	var payload errorResponse
	decodeResponse(t, response, &payload)
	if payload.Field != "name" {
		t.Fatalf("Field = %q, want name", payload.Field)
	}
	if payload.Rule != "required" {
		t.Fatalf("Rule = %q, want required", payload.Rule)
	}
	if len(payload.Errors) != 0 {
		t.Fatalf("len(Errors) = %d, want 0", len(payload.Errors))
	}
}

func TestWriteErrorMapsJoinedValidationErrors(t *testing.T) {
	response := writeErrorForTest(errors.Join(
		&apperr.ValidationError{Field: "name", Rule: "required"},
		&apperr.ValidationError{Field: "status", Rule: "supported_value"},
	))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	var payload errorResponse
	decodeResponse(t, response, &payload)
	if len(payload.Errors) != 2 {
		t.Fatalf("len(Errors) = %d, want 2", len(payload.Errors))
	}
	if payload.Errors[0].Field != "name" {
		t.Fatalf("first field = %q, want name", payload.Errors[0].Field)
	}
	if payload.Errors[1].Field != "status" {
		t.Fatalf("second field = %q, want status", payload.Errors[1].Field)
	}
}

func TestWriteErrorMapsAppErrorsToStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{
			name:       "not found",
			err:        apperr.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantError:  "not found",
		},
		{
			name:       "conflict",
			err:        apperr.ErrConflict,
			wantStatus: http.StatusConflict,
			wantError:  "conflict",
		},
		{
			name:       "internal",
			err:        errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := writeErrorForTest(tt.err)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}

			var payload errorResponse
			decodeResponse(t, response, &payload)
			if payload.Error != tt.wantError {
				t.Fatalf("Error = %q, want %q", payload.Error, tt.wantError)
			}
		})
	}
}

func TestCollectValidationErrorsFlattensJoinedErrors(t *testing.T) {
	err := errors.Join(
		errors.New("plain error"),
		errors.Join(
			&apperr.ValidationError{Field: "name", Rule: "required"},
			&apperr.ValidationError{Field: "status", Rule: "supported_value"},
		),
	)

	validationErrors := collectValidationErrors(err)
	if len(validationErrors) != 2 {
		t.Fatalf("len(validationErrors) = %d, want 2", len(validationErrors))
	}
	if validationErrors[0].Field != "name" {
		t.Fatalf("first field = %q, want name", validationErrors[0].Field)
	}
	if validationErrors[1].Field != "status" {
		t.Fatalf("second field = %q, want status", validationErrors[1].Field)
	}
}

func TestWriteBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)

	writeBadRequest(ctx, "invalid request body")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	var payload errorResponse
	decodeResponse(t, response, &payload)
	if payload.Error != "invalid request body" {
		t.Fatalf("Error = %q, want invalid request body", payload.Error)
	}
}

func writeErrorForTest(err error) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	writeError(ctx, err)
	return response
}
