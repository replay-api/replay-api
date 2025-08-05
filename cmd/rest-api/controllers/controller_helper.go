package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	common "github.com/replay-api/replay-api/pkg/domain"
)

// ControllerHelper provides utility methods for controllers
type ControllerHelper struct{}

// NewControllerHelper creates a new controller helper
func NewControllerHelper() *ControllerHelper {
	return &ControllerHelper{}
}

// DecodeJSONRequest decodes JSON request body into the provided struct
func (h *ControllerHelper) DecodeJSONRequest(w http.ResponseWriter, r *http.Request, dest interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request", "err", err)
		apiErr := common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return common.WriteErrorResponse(w, apiErr)
	}
	return nil
}

// DecodeJSONRequestWithContext decodes JSON and stores errors in context for middleware handling
func (h *ControllerHelper) DecodeJSONRequestWithContext(r *http.Request, dest interface{}) (context.Context, error) {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request", "err", err)
		apiErr := common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return common.SetError(r.Context(), apiErr), err
	}
	return r.Context(), nil
}

// HandleError processes errors and writes appropriate responses
func (h *ControllerHelper) HandleError(w http.ResponseWriter, r *http.Request, err error, logMessage string) bool {
	if err == nil {
		return false
	}

	slog.ErrorContext(r.Context(), logMessage, "err", err)
	apiErr := common.ErrorFromString(err)
	if writeErr := common.WriteErrorResponse(w, apiErr); writeErr != nil {
		slog.ErrorContext(r.Context(), "Failed to write error response", "error", writeErr)
	}
	return true
}

// HandleErrorWithContext stores error in context for middleware handling
func (h *ControllerHelper) HandleErrorWithContext(r *http.Request, err error, logMessage string) (context.Context, bool) {
	if err == nil {
		return r.Context(), false
	}

	slog.ErrorContext(r.Context(), logMessage, "err", err)
	return common.SetError(r.Context(), err), true
}

// HandleBusinessLogicError handles domain-specific business logic errors
func (h *ControllerHelper) HandleBusinessLogicError(w http.ResponseWriter, r *http.Request, err error, operation string) bool {
	if err == nil {
		return false
	}

	slog.ErrorContext(r.Context(), "Business logic error in "+operation, "err", err)

	// Map business logic errors to appropriate HTTP status codes
	var apiErr *common.APIError
	errMsg := err.Error()
	switch {
	case errMsg == "Unauthorized":
		apiErr = common.ErrUnauthorized
	case contains(errMsg, "already exists"):
		apiErr = common.NewAPIError(http.StatusConflict, "CONFLICT", errMsg)
	case contains(errMsg, "not found"):
		apiErr = common.NewAPIError(http.StatusNotFound, "NOT_FOUND", errMsg)
	case contains(errMsg, "invalid") || contains(errMsg, "bad"):
		apiErr = common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", errMsg)
	default:
		apiErr = common.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", errMsg)
	}

	if writeErr := common.WriteErrorResponse(w, apiErr); writeErr != nil {
		slog.ErrorContext(r.Context(), "Failed to write error response", "error", writeErr)
	}
	return true
}

// WriteSuccess writes a successful response
func (h *ControllerHelper) WriteSuccess(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
	if err := common.WriteSuccessResponse(w, data, statusCode); err != nil {
		slog.ErrorContext(r.Context(), "Failed to encode response", "err", err)
	}
}

// WriteCreated writes a successful creation response
func (h *ControllerHelper) WriteCreated(w http.ResponseWriter, r *http.Request, data interface{}) {
	h.WriteSuccess(w, r, data, http.StatusCreated)
}

// WriteOK writes a successful OK response
func (h *ControllerHelper) WriteOK(w http.ResponseWriter, r *http.Request, data interface{}) {
	h.WriteSuccess(w, r, data, http.StatusOK)
}

// WriteNoContent writes a 204 No Content response
func (h *ControllerHelper) WriteNoContent(w http.ResponseWriter, r *http.Request) {
	h.WriteSuccess(w, r, nil, http.StatusNoContent)
}

// WriteBadRequest writes a standardized 400 Bad Request response
func (h *ControllerHelper) WriteBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	apiErr := common.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", message)
	if err := common.WriteErrorResponse(w, apiErr); err != nil {
		slog.ErrorContext(r.Context(), "Failed to write error response", "error", err)
	}
}

// contains is a helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(containsAtIndex(s, substr, 0) || containsInMiddle(s, substr))))
}

func containsAtIndex(s, substr string, index int) bool {
	if index+len(substr) > len(s) {
		return false
	}
	for i := 0; i < len(substr); i++ {
		if toLower(s[index+i]) != toLower(substr[i]) {
			return false
		}
	}
	return true
}

func containsInMiddle(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if containsAtIndex(s, substr, i) {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
