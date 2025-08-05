package common

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// ErrorContextKey is used to store errors in the request context
type ErrorContextKey struct{}

// SetError stores an error in the request context for the error middleware to handle
func SetError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ErrorContextKey{}, err)
}

// GetError retrieves an error from the request context
func GetError(ctx context.Context) error {
	if err, ok := ctx.Value(ErrorContextKey{}).(error); ok {
		return err
	}
	return nil
}

// APIError represents a structured API error
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// Common API errors
var (
	ErrUnauthorized   = NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	ErrForbidden      = NewAPIError(http.StatusForbidden, "FORBIDDEN", "Forbidden")
	ErrNotFound       = NewAPIError(http.StatusNotFound, "NOT_FOUND", "Resource not found")
	ErrBadRequest     = NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Bad request")
	ErrConflict       = NewAPIError(http.StatusConflict, "CONFLICT", "Resource already exists")
	ErrInternalServer = NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error")
)

// ErrorFromString creates an APIError from a string message by detecting common patterns
func ErrorFromString(err error) *APIError {
	if err == nil {
		return nil
	}

	message := err.Error()

	switch {
	case message == "Unauthorized":
		return ErrUnauthorized
	case message == "Forbidden":
		return ErrForbidden
	case strings.Contains(strings.ToLower(message), "not found"):
		return NewAPIError(http.StatusNotFound, "NOT_FOUND", message)
	case strings.Contains(strings.ToLower(message), "already exists"):
		return NewAPIError(http.StatusConflict, "CONFLICT", message)
	case strings.Contains(strings.ToLower(message), "bad request"):
		return NewAPIError(http.StatusBadRequest, "BAD_REQUEST", message)
	default:
		return NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
	}
}

// WriteErrorResponse writes an API error as JSON response
func WriteErrorResponse(w http.ResponseWriter, apiErr *APIError) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)

	response := map[string]string{
		"code":  apiErr.Code,
		"error": apiErr.Message,
	}

	return json.NewEncoder(w).Encode(response)
}

// WriteSuccessResponse writes a successful response with proper headers
func WriteSuccessResponse(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		return json.NewEncoder(w).Encode(data)
	}
	return nil
}
