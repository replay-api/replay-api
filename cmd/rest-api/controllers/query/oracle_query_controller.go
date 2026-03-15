package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// OracleQueryController handles read operations for oracle results
type OracleQueryController struct {
	queryHandler oracle_in.OracleQueryHandler
}

// NewOracleQueryController creates a new controller resolved from the DI container
func NewOracleQueryController(c container.Container) *OracleQueryController {
	var queryHandler oracle_in.OracleQueryHandler

	if err := c.Resolve(&queryHandler); err != nil {
		slog.Warn("OracleQueryHandler not available", "error", err)
	}

	return &OracleQueryController{
		queryHandler: queryHandler,
	}
}

// GetOracleResultHandler handles GET /oracle/results/{id}
func (c *OracleQueryController) GetOracleResultHandler(w http.ResponseWriter, r *http.Request) {
	if c.queryHandler == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid oracle result id", http.StatusBadRequest)
		return
	}

	query := oracle_in.GetOracleResultQuery{OracleResultID: id}
	result, err := c.queryHandler.GetOracleResult(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get oracle result", "id", id, "error", err)
		http.Error(w, "oracle result not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetOracleResultByMatchHandler handles GET /oracle/results/by-match/{matchId}
func (c *OracleQueryController) GetOracleResultByMatchHandler(w http.ResponseWriter, r *http.Request) {
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

	query := oracle_in.GetOracleResultByMatchIDQuery{MatchID: matchID}
	result, err := c.queryHandler.GetOracleResultByMatchID(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get oracle result by match", "matchId", matchID, "error", err)
		http.Error(w, "oracle result not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListOracleResultsHandler handles GET /oracle/results
func (c *OracleQueryController) ListOracleResultsHandler(w http.ResponseWriter, r *http.Request) {
	if c.queryHandler == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	queryParams := r.URL.Query()

	page := 0
	if p := queryParams.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed >= 0 {
			page = parsed
		}
	}

	pageSize := 20
	if ps := queryParams.Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	query := oracle_in.ListOracleResultsQuery{
		Page:     page,
		PageSize: pageSize,
	}

	if gameID := queryParams.Get("game_id"); gameID != "" {
		gid := replay_common.GameIDKey(gameID)
		query.GameID = &gid
	}
	if status := queryParams.Get("status"); status != "" {
		s := oracle_vo.OracleStatus(status)
		query.Status = &s
	}

	result, err := c.queryHandler.ListOracleResults(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list oracle results", "error", err)
		http.Error(w, "failed to list oracle results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetSubmissionsHandler handles GET /oracle/results/{id}/submissions
func (c *OracleQueryController) GetSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	if c.queryHandler == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid oracle result id", http.StatusBadRequest)
		return
	}

	query := oracle_in.GetSubmissionsQuery{OracleResultID: id}
	result, err := c.queryHandler.GetSubmissionsForResult(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get submissions", "id", id, "error", err)
		http.Error(w, "submissions not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetPublicationStatusHandler handles GET /oracle/results/{id}/publications
func (c *OracleQueryController) GetPublicationStatusHandler(w http.ResponseWriter, r *http.Request) {
	if c.queryHandler == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid oracle result id", http.StatusBadRequest)
		return
	}

	query := oracle_in.GetPublicationStatusQuery{OracleResultID: id}
	result, err := c.queryHandler.GetPublicationStatus(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get publication status", "id", id, "error", err)
		http.Error(w, "publication status not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
