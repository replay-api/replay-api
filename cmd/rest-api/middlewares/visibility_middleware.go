package middlewares

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entities "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type ResourceType string

const (
	ResourceTypeReplay        ResourceType = "replay"
	ResourceTypeSquad         ResourceType = "squad"
	ResourceTypePlayerProfile ResourceType = "player_profile"
)

type AdminBypassMode int

const (
	AdminBypassDisabled AdminBypassMode = iota
)

type VisibilityConfig struct {
	ResourceType      ResourceType
	ResourceIDParam   string
	AllowPublic       bool
	CheckShareToken   bool
	AdminBypass       AdminBypassMode
}

type ResourceReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (interface{}, error)
}

type ShareTokenValidator interface {
	ValidateToken(ctx context.Context, resourceID uuid.UUID, resourceType string) (bool, error)
}

type VisibilityMiddleware struct {
	readers             map[ResourceType]ResourceReader
	shareTokenValidator ShareTokenValidator
}

func NewVisibilityMiddleware(
	readers map[ResourceType]ResourceReader,
	shareTokenValidator ShareTokenValidator,
) *VisibilityMiddleware {
	return &VisibilityMiddleware{
		readers:             readers,
		shareTokenValidator: shareTokenValidator,
	}
}

func (m *VisibilityMiddleware) CheckVisibility(config VisibilityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			vars := mux.Vars(r)

			resourceIDStr := vars[config.ResourceIDParam]
			if resourceIDStr == "" {
				common.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Resource ID is required", "")
				return
			}

			resourceID, err := uuid.Parse(resourceIDStr)
			if err != nil {
				common.WriteError(w, http.StatusBadRequest, "INVALID_RESOURCE_ID", "Invalid resource ID format", err.Error())
				return
			}

			reader, exists := m.readers[config.ResourceType]
			if !exists {
				slog.ErrorContext(ctx, "No reader configured for resource type", "resource_type", config.ResourceType)
				common.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Server configuration error", "")
				return
			}

			resource, err := reader.GetByID(ctx, resourceID)
			if err != nil {
				if _, ok := err.(*common.ErrNotFound); ok {
					common.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", "")
				} else {
					slog.ErrorContext(ctx, "Failed to fetch resource", "resource_type", config.ResourceType, "resource_id", resourceID, "error", err)
					common.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch resource", "")
				}
				return
			}

			if !m.canAccessResource(ctx, resource, config) {
				common.WriteError(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource", "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *VisibilityMiddleware) canAccessResource(ctx context.Context, resource interface{}, config VisibilityConfig) bool {
	baseEntity := m.extractBaseEntity(resource)
	if baseEntity == nil {
		slog.ErrorContext(ctx, "Resource does not contain BaseEntity", "resource_type", config.ResourceType)
		return false
	}

	visibilityType := baseEntity.VisibilityType

	if config.AllowPublic && visibilityType == common.PublicVisibilityTypeKey {
		return true
	}

	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return false
	}

	currentUser := common.GetResourceOwner(ctx)
	resourceOwner := baseEntity.ResourceOwner

	if resourceOwner.UserID == currentUser.UserID {
		return true
	}

	if resourceOwner.GroupID == currentUser.GroupID && currentUser.GroupID != uuid.Nil {
		return true
	}

	if config.CheckShareToken && m.shareTokenValidator != nil {
		hasValidToken, err := m.shareTokenValidator.ValidateToken(ctx, baseEntity.ID, string(config.ResourceType))
		if err != nil {
			slog.WarnContext(ctx, "Failed to validate share token", "resource_id", baseEntity.ID, "error", err)
			return false
		}
		if hasValidToken {
			return true
		}
	}

	return false
}

func (m *VisibilityMiddleware) extractBaseEntity(resource interface{}) *common.BaseEntity {
	switch r := resource.(type) {
	case *replay_entities.ReplayFile:
		return &common.BaseEntity{
			ID:            r.ID,
			ResourceOwner: r.ResourceOwner,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		}
	case *replay_entities.Match:
		return &common.BaseEntity{
			ID:            r.ID,
			ResourceOwner: r.ResourceOwner,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		}
	case replay_entities.ReplayFile:
		return &common.BaseEntity{
			ID:            r.ID,
			ResourceOwner: r.ResourceOwner,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		}
	case replay_entities.Match:
		return &common.BaseEntity{
			ID:            r.ID,
			ResourceOwner: r.ResourceOwner,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		}
	default:
		if hasBaseEntity, ok := resource.(interface{ GetBaseEntity() *common.BaseEntity }); ok {
			return hasBaseEntity.GetBaseEntity()
		}
		return nil
	}
}
