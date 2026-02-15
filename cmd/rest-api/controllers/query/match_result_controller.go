package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// MatchResultQueryController handles read operations for match results
type MatchResultQueryController struct {
	queryHandler scores_in.MatchResultQueryHandler
}

// NewMatchResultQueryController creates a new controller resolved from the DI container
func NewMatchResultQueryController(c container.Container) *MatchResultQueryController {
	var queryHandler scores_in.MatchResultQueryHandler

	if err := c.Resolve(&queryHandler); err != nil {
		slog.Warn("MatchResultQueryHandler not available", "error", err)
	}

	return &MatchResultQueryController{
		queryHandler: queryHandler,
	}
}

// GetMatchResultHandler handles GET /scores/match-results/{id}
func (c *MatchResultQueryController) GetMatchResultHandler(w http.ResponseWriter, r *http.Request) {
	if c.queryHandler == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid match result id", http.StatusBadRequest)
		return
	}

	query := scores_in.GetMatchResultQuery{MatchResultID: id}
	result, err := c.queryHandler.GetMatchResult(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get match result", "id", id, "error", err)
		http.Error(w, "match result not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetMatchResultByMatchHandler handles GET /scores/match-results/by-match/{matchId}
func (c *MatchResultQueryController) GetMatchResultByMatchHandler(w http.ResponseWriter, r *http.Request) {
	if c.queryHandler == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	matchID, err := uuid.Parse(vars["matchId"])
	if err != nil {
		http.Error(w, "invalid match id", http.StatusBadRequest)
		return
	}

	query := scores_in.GetMatchResultByMatchIDQuery{MatchID: matchID}
	result, err := c.queryHandler.GetMatchResultByMatchID(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get match result by matchID", "match_id", matchID, "error", err)
		http.Error(w, "match result not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListMatchResultsHandler handles GET /scores/match-results
func (c *MatchResultQueryController) ListMatchResultsHandler(w http.ResponseWriter, r *http.Request) {
	if c.queryHandler == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Parse query parameters
	query := scores_in.ListMatchResultsQuery{
		Page:     0,
		PageSize: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p >= 0 {
			query.Page = p
		}
	}
	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			query.PageSize = ps
		}
	}
	if gameID := r.URL.Query().Get("game_id"); gameID != "" {
		gid := replay_common.GameIDKey(gameID)
		query.GameID = &gid
	}
	if status := r.URL.Query().Get("status"); status != "" {
		s := scores_vo.ResultStatus(status)
		query.Status = &s
	}
	if tournamentID := r.URL.Query().Get("tournament_id"); tournamentID != "" {
		if tid, err := uuid.Parse(tournamentID); err == nil {
			query.TournamentID = &tid
		}
	}
	if sessionID := r.URL.Query().Get("matchmaking_session_id"); sessionID != "" {
		if sid, err := uuid.Parse(sessionID); err == nil {
			query.MatchmakingSessionID = &sid
		}
	}
	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		if pid, err := uuid.Parse(playerID); err == nil {
			query.PlayerID = &pid
		}
	}
	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		if tid, err := uuid.Parse(teamID); err == nil {
			query.TeamID = &tid
		}
	}

	result, err := c.queryHandler.ListMatchResults(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list match results", "error", err)
		http.Error(w, "failed to list match results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetTournamentResultsHandler handles GET /scores/tournaments/{tournamentId}/results
func (c *MatchResultQueryController) GetTournamentResultsHandler(w http.ResponseWriter, r *http.Request) {
	if c.queryHandler == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	tournamentID, err := uuid.Parse(vars["tournamentId"])
	if err != nil {
		http.Error(w, "invalid tournament id", http.StatusBadRequest)
		return
	}

	query := scores_in.GetTournamentResultsQuery{TournamentID: tournamentID}
	result, err := c.queryHandler.GetMatchResultsByTournament(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get tournament results", "tournament_id", tournamentID, "error", err)
		http.Error(w, "failed to get tournament results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
