package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type LobbyController struct {
	container     container.Container
	lobbyCommand  matchmaking_in.LobbyCommand
}

func NewLobbyController(container container.Container, lobbyCommand matchmaking_in.LobbyCommand) *LobbyController {
	return &LobbyController{
		container:    container,
		lobbyCommand: lobbyCommand,
	}
}

// Request/Response DTOs
type CreateLobbyRequest struct {
	CreatorID        string `json:"creator_id"`
	GameID           string `json:"game_id"`
	Region           string `json:"region"`
	Tier             string `json:"tier"`
	DistributionRule string `json:"distribution_rule"`
	MaxPlayers       int    `json:"max_players"`
	AutoFill         bool   `json:"auto_fill"`
	InviteOnly       bool   `json:"invite_only"`
}

type CreateLobbyResponse struct {
	LobbyID   string `json:"lobby_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type JoinLobbyRequest struct {
	PlayerID string `json:"player_id"`
	MMR      int    `json:"mmr"`
}

type LobbyActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	LobbyID string `json:"lobby_id,omitempty"`
}

// CreateLobbyHandler handles POST /api/lobbies
func (ctrl *LobbyController) CreateLobbyHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Derive creator ID from authenticated context, not from request body
		ctx := r.Context()
		authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}
		creatorID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
		if !ok || creatorID == uuid.Nil {
			http.Error(w, `{"success":false,"error":"User identity not found"}`, http.StatusUnauthorized)
			return
		}

		var req CreateLobbyRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			slog.ErrorContext(apiContext, "failed to decode create lobby request", "error", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate distribution rule
		var distributionRule matchmaking_vo.DistributionRule
		switch req.DistributionRule {
		case "winner_takes_all":
			distributionRule = matchmaking_vo.DistributionRuleWinnerTakesAll
		case "top_three_split":
			distributionRule = matchmaking_vo.DistributionRuleTopThreeSplit
		default:
			distributionRule = matchmaking_vo.DistributionRuleWinnerTakesAll
		}

		// Create lobby command
		cmd := matchmaking_in.CreateLobbyCommand{
			CreatorID:        creatorID,
			GameID:           req.GameID,
			Region:           req.Region,
			Tier:             req.Tier,
			DistributionRule: distributionRule,
			MaxPlayers:       req.MaxPlayers,
			AutoFill:         req.AutoFill,
			InviteOnly:       req.InviteOnly,
		}

		lobby, err := ctrl.lobbyCommand.CreateLobby(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create lobby", "error", err)
			http.Error(w, `{"success":false,"error":"Failed to create lobby"}`, http.StatusInternalServerError)
			return
		}

		response := CreateLobbyResponse{
			LobbyID:   lobby.ID.String(),
			Status:    string(lobby.Status),
			CreatedAt: lobby.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		slog.InfoContext(apiContext, "lobby created successfully", "lobby_id", lobby.ID)
	}
}

// JoinLobbyHandler handles POST /api/lobbies/{lobby_id}/join
func (ctrl *LobbyController) JoinLobbyHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Derive player ID from authenticated context
		ctx := r.Context()
		authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}
		playerID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
		if !ok || playerID == uuid.Nil {
			http.Error(w, `{"success":false,"error":"User identity not found"}`, http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		lobbyID, err := uuid.Parse(vars["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		var req JoinLobbyRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			slog.ErrorContext(apiContext, "failed to decode join lobby request", "error", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := matchmaking_in.JoinLobbyCommand{
			LobbyID:  lobbyID,
			PlayerID: playerID, // From auth context, NOT from request body
			MMR:      req.MMR,
		}

		err = ctrl.lobbyCommand.JoinLobby(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to join lobby", "lobby_id", lobbyID, "player_id", playerID, "error", err)
			http.Error(w, `{"success":false,"error":"Failed to join lobby"}`, http.StatusBadRequest)
			return
		}

		response := LobbyActionResponse{
			Success: true,
			Message: "joined lobby successfully",
			LobbyID: lobbyID.String(),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		slog.InfoContext(apiContext, "player joined lobby", "lobby_id", lobbyID, "player_id", playerID)
	}
}

// LeaveLobbyHandler handles DELETE /api/lobbies/{lobby_id}/leave
func (ctrl *LobbyController) LeaveLobbyHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Derive player ID from authenticated context
		ctx := r.Context()
		authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}
		playerID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
		if !ok || playerID == uuid.Nil {
			http.Error(w, `{"success":false,"error":"User identity not found"}`, http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		lobbyID, err := uuid.Parse(vars["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		cmd := matchmaking_in.LeaveLobbyCommand{
			LobbyID:  lobbyID,
			PlayerID: playerID, // From auth context, NOT from query string
		}

		err = ctrl.lobbyCommand.LeaveLobby(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to leave lobby", "lobby_id", lobbyID, "player_id", playerID, "error", err)
			http.Error(w, `{"success":false,"error":"Failed to leave lobby"}`, http.StatusBadRequest)
			return
		}

		response := LobbyActionResponse{
			Success: true,
			Message: "left lobby successfully",
			LobbyID: lobbyID.String(),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		slog.InfoContext(apiContext, "player left lobby", "lobby_id", lobbyID, "player_id", playerID)
	}
}

// SetPlayerReadyHandler handles PUT /api/lobbies/{lobby_id}/ready
func (ctrl *LobbyController) SetPlayerReadyHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Derive player ID from authenticated context
		ctx := r.Context()
		authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}
		playerID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
		if !ok || playerID == uuid.Nil {
			http.Error(w, `{"success":false,"error":"User identity not found"}`, http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		lobbyID, err := uuid.Parse(vars["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		var reqBody struct {
			IsReady bool `json:"is_ready"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&reqBody); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := matchmaking_in.SetPlayerReadyCommand{
			LobbyID:  lobbyID,
			PlayerID: playerID, // From auth context, NOT from request body
			IsReady:  reqBody.IsReady,
		}

		err = ctrl.lobbyCommand.SetPlayerReady(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to set player ready", "lobby_id", lobbyID, "player_id", playerID, "error", err)
			http.Error(w, `{"success":false,"error":"Failed to update ready status"}`, http.StatusBadRequest)
			return
		}

		response := LobbyActionResponse{
			Success: true,
			Message: "ready status updated",
			LobbyID: lobbyID.String(),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		slog.InfoContext(apiContext, "player ready status updated", "lobby_id", lobbyID, "player_id", playerID, "is_ready", reqBody.IsReady)
	}
}

// StartMatchHandler handles POST /api/lobbies/{lobby_id}/start
func (ctrl *LobbyController) StartMatchHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Require authentication
		ctx := r.Context()
		authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		lobbyID, err := uuid.Parse(vars["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		matchID, err := ctrl.lobbyCommand.StartMatch(r.Context(), lobbyID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to start match", "lobby_id", lobbyID, "error", err)
			http.Error(w, `{"success":false,"error":"Failed to start match"}`, http.StatusBadRequest)
			return
		}

		response := struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			LobbyID string `json:"lobby_id"`
			MatchID string `json:"match_id"`
		}{
			Success: true,
			Message: "match started",
			LobbyID: lobbyID.String(),
			MatchID: matchID.String(),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		slog.InfoContext(apiContext, "match started", "lobby_id", lobbyID)
	}
}

// CancelLobbyHandler handles DELETE /api/lobbies/{lobby_id}
func (ctrl *LobbyController) CancelLobbyHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Require authentication
		ctx := r.Context()
		authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		lobbyID, err := uuid.Parse(vars["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		var reqBody struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&reqBody); err != nil {
			// If no body provided, use default reason
			reqBody.Reason = "cancelled by creator"
		}

		if reqBody.Reason == "" {
			reqBody.Reason = "cancelled by creator"
		}

		err = ctrl.lobbyCommand.CancelLobby(r.Context(), lobbyID, reqBody.Reason)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to cancel lobby", "lobby_id", lobbyID, "error", err)
			http.Error(w, `{"success":false,"error":"Failed to cancel lobby"}`, http.StatusBadRequest)
			return
		}

		response := LobbyActionResponse{
			Success: true,
			Message: "lobby cancelled and refunds issued",
			LobbyID: lobbyID.String(),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		slog.InfoContext(apiContext, "lobby cancelled", "lobby_id", lobbyID)
	}
}
