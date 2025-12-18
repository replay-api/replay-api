package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

type PlayerProfileQueryController struct {
	controllers.DefaultSearchController[squad_entities.PlayerProfile]
	statisticsReader squad_in.PlayerStatisticsReader
}

func NewPlayerProfileQueryController(c container.Container) *PlayerProfileQueryController {
	var queryService squad_in.PlayerProfileReader

	err := c.Resolve(&queryService)

	if err != nil {
		panic(err)
	}

	var statsReader squad_in.PlayerStatisticsReader
	if err := c.Resolve(&statsReader); err != nil {
		slog.Warn("PlayerStatisticsReader not registered, stats endpoint will not work", "error", err)
	}

	baseController := controllers.NewDefaultSearchController(queryService)

	return &PlayerProfileQueryController{
		DefaultSearchController: *baseController,
		statisticsReader:        statsReader,
	}
}

// GetPlayerStatsHandler handles GET /players/{id}/stats
func (c *PlayerProfileQueryController) GetPlayerStatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerIDStr := vars["id"]

	if playerIDStr == "" {
		http.Error(w, "player_id is required", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		slog.Error("GetPlayerStatsHandler: invalid player_id", "player_id", playerIDStr, "error", err)
		http.Error(w, "invalid player_id format", http.StatusBadRequest)
		return
	}

	if c.statisticsReader == nil {
		http.Error(w, "player statistics service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse optional game_id filter
	var gameID *common.GameIDKey
	if gameIDStr := r.URL.Query().Get("game_id"); gameIDStr != "" {
		gid := common.GameIDKey(gameIDStr)
		gameID = &gid
	}

	stats, err := c.statisticsReader.GetPlayerStatistics(r.Context(), playerID, gameID)
	if err != nil {
		slog.Error("GetPlayerStatsHandler: error getting player statistics", "error", err, "player_id", playerID)
		http.Error(w, "error fetching player statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("GetPlayerStatsHandler: error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
