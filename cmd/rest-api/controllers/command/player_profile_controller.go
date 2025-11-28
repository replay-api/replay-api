package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
)

type PlayerProfileController struct {
	container               container.Container
	createPlayerUseCase     squad_in.CreatePlayerProfileCommandHandler
	updatePlayerUseCase     squad_in.UpdatePlayerProfileCommandHandler
	deletePlayerUseCase     squad_in.DeletePlayerProfileCommandHandler
	playerProfileReader     squad_out.PlayerProfileReader
}

func NewPlayerProfileController(container container.Container) *PlayerProfileController {
	var createPlayerUseCase squad_in.CreatePlayerProfileCommandHandler
	if err := container.Resolve(&createPlayerUseCase); err != nil {
		slog.Error("Failed to resolve CreatePlayerProfileCommandHandler", "error", err)
		panic(err)
	}

	var updatePlayerUseCase squad_in.UpdatePlayerProfileCommandHandler
	if err := container.Resolve(&updatePlayerUseCase); err != nil {
		slog.Error("Failed to resolve UpdatePlayerProfileCommandHandler", "error", err)
		panic(err)
	}

	var deletePlayerUseCase squad_in.DeletePlayerProfileCommandHandler
	if err := container.Resolve(&deletePlayerUseCase); err != nil {
		slog.Error("Failed to resolve DeletePlayerProfileCommandHandler", "error", err)
		panic(err)
	}

	var playerProfileReader squad_out.PlayerProfileReader
	if err := container.Resolve(&playerProfileReader); err != nil {
		slog.Error("Failed to resolve PlayerProfileReader", "error", err)
		panic(err)
	}

	return &PlayerProfileController{
		container:           container,
		createPlayerUseCase: createPlayerUseCase,
		updatePlayerUseCase: updatePlayerUseCase,
		deletePlayerUseCase: deletePlayerUseCase,
		playerProfileReader: playerProfileReader,
	}
}

func (ctrl *PlayerProfileController) CreatePlayerProfileHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var createPlayerCommand squad_in.CreatePlayerProfileCommand
		err := json.NewDecoder(r.Body).Decode(&createPlayerCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to decode request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		player, err := ctrl.createPlayerUseCase.Exec(r.Context(), createPlayerCommand)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create player profile",
				"error", err,
				"nickname", createPlayerCommand.Nickname,
				"game_id", createPlayerCommand.GameID)
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			} else if strings.Contains(err.Error(), "already exists") {
				w.WriteHeader(http.StatusConflict)
				errorJSON := map[string]string{
					"code":  "CONFLICT",
					"error": err.Error(),
				}
				json.NewEncoder(w).Encode(errorJSON)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		slog.InfoContext(r.Context(), "Player profile created successfully",
			"player_id", player.ID,
			"nickname", player.Nickname,
			"game_id", player.GameID,
			"user_id", r.Context().Value(common.UserIDKey))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(player)
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

		idUUID, err := uuid.Parse(profileID)
		if err != nil {
			slog.WarnContext(r.Context(), "Invalid profile ID format", "profile_id", profileID, "error", err)
			http.Error(w, "invalid profile_id format", http.StatusBadRequest)
			return
		}

		valueParams := []common.SearchableValue{{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator}}
		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := ctrl.playerProfileReader.Search(r.Context(), search)
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

		slog.InfoContext(r.Context(), "Player profile retrieved successfully", "profile_id", idUUID, "nickname", results[0].Nickname)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results[0])
	}
}

type UpdatePlayerProfileRequest struct {
	Nickname        string   `json:"nickname"`
	Base64Avatar    string   `json:"base64_avatar"`
	AvatarExtension string   `json:"avatar_extension"`
	SlugURI         string   `json:"slug_uri"`
	Description     string   `json:"description"`
	Roles           []string `json:"roles"`
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

		idUUID, err := uuid.Parse(profileID)
		if err != nil {
			slog.WarnContext(r.Context(), "Invalid profile ID format for update", "profile_id", profileID, "error", err)
			http.Error(w, "invalid profile_id format", http.StatusBadRequest)
			return
		}

		updateCmd := squad_in.UpdatePlayerCommand{
			PlayerID:        idUUID,
			Nickname:        req.Nickname,
			Base64Avatar:    req.Base64Avatar,
			AvatarExtension: req.AvatarExtension,
			SlugURI:         req.SlugURI,
			Roles:           req.Roles,
			Description:     req.Description,
		}

		updatedProfile, err := ctrl.updatePlayerUseCase.Exec(r.Context(), updateCmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to update player profile",
				"error", err,
				"profile_id", idUUID,
				"user_id", r.Context().Value(common.UserIDKey))
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			} else if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
			} else if strings.Contains(err.Error(), "already exists") {
				w.WriteHeader(http.StatusConflict)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		slog.InfoContext(r.Context(), "Player profile updated successfully",
			"profile_id", idUUID,
			"nickname", updatedProfile.Nickname,
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

		profileUUID, err := uuid.Parse(profileID)
		if err != nil {
			slog.WarnContext(r.Context(), "Invalid profile ID format for delete", "profile_id", profileID, "error", err)
			http.Error(w, "invalid profile_id format", http.StatusBadRequest)
			return
		}

		err = ctrl.deletePlayerUseCase.Exec(r.Context(), profileUUID)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to delete player profile",
				"error", err,
				"profile_id", profileUUID,
				"user_id", r.Context().Value(common.UserIDKey))
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			} else if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		slog.InfoContext(r.Context(), "Player profile deleted successfully",
			"profile_id", profileUUID,
			"user_id", r.Context().Value(common.UserIDKey))

		w.WriteHeader(http.StatusNoContent)
	}
}
