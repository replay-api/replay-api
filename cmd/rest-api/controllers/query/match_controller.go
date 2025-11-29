package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/gorilla/mux"
	controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
)

type MatchQueryController struct {
	controllers.DefaultSearchController[replay_entity.Match]
	matchReader replay_in.MatchReader
	eventReader replay_in.EventReader
	teamReader  replay_in.TeamReader
	roundReader replay_in.RoundReader
}

func NewMatchQueryController(c container.Container) *MatchQueryController {
	var matchReader replay_in.MatchReader
	var eventReader replay_in.EventReader
	var teamReader replay_in.TeamReader
	var roundReader replay_in.RoundReader

	if err := c.Resolve(&matchReader); err != nil {
		panic(err)
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

	baseController := controllers.NewDefaultSearchController(matchReader)

	return &MatchQueryController{
		DefaultSearchController: *baseController,
		matchReader:             matchReader,
		eventReader:             eventReader,
		teamReader:              teamReader,
		roundReader:             roundReader,
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
	searchParams := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{Field: "id", Values: []interface{}{matchID}, Operator: common.EqualsOperator},
						{Field: "game_id", Values: []interface{}{gameID}, Operator: common.EqualsOperator},
					},
				},
			},
			AggregationClause: common.AndAggregationClause,
		},
	}

	resultOptions := common.SearchResultOptions{
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
