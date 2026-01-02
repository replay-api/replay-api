package middlewares

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	common "github.com/replay-api/replay-api/pkg/domain"
)

type AuthMiddleware struct {
	// Public paths that don't require authentication
	publicPaths map[string]bool
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		publicPaths: map[string]bool{
			"/health":            true,
			"/ready":             true,
			"/coverage":          true,
			"/metrics":           true,
			"/onboarding/steam":  true,
			"/onboarding/google": true,
			"/onboarding/email":  true,
			"/auth/login":        true,
			"/auth/logout":       true, // Logout should work even with expired tokens
			"/auth/guest":        true, // Guest token creation doesn't require auth
			"/webhooks/stripe":   true,
		},
	}
}

// publicPostPaths returns paths that allow POST without authentication
func (am *AuthMiddleware) isPublicPostPath(path string) bool {
	publicPostPaths := map[string]bool{
		"/onboarding/steam":  true,
		"/onboarding/google": true,
		"/onboarding/email":  true,
		"/auth/login":        true,
		"/auth/logout":       true, // Logout should work even with expired tokens
		"/auth/guest":        true, // Guest token creation doesn't require auth
		"/webhooks/stripe":   true,
	}
	return publicPostPaths[path]
}

// isPublicPath checks if the given path is a public endpoint
func (am *AuthMiddleware) isPublicPath(path string) bool {
	// Check exact match
	if am.publicPaths[path] {
		return true
	}

	// Check for public path prefixes
	publicPrefixes := []string{
		"/search/",
		"/tournaments",
		"/players",
		"/squads",
	}

	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			// GET requests on these paths are public
			return true
		}
	}

	return false
}

// isReadOnlyMethod returns true for methods that might be allowed on some public endpoints
func isReadOnlyMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := r.URL.Path
		method := r.Method

		// OPTIONS requests are always allowed (CORS preflight)
		if method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Public paths don't require authentication for read methods
		if am.isPublicPath(path) && isReadOnlyMethod(method) {
			next.ServeHTTP(w, r)
			return
		}

		// Allow POST for public post paths (onboarding, login, webhooks)
		if method == http.MethodPost && am.isPublicPostPath(path) {
			next.ServeHTTP(w, r)
			return
		}

		// Check if user is authenticated via ResourceContextMiddleware
		authenticated, ok := ctx.Value(common.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			// Protected endpoints require authentication
			protectedPrefixes := []string{
				"/wallet/",
				"/payments",
				"/api/lobbies",
				"/match-making/",
			}

			requiresAuth := false
			for _, prefix := range protectedPrefixes {
				if strings.HasPrefix(path, prefix) {
					requiresAuth = true
					break
				}
			}

			// POST/PUT/DELETE on most endpoints require auth
			if !isReadOnlyMethod(method) {
				requiresAuth = true
			}

			if requiresAuth {
				slog.WarnContext(ctx, "unauthorized access attempt",
					"path", path,
					"method", method,
					"authenticated", authenticated,
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Authentication required",
					"code":    "UNAUTHORIZED",
				})
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth is a middleware that strictly requires authentication
// Use this for sensitive endpoints like wallet operations
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authenticated, ok := ctx.Value(common.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			slog.WarnContext(ctx, "unauthorized access to protected endpoint",
				"path", r.URL.Path,
				"method", r.Method,
			)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Authentication required",
				"code":    "UNAUTHORIZED",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
