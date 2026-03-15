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
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// MatchResultCommandController handles write operations for match results
type MatchResultCommandController struct {
	commandHandler scores_in.MatchResultCommandHandler
}

// NewMatchResultCommandController creates a new controller resolving deps from DI container
func NewMatchResultCommandController(c container.Container) *MatchResultCommandController {
	var commandHandler scores_in.MatchResultCommandHandler

	if err := c.Resolve(&commandHandler); err != nil {
		slog.Warn("MatchResultCommandHandler not available", "error", err)
	}

	return &MatchResultCommandController{
		commandHandler: commandHandler,
	}
}

// --- Request DTOs ---

type SubmitMatchResultRequest struct {
	MatchID              string             `json:"match_id"`
	TournamentID         *string            `json:"tournament_id,omitempty"`
	MatchmakingSessionID *string            `json:"matchmaking_session_id,omitempty"`
	GameID               string             `json:"game_id"`
	MapName              string             `json:"map_name"`
	Mode                 string             `json:"mode"`
	Source               string             `json:"source"`
	TeamResults          []TeamResultReq    `json:"team_results"`
	PlayerResults        []PlayerResultReq  `json:"player_results"`
	PlayedAt             string             `json:"played_at"`
	DurationMinutes      int                `json:"duration_minutes"`
	RoundsPlayed         int                `json:"rounds_played"`
}

type TeamResultReq struct {
	TeamID   string   `json:"team_id"`
	TeamName string   `json:"team_name"`
	Score    int      `json:"score"`
	Players  []string `json:"players"`
}

type PlayerResultReq struct {
	PlayerID string                 `json:"player_id"`
	TeamID   string                 `json:"team_id"`
	Score    int                    `json:"score"`
	Kills    int                    `json:"kills"`
	Deaths   int                    `json:"deaths"`
	Assists  int                    `json:"assists"`
	Rating   float64                `json:"rating"`
	IsMVP    bool                   `json:"is_mvp"`
	Stats    map[string]interface{} `json:"stats,omitempty"`
}

type DisputeMatchResultRequest struct {
	Reason string `json:"reason"`
}

type ConciliateMatchResultRequest struct {
	Notes               string          `json:"notes"`
	AdjustedTeamResults []TeamResultReq `json:"adjusted_team_results,omitempty"`
}

type VerifyMatchResultRequest struct {
	VerificationMethod string `json:"verification_method"`
}

type CancelMatchResultRequest struct {
	Reason string `json:"reason"`
}

// --- Handlers ---

// SubmitMatchResultHandler handles POST /scores/match-results
func (c *MatchResultCommandController) SubmitMatchResultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		// Limit request body to 1MB to prevent abuse
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req SubmitMatchResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		matchID, err := uuid.Parse(req.MatchID)
		if err != nil {
			http.Error(w, "invalid match_id", http.StatusBadRequest)
			return
		}

		playedAt, err := time.Parse(time.RFC3339, req.PlayedAt)
		if err != nil {
			http.Error(w, "invalid played_at format (expected RFC3339)", http.StatusBadRequest)
			return
		}

		teamResults, err := parseTeamResults(req.TeamResults)
		if err != nil {
			http.Error(w, "invalid team results format", http.StatusBadRequest)
			return
		}

		playerResults, err := parsePlayerResults(req.PlayerResults)
		if err != nil {
			http.Error(w, "invalid player results format", http.StatusBadRequest)
			return
		}

		cmd := scores_in.SubmitMatchResultCommand{
			MatchID:       matchID,
			GameID:        replay_common.GameIDKey(req.GameID),
			MapName:       req.MapName,
			Mode:          req.Mode,
			Source:        scores_vo.ScoreSource(req.Source),
			TeamResults:   teamResults,
			PlayerResults: playerResults,
			PlayedAt:      playedAt,
			Duration:      time.Duration(req.DurationMinutes) * time.Minute,
			RoundsPlayed:  req.RoundsPlayed,
		}

		if req.TournamentID != nil {
			tid, err := uuid.Parse(*req.TournamentID)
			if err != nil {
				http.Error(w, "invalid tournament_id", http.StatusBadRequest)
				return
			}
			cmd.TournamentID = &tid
		}

		if req.MatchmakingSessionID != nil {
			sid, err := uuid.Parse(*req.MatchmakingSessionID)
			if err != nil {
				http.Error(w, "invalid matchmaking_session_id", http.StatusBadRequest)
				return
			}
			cmd.MatchmakingSessionID = &sid
		}

		result, err := c.commandHandler.SubmitMatchResult(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to submit match result", "error", err)
			if isForbiddenError(err) {
				writeJSONError(w, sanitizeForbiddenError(err), http.StatusForbidden)
				return
			}
			writeJSONError(w, "Failed to submit match result", http.StatusBadRequest)
			return
		}

		dto := scores_in.MatchResultToDTO(result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(dto)
	}
}

// VerifyMatchResultHandler handles PUT /scores/match-results/{id}/verify
func (c *MatchResultCommandController) VerifyMatchResultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid match result id", http.StatusBadRequest)
			return
		}

		var req VerifyMatchResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := scores_in.VerifyMatchResultCommand{
			MatchResultID:      id,
			VerificationMethod: scores_vo.VerificationMethod(req.VerificationMethod),
		}

		if err := c.commandHandler.VerifyMatchResult(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to verify match result", "error", err)
			if isForbiddenError(err) {
				writeJSONError(w, sanitizeForbiddenError(err), http.StatusForbidden)
				return
			}
			writeJSONError(w, "Failed to verify result", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "verified"})
	}
}

// DisputeMatchResultHandler handles PUT /scores/match-results/{id}/dispute
func (c *MatchResultCommandController) DisputeMatchResultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid match result id", http.StatusBadRequest)
			return
		}

		var req DisputeMatchResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := scores_in.DisputeMatchResultCommand{
			MatchResultID: id,
			Reason:        req.Reason,
		}

		if err := c.commandHandler.DisputeMatchResult(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to dispute match result", "error", err)
			if isForbiddenError(err) {
				writeJSONError(w, sanitizeForbiddenError(err), http.StatusForbidden)
				return
			}
			writeJSONError(w, "Failed to submit dispute", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "disputed"})
	}
}

// ConciliateMatchResultHandler handles PUT /scores/match-results/{id}/conciliate
func (c *MatchResultCommandController) ConciliateMatchResultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid match result id", http.StatusBadRequest)
			return
		}

		var req ConciliateMatchResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		adjustedTeams, err := parseTeamResults(req.AdjustedTeamResults)
		if err != nil {
			http.Error(w, "invalid adjusted team results format", http.StatusBadRequest)
			return
		}

		cmd := scores_in.ConciliateMatchResultCommand{
			MatchResultID:       id,
			Notes:               req.Notes,
			AdjustedTeamResults: adjustedTeams,
		}

		if err := c.commandHandler.ConciliateMatchResult(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to conciliate match result", "error", err)
			if isForbiddenError(err) {
				writeJSONError(w, sanitizeForbiddenError(err), http.StatusForbidden)
				return
			}
			writeJSONError(w, "Failed to conciliate", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "conciliated"})
	}
}

// FinalizeMatchResultHandler handles PUT /scores/match-results/{id}/finalize
func (c *MatchResultCommandController) FinalizeMatchResultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid match result id", http.StatusBadRequest)
			return
		}

		cmd := scores_in.FinalizeMatchResultCommand{
			MatchResultID: id,
		}

		if err := c.commandHandler.FinalizeMatchResult(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to finalize match result", "error", err)
			if isForbiddenError(err) {
				writeJSONError(w, sanitizeForbiddenError(err), http.StatusForbidden)
				return
			}
			writeJSONError(w, "Failed to finalize result", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "finalized"})
	}
}

// CancelMatchResultHandler handles PUT /scores/match-results/{id}/cancel
func (c *MatchResultCommandController) CancelMatchResultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.commandHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid match result id", http.StatusBadRequest)
			return
		}

		var req CancelMatchResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		cmd := scores_in.CancelMatchResultCommand{
			MatchResultID: id,
			Reason:        req.Reason,
		}

		if err := c.commandHandler.CancelMatchResult(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to cancel match result", "error", err)
			if isForbiddenError(err) {
				writeJSONError(w, sanitizeForbiddenError(err), http.StatusForbidden)
				return
			}
			writeJSONError(w, "Failed to cancel match", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
	}
}

// --- Internal helpers ---

func parseTeamResults(reqs []TeamResultReq) ([]scores_entities.TeamResult, error) {
	results := make([]scores_entities.TeamResult, len(reqs))
	for i, req := range reqs {
		teamID, err := uuid.Parse(req.TeamID)
		if err != nil {
			return nil, fmt.Errorf("invalid team_id: %s", req.TeamID)
		}

		players := make([]uuid.UUID, len(req.Players))
		for j, p := range req.Players {
			pid, err := uuid.Parse(p)
			if err != nil {
				return nil, fmt.Errorf("invalid player_id in team %s: %s", req.TeamID, p)
			}
			players[j] = pid
		}

		results[i] = scores_entities.TeamResult{
			TeamID:   teamID,
			TeamName: req.TeamName,
			Score:    req.Score,
			Players:  players,
		}
	}
	return results, nil
}

func parsePlayerResults(reqs []PlayerResultReq) ([]scores_entities.PlayerResult, error) {
	results := make([]scores_entities.PlayerResult, len(reqs))
	for i, req := range reqs {
		playerID, err := uuid.Parse(req.PlayerID)
		if err != nil {
			return nil, fmt.Errorf("invalid player_id: %s", req.PlayerID)
		}

		teamID, err := uuid.Parse(req.TeamID)
		if err != nil {
			return nil, fmt.Errorf("invalid team_id: %s", req.TeamID)
		}

		results[i] = scores_entities.PlayerResult{
			PlayerID: playerID,
			TeamID:   teamID,
			Score:    req.Score,
			Kills:    req.Kills,
			Deaths:   req.Deaths,
			Assists:  req.Assists,
			Rating:   req.Rating,
			IsMVP:    req.IsMVP,
			Stats:    req.Stats,
		}
	}
	return results, nil
}

// isForbiddenError checks if the error is a permission/authorization error
func isForbiddenError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "forbidden:") || strings.Contains(msg, "insufficient permissions")
}

// sanitizeForbiddenError returns a safe error message for 403 responses
// without leaking internal details like user IDs or tournament IDs
func sanitizeForbiddenError(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "forbidden:") {
		// Extract the part after "forbidden:" which is the user-facing message
		parts := strings.SplitN(msg, "forbidden:", 2)
		if len(parts) == 2 {
			return "Forbidden:" + parts[1]
		}
	}
	return "Forbidden: insufficient permissions"
}

// writeJSONError writes a standardized JSON error response
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
		"code":    http.StatusText(statusCode),
	})
}
