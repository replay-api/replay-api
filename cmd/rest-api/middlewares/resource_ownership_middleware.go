package middlewares

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// ResourceOwnershipConfig defines configuration for resource ownership validation
type ResourceOwnershipConfig struct {
	// RequireAuthentication ensures user is authenticated
	RequireAuthentication bool
	// RequireOwnership ensures user owns the resource
	RequireOwnership bool
	// AllowedRoles defines roles that bypass ownership checks (e.g., admin, moderator)
	AllowedRoles []string
	// ResourceIDParam is the URL parameter name for the resource ID (default: "id")
	ResourceIDParam string
	// ResourceType for logging and error messages
	ResourceType common.ResourceType
}

// DefaultOwnershipConfig returns a standard ownership configuration
func DefaultOwnershipConfig(resourceType common.ResourceType) ResourceOwnershipConfig {
	return ResourceOwnershipConfig{
		RequireAuthentication: true,
		RequireOwnership:      true,
		AllowedRoles:          []string{},
		ResourceIDParam:       "id",
		ResourceType:          resourceType,
	}
}

// ResourceOwnershipMiddleware creates middleware that validates resource ownership
func ResourceOwnershipMiddleware(config ResourceOwnershipConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1. Authentication check
			if config.RequireAuthentication {
				isAuthenticated := ctx.Value(common.AuthenticatedKey)
				if isAuthenticated == nil || !isAuthenticated.(bool) {
					slog.WarnContext(ctx, "Unauthorized access attempt",
						"resource_type", config.ResourceType,
						"path", r.URL.Path,
						"method", r.Method)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			// 2. Ownership validation (if resource ID is present)
			if config.RequireOwnership {
				vars := mux.Vars(r)
				resourceIDStr := vars[config.ResourceIDParam]

				if resourceIDStr != "" {
					resourceID, err := uuid.Parse(resourceIDStr)
					if err != nil {
						slog.WarnContext(ctx, "Invalid resource ID format",
							"resource_type", config.ResourceType,
							"resource_id", resourceIDStr,
							"error", err)
						http.Error(w, "Invalid resource ID", http.StatusBadRequest)
						return
					}

					if !isUserAuthorizedForResource(ctx, resourceID, config) {
						slog.WarnContext(ctx, "Forbidden: user not authorized for resource",
							"resource_type", config.ResourceType,
							"resource_id", resourceID,
							"user_id", ctx.Value(common.UserIDKey))
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}

					// Store resource ID in context for handler use
					ctx = context.WithValue(ctx, common.ResourceIDKey, resourceID)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isUserAuthorizedForResource checks if the user has access to the resource
func isUserAuthorizedForResource(ctx context.Context, _ uuid.UUID, _ ResourceOwnershipConfig) bool {
	userID, ok := ctx.Value(common.UserIDKey).(uuid.UUID)
	if !ok {
		return false
	}

	resourceOwner := common.GetResourceOwner(ctx)

	// Check if user is the resource owner
	if resourceOwner.UserID == userID {
		return true
	}

	// Check if user's group matches resource owner's group
	groupID, ok := ctx.Value(common.GroupIDKey).(uuid.UUID)
	if ok && resourceOwner.GroupID == groupID {
		return true
	}

	// TODO: Add role-based checks when role system is implemented
	// For now, if allowed roles are configured, we'd check them here

	return false
}

// RequireAuthentication is a simple middleware that only checks authentication
func RequireAuthentication() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			isAuthenticated := ctx.Value(common.AuthenticatedKey)

			if isAuthenticated == nil || !isAuthenticated.(bool) {
				slog.WarnContext(ctx, "Authentication required",
					"path", r.URL.Path,
					"method", r.Method)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwnerOrAdmin creates middleware that requires user to be owner or have admin role
// This is used for squad updates where members with admin role can also modify
func RequireOwnerOrAdminRole(resourceType common.ResourceType, roleChecker func(ctx context.Context, resourceID uuid.UUID, userID uuid.UUID) (bool, error)) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Check authentication first
			isAuthenticated := ctx.Value(common.AuthenticatedKey)
			if isAuthenticated == nil || !isAuthenticated.(bool) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get resource ID from URL
			vars := mux.Vars(r)
			resourceIDStr := vars["id"]
			if resourceIDStr == "" {
				// No resource ID, proceed (for create operations)
				next.ServeHTTP(w, r)
				return
			}

			resourceID, err := uuid.Parse(resourceIDStr)
			if err != nil {
				http.Error(w, "Invalid resource ID", http.StatusBadRequest)
				return
			}

			userID, ok := ctx.Value(common.UserIDKey).(uuid.UUID)
			if !ok {
				http.Error(w, "User ID not found", http.StatusUnauthorized)
				return
			}

			// Check if user is owner via resource owner
			resourceOwner := common.GetResourceOwner(ctx)
			if resourceOwner.UserID == userID {
				// User is owner
				ctx = context.WithValue(ctx, common.ResourceIDKey, resourceID)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has admin role for this resource
			if roleChecker != nil {
				hasAdminRole, err := roleChecker(ctx, resourceID, userID)
				if err != nil {
					slog.ErrorContext(ctx, "Error checking user role",
						"resource_type", resourceType,
						"resource_id", resourceID,
						"user_id", userID,
						"error", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				if hasAdminRole {
					// User has admin role
					ctx = context.WithValue(ctx, common.ResourceIDKey, resourceID)
					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
					return
				}
			}

			// User is neither owner nor admin
			slog.WarnContext(ctx, "Forbidden: user lacks required role",
				"resource_type", resourceType,
				"resource_id", resourceID,
				"user_id", userID)
			http.Error(w, fmt.Sprintf("Forbidden: you must be the owner or have admin role for this %s", resourceType), http.StatusForbidden)
		})
	}
}

// ResourceVisibilityMiddleware validates that the user can access resources based on visibility rules
func ResourceVisibilityMiddleware(resourceType common.ResourceType) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get intended audience from context (set by search middleware)
			intendedAudience := ctx.Value(common.IntendedAudienceCtxKey)
			if intendedAudience == nil {
				// Default to user-level audience
				intendedAudience = common.UserAudienceIDKey
			}

			slog.DebugContext(ctx, "Visibility check",
				"resource_type", resourceType,
				"intended_audience", intendedAudience,
				"path", r.URL.Path)

			// Visibility validation is handled in the repository layer via Search
			// This middleware just ensures the context has the intended audience set
			next.ServeHTTP(w, r)
		})
	}
}
