package middlewares

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

const (
	// APIKeyHeader is the header used for service-to-service API key authentication
	APIKeyHeader = "X-API-Key"
)

// APIKeyMiddleware validates API key for internal/system endpoints (e.g., oracle pipeline).
// This provides authentication for automated services (workers, cron jobs) that don't have
// user-level RID tokens but need to perform system-level operations.
//
// SECURITY: The API key is loaded from the ORACLE_API_KEY environment variable,
// which must be sourced from a Kubernetes Secret. Never hardcode API keys.
type APIKeyMiddleware struct {
	apiKey string
}

// NewAPIKeyMiddleware creates a new API key middleware.
// If ORACLE_API_KEY is not set, all requests will be rejected (fail-closed).
func NewAPIKeyMiddleware() *APIKeyMiddleware {
	key := os.Getenv("ORACLE_API_KEY")
	if key == "" {
		slog.Warn("ORACLE_API_KEY not configured — oracle API-key auth will reject all requests (fail-closed)")
	}
	return &APIKeyMiddleware{
		apiKey: key,
	}
}

// RequireAPIKeyOrAuth creates middleware that requires EITHER a valid API key OR an authenticated session.
// This allows both automated systems (via API key) and authenticated users (via RID token) to access
// oracle endpoints.
//
// When authenticated via API key, a system-level resource owner context is injected,
// giving the request platform-level access for RLS-filtered queries.
func (m *APIKeyMiddleware) RequireAPIKeyOrAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check if already authenticated via RID token (user session)
		if authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool); ok && authenticated {
			next.ServeHTTP(w, r)
			return
		}

		// Check API key header
		providedKey := r.Header.Get(APIKeyHeader)
		if providedKey == "" {
			slog.WarnContext(ctx, "oracle endpoint accessed without authentication or API key",
				"path", r.URL.Path,
				"method", r.Method,
				"remote_addr", r.RemoteAddr,
			)
			writeUnauthorized(w, "Authentication required: provide X-API-Key header or valid session")
			return
		}

		// Fail-closed: if API key is not configured, reject
		if m.apiKey == "" {
			slog.ErrorContext(ctx, "ORACLE_API_KEY not configured, rejecting API-key auth attempt",
				"path", r.URL.Path,
			)
			writeUnauthorized(w, "Service configuration error")
			return
		}

		// Constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(m.apiKey)) != 1 {
			slog.WarnContext(ctx, "invalid API key provided for oracle endpoint",
				"path", r.URL.Path,
				"method", r.Method,
				"remote_addr", r.RemoteAddr,
			)
			writeUnauthorized(w, "Invalid API key")
			return
		}

		// API key is valid — set system-level resource owner context
		systemUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		ctx = context.WithValue(ctx, shared.TenantIDKey, replay_common.TeamPROTenantID)
		ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
		ctx = context.WithValue(ctx, shared.UserIDKey, systemUserID)
		ctx = context.WithValue(ctx, shared.GroupIDKey, systemUserID)
		ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

		slog.InfoContext(ctx, "oracle endpoint authenticated via API key",
			"path", r.URL.Path,
			"method", r.Method,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
		"code":    "UNAUTHORIZED",
	})
}
