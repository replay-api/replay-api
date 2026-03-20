package cmd_controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
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

type GetLobbyResponse struct {
	Lobby *matchmaking_entities.MatchmakingLobby `json:"lobby"`
}

type CommitmentPlayerSummary struct {
	PlayerID    string `json:"player_id"`
	Status      string `json:"status"`
	RespondedAt string `json:"responded_at,omitempty"`
	ExpiresAt   string `json:"expires_at"`
}

type CommitmentSummaryResponse struct {
	LobbyID              string                    `json:"lobby_id"`
	TotalPlayers         int                       `json:"total_players"`
	ConfirmedCount       int                       `json:"confirmed_count"`
	PendingCount         int                       `json:"pending_count"`
	DeclinedCount        int                       `json:"declined_count"`
	ExpiredCount         int                       `json:"expired_count"`
	AllConfirmed         bool                      `json:"all_confirmed"`
	HasDeclinedOrExpired bool                      `json:"has_declined_or_expired"`
	Commitments          []CommitmentPlayerSummary `json:"commitments"`
}

type CommitmentConfirmResponse struct {
	Commitment map[string]string          `json:"commitment"`
	Summary    CommitmentSummaryResponse  `json:"summary"`
	AllReady   bool                       `json:"all_ready"`
}

type GameConnectionInfoResponse struct {
	LobbyID     string `json:"lobby_id"`
	MatchID     string `json:"match_id"`
	GameID      string `json:"game_id"`
	Region      string `json:"region"`
	ServerURL   string `json:"server_url,omitempty"`
	ServerIP    string `json:"server_ip,omitempty"`
	Port        int    `json:"port,omitempty"`
	Passcode    string `json:"passcode,omitempty"`
	QRCodeData  string `json:"qr_code_data,omitempty"`
	DeepLink    string `json:"deep_link,omitempty"`
	Instructions string `json:"instructions"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

func (ctrl *LobbyController) GetLobbyHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		playerID, ok := getAuthenticatedUserID(r.Context())
		if !ok {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		lobbyRepo, err := ctrl.resolveLobbyRepository()
		if err != nil {
			http.Error(w, `{"success":false,"error":"Failed to resolve lobby repository"}`, http.StatusInternalServerError)
			return
		}

		lobby, err := lobbyRepo.FindByID(r.Context(), lobbyID)
		if err != nil || lobby == nil {
			http.Error(w, `{"success":false,"error":"Lobby not found"}`, http.StatusNotFound)
			return
		}
		if !isLobbyParticipant(lobby, playerID) {
			http.Error(w, `{"success":false,"error":"Forbidden"}`, http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(GetLobbyResponse{Lobby: lobby})
		slog.InfoContext(apiContext, "lobby fetched", "lobby_id", lobbyID, "player_id", playerID)
	}
}

func (ctrl *LobbyController) ConfirmReadinessHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		playerID, ok := getAuthenticatedUserID(r.Context())
		if !ok {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		if err := ctrl.lobbyCommand.SetPlayerReady(r.Context(), matchmaking_in.SetPlayerReadyCommand{
			LobbyID:  lobbyID,
			PlayerID: playerID,
			IsReady:  true,
		}); err != nil {
			slog.ErrorContext(r.Context(), "failed to confirm readiness", "lobby_id", lobbyID, "player_id", playerID, "error", err)
			http.Error(w, `{"success":false,"error":"Failed to confirm readiness"}`, http.StatusBadRequest)
			return
		}

		lobby, err := ctrl.loadLobbyForParticipant(r.Context(), lobbyID, playerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		summary := buildCommitmentSummary(lobby)
		response := CommitmentConfirmResponse{
			Commitment: map[string]string{
				"id":           lobbyID.String() + ":" + playerID.String(),
				"lobby_id":     lobbyID.String(),
				"player_id":    playerID.String(),
				"status":       "confirmed",
				"responded_at": lobby.UpdatedAt.Format(time.RFC3339),
			},
			Summary:  summary,
			AllReady: summary.AllConfirmed,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

func (ctrl *LobbyController) DeclineReadinessHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		playerID, ok := getAuthenticatedUserID(r.Context())
		if !ok {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		lobbyRepo, err := ctrl.resolveLobbyRepository()
		if err != nil {
			http.Error(w, `{"success":false,"error":"Failed to resolve lobby repository"}`, http.StatusInternalServerError)
			return
		}

		lobby, err := ctrl.loadLobbyForParticipant(r.Context(), lobbyID, playerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		declinedPlayers := readStringSliceMetadata(lobby.Metadata, "declined_players")
		declinedPlayers = appendUniqueString(declinedPlayers, playerID.String())
		if lobby.Metadata == nil {
			lobby.Metadata = make(map[string]any)
		}
		lobby.Metadata["declined_players"] = declinedPlayers
		if updateErr := lobbyRepo.Update(r.Context(), lobby); updateErr != nil {
			http.Error(w, `{"success":false,"error":"Failed to persist decline state"}`, http.StatusInternalServerError)
			return
		}

		reason := "player_declined:" + playerID.String()
		if err := ctrl.lobbyCommand.CancelLobby(r.Context(), lobbyID, reason); err != nil {
			slog.ErrorContext(r.Context(), "failed to decline readiness", "lobby_id", lobbyID, "player_id", playerID, "error", err)
			http.Error(w, `{"success":false,"error":"Failed to decline readiness"}`, http.StatusBadRequest)
			return
		}

		updatedLobby, _ := lobbyRepo.FindByID(r.Context(), lobbyID)
		if updatedLobby == nil {
			updatedLobby = lobby
		}
		summary := buildCommitmentSummary(updatedLobby)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(summary)
		slog.InfoContext(apiContext, "player declined readiness", "lobby_id", lobbyID, "player_id", playerID)
	}
}

func (ctrl *LobbyController) GetCommitmentSummaryHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		playerID, ok := getAuthenticatedUserID(r.Context())
		if !ok {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		lobby, err := ctrl.loadLobbyForParticipant(r.Context(), lobbyID, playerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(buildCommitmentSummary(lobby))
	}
}

func (ctrl *LobbyController) GetGameConnectionInfoHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		playerID, ok := getAuthenticatedUserID(r.Context())
		if !ok {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			http.Error(w, "invalid lobby_id", http.StatusBadRequest)
			return
		}

		lobby, err := ctrl.loadLobbyForParticipant(r.Context(), lobbyID, playerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		summary := buildCommitmentSummary(lobby)
		playerStatus := summary.statusForPlayer(playerID.String())
		if playerStatus != "confirmed" && playerStatus != "expired" && lobby.Status != matchmaking_entities.LobbyStatusStarted {
			http.Error(w, `{"success":false,"error":"You must confirm readiness first"}`, http.StatusPreconditionFailed)
			return
		}
		if lobby.MatchID == nil {
			http.Error(w, `{"success":false,"error":"Game connection info not available yet"}`, http.StatusNotFound)
			return
		}

		deepLink := readStringMetadata(lobby.Metadata, "deep_link")
		passcode := readStringMetadata(lobby.Metadata, "match_code")
		if deepLink == "" {
			deepLink = "leetgaming://matches/" + lobby.MatchID.String() + "/join?code=" + strings.ToUpper(lobby.MatchID.String()[:8])
		}
		if passcode == "" {
			passcode = strings.ToUpper(lobby.MatchID.String()[:8])
		}

		response := GameConnectionInfoResponse{
			LobbyID:      lobby.ID.String(),
			MatchID:      lobby.MatchID.String(),
			GameID:       lobby.GameID,
			Region:       lobby.Region,
			ServerURL:    "https://join.leetgaming.pro/matches/" + lobby.MatchID.String(),
			Port:         27015,
			Passcode:     passcode,
			QRCodeData:   deepLink,
			DeepLink:     deepLink,
			Instructions: "Open the match link or scan the QR code to join the server.",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
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

// InviteToLobbyHandler handles POST /api/lobbies/{lobby_id}/invite
func (ctrl *LobbyController) InviteToLobbyHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Derive inviter ID from authenticated context
		ctx := r.Context()
		authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"success":false,"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}
		inviterID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
		if !ok || inviterID == uuid.Nil {
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
			InviteeID string `json:"invitee_id"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&reqBody); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		inviteeID, err := uuid.Parse(reqBody.InviteeID)
		if err != nil {
			http.Error(w, `{"success":false,"error":"Invalid invitee_id"}`, http.StatusBadRequest)
			return
		}

		cmd := matchmaking_in.InviteToLobbyCommand{
			LobbyID:   lobbyID,
			InviterID: inviterID,
			InviteeID: inviteeID,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, `{"success":false,"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		err = ctrl.lobbyCommand.InviteToLobby(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to invite to lobby", "lobby_id", lobbyID, "invitee", inviteeID, "error", err)
			http.Error(w, `{"success":false,"error":"Failed to send invite"}`, http.StatusBadRequest)
			return
		}

		response := LobbyActionResponse{
			Success: true,
			Message: "invite sent",
			LobbyID: lobbyID.String(),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		slog.InfoContext(apiContext, "lobby invite sent", "lobby_id", lobbyID, "invitee", inviteeID)
	}
}

func getAuthenticatedUserID(ctx context.Context) (uuid.UUID, bool) {
	authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
	if !ok || !authenticated {
		return uuid.Nil, false
	}
	playerID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
	if !ok || playerID == uuid.Nil {
		return uuid.Nil, false
	}
	return playerID, true
}

func (ctrl *LobbyController) resolveLobbyRepository() (matchmaking_out.LobbyRepository, error) {
	var lobbyRepo matchmaking_out.LobbyRepository
	if err := ctrl.container.Resolve(&lobbyRepo); err != nil {
		return nil, err
	}
	return lobbyRepo, nil
}

func (ctrl *LobbyController) loadLobbyForParticipant(ctx context.Context, lobbyID, playerID uuid.UUID) (*matchmaking_entities.MatchmakingLobby, error) {
	lobbyRepo, err := ctrl.resolveLobbyRepository()
	if err != nil {
		return nil, err
	}

	lobby, err := lobbyRepo.FindByID(ctx, lobbyID)
	if err != nil || lobby == nil {
		return nil, err
	}
	if !isLobbyParticipant(lobby, playerID) {
		return nil, fmt.Errorf("forbidden")
	}
	return lobby, nil
}

func isLobbyParticipant(lobby *matchmaking_entities.MatchmakingLobby, playerID uuid.UUID) bool {
	if lobby == nil {
		return false
	}
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil && *slot.PlayerID == playerID {
			return true
		}
	}
	return false
}

func buildCommitmentSummary(lobby *matchmaking_entities.MatchmakingLobby) CommitmentSummaryResponse {
	summary := CommitmentSummaryResponse{}
	if lobby == nil {
		return summary
	}

	expiresAt := lobby.UpdatedAt.Add(30 * time.Second).Format(time.RFC3339)
	if lobby.ReadyCheckEnd != nil {
		expiresAt = lobby.ReadyCheckEnd.Format(time.RFC3339)
	}

	declinedPlayers := readStringSliceMetadata(lobby.Metadata, "declined_players")
	declinedSet := make(map[string]bool, len(declinedPlayers))
	for _, playerID := range declinedPlayers {
		declinedSet[playerID] = true
	}

	summary.LobbyID = lobby.ID.String()
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID == nil {
			continue
		}

		status := "pending"
		switch {
		case declinedSet[slot.PlayerID.String()]:
			status = "declined"
			summary.DeclinedCount++
		case lobby.CancelReason == "ready_check_timeout" && !slot.IsReady:
			status = "expired"
			summary.ExpiredCount++
		case slot.IsReady:
			status = "confirmed"
			summary.ConfirmedCount++
		default:
			summary.PendingCount++
		}

		respondedAt := ""
		if status == "confirmed" || status == "declined" {
			respondedAt = lobby.UpdatedAt.Format(time.RFC3339)
		}

		summary.Commitments = append(summary.Commitments, CommitmentPlayerSummary{
			PlayerID:    slot.PlayerID.String(),
			Status:      status,
			RespondedAt: respondedAt,
			ExpiresAt:   expiresAt,
		})
		summary.TotalPlayers++
	}

	summary.AllConfirmed = summary.TotalPlayers > 0 && summary.ConfirmedCount == summary.TotalPlayers
	summary.HasDeclinedOrExpired = summary.DeclinedCount > 0 || summary.ExpiredCount > 0
	return summary
}

func (s CommitmentSummaryResponse) statusForPlayer(playerID string) string {
	for _, commitment := range s.Commitments {
		if commitment.PlayerID == playerID {
			return commitment.Status
		}
	}
	return ""
}

func readStringMetadata(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	stringValue, ok := value.(string)
	if ok {
		return stringValue
	}
	return ""
}

func readStringSliceMetadata(metadata map[string]any, key string) []string {
	if metadata == nil {
		return nil
	}
	value, ok := metadata[key]
	if !ok {
		return nil
	}

	result := make([]string, 0)
	switch typed := value.(type) {
	case []string:
		return append(result, typed...)
	case []any:
		for _, item := range typed {
			if stringValue, ok := item.(string); ok {
				result = append(result, stringValue)
			}
		}
	}

	return result
}

func appendUniqueString(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}
