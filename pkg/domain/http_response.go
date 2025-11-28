package common

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type HTTPResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDTO   `json:"error,omitempty"`
}

type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := HTTPResponse{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

func WriteError(w http.ResponseWriter, status int, code string, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := HTTPResponse{
		Success: false,
		Error: &ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode error response", "error", err)
	}
}

func WriteErrorFromDomainError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *ErrUnauthorized:
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", e.Error(), "")
	case *ErrNotFound:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", e.Error(), "")
	case *ErrAlreadyExists:
		WriteError(w, http.StatusConflict, "ALREADY_EXISTS", e.Error(), "")
	case *ErrInvalidInput:
		WriteError(w, http.StatusBadRequest, "INVALID_INPUT", e.Error(), "")
	case *ErrForbidden:
		WriteError(w, http.StatusForbidden, "FORBIDDEN", e.Error(), "")
	default:
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", err.Error())
	}
}

func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, data)
}

func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, data)
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
