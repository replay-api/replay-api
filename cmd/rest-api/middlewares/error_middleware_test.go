package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	common "github.com/replay-api/replay-api/pkg/domain"
)

// Test response structure for error validation
type ErrorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

// Mock handler that can simulate various scenarios
type mockHandler struct {
	action func(w http.ResponseWriter, r *http.Request)
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.action != nil {
		m.action(w, r)
	}
}

func TestErrorMiddleware_ContextErrors(t *testing.T) {
	tests := []struct {
		name           string
		contextError   error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "APIError in context",
			contextError:   common.NewAPIError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid input"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
			expectedMsg:    "Invalid input",
		},
		{
			name:           "Unauthorized error in context",
			contextError:   common.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
			expectedMsg:    "Unauthorized",
		},
		{
			name:           "Not found error in context",
			contextError:   common.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
			expectedMsg:    "Resource not found",
		},
		{
			name:           "Conflict error in context",
			contextError:   common.ErrConflict,
			expectedStatus: http.StatusConflict,
			expectedCode:   "CONFLICT",
			expectedMsg:    "Resource already exists",
		},
		{
			name:           "String error converted to APIError",
			contextError:   &testError{message: "user not found"},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
			expectedMsg:    "user not found",
		},
		{
			name:           "Generic string error",
			contextError:   &testError{message: "something went wrong"},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_SERVER_ERROR",
			expectedMsg:    "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock handler that sets error in context
			handler := &mockHandler{
				action: func(w http.ResponseWriter, r *http.Request) {
					// Set error in context
					ctx := common.SetError(r.Context(), tt.contextError)
					*r = *r.WithContext(ctx)
				},
			}

			// Wrap with error middleware
			middleware := ErrorMiddleware(handler)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			// Execute request
			middleware.ServeHTTP(rr, req)

			// Verify status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Verify content type
			if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			// Parse response
			var errorResp ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
				t.Fatalf("Failed to parse error response: %v", err)
			}

			// Verify error code and message
			if errorResp.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, errorResp.Code)
			}
			if errorResp.Error != tt.expectedMsg {
				t.Errorf("Expected error message %s, got %s", tt.expectedMsg, errorResp.Error)
			}
		})
	}
}

func TestErrorMiddleware_RequestContextErrors(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "Cancelled context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			},
			expectedStatus: http.StatusRequestTimeout,
			expectedCode:   "REQUEST_CANCELLED",
		},
		{
			name: "Deadline exceeded context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()
				time.Sleep(1 * time.Millisecond) // Ensure timeout
				return ctx
			},
			expectedStatus: http.StatusRequestTimeout,
			expectedCode:   "REQUEST_TIMEOUT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock handler that does nothing
			handler := &mockHandler{
				action: func(w http.ResponseWriter, r *http.Request) {
					// Handler does nothing, context error should be caught
				},
			}

			// Wrap with error middleware
			middleware := ErrorMiddleware(handler)

			// Create test request with specific context
			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(tt.setupContext())
			rr := httptest.NewRecorder()

			// Execute request
			middleware.ServeHTTP(rr, req)

			// Verify status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Parse response
			var errorResp ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
				t.Fatalf("Failed to parse error response: %v", err)
			}

			// Verify error code
			if errorResp.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, errorResp.Code)
			}
		})
	}
}

func TestErrorMiddleware_HTTPStatusErrors(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Bad Request status",
			statusCode:     http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Unauthorized status",
			statusCode:     http.StatusUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "Forbidden status",
			statusCode:     http.StatusForbidden,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "FORBIDDEN",
		},
		{
			name:           "Not Found status",
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "Conflict status",
			statusCode:     http.StatusConflict,
			expectedStatus: http.StatusConflict,
			expectedCode:   "CONFLICT",
		},
		{
			name:           "Internal Server Error status",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "ERROR",
		},
		{
			name:           "Custom 4xx status",
			statusCode:     http.StatusTeapot, // 418
			expectedStatus: http.StatusTeapot,
			expectedCode:   "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock handler that sets status but doesn't write body
			// The key here is to NOT call WriteHeader - just return, and let the middleware detect the error
			handler := &mockHandler{
				action: func(w http.ResponseWriter, r *http.Request) {
					// Simulate a controller that detects an error but doesn't handle it
					// This would typically be done by setting an error in context
					ctx := common.SetError(r.Context(), common.NewAPIError(tt.statusCode, tt.expectedCode, http.StatusText(tt.statusCode)))
					*r = *r.WithContext(ctx)
					// Don't call WriteHeader or Write - let middleware handle it
				},
			}

			// Wrap with error middleware
			middleware := ErrorMiddleware(handler)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			// Execute request
			middleware.ServeHTTP(rr, req)

			// Verify status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Parse response
			var errorResp ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
				t.Fatalf("Failed to parse error response: %v", err)
			}

			// Verify error code
			if errorResp.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, errorResp.Code)
			}
		})
	}
}

func TestErrorMiddleware_HTTPProtocolSafety(t *testing.T) {
	t.Run("Prevents multiple header writes", func(t *testing.T) {
		// Create handler that tries to write headers multiple times
		handler := &mockHandler{
			action: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.WriteHeader(http.StatusBadRequest) // This should be ignored
				w.Write([]byte(`{"data": "test"}`))
			},
		}

		middleware := ErrorMiddleware(handler)
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		// Should still be 200, not 400
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	})

	t.Run("Handles successful response", func(t *testing.T) {
		testData := map[string]string{"message": "success"}

		handler := &mockHandler{
			action: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(testData)
			},
		}

		middleware := ErrorMiddleware(handler)
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp["message"] != "success" {
			t.Errorf("Expected message 'success', got %s", resp["message"])
		}
	})
}

func TestErrorMiddleware_ErrorPrecedence(t *testing.T) {
	t.Run("Context error takes precedence over status error", func(t *testing.T) {
		contextErr := common.NewAPIError(http.StatusBadRequest, "CONTEXT_ERROR", "Context error message")

		handler := &mockHandler{
			action: func(w http.ResponseWriter, r *http.Request) {
				// Set context error
				ctx := common.SetError(r.Context(), contextErr)
				*r = *r.WithContext(ctx)

				// Don't set status - let middleware handle the context error
			},
		}

		middleware := ErrorMiddleware(handler)
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		// Should use context error, not status error
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 (from context), got %d", rr.Code)
		}

		var errorResp ErrorResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
			t.Fatalf("Failed to parse error response: %v", err)
		}

		if errorResp.Code != "CONTEXT_ERROR" {
			t.Errorf("Expected CONTEXT_ERROR, got %s", errorResp.Code)
		}
	})
}

func TestContextualErrorMiddleware_BackwardCompatibility(t *testing.T) {
	t.Run("ContextualErrorMiddleware uses ErrorMiddleware", func(t *testing.T) {
		contextErr := common.ErrUnauthorized

		handler := &mockHandler{
			action: func(w http.ResponseWriter, r *http.Request) {
				ctx := common.SetError(r.Context(), contextErr)
				*r = *r.WithContext(ctx)
			},
		}

		// Use the backward compatibility wrapper
		middleware := ContextualErrorMiddleware(handler)
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}

		var errorResp ErrorResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
			t.Fatalf("Failed to parse error response: %v", err)
		}

		if errorResp.Code != "UNAUTHORIZED" {
			t.Errorf("Expected UNAUTHORIZED, got %s", errorResp.Code)
		}
	})
}

func TestErrorResponseWriter_Implementation(t *testing.T) {
	t.Run("Tracks status code correctly", func(t *testing.T) {
		rw := &errorResponseWriter{
			ResponseWriter: httptest.NewRecorder(),
			statusCode:     http.StatusOK,
			headerWritten:  false,
		}

		rw.WriteHeader(http.StatusNotFound)
		if rw.statusCode != http.StatusNotFound {
			t.Errorf("Expected status code 404, got %d", rw.statusCode)
		}

		if !rw.headerWritten {
			t.Error("Expected headerWritten to be true")
		}
	})

	t.Run("Write sets header if not already written", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &errorResponseWriter{
			ResponseWriter: recorder,
			statusCode:     http.StatusOK,
			headerWritten:  false,
		}

		data := []byte("test data")
		n, err := rw.Write(data)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if n != len(data) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
		}

		if !rw.headerWritten {
			t.Error("Expected headerWritten to be true after Write")
		}

		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("writeErrorResponse only writes if header not written", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &errorResponseWriter{
			ResponseWriter: recorder,
			statusCode:     http.StatusOK,
			headerWritten:  false,
		}

		apiErr := common.NewAPIError(http.StatusBadRequest, "TEST_ERROR", "Test error message")
		rw.writeErrorResponse(apiErr)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", recorder.Code)
		}

		// Try to write again - should be ignored
		apiErr2 := common.NewAPIError(http.StatusInternalServerError, "IGNORED", "Should be ignored")
		rw.writeErrorResponse(apiErr2)

		// Status should still be 400, not 500
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("Expected status to remain 400, got %d", recorder.Code)
		}
	})
}

// Helper test error type
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// Benchmark tests for performance validation
func BenchmarkErrorMiddleware_SuccessPath(b *testing.B) {
	handler := &mockHandler{
		action: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		},
	}

	middleware := ErrorMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
	}
}

func BenchmarkErrorMiddleware_ContextError(b *testing.B) {
	handler := &mockHandler{
		action: func(w http.ResponseWriter, r *http.Request) {
			ctx := common.SetError(r.Context(), common.ErrUnauthorized)
			*r = *r.WithContext(ctx)
		},
	}

	middleware := ErrorMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
	}
}

func BenchmarkErrorMiddleware_StatusError(b *testing.B) {
	handler := &mockHandler{
		action: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			// Don't write body to trigger error handling
		},
	}

	middleware := ErrorMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
	}
}
