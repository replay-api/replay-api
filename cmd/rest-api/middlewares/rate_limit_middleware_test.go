package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitMiddleware_SkipsOptionsRequests(t *testing.T) {
	middleware := NewRateLimitMiddleware()
	
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// OPTIONS requests should never be rate limited
	for i := 0; i < 200; i++ {
		req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Fatalf("OPTIONS request %d was rate limited, expected 200 got %d", i, rec.Code)
		}
	}
}

func TestRateLimitMiddleware_RateLimitsGETRequests(t *testing.T) {
	// Create middleware with low limit for testing
	middleware := NewRateLimitMiddlewareWithConfig(5, time.Minute)
	
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	var successCount, rateLimitedCount int
	
	// Make requests until we hit the limit
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "192.168.1.2:1234"
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code == http.StatusOK {
			successCount++
		} else if rec.Code == http.StatusTooManyRequests {
			rateLimitedCount++
		}
	}
	
	if successCount != 5 {
		t.Errorf("Expected 5 successful requests, got %d", successCount)
	}
	if rateLimitedCount != 5 {
		t.Errorf("Expected 5 rate limited requests, got %d", rateLimitedCount)
	}
}

func TestRateLimitMiddleware_ReturnsDetailedError(t *testing.T) {
	middleware := NewRateLimitMiddlewareWithConfig(1, time.Minute)
	
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request - should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req1.RemoteAddr = "192.168.1.3:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	
	if rec1.Code != http.StatusOK {
		t.Fatalf("First request should succeed, got %d", rec1.Code)
	}

	// Second request - should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.RemoteAddr = "192.168.1.3:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("Second request should be rate limited, got %d", rec2.Code)
	}

	// Check response body
	var errorResponse RateLimitErrorResponse
	if err := json.NewDecoder(rec2.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse.Success {
		t.Error("Expected success=false in error response")
	}
	if errorResponse.Error == nil {
		t.Fatal("Expected error object in response")
	}
	if errorResponse.Error.Code != "RATE_LIMIT_EXCEEDED" {
		t.Errorf("Expected code 'RATE_LIMIT_EXCEEDED', got '%s'", errorResponse.Error.Code)
	}
	if errorResponse.Error.Limit != 1 {
		t.Errorf("Expected limit=1, got %d", errorResponse.Error.Limit)
	}
	if errorResponse.Error.Remaining != 0 {
		t.Errorf("Expected remaining=0, got %d", errorResponse.Error.Remaining)
	}
	if errorResponse.Error.RetryAfter <= 0 {
		t.Error("Expected retry_after_seconds > 0")
	}

	// Check headers
	if rec2.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header")
	}
	if rec2.Header().Get("X-RateLimit-Limit") != "1" {
		t.Errorf("Expected X-RateLimit-Limit=1, got '%s'", rec2.Header().Get("X-RateLimit-Limit"))
	}
	if rec2.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("Expected X-RateLimit-Remaining=0, got '%s'", rec2.Header().Get("X-RateLimit-Remaining"))
	}
	if rec2.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("Expected X-RateLimit-Reset header")
	}
}

func TestRateLimitMiddleware_HeadersOnSuccess(t *testing.T) {
	middleware := NewRateLimitMiddlewareWithConfig(10, time.Minute)
	
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "192.168.1.4:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rec.Code)
	}

	// Check rate limit headers are present on successful requests
	if rec.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("Expected X-RateLimit-Limit=10, got '%s'", rec.Header().Get("X-RateLimit-Limit"))
	}
	if rec.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("Expected X-RateLimit-Remaining header")
	}
	if rec.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("Expected X-RateLimit-Reset header")
	}
}

func TestRateLimiter_Check(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		result := rl.Check("test-client")
		if !result.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
		if result.Limit != 3 {
			t.Errorf("Expected limit=3, got %d", result.Limit)
		}
		// Remaining decreases each time
		expectedRemaining := 2 - i
		if result.Remaining != expectedRemaining {
			t.Errorf("Request %d: expected remaining=%d, got %d", i+1, expectedRemaining, result.Remaining)
		}
	}

	// 4th request should be denied
	result := rl.Check("test-client")
	if result.Allowed {
		t.Error("4th request should be denied")
	}
	if result.Remaining != 0 {
		t.Errorf("Expected remaining=0, got %d", result.Remaining)
	}
	if result.RetryAfterSec < 1 {
		t.Error("Expected RetryAfterSec >= 1")
	}
}

func TestRateLimiter_DifferentClients(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	// Client A uses 2 tokens
	for i := 0; i < 2; i++ {
		result := rl.Check("client-a")
		if !result.Allowed {
			t.Errorf("Client A request %d should be allowed", i+1)
		}
	}

	// Client A is now rate limited
	result := rl.Check("client-a")
	if result.Allowed {
		t.Error("Client A should be rate limited")
	}

	// Client B should still have full quota
	result = rl.Check("client-b")
	if !result.Allowed {
		t.Error("Client B should be allowed")
	}
	if result.Remaining != 1 {
		t.Errorf("Client B expected remaining=1, got %d", result.Remaining)
	}
}
