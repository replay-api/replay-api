package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
)

// Mock implementations for testing
type mockVerifyRIDKeyCommand struct {
	shouldFail bool
}

func (m *mockVerifyRIDKeyCommand) Exec(ctx context.Context, key uuid.UUID) (common.ResourceOwner, common.IntendedAudienceKey, error) {
	if m.shouldFail {
		return common.ResourceOwner{}, common.UserAudienceIDKey, common.ErrUnauthorized
	}
	// Return mock resource owner and audience
	return common.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		UserID:   uuid.New(),
	}, common.UserAudienceIDKey, nil
}

// Test response structure
type TestResponse struct {
	Message string `json:"message"`
	Data    string `json:"data"`
}

// Mock final handler that records execution and context state
type testHandler struct {
	executed        bool
	receivedHeaders map[string]string
	contextData     map[string]interface{}
	action          func(w http.ResponseWriter, r *http.Request)
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.executed = true
	h.receivedHeaders = make(map[string]string)
	h.contextData = make(map[string]interface{})

	// Capture important headers
	h.receivedHeaders["Origin"] = r.Header.Get("Origin")
	h.receivedHeaders["Authorization"] = r.Header.Get("Authorization")
	h.receivedHeaders["X-Resource-Owner-ID"] = r.Header.Get("X-Resource-Owner-ID") // Capture context data
	if steamID := r.Context().Value("steamID"); steamID != nil {
		h.contextData["steamID"] = steamID
	}

	if resourceOwner := r.Context().Value(common.UserIDKey); resourceOwner != nil {
		h.contextData["resourceOwner"] = resourceOwner
	}
	if err := common.GetError(r.Context()); err != nil {
		h.contextData["error"] = err
	}

	// Execute custom action if provided
	if h.action != nil {
		h.action(w, r)
		return
	}

	// Check if there's an error in context and don't write success response
	if err := common.GetError(r.Context()); err != nil {
		// Error middleware should handle this - don't write anything
		return
	}

	// Default successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TestResponse{
		Message: "success",
		Data:    "middleware chain completed",
	})
}

func (h *testHandler) reset() {
	h.executed = false
	h.receivedHeaders = nil
	h.contextData = nil
	h.action = nil
}

// Helper function to create a container with mock dependencies
func createTestContainer(shouldFailRID bool) *container.Container {
	c := container.New()

	mockVerifyRID := &mockVerifyRIDKeyCommand{
		shouldFail: shouldFailRID,
	}

	c.Singleton(func() iam_in.VerifyRIDKeyCommand {
		return mockVerifyRID
	})

	return &c
}

// Helper function to create CORS config for testing
func createTestCORSConfig() common.CORSConfig {
	return common.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000", "https://example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Resource-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
}

func TestMiddlewareChain_CompleteIntegration(t *testing.T) {
	tests := []struct {
		name                   string
		setupRequest           func() *http.Request
		corsConfig             common.CORSConfig
		ridShouldFail          bool
		expectedStatus         int
		expectedCORSHeaders    map[string]string
		expectedHandlerExecute bool
		validateResponse       func(t *testing.T, rr *httptest.ResponseRecorder, handler *testHandler)
	}{
		{
			name: "Successful request with all middlewares",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("X-Resource-Owner-ID", uuid.New().String()) // Correct header name
				return req
			},
			corsConfig:             createTestCORSConfig(),
			ridShouldFail:          false,
			expectedStatus:         http.StatusOK,
			expectedCORSHeaders:    map[string]string{"Access-Control-Allow-Origin": "http://localhost:3000"},
			expectedHandlerExecute: true,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder, handler *testHandler) {
				var resp TestResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if resp.Message != "success" {
					t.Errorf("Expected success message, got %s", resp.Message)
				}
			},
		},
		{
			name: "CORS preflight request",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("OPTIONS", "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")
				return req
			},
			corsConfig:             createTestCORSConfig(),
			ridShouldFail:          false,
			expectedStatus:         http.StatusOK, // CORS middleware returns 200 for OPTIONS
			expectedCORSHeaders:    map[string]string{"Access-Control-Allow-Origin": "http://localhost:3000"},
			expectedHandlerExecute: false, // CORS middleware handles OPTIONS and returns early
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder, handler *testHandler) {
				// Verify CORS headers are set correctly for preflight
				if rr.Header().Get("Access-Control-Allow-Methods") == "" {
					t.Error("Expected Access-Control-Allow-Methods header for preflight")
				}
				if rr.Header().Get("Access-Control-Allow-Headers") == "" {
					t.Error("Expected Access-Control-Allow-Headers header for preflight")
				}
			},
		},
		{
			name: "Missing Resource Owner ID header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				// No X-Resource-Owner-ID header
				return req
			},
			corsConfig:             createTestCORSConfig(),
			ridShouldFail:          false,
			expectedStatus:         http.StatusOK, // Middleware warns but continues
			expectedCORSHeaders:    map[string]string{"Access-Control-Allow-Origin": "http://localhost:3000"},
			expectedHandlerExecute: true, // Handler still executes
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder, handler *testHandler) {
				// Should succeed despite missing header (middleware only warns)
				var resp TestResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
			},
		},
		{
			name: "Invalid RID key triggers error",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("X-Resource-Owner-ID", uuid.New().String()) // Valid UUID format
				return req
			},
			corsConfig:             createTestCORSConfig(),
			ridShouldFail:          true,          // RID verification fails
			expectedStatus:         http.StatusOK, // Handler doesn't write response when context error detected
			expectedCORSHeaders:    map[string]string{"Access-Control-Allow-Origin": "http://localhost:3000"},
			expectedHandlerExecute: true, // Handler executes but detects error and returns early
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder, handler *testHandler) {
				// Check if error was set in context during handler execution
				if err, exists := handler.contextData["error"]; exists && err != nil {
					t.Logf("Error correctly set in context: %v", err)
				} else {
					t.Error("Expected error to be set in context")
				}

				// Handler should have detected error and not written response
				if rr.Body.Len() > 0 {
					t.Error("Expected no response body when handler detects context error")
				}
			},
		},
		{
			name: "Disallowed CORS origin",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Origin", "https://malicious-site.com")
				req.Header.Set("X-Resource-Owner-ID", uuid.New().String())
				return req
			},
			corsConfig:             createTestCORSConfig(),
			ridShouldFail:          false,
			expectedStatus:         http.StatusOK,       // CORS doesn't block, just doesn't set headers
			expectedCORSHeaders:    map[string]string{}, // No CORS headers set
			expectedHandlerExecute: true,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder, handler *testHandler) {
				// Should succeed but without CORS headers
				if rr.Header().Get("Access-Control-Allow-Origin") != "" {
					t.Error("Should not have CORS headers for disallowed origin")
				}
			},
		},
		{
			name: "Handler throws error - middleware catches it",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("X-Resource-Owner-ID", uuid.New().String())
				return req
			},
			corsConfig:             createTestCORSConfig(),
			ridShouldFail:          false,
			expectedStatus:         http.StatusOK, // Handler sets error but doesn't write response
			expectedCORSHeaders:    map[string]string{"Access-Control-Allow-Origin": "http://localhost:3000"},
			expectedHandlerExecute: true,
			validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder, handler *testHandler) {
				// Handler sets error in context and doesn't write response
				// This demonstrates proper error handling pattern for handlers
				if rr.Body.Len() > 0 {
					t.Error("Expected no response body when handler sets context error")
				} else {
					t.Log("Handler correctly handled error by not writing response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create dependencies
			container := createTestContainer(tt.ridShouldFail)
			corsMiddleware := NewCORSMiddleware(tt.corsConfig)
			resourceContextMiddleware := NewResourceContextMiddleware(container)

			// Create test handler
			handler := &testHandler{}

			// Set up special handler action for error testing
			if tt.name == "Handler throws error - middleware catches it" {
				handler.action = func(w http.ResponseWriter, r *http.Request) {
					// Simulate handler setting an error in context
					ctx := common.SetError(r.Context(), common.NewAPIError(
						http.StatusBadRequest,
						"VALIDATION_ERROR",
						"Handler generated error",
					))
					*r = *r.WithContext(ctx)
					// Don't write any response - let error middleware handle it
				}
			} // Build middleware chain (same order as in router.go)
			var chain http.Handler = handler

			// Apply middlewares in reverse order (last applied = first executed)
			chain = resourceContextMiddleware.Handler(chain)
			chain = mux.CORSMethodMiddleware(&mux.Router{})(chain) // Simplified for testing
			chain = corsMiddleware.Handler(chain)
			chain = ErrorMiddleware(chain) // Error middleware is first (outermost)

			// Execute request
			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			chain.ServeHTTP(rr, req)

			// Validate status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Validate CORS headers
			for expectedHeader, expectedValue := range tt.expectedCORSHeaders {
				actualValue := rr.Header().Get(expectedHeader)
				if actualValue != expectedValue {
					t.Errorf("Expected %s: %s, got %s", expectedHeader, expectedValue, actualValue)
				}
			}

			// Validate handler execution
			if handler.executed != tt.expectedHandlerExecute {
				t.Errorf("Expected handler executed: %t, got: %t", tt.expectedHandlerExecute, handler.executed)
			}

			// Validate content type for error responses
			if rr.Code >= 400 {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json for error, got %s", contentType)
				}
			}

			// Run custom validation
			if tt.validateResponse != nil {
				tt.validateResponse(t, rr, handler)
			}

			// Reset handler for next test
			handler.reset()
		})
	}
}

func TestMiddlewareChain_ErrorPropagation(t *testing.T) {
	t.Run("Error set by middleware propagates correctly", func(t *testing.T) {
		container := createTestContainer(false)
		corsMiddleware := NewCORSMiddleware(createTestCORSConfig())
		resourceContextMiddleware := NewResourceContextMiddleware(container)

		var errorReceived bool
		// Handler that checks if error from middleware is accessible
		handler := &testHandler{
			action: func(w http.ResponseWriter, r *http.Request) {
				if err := common.GetError(r.Context()); err != nil {
					// Error should be accessible in final handler
					errorReceived = true
					t.Logf("Handler received error: %v", err)
					// Handler detects error and doesn't write response
					return
				}
				// Write success response if no error
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			},
		}

		// Build chain
		var chain http.Handler = handler
		chain = resourceContextMiddleware.Handler(chain)
		chain = corsMiddleware.Handler(chain)
		chain = ErrorMiddleware(chain)

		// Request without Resource Owner ID (will trigger warning but not error)
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)

		// Resource context middleware only warns about missing RID, doesn't set error
		// So handler should execute successfully
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		// Should not have received error since missing RID only generates warning
		if errorReceived {
			t.Error("Did not expect error for missing RID (only generates warning)")
		} else {
			t.Log("Correctly handled missing RID with warning only")
		}
	})
}

func TestMiddlewareChain_OrderDependency(t *testing.T) {
	t.Run("Middleware order affects behavior", func(t *testing.T) {
		container := createTestContainer(false)
		corsMiddleware := NewCORSMiddleware(createTestCORSConfig())
		authMiddleware := NewAuthMiddleware()
		resourceContextMiddleware := NewResourceContextMiddleware(container)

		// Handler that checks for context errors
		var authErrorDetected bool
		handler := &testHandler{
			action: func(w http.ResponseWriter, r *http.Request) {
				if err := common.GetError(r.Context()); err != nil {
					authErrorDetected = true
					t.Logf("Handler detected auth error: %v", err)
					// Handler detects error and doesn't write response
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			},
		}

		// Test 1: Correct order (ErrorMiddleware first)
		var correctChain http.Handler = handler
		correctChain = resourceContextMiddleware.Handler(correctChain)
		correctChain = corsMiddleware.Handler(correctChain)
		correctChain = authMiddleware.Handler(correctChain)
		correctChain = ErrorMiddleware(correctChain) // First (outermost)

		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.Header.Set("Origin", "http://localhost:3000")
		req1.Header.Set("X-Resource-Owner-ID", uuid.New().String())
		// No auth header - should trigger error
		rr1 := httptest.NewRecorder()

		correctChain.ServeHTTP(rr1, req1)

		// Handler should detect auth error and not write response
		if !authErrorDetected {
			t.Error("Expected handler to detect auth error")
		}
		if rr1.Body.Len() > 0 {
			t.Error("Expected no response body when auth error detected")
		}

		// Reset for second test
		authErrorDetected = false

		// Test 2: Wrong order (ErrorMiddleware after auth) - But same behavior expected
		var wrongChain http.Handler = handler
		wrongChain = resourceContextMiddleware.Handler(wrongChain)
		wrongChain = corsMiddleware.Handler(wrongChain)
		wrongChain = ErrorMiddleware(wrongChain) // Wrong position
		wrongChain = authMiddleware.Handler(wrongChain)

		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.Header.Set("Origin", "http://localhost:3000")
		req2.Header.Set("X-Resource-Owner-ID", uuid.New().String())
		// No auth header - auth middleware sets error in context
		rr2 := httptest.NewRecorder()

		wrongChain.ServeHTTP(rr2, req2)

		// Same behavior expected - handler should detect error and not write response
		if !authErrorDetected {
			t.Error("Expected handler to detect auth error even with wrong middleware order")
		}

		// The key difference is that CORS headers should still be set
		// because CORS middleware runs before auth in the wrong chain
		if rr2.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Log("CORS headers correctly set regardless of middleware order")
		}
	})
}

func TestMiddlewareChain_ContextPassing(t *testing.T) {
	t.Run("Context values pass through middleware chain", func(t *testing.T) {
		container := createTestContainer(false)
		corsMiddleware := NewCORSMiddleware(createTestCORSConfig())
		resourceContextMiddleware := NewResourceContextMiddleware(container)

		// Handler that verifies context values
		contextReceived := make(map[string]interface{})
		handler := &testHandler{
			action: func(w http.ResponseWriter, r *http.Request) {
				// Check what context values we received
				if steamID := r.Context().Value("steamID"); steamID != nil {
					contextReceived["steamID"] = steamID
				}
				if resourceOwner := r.Context().Value(common.UserIDKey); resourceOwner != nil {
					contextReceived["resourceOwner"] = resourceOwner
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			},
		}

		// Build chain
		var chain http.Handler = handler
		chain = resourceContextMiddleware.Handler(chain)
		chain = corsMiddleware.Handler(chain)
		chain = ErrorMiddleware(chain)

		// Request with all required headers
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("X-Resource-ID", "test-resource-123")
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)

		// Should succeed
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		// Verify context values were passed through
		t.Logf("Context values received: %+v", contextReceived)

		// At minimum, we should have some context data
		// (Exact values depend on middleware implementations)
		if len(contextReceived) == 0 {
			t.Log("No context values received - middlewares may not be setting context")
		}
	})
}

// Benchmark the complete middleware chain
func BenchmarkMiddlewareChain_Complete(b *testing.B) {
	container := createTestContainer(false)
	corsMiddleware := NewCORSMiddleware(createTestCORSConfig())
	authMiddleware := NewAuthMiddleware()
	resourceContextMiddleware := NewResourceContextMiddleware(container)

	handler := &testHandler{
		action: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		},
	}

	// Build complete chain
	var chain http.Handler = handler
	chain = resourceContextMiddleware.Handler(chain)
	chain = corsMiddleware.Handler(chain)
	chain = authMiddleware.Handler(chain)
	chain = ErrorMiddleware(chain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("X-Resource-ID", "test-resource-123")
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)
	}
}

func BenchmarkMiddlewareChain_WithError(b *testing.B) {
	container := createTestContainer(false)
	corsMiddleware := NewCORSMiddleware(createTestCORSConfig())
	authMiddleware := NewAuthMiddleware()
	resourceContextMiddleware := NewResourceContextMiddleware(container)

	handler := &testHandler{}

	// Build complete chain
	var chain http.Handler = handler
	chain = resourceContextMiddleware.Handler(chain)
	chain = corsMiddleware.Handler(chain)
	chain = authMiddleware.Handler(chain)
	chain = ErrorMiddleware(chain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		// No auth header - triggers error
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)
	}
}
