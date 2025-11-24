package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/out"
)

type ShareTokenController struct {
	commandService replay_in.ShareTokenCommand
	tokenReader    replay_out.ShareTokenReader
}

func NewShareTokenController(c container.Container) *ShareTokenController {
	var commandService replay_in.ShareTokenCommand
	var tokenReader replay_out.ShareTokenReader

	if err := c.Resolve(&commandService); err != nil {
		panic(err)
	}
	if err := c.Resolve(&tokenReader); err != nil {
		panic(err)
	}

	return &ShareTokenController{
		commandService: commandService,
		tokenReader:    tokenReader,
	}
}

type CreateShareTokenRequest struct {
	ResourceID    uuid.UUID                           `json:"resource_id"`
	ResourceType  replay_entity.SharingResourceType   `json:"resource_type"`
	ExpiresAt     *time.Time                          `json:"expires_at,omitempty"`
	ResourceOwner common.ResourceOwner                `json:"resource_owner"`
}

type ShareTokenResponse struct {
	Token         uuid.UUID                          `json:"token"`
	ResourceID    uuid.UUID                          `json:"resource_id"`
	ResourceType  replay_entity.SharingResourceType  `json:"resource_type"`
	ExpiresAt     time.Time                          `json:"expires_at"`
	Uri           string                             `json:"uri"`
	Status        replay_entity.ShareTokenStatus     `json:"status"`
	CreatedAt     time.Time                          `json:"created_at"`
}

// CreateShareToken handles POST /share-tokens
func (c *ShareTokenController) CreateShareToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateShareTokenRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.ErrorContext(ctx, "CreateShareToken: error decoding request", "error", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.ResourceID == uuid.Nil || req.ResourceType == "" {
			slog.ErrorContext(ctx, "CreateShareToken: missing required fields")
			http.Error(w, "resource_id and resource_type are required", http.StatusBadRequest)
			return
		}

		// Generate URI for the shared resource
		uri := generateShareURI(req.ResourceType, req.ResourceID)

		token := &replay_entity.ShareToken{
			ResourceID:    req.ResourceID,
			ResourceType:  req.ResourceType,
			Uri:           uri,
			EntityType:    string(req.ResourceType),
			Status:        replay_entity.ShareTokenStatusActive,
			ResourceOwner: req.ResourceOwner,
		}

		if req.ExpiresAt != nil {
			token.ExpiresAt = *req.ExpiresAt
		}

		if err := c.commandService.Create(r.Context(), token); err != nil {
			slog.ErrorContext(ctx, "CreateShareToken: error creating token", "error", err)
			http.Error(w, "error creating share token", http.StatusInternalServerError)
			return
		}

		response := ShareTokenResponse{
			Token:        token.ID,
			ResourceID:   token.ResourceID,
			ResourceType: token.ResourceType,
			ExpiresAt:    token.ExpiresAt,
			Uri:          token.Uri,
			Status:       token.Status,
			CreatedAt:    token.CreatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

		slog.InfoContext(ctx, "Share token created", "token_id", token.ID, "resource_id", req.ResourceID)
	}
}

// GetShareToken handles GET /share-tokens/{token}
func (c *ShareTokenController) GetShareToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenIDStr := vars["token"]

		tokenID, err := uuid.Parse(tokenIDStr)
		if err != nil {
			slog.ErrorContext(ctx, "GetShareToken: invalid token ID", "error", err)
			http.Error(w, "invalid token ID", http.StatusBadRequest)
			return
		}

		token, err := c.tokenReader.FindByToken(r.Context(), tokenID)
		if err != nil {
			slog.ErrorContext(ctx, "GetShareToken: error fetching token", "error", err)
			http.Error(w, "error fetching token", http.StatusInternalServerError)
			return
		}

		if token == nil {
			slog.WarnContext(ctx, "GetShareToken: token not found", "token_id", tokenID)
			http.Error(w, "token not found", http.StatusNotFound)
			return
		}

		// Check if token is expired
		if token.ExpiresAt.Before(time.Now()) && token.Status == replay_entity.ShareTokenStatusActive {
			token.Status = replay_entity.ShareTokenStatusExpired
			_ = c.commandService.Update(r.Context(), token)
		}

		if token.Status != replay_entity.ShareTokenStatusActive {
			slog.WarnContext(ctx, "GetShareToken: token not active", "token_id", tokenID, "status", token.Status)
			http.Error(w, "token is not active", http.StatusForbidden)
			return
		}

		response := ShareTokenResponse{
			Token:        token.ID,
			ResourceID:   token.ResourceID,
			ResourceType: token.ResourceType,
			ExpiresAt:    token.ExpiresAt,
			Uri:          token.Uri,
			Status:       token.Status,
			CreatedAt:    token.CreatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// RevokeShareToken handles DELETE /share-tokens/{token}
func (c *ShareTokenController) RevokeShareToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenIDStr := vars["token"]

		tokenID, err := uuid.Parse(tokenIDStr)
		if err != nil {
			slog.ErrorContext(ctx, "RevokeShareToken: invalid token ID", "error", err)
			http.Error(w, "invalid token ID", http.StatusBadRequest)
			return
		}

		if err := c.commandService.Revoke(r.Context(), tokenID); err != nil {
			slog.ErrorContext(ctx, "RevokeShareToken: error revoking token", "error", err)
			http.Error(w, "error revoking token", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		slog.InfoContext(ctx, "Share token revoked", "token_id", tokenID)
	}
}

// ListShareTokens handles GET /share-tokens
func (c *ShareTokenController) ListShareTokens(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get resource_id from query params
		resourceIDStr := r.URL.Query().Get("resource_id")
		
		if resourceIDStr == "" {
			http.Error(w, "resource_id query parameter is required", http.StatusBadRequest)
			return
		}

		resourceID, err := uuid.Parse(resourceIDStr)
		if err != nil {
			slog.ErrorContext(ctx, "ListShareTokens: invalid resource ID", "error", err)
			http.Error(w, "invalid resource ID", http.StatusBadRequest)
			return
		}

		tokens, err := c.tokenReader.FindByResourceID(r.Context(), resourceID)
		if err != nil {
			slog.ErrorContext(ctx, "ListShareTokens: error fetching tokens", "error", err)
			http.Error(w, "error fetching tokens", http.StatusInternalServerError)
			return
		}

		var response []ShareTokenResponse
		for _, token := range tokens {
			response = append(response, ShareTokenResponse{
				Token:        token.ID,
				ResourceID:   token.ResourceID,
				ResourceType: token.ResourceType,
				ExpiresAt:    token.ExpiresAt,
				Uri:          token.Uri,
				Status:       token.Status,
				CreatedAt:    token.CreatedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func generateShareURI(resourceType replay_entity.SharingResourceType, resourceID uuid.UUID) string {
	baseURL := "https://leetgaming.pro" // TODO: Make this configurable

	switch resourceType {
	case replay_entity.SharingResourceContentTypeMatchStats:
		return baseURL + "/match/" + resourceID.String()
	case replay_entity.SharingResourceTypeReplayFileContent:
		return baseURL + "/replays/" + resourceID.String()
	case replay_entity.SharingResourceContentTypePlayerStats:
		return baseURL + "/players/" + resourceID.String()
	case replay_entity.SharingResourceContentTypeTeamStats:
		return baseURL + "/teams/" + resourceID.String()
	default:
		return baseURL + "/shared/" + resourceID.String()
	}
}
