package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlayerProfileController struct {
	container container.Container
}

func NewPlayerProfileController(container container.Container) *PlayerProfileController {
	return &PlayerProfileController{container: container}
}

func (ctrl *PlayerProfileController) CreatePlayerProfileHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var createPlayerCommand squad_in.CreatePlayerProfileCommand
		err := json.NewDecoder(r.Body).Decode(&createPlayerCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to decode request", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var createPlayerCommandHandler squad_in.CreatePlayerProfileCommandHandler
		err = ctrl.container.Resolve(&createPlayerCommandHandler)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve CreatePlayerProfileCommandHandler", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		player, err := createPlayerCommandHandler.Exec(r.Context(), createPlayerCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create player profile",
				"error", err,
				"username", createPlayerCommand.Username,
				"game_id", createPlayerCommand.GameID)
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			} else if strings.Contains(err.Error(), "already exists") {
				w.WriteHeader(http.StatusConflict)
				errorJSON := map[string]string{
					"code":  "CONFLICT",
					"error": err.Error(),
				}

				err = json.NewEncoder(w).Encode(errorJSON)
				if err != nil {
					slog.ErrorContext(r.Context(), "Failed to encode response", "error", err)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		slog.InfoContext(r.Context(), "Player profile created successfully",
			"player_id", player.ID,
			"username", player.Username,
			"game_id", player.GameID,
			"user_id", r.Context().Value(common.UserIDKey))

		err = json.NewEncoder(w).Encode(player)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to encode response",
				"error", err,
				"player_id", player.ID)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
	}
}

// GetPlayerProfileHandler handles GET /players/{id}
func (ctrl *PlayerProfileController) GetPlayerProfileHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		profileID := vars["id"]

		if profileID == "" {
			http.Error(w, "profile_id is required", http.StatusBadRequest)
			return
		}

		var profileReader squad_out.PlayerProfileReader
		if err := ctrl.container.Resolve(&profileReader); err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve PlayerProfileReader", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(profileID)
		if err != nil {
			slog.WarnContext(r.Context(), "Invalid profile ID format", "profile_id", profileID, "error", err)
			http.Error(w, "invalid profile_id format", http.StatusBadRequest)
			return
		}

		valueParams := []common.SearchableValue{{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator}}
		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := profileReader.Search(r.Context(), search)
		if err != nil {
			slog.ErrorContext(r.Context(), "Error fetching player profile", "error", err, "profile_id", idUUID)
			http.Error(w, "error fetching profile", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			slog.WarnContext(r.Context(), "Player profile not found", "profile_id", idUUID)
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}

		slog.InfoContext(r.Context(), "Player profile retrieved successfully", "profile_id", idUUID, "username", results[0].Username)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results[0])
	}
}

type UpdatePlayerProfileRequest struct {
	Nickname    string   `json:"nickname"`
	AvatarURI   string   `json:"avatar_uri"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
}

// UpdatePlayerProfileHandler handles PUT /players/{id}
func (ctrl *PlayerProfileController) UpdatePlayerProfileHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		profileID := vars["id"]

		if profileID == "" {
			http.Error(w, "profile_id is required", http.StatusBadRequest)
			return
		}

		var req UpdatePlayerProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.ErrorContext(r.Context(), "Failed to decode update request", "error", err, "profile_id", profileID)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		var profileReader squad_out.PlayerProfileReader
		if err := ctrl.container.Resolve(&profileReader); err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve PlayerProfileReader", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(profileID)
		if err != nil {
			slog.WarnContext(r.Context(), "Invalid profile ID format for update", "profile_id", profileID, "error", err)
			http.Error(w, "invalid profile_id format", http.StatusBadRequest)
			return
		}

		valueParams := []common.SearchableValue{{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator}}
		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := profileReader.Search(r.Context(), search)
		if err != nil || len(results) == 0 {
			slog.WarnContext(r.Context(), "Player profile not found for update", "profile_id", idUUID, "error", err)
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}

		profile := results[0]

		// Update fields
		if req.Nickname != "" {
			profile.Nickname = req.Nickname
		}
		if req.AvatarURI != "" {
			profile.Avatar = req.AvatarURI
		}
		if req.Description != "" {
			profile.Description = req.Description
		}
		if req.Roles != nil {
			profile.Roles = req.Roles
		}
		profile.UpdatedAt = time.Now()

		var profileWriter squad_out.PlayerProfileWriter
		if err := ctrl.container.Resolve(&profileWriter); err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve PlayerProfileWriter", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		updatedProfile, err := profileWriter.Update(r.Context(), &profile)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to update player profile",
				"error", err,
				"profile_id", idUUID,
				"user_id", r.Context().Value(common.UserIDKey))
			http.Error(w, "failed to update profile", http.StatusInternalServerError)
			return
		}

		slog.InfoContext(r.Context(), "Player profile updated successfully",
			"profile_id", idUUID,
			"username", updatedProfile.Username,
			"user_id", r.Context().Value(common.UserIDKey))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedProfile)
	}
}

// DeletePlayerProfileHandler handles DELETE /players/{id}
func (ctrl *PlayerProfileController) DeletePlayerProfileHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		profileID := vars["id"]

		if profileID == "" {
			http.Error(w, "profile_id is required", http.StatusBadRequest)
			return
		}

		var profileWriter squad_out.PlayerProfileWriter
		if err := ctrl.container.Resolve(&profileWriter); err != nil {
			slog.ErrorContext(r.Context(), "Failed to resolve PlayerProfileWriter", "error", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		profileUUID, err := uuid.Parse(profileID)
		if err != nil {
			slog.WarnContext(r.Context(), "Invalid profile ID format for delete", "profile_id", profileID, "error", err)
			http.Error(w, "invalid profile_id format", http.StatusBadRequest)
			return
		}

		if err := profileWriter.Delete(r.Context(), profileUUID); err != nil {
			if err == mongo.ErrNoDocuments {
				slog.WarnContext(r.Context(), "Player profile not found for delete", "profile_id", profileUUID)
				http.Error(w, "profile not found", http.StatusNotFound)
				return
			}
			slog.ErrorContext(r.Context(), "Failed to delete player profile",
				"error", err,
				"profile_id", profileUUID,
				"user_id", r.Context().Value(common.UserIDKey))
			http.Error(w, "failed to delete profile", http.StatusInternalServerError)
			return
		}

		slog.InfoContext(r.Context(), "Player profile deleted successfully",
			"profile_id", profileUUID,
			"user_id", r.Context().Value(common.UserIDKey))

		w.WriteHeader(http.StatusNoContent)
	}
}
