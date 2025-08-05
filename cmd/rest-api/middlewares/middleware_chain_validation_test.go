package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// Test that validates the actual middleware chain behavior as implemented
func TestMiddlewareChain_ActualBehavior(t *testing.T) {
	// Create test dependencies
	container := createTestContainer(false)
	corsMiddleware := NewCORSMiddleware(createTestCORSConfig())
	resourceContextMiddleware := NewResourceContextMiddleware(container)

	t.Run("Complete middleware chain with successful request", func(t *testing.T) {
		// Handler that captures context and writes success
		var capturedContext map[string]interface{}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = make(map[string]interface{})

			// Capture context values that middlewares should set
			if tenantID := r.Context().Value(common.TenantIDKey); tenantID != nil {
				capturedContext["tenantID"] = tenantID
			}
			if clientID := r.Context().Value(common.ClientIDKey); clientID != nil {
				capturedContext["clientID"] = clientID
			}
			if userID := r.Context().Value(common.UserIDKey); userID != nil {
				capturedContext["userID"] = userID
			}
			if authenticated := r.Context().Value(common.AuthenticatedKey); authenticated != nil {
				capturedContext["authenticated"] = authenticated
			}
			if audience := r.Context().Value(common.AudienceKey); audience != nil {
				capturedContext["audience"] = audience
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		})

		// Build the middleware chain (same order as router.go)
		var chain http.Handler = handler
		chain = resourceContextMiddleware.Handler(chain)
		chain = mux.CORSMethodMiddleware(&mux.Router{})(chain)
		chain = corsMiddleware.Handler(chain)
		chain = ErrorMiddleware(chain)

		// Make request with proper headers
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("X-Resource-Owner-ID", uuid.New().String())
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)

		// Validate response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		// Validate CORS headers
		if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Errorf("Expected CORS origin header, got %s", rr.Header().Get("Access-Control-Allow-Origin"))
		}

		// Validate response body
		var resp map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if resp["status"] != "success" {
			t.Errorf("Expected status success, got %s", resp["status"])
		}

		// Validate context values were set by middlewares
		if capturedContext["tenantID"] == nil {
			t.Error("Expected tenantID to be set in context")
		}
		if capturedContext["clientID"] == nil {
			t.Error("Expected clientID to be set in context")
		}
		if capturedContext["userID"] == nil {
			t.Error("Expected userID to be set in context")
		}
		if authenticated, ok := capturedContext["authenticated"].(bool); !ok || !authenticated {
			t.Error("Expected authenticated to be true")
		}
		if capturedContext["audience"] == nil {
			t.Error("Expected audience to be set in context")
		}

		t.Logf("Captured context: %+v", capturedContext)
	})

	t.Run("CORS preflight request handling", func(t *testing.T) {
		var handlerCalled bool
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		// Build chain
		var chain http.Handler = handler
		chain = resourceContextMiddleware.Handler(chain)
		chain = corsMiddleware.Handler(chain)
		chain = ErrorMiddleware(chain)

		// Make OPTIONS request
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)

		// Validate CORS preflight response
		if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Error("Expected CORS origin header for preflight")
		}
		if rr.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("Expected CORS methods header for preflight")
		}
		if rr.Header().Get("Access-Control-Allow-Headers") == "" {
			t.Error("Expected CORS headers header for preflight")
		}

		// CORS middleware should handle OPTIONS and return early
		if handlerCalled {
			t.Error("Handler should not be called for CORS preflight")
		}
	})

	t.Run("Context error propagation", func(t *testing.T) {
		// Create middleware chain with RID verification that fails
		containerFail := createTestContainer(true) // RID will fail
		resourceContextMiddlewareFail := NewResourceContextMiddleware(containerFail)

		var handlerContextError error
		var handlerWasCalled bool
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerWasCalled = true
			// Check if error was set in context by middleware
			handlerContextError = common.GetError(r.Context())

			// If there's an error, don't write response (let error middleware handle)
			if handlerContextError != nil {
				t.Logf("Handler detected context error: %v", handlerContextError)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		})

		// Build chain
		var chain http.Handler = handler
		chain = resourceContextMiddlewareFail.Handler(chain)
		chain = corsMiddleware.Handler(chain)
		chain = ErrorMiddleware(chain)

		// Make request that will trigger RID verification failure
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("X-Resource-Owner-ID", uuid.New().String())
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)

		// Validate that error was properly propagated through context
		if !handlerWasCalled {
			t.Error("Expected handler to be called")
		}

		if handlerContextError == nil {
			t.Error("Expected error to be set in context by middleware")
		} else {
			t.Logf("Error correctly propagated: %v", handlerContextError)
		}

		// Since handler detected error and returned early, should have no content
		if rr.Body.Len() > 0 {
			t.Errorf("Expected no response body when handler detects context error, got: %s", rr.Body.String())
		}

		// Should still have CORS headers
		if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Error("Expected CORS headers even with error")
		}

		// This test validates that context error propagation works properly
		// In production, controllers should check for context errors and handle appropriately
		t.Log("Context error propagation working as expected - handlers receive context errors")
	})

	t.Run("Missing resource owner ID handling", func(t *testing.T) {
		var capturedContext map[string]interface{}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = make(map[string]interface{})
			if authenticated := r.Context().Value(common.AuthenticatedKey); authenticated != nil {
				capturedContext["authenticated"] = authenticated
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		})

		// Build chain
		var chain http.Handler = handler
		chain = resourceContextMiddleware.Handler(chain)
		chain = corsMiddleware.Handler(chain)
		chain = ErrorMiddleware(chain)

		// Make request WITHOUT X-Resource-Owner-ID header
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		// No X-Resource-Owner-ID header
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)

		// Should succeed but with authenticated = false
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		if authenticated, ok := capturedContext["authenticated"].(bool); !ok || authenticated {
			t.Error("Expected authenticated to be false when no resource owner ID")
		}
	})
}

// Benchmark the actual middleware chain performance
func BenchmarkMiddlewareChain_RealWorld(b *testing.B) {
	container := createTestContainer(false)
	corsMiddleware := NewCORSMiddleware(createTestCORSConfig())
	resourceContextMiddleware := NewResourceContextMiddleware(container)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Build complete chain
	var chain http.Handler = handler
	chain = resourceContextMiddleware.Handler(chain)
	chain = mux.CORSMethodMiddleware(&mux.Router{})(chain)
	chain = corsMiddleware.Handler(chain)
	chain = ErrorMiddleware(chain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("X-Resource-Owner-ID", uuid.New().String())
		rr := httptest.NewRecorder()

		chain.ServeHTTP(rr, req)
	}
}
