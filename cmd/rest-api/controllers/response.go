package controllers

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents a standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *APIMeta    `json:"meta,omitempty"`
}

// APIError represents an error in the API response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// APIMeta contains pagination and other metadata
type APIMeta struct {
	Total  int `json:"total,omitempty"`
	Page   int `json:"page,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// Common error codes
const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeValidation          = "VALIDATION_ERROR"
	ErrCodeInternalServer      = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeRateLimited         = "RATE_LIMITED"
	ErrCodeInvalidID           = "INVALID_ID"
	ErrCodeMissingParam        = "MISSING_PARAMETER"
)

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteSuccess writes a successful API response
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// WriteCreated writes a successful creation response
func WriteCreated(w http.ResponseWriter, data interface{}, location string) {
	if location != "" {
		w.Header().Set("Location", location)
	}
	WriteJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

// WriteSuccessWithMeta writes a successful response with pagination metadata
func WriteSuccessWithMeta(w http.ResponseWriter, data interface{}, meta *APIMeta) {
	WriteJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, code, message string, details ...string) {
	apiErr := &APIError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		apiErr.Details = details[0]
	}
	
	WriteJSON(w, status, APIResponse{
		Success: false,
		Error:   apiErr,
	})
}

// WriteBadRequest writes a 400 Bad Request response
func WriteBadRequest(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusBadRequest, ErrCodeBadRequest, message, details...)
}

// WriteUnauthorized writes a 401 Unauthorized response
func WriteUnauthorized(w http.ResponseWriter, message ...string) {
	msg := "Authentication required"
	if len(message) > 0 {
		msg = message[0]
	}
	WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, msg)
}

// WriteForbidden writes a 403 Forbidden response
func WriteForbidden(w http.ResponseWriter, message ...string) {
	msg := "Access denied"
	if len(message) > 0 {
		msg = message[0]
	}
	WriteError(w, http.StatusForbidden, ErrCodeForbidden, msg)
}

// WriteNotFound writes a 404 Not Found response
func WriteNotFound(w http.ResponseWriter, resource string) {
	WriteError(w, http.StatusNotFound, ErrCodeNotFound, resource+" not found")
}

// WriteConflict writes a 409 Conflict response
func WriteConflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, ErrCodeConflict, message)
}

// WriteValidationError writes a 422 Unprocessable Entity response
func WriteValidationError(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusUnprocessableEntity, ErrCodeValidation, message, details...)
}

// WriteInternalError writes a 500 Internal Server Error response
func WriteInternalError(w http.ResponseWriter, message ...string) {
	msg := "An internal error occurred"
	if len(message) > 0 {
		msg = message[0]
	}
	WriteError(w, http.StatusInternalServerError, ErrCodeInternalServer, msg)
}

// WriteServiceUnavailable writes a 503 Service Unavailable response
func WriteServiceUnavailable(w http.ResponseWriter, message ...string) {
	msg := "Service temporarily unavailable"
	if len(message) > 0 {
		msg = message[0]
	}
	WriteError(w, http.StatusServiceUnavailable, ErrCodeServiceUnavailable, msg)
}

// WriteTooManyRequests writes a 429 Too Many Requests response
func WriteTooManyRequests(w http.ResponseWriter, retryAfter int) {
	w.Header().Set("Retry-After", string(rune(retryAfter)))
	WriteError(w, http.StatusTooManyRequests, ErrCodeRateLimited, "Rate limit exceeded")
}

// WriteNoContent writes a 204 No Content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

