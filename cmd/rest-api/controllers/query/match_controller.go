package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

type MatchQueryController struct {
	controllers.DefaultSearchController[replay_entity.Match]
	matchReader  replay_in.MatchReader
	eventReader  replay_in.EventReader
	teamReader   replay_in.TeamReader
	roundReader  replay_in.RoundReader
	squadReader  squad_in.SquadReader
}

func NewMatchQueryController(c container.Container) *MatchQueryController {
	var matchReader replay_in.MatchReader
	var eventReader replay_in.EventReader
	var teamReader replay_in.TeamReader
	var roundReader replay_in.RoundReader
	var squadReader squad_in.SquadReader

	if err := c.Resolve(&matchReader); err != nil {
		slog.Warn("MatchReader not available - match queries will be disabled", "error", err)
	}
	if err := c.Resolve(&eventReader); err != nil {
		slog.Warn("EventReader not available", "error", err)
	}
	if err := c.Resolve(&teamReader); err != nil {
		slog.Warn("TeamReader not available", "error", err)
	}
	if err := c.Resolve(&roundReader); err != nil {
		slog.Warn("RoundReader not available", "error", err)
	}
	if err := c.Resolve(&squadReader); err != nil {
		slog.Warn("SquadReader not available", "error", err)
	}

	baseController := controllers.NewDefaultSearchController(matchReader)

	return &MatchQueryController{
		DefaultSearchController: *baseController,
		matchReader:             matchReader,
		eventReader:             eventReader,
		teamReader:              teamReader,
		roundReader:             roundReader,
		squadReader:             squadReader,
	}
}

// GetMatchDetailHandler handles GET /games/{game_id}/match/{match_id}
func (c *MatchQueryController) GetMatchDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["match_id"]
	gameID := vars["game_id"]

	if matchID == "" || gameID == "" {
		slog.Error("GetMatchDetailHandler: missing match_id or game_id")
		http.Error(w, "match_id and game_id are required", http.StatusBadRequest)
		return
	}

	slog.Info("Fetching match detail", "match_id", matchID, "game_id", gameID)

	// Search for the specific match
	searchParams := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{Field: "id", Values: []interface{}{matchID}, Operator: shared.EqualsOperator},
						{Field: "game_id", Values: []interface{}{gameID}, Operator: shared.EqualsOperator},
					},
				},
			},
			AggregationClause: shared.AndAggregationClause,
		},
	}

	resultOptions := shared.SearchResultOptions{
		Limit: 1,
		Skip:  0,
	}

	compiledSearch, err := c.matchReader.Compile(r.Context(), searchParams, resultOptions)
	if err != nil {
		slog.Error("GetMatchDetailHandler: error compiling search", "error", err)
		http.Error(w, "invalid search parameters", http.StatusBadRequest)
		return
	}

	results, err := c.matchReader.Search(r.Context(), *compiledSearch)
	if err != nil {
		slog.Error("GetMatchDetailHandler: error searching match", "error", err)
		http.Error(w, "error fetching match", http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		slog.Warn("GetMatchDetailHandler: match not found", "match_id", matchID)
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	match := results[0]

	// Optionally fetch related data (events, rounds)
	matchDetail := map[string]any{
		"match": match,
		"metadata": map[string]any{
			"has_events": len(match.Events) > 0,
			"has_rounds": len(match.Scoreboard.TeamScoreboards) > 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(matchDetail)
}

// GetPlayerMatchHistoryHandler handles GET /matches/player/{player_id}
// Returns paginated match history for a specific player
func (c *MatchQueryController) GetPlayerMatchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerIDStr := vars["player_id"]

	if playerIDStr == "" {
		slog.Error("GetPlayerMatchHistoryHandler: missing player_id")
		http.Error(w, "player_id is required", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		slog.Error("GetPlayerMatchHistoryHandler: invalid player_id", "player_id", playerIDStr, "error", err)
		http.Error(w, "invalid player_id format", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	slog.Info("Fetching player match history", "player_id", playerID, "limit", limit, "offset", offset)

	// Search for matches containing this player in the scoreboard
	searchParams := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field:    "scoreboard.team_scoreboards.players._id",
							Values:   []interface{}{playerID.String()},
							Operator: shared.EqualsOperator,
						},
					},
				},
			},
			AggregationClause: shared.AndAggregationClause,
		},
	}

	resultOptions := shared.SearchResultOptions{
		Limit: uint(limit),  // #nosec G115 - limit is validated to be non-negative
		Skip:  uint(offset), // #nosec G115 - offset is validated to be non-negative
	}

	compiledSearch, err := c.matchReader.Compile(r.Context(), searchParams, resultOptions)
	if err != nil {
		slog.Error("GetPlayerMatchHistoryHandler: error compiling search", "error", err)
		http.Error(w, "invalid search parameters", http.StatusBadRequest)
		return
	}

	results, err := c.matchReader.Search(r.Context(), *compiledSearch)
	if err != nil {
		slog.Error("GetPlayerMatchHistoryHandler: error searching matches", "error", err)
		http.Error(w, "error fetching match history", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"data":        results,
		"player_id":   playerID,
		"total":       len(results),
		"limit":       limit,
		"offset":      offset,
		"next_offset": nil,
	}

	if len(results) == limit {
		response["next_offset"] = offset + limit
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetSquadMatchHistoryHandler handles GET /matches/squad/{squad_id}
// Returns paginated match history for all members of a squad
func (c *MatchQueryController) GetSquadMatchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	squadIDStr := vars["squad_id"]

	if squadIDStr == "" {
		slog.Error("GetSquadMatchHistoryHandler: missing squad_id")
		http.Error(w, "squad_id is required", http.StatusBadRequest)
		return
	}

	squadID, err := uuid.Parse(squadIDStr)
	if err != nil {
		slog.Error("GetSquadMatchHistoryHandler: invalid squad_id", "squad_id", squadIDStr, "error", err)
		http.Error(w, "invalid squad_id format", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	slog.Info("Fetching squad match history", "squad_id", squadID, "limit", limit, "offset", offset)

	// First, get the squad to extract member player IDs
	if c.squadReader == nil {
		slog.Error("GetSquadMatchHistoryHandler: squadReader not available")
		http.Error(w, "squad service unavailable", http.StatusServiceUnavailable)
		return
	}

	squadSearchParams := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field:    "ID",
							Values:   []interface{}{squadID.String()},
							Operator: shared.EqualsOperator,
						},
					},
				},
			},
			AggregationClause: shared.AndAggregationClause,
		},
	}

	squadResultOptions := shared.SearchResultOptions{
		Limit: 1,
		Skip:  0,
	}

	compiledSquadSearch, err := c.squadReader.Compile(r.Context(), squadSearchParams, squadResultOptions)
	if err != nil {
		slog.Error("GetSquadMatchHistoryHandler: error compiling squad search", "error", err)
		http.Error(w, "invalid squad search parameters", http.StatusBadRequest)
		return
	}

	squads, err := c.squadReader.Search(r.Context(), *compiledSquadSearch)
	if err != nil {
		slog.Error("GetSquadMatchHistoryHandler: error searching squad", "error", err)
		http.Error(w, "error fetching squad", http.StatusInternalServerError)
		return
	}

	if len(squads) == 0 {
		slog.Warn("GetSquadMatchHistoryHandler: squad not found", "squad_id", squadID)
		http.Error(w, "squad not found", http.StatusNotFound)
		return
	}

	squad := squads[0]

	// Extract player profile IDs from squad membership
	var playerIDs []interface{}
	for _, member := range squad.Membership {
		playerIDs = append(playerIDs, member.PlayerProfileID.String())
	}

	if len(playerIDs) == 0 {
		response := map[string]any{
			"data":        []replay_entity.Match{},
			"squad_id":    squadID,
			"squad_name":  squad.Name,
			"total":       0,
			"limit":       limit,
			"offset":      offset,
			"next_offset": nil,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	// Search for matches containing any of the squad members
	searchParams := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field:    "scoreboard.team_scoreboards.players._id",
							Values:   playerIDs,
							Operator: shared.InOperator,
						},
					},
				},
			},
			AggregationClause: shared.AndAggregationClause,
		},
	}

	resultOptions := shared.SearchResultOptions{
		Limit: uint(limit),  // #nosec G115 - limit is validated to be non-negative
		Skip:  uint(offset), // #nosec G115 - offset is validated to be non-negative
	}

	compiledSearch, err := c.matchReader.Compile(r.Context(), searchParams, resultOptions)
	if err != nil {
		slog.Error("GetSquadMatchHistoryHandler: error compiling match search", "error", err)
		http.Error(w, "invalid search parameters", http.StatusBadRequest)
		return
	}

	results, err := c.matchReader.Search(r.Context(), *compiledSearch)
	if err != nil {
		slog.Error("GetSquadMatchHistoryHandler: error searching matches", "error", err)
		http.Error(w, "error fetching match history", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"data":        results,
		"squad_id":    squadID,
		"squad_name":  squad.Name,
		"total":       len(results),
		"limit":       limit,
		"offset":      offset,
		"next_offset": nil,
	}

	if len(results) == limit {
		response["next_offset"] = offset + limit
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
