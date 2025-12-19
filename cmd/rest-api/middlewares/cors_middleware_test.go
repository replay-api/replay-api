package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCORSMiddleware_DefaultOrigins validates default origin configuration
// Business context: CORS must allow development origins by default for
// local development without additional configuration.
func TestCORSMiddleware_DefaultOrigins(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("CORS_ALLOWED_ORIGIN")

	m := NewCORSMiddleware()

	// Default development origins should be allowed
	assert.True(t, m.isOriginAllowed("http://localhost:3030"))
	assert.True(t, m.isOriginAllowed("http://localhost:3000"))
	assert.False(t, m.isOriginAllowed("https://evil.com"))
}

// TestCORSMiddleware_MultipleOrigins validates comma-separated origins
// Business context: E2E testing requires multiple origins (frontend, test runner).
// CORS_ALLOWED_ORIGINS supports comma-separated list for flexibility.
func TestCORSMiddleware_MultipleOrigins(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://leetgaming.gg, https://staging.leetgaming.gg, http://localhost:3000")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	m := NewCORSMiddleware()

	assert.True(t, m.isOriginAllowed("https://leetgaming.gg"))
	assert.True(t, m.isOriginAllowed("https://staging.leetgaming.gg"))
	assert.True(t, m.isOriginAllowed("http://localhost:3000"))
	assert.False(t, m.isOriginAllowed("https://malicious.com"))
}

// TestCORSMiddleware_SingleOriginFallback validates CORS_ALLOWED_ORIGIN env
// Business context: Backward compatibility with single origin configuration
// for simpler deployments.
func TestCORSMiddleware_SingleOriginFallback(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Setenv("CORS_ALLOWED_ORIGIN", "https://production.leetgaming.gg")
	defer os.Unsetenv("CORS_ALLOWED_ORIGIN")

	m := NewCORSMiddleware()

	assert.True(t, m.isOriginAllowed("https://production.leetgaming.gg"))
	assert.Equal(t, "https://production.leetgaming.gg", m.defaultOrigin)
}

// TestCORSMiddleware_Handler_SetsCorrectHeaders validates CORS header setting
func TestCORSMiddleware_Handler_SetsCorrectHeaders(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("CORS_ALLOWED_ORIGIN")

	m := NewCORSMiddleware()

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3030")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, "http://localhost:3030", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
}

// TestCORSMiddleware_Handler_PreflightRequest validates OPTIONS handling
// Business context: Browser preflight requests must receive proper CORS headers
// and return 200 OK without invoking the downstream handler.
func TestCORSMiddleware_Handler_PreflightRequest(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("CORS_ALLOWED_ORIGIN")

	m := NewCORSMiddleware()

	handlerCalled := false
	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3030")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, handlerCalled, "Handler should not be called for OPTIONS")
	assert.Equal(t, "http://localhost:3030", rr.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORSMiddleware_Handler_UnknownOrigin validates unknown origin behavior
// Business context: Requests from unknown origins still get CORS headers
// but use the default origin to maintain security.
func TestCORSMiddleware_Handler_UnknownOrigin(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("CORS_ALLOWED_ORIGIN")

	m := NewCORSMiddleware()

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "https://unknown.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Unknown origin should fall back to default
	assert.Equal(t, "http://localhost:3030", rr.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORSMiddleware_Handler_MatchingOrigin validates dynamic origin matching
// Business context: When the request Origin header matches an allowed origin,
// that specific origin is returned (not the default), enabling proper
// cross-origin requests from multiple allowed domains.
func TestCORSMiddleware_Handler_MatchingOrigin(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://leetgaming.gg, https://api.leetgaming.gg")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	m := NewCORSMiddleware()

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "https://api.leetgaming.gg")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should return the matching origin
	assert.Equal(t, "https://api.leetgaming.gg", rr.Header().Get("Access-Control-Allow-Origin"))
}

