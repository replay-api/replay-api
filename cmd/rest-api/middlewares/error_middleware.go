package middlewares

import (
	"context"
	"log/slog"
	"net/http"

	common "github.com/replay-api/replay-api/pkg/domain"
)

// ErrorMiddleware handles all types of errors with proper header management
func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response recorder to capture status and prevent multiple header writes
		rw := &errorResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			headerWritten:  false,
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Check for errors in context first (highest priority)
		if err := common.GetError(r.Context()); err != nil && !rw.headerWritten {
			slog.ErrorContext(r.Context(), "Handling context error", "error", err)
			var apiErr *common.APIError
			if existingAPIErr, ok := err.(*common.APIError); ok {
				apiErr = existingAPIErr
			} else {
				apiErr = common.ErrorFromString(err)
			}

			rw.writeErrorResponse(apiErr)
			return
		}

		// Check for request context cancellation/timeout errors
		if ctxErr := r.Context().Err(); ctxErr != nil && !rw.headerWritten {
			slog.ErrorContext(r.Context(), "Request context error", "error", ctxErr)

			var apiErr *common.APIError
			switch ctxErr {
			case context.Canceled:
				apiErr = common.NewAPIError(http.StatusRequestTimeout, "REQUEST_CANCELLED", "Request was cancelled")
			case context.DeadlineExceeded:
				apiErr = common.NewAPIError(http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timeout")
			default:
				apiErr = common.NewAPIError(http.StatusInternalServerError, "CONTEXT_ERROR", ctxErr.Error())
			}

			rw.writeErrorResponse(apiErr)
			return
		}

		// Check if an error status was set but no body was written
		if rw.statusCode >= 400 && !rw.headerWritten {
			slog.ErrorContext(r.Context(), "Error status without response body", "status", rw.statusCode)

			var apiErr *common.APIError
			switch rw.statusCode {
			case http.StatusUnauthorized:
				apiErr = common.ErrUnauthorized
			case http.StatusForbidden:
				apiErr = common.ErrForbidden
			case http.StatusNotFound:
				apiErr = common.ErrNotFound
			case http.StatusBadRequest:
				apiErr = common.ErrBadRequest
			case http.StatusConflict:
				apiErr = common.ErrConflict
			default:
				apiErr = common.NewAPIError(rw.statusCode, "ERROR", http.StatusText(rw.statusCode))
			}

			rw.writeErrorResponse(apiErr)
			return
		}

		// Log successful requests
		if rw.statusCode < 400 {
			slog.InfoContext(r.Context(), "Request completed successfully", "status", rw.statusCode)
		}
	})
}

// ContextualErrorMiddleware is kept for backward compatibility, but now uses ErrorMiddleware
func ContextualErrorMiddleware(next http.Handler) http.Handler {
	return ErrorMiddleware(next)
}

// errorResponseWriter wraps http.ResponseWriter to track status and prevent multiple header writes
type errorResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func (rw *errorResponseWriter) WriteHeader(statusCode int) {
	if !rw.headerWritten {
		rw.statusCode = statusCode
		rw.headerWritten = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *errorResponseWriter) Write(data []byte) (int, error) {
	if !rw.headerWritten {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(data)
}

func (rw *errorResponseWriter) writeErrorResponse(apiErr *common.APIError) {
	if !rw.headerWritten {
		if err := common.WriteErrorResponse(rw.ResponseWriter, apiErr); err != nil {
			slog.Error("Failed to write error response", "error", err)
		}
	}
}
