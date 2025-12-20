package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
	challenge_in "github.com/replay-api/replay-api/pkg/domain/challenge/ports/in"
)

// ChallengeController handles challenge-related HTTP requests
type ChallengeController struct {
	container              container.Container
	createChallengeHandler challenge_in.CreateChallengeCommandHandler
	addEvidenceHandler     challenge_in.AddEvidenceCommandHandler
	voteHandler            challenge_in.VoteOnChallengeCommandHandler
	resolveHandler         challenge_in.ResolveChallengeCommandHandler
	cancelHandler          challenge_in.CancelChallengeCommandHandler
	queryService           challenge_in.ChallengeQueryService
}

// NewChallengeController creates a new challenge controller
func NewChallengeController(container container.Container) *ChallengeController {
	ctrl := &ChallengeController{container: container}

	// Resolve command handlers
	if err := container.Resolve(&ctrl.createChallengeHandler); err != nil {
		slog.Error("Failed to resolve CreateChallengeCommandHandler", "err", err)
	}
	if err := container.Resolve(&ctrl.addEvidenceHandler); err != nil {
		slog.Error("Failed to resolve AddEvidenceCommandHandler", "err", err)
	}
	if err := container.Resolve(&ctrl.voteHandler); err != nil {
		slog.Error("Failed to resolve VoteOnChallengeCommandHandler", "err", err)
	}
	if err := container.Resolve(&ctrl.resolveHandler); err != nil {
		slog.Error("Failed to resolve ResolveChallengeCommandHandler", "err", err)
	}
	if err := container.Resolve(&ctrl.cancelHandler); err != nil {
		slog.Error("Failed to resolve CancelChallengeCommandHandler", "err", err)
	}

	// Resolve query service
	if err := container.Resolve(&ctrl.queryService); err != nil {
		slog.Error("Failed to resolve ChallengeQueryService", "err", err)
	}

	return ctrl
}

// Request/Response DTOs
type CreateChallengeRequest struct {
	MatchID          string  `json:"match_id"`
	RoundNumber      *int    `json:"round_number,omitempty"`
	ChallengerTeamID *string `json:"challenger_team_id,omitempty"`
	GameID           string  `json:"game_id"`
	LobbyID          *string `json:"lobby_id,omitempty"`
	TournamentID     *string `json:"tournament_id,omitempty"`
	Type             string  `json:"type"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Priority         string  `json:"priority"`
}

type AddEvidenceRequest struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	Description string `json:"description"`
	StartTick   *int64 `json:"start_tick,omitempty"`
	EndTick     *int64 `json:"end_tick,omitempty"`
}

type VoteRequest struct {
	VoteType string `json:"vote_type"`
	Reason   string `json:"reason,omitempty"`
}

type ResolveRequest struct {
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Notes      string `json:"notes"`
}

type CancelRequest struct {
	Reason string `json:"reason"`
}

type ChallengeResponse struct {
	ID               string  `json:"id"`
	MatchID          string  `json:"match_id"`
	RoundNumber      *int    `json:"round_number,omitempty"`
	ChallengerID     string  `json:"challenger_id"`
	ChallengerTeamID *string `json:"challenger_team_id,omitempty"`
	GameID           string  `json:"game_id"`
	Type             string  `json:"type"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Priority         string  `json:"priority"`
	Status           string  `json:"status"`
	Resolution       string  `json:"resolution"`
	EvidenceCount    int     `json:"evidence_count"`
	VoteCount        int     `json:"vote_count"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func toChallengeResponse(c *challenge_entities.Challenge) ChallengeResponse {
	resp := ChallengeResponse{
		ID:            c.ID.String(),
		MatchID:       c.MatchID.String(),
		RoundNumber:   c.RoundNumber,
		ChallengerID:  c.ChallengerID.String(),
		GameID:        c.GameID,
		Type:          string(c.Type),
		Title:         c.Title,
		Description:   c.Description,
		Priority:      string(c.Priority),
		Status:        string(c.Status),
		Resolution:    string(c.Resolution),
		EvidenceCount: len(c.Evidence),
		VoteCount:     len(c.Votes),
		CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if c.ChallengerTeamID != nil {
		teamID := c.ChallengerTeamID.String()
		resp.ChallengerTeamID = &teamID
	}
	return resp
}

// CreateChallengeHandler handles POST /api/challenges
func (ctrl *ChallengeController) CreateChallengeHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var req CreateChallengeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.ErrorContext(apiContext, "failed to decode create challenge request", "error", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Get authenticated user
		ctx := r.Context()
		resourceOwner := common.GetResourceOwner(ctx)
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		matchID, err := uuid.Parse(req.MatchID)
		if err != nil {
			http.Error(w, "invalid match_id", http.StatusBadRequest)
			return
		}

		// Build command
		cmd := challenge_in.CreateChallengeCommand{
			MatchID:      matchID,
			ChallengerID: resourceOwner.UserID,
			GameID:       req.GameID,
			Type:         challenge_entities.ChallengeType(req.Type),
			Title:        req.Title,
			Description:  req.Description,
			Priority:     challenge_entities.ChallengePriority(req.Priority),
			RoundNumber:  req.RoundNumber,
		}

		// Optional fields
		if req.ChallengerTeamID != nil {
			teamID, err := uuid.Parse(*req.ChallengerTeamID)
			if err == nil {
				cmd.ChallengerTeamID = &teamID
			}
		}
		if req.LobbyID != nil {
			lobbyID, err := uuid.Parse(*req.LobbyID)
			if err == nil {
				cmd.LobbyID = &lobbyID
			}
		}
		if req.TournamentID != nil {
			tournamentID, err := uuid.Parse(*req.TournamentID)
			if err == nil {
				cmd.TournamentID = &tournamentID
			}
		}

		// Defaults
		if cmd.Priority == "" {
			cmd.Priority = challenge_entities.ChallengePriorityNormal
		}
		if cmd.GameID == "" {
			cmd.GameID = "cs2"
		}

		// Execute command
		if ctrl.createChallengeHandler == nil {
			http.Error(w, "challenge handler not available", http.StatusServiceUnavailable)
			return
		}

		challenge, err := ctrl.createChallengeHandler.Exec(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create challenge", "err", err)
			statusCode := http.StatusBadRequest
			if err.Error() == "Unauthorized" {
				statusCode = http.StatusUnauthorized
			} else if err.Error() == "Forbidden" {
				statusCode = http.StatusForbidden
			}
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(toChallengeResponse(challenge))
		slog.InfoContext(ctx, "challenge created successfully", "challenge_id", challenge.ID)
	}
}

// GetChallengeHandler handles GET /api/challenges/{id}
func (ctrl *ChallengeController) GetChallengeHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		challengeID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid challenge id", http.StatusBadRequest)
			return
		}

		if ctrl.queryService == nil {
			http.Error(w, "query service not available", http.StatusServiceUnavailable)
			return
		}

		challenge, err := ctrl.queryService.GetByID(r.Context(), challenge_in.GetChallengeByIDQuery{
			ChallengeID: challengeID,
		})
		if err != nil {
			slog.ErrorContext(apiContext, "Failed to get challenge", "err", err)
			http.Error(w, "failed to get challenge", http.StatusInternalServerError)
			return
		}

		if challenge == nil {
			http.Error(w, "challenge not found", http.StatusNotFound)
			return
		}

		_ = json.NewEncoder(w).Encode(toChallengeResponse(challenge))
	}
}

// GetChallengesByMatchHandler handles GET /api/matches/{match_id}/challenges
func (ctrl *ChallengeController) GetChallengesByMatchHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		matchID, err := uuid.Parse(vars["match_id"])
		if err != nil {
			http.Error(w, "invalid match id", http.StatusBadRequest)
			return
		}

		if ctrl.queryService == nil {
			http.Error(w, "query service not available", http.StatusServiceUnavailable)
			return
		}

		challenges, err := ctrl.queryService.GetByMatch(r.Context(), challenge_in.GetChallengesByMatchQuery{
			MatchID: matchID,
		})
		if err != nil {
			slog.ErrorContext(apiContext, "Failed to get challenges", "err", err)
			http.Error(w, "failed to get challenges", http.StatusInternalServerError)
			return
		}

		var responses []ChallengeResponse
		for _, c := range challenges {
			responses = append(responses, toChallengeResponse(c))
		}

		_ = json.NewEncoder(w).Encode(responses)
	}
}

// AddEvidenceHandler handles POST /api/challenges/{id}/evidence
func (ctrl *ChallengeController) AddEvidenceHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		challengeID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid challenge id", http.StatusBadRequest)
			return
		}

		var req AddEvidenceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := challenge_in.AddEvidenceCommand{
			ChallengeID: challengeID,
			Type:        req.Type,
			URL:         req.URL,
			Description: req.Description,
			StartTick:   req.StartTick,
			EndTick:     req.EndTick,
		}

		if ctrl.addEvidenceHandler == nil {
			http.Error(w, "evidence handler not available", http.StatusServiceUnavailable)
			return
		}

		challenge, err := ctrl.addEvidenceHandler.Exec(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(apiContext, "Failed to add evidence", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(toChallengeResponse(challenge))
	}
}

// VoteHandler handles POST /api/challenges/{id}/vote
func (ctrl *ChallengeController) VoteHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		challengeID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid challenge id", http.StatusBadRequest)
			return
		}

		// Get authenticated user
		ctx := r.Context()
		resourceOwner := common.GetResourceOwner(ctx)
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req VoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := challenge_in.VoteOnChallengeCommand{
			ChallengeID: challengeID,
			PlayerID:    resourceOwner.UserID,
			VoteType:    req.VoteType,
			Reason:      req.Reason,
		}

		if ctrl.voteHandler == nil {
			http.Error(w, "vote handler not available", http.StatusServiceUnavailable)
			return
		}

		challenge, err := ctrl.voteHandler.Exec(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to vote", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(toChallengeResponse(challenge))
	}
}

// ResolveHandler handles POST /api/challenges/{id}/resolve
func (ctrl *ChallengeController) ResolveHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		challengeID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid challenge id", http.StatusBadRequest)
			return
		}

		// Get authenticated user (admin)
		ctx := r.Context()
		resourceOwner := common.GetResourceOwner(ctx)
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req ResolveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := challenge_in.ResolveChallengeCommand{
			ChallengeID: challengeID,
			AdminID:     resourceOwner.UserID,
			Decision:    req.Decision,
			Resolution:  challenge_entities.ChallengeResolution(req.Resolution),
			Notes:       req.Notes,
		}

		if ctrl.resolveHandler == nil {
			http.Error(w, "resolve handler not available", http.StatusServiceUnavailable)
			return
		}

		challenge, err := ctrl.resolveHandler.Exec(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to resolve challenge", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(toChallengeResponse(challenge))
	}
}

// CancelHandler handles DELETE /api/challenges/{id}
func (ctrl *ChallengeController) CancelHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		challengeID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid challenge id", http.StatusBadRequest)
			return
		}

		// Get authenticated user
		ctx := r.Context()
		resourceOwner := common.GetResourceOwner(ctx)
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req CancelRequest
		_ = json.NewDecoder(r.Body).Decode(&req) // Optional body

		cmd := challenge_in.CancelChallengeCommand{
			ChallengeID: challengeID,
			CancellerID: resourceOwner.UserID,
			Reason:      req.Reason,
		}

		if ctrl.cancelHandler == nil {
			http.Error(w, "cancel handler not available", http.StatusServiceUnavailable)
			return
		}

		challenge, err := ctrl.cancelHandler.Exec(ctx, cmd)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to cancel challenge", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(toChallengeResponse(challenge))
	}
}

// GetPendingChallengesHandler handles GET /api/challenges/pending
func (ctrl *ChallengeController) GetPendingChallengesHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if ctrl.queryService == nil {
			http.Error(w, "query service not available", http.StatusServiceUnavailable)
			return
		}

		// Parse query params
		query := challenge_in.GetPendingChallengesQuery{
			Limit: 50,
		}

		gameID := r.URL.Query().Get("game_id")
		if gameID != "" {
			query.GameID = &gameID
		}

		priorityStr := r.URL.Query().Get("priority")
		if priorityStr != "" {
			priority := challenge_entities.ChallengePriority(priorityStr)
			query.Priority = &priority
		}

		challenges, err := ctrl.queryService.GetPendingChallenges(r.Context(), query)
		if err != nil {
			slog.ErrorContext(apiContext, "Failed to get pending challenges", "err", err)
			http.Error(w, "failed to get pending challenges", http.StatusInternalServerError)
			return
		}

		var responses []ChallengeResponse
		for _, c := range challenges {
			responses = append(responses, toChallengeResponse(c))
		}

		_ = json.NewEncoder(w).Encode(responses)
	}
}

