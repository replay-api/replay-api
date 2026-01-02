package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/gorilla/mux"
	controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
)

// ReplayFileQueryController handles GET requests for replay files
type ReplayFileQueryController struct {
	controllers.DefaultSearchController[replay_entity.ReplayFile]
	replayFileReader replay_in.ReplayFileReader
}

// NewReplayFileQueryController creates a new ReplayFileQueryController
func NewReplayFileQueryController(c container.Container) *ReplayFileQueryController {
	var replayFileReader replay_in.ReplayFileReader

	if err := c.Resolve(&replayFileReader); err != nil {
		slog.Warn("ReplayFileReader not available - replay file queries will be disabled", "error", err)
	}

	baseController := controllers.NewDefaultSearchController(replayFileReader)

	return &ReplayFileQueryController{
		DefaultSearchController: *baseController,
		replayFileReader:        replayFileReader,
	}
}

// ListReplayFilesHandler handles GET /games/{game_id}/replays
// Supports query parameters: q (search term), player_id, squad_id, status, visibility, limit, offset
func (c *ReplayFileQueryController) ListReplayFilesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["game_id"]

	// Get query parameters
	query := r.URL.Query()
	searchTerm := query.Get("q")
	playerID := query.Get("player_id")
	squadID := query.Get("squad_id")
	status := query.Get("status")
	visibility := query.Get("visibility")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	// Parse pagination
	limit := 20 // default
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	slog.Info("ListReplayFilesHandler", "game_id", gameID, "q", searchTerm, "player_id", playerID, "limit", limit, "offset", offset)

	// Build value parameters for search
	var valueParams []common.SearchableValue

	// Always filter by game_id if provided
	if gameID != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "GameID",
			Values:   []interface{}{gameID},
			Operator: common.EqualsOperator,
		})
	}

	// Filter by player_id if provided
	if playerID != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "UploadedBy",
			Values:   []interface{}{playerID},
			Operator: common.EqualsOperator,
		})
	}

	// Filter by squad_id if provided
	if squadID != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "SquadID",
			Values:   []interface{}{squadID},
			Operator: common.EqualsOperator,
		})
	}

	// Filter by status if provided
	if status != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "Status",
			Values:   []interface{}{status},
			Operator: common.EqualsOperator,
		})
	}

	// Filter by visibility if provided
	if visibility != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "Visibility",
			Values:   []interface{}{visibility},
			Operator: common.EqualsOperator,
		})
	}

	// Add search term if provided (search in filename)
	if searchTerm != "" {
		valueParams = append(valueParams, common.SearchableValue{
			Field:    "FileName",
			Values:   []interface{}{searchTerm},
			Operator: common.ContainsOperator,
		})
	}

	// Build the search using the helper function
	resultOptions := common.SearchResultOptions{
		Skip:  uint(offset),
		Limit: uint(limit),
	}

	search := common.NewSearchByValues(r.Context(), valueParams, resultOptions, common.UserAudienceIDKey)

	// Execute search
	results, err := c.replayFileReader.Search(r.Context(), search)
	if err != nil {
		slog.Error("ListReplayFilesHandler: search error", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to search replay files"})
		return
	}

	// Return empty array instead of null
	if results == nil {
		results = []replay_entity.ReplayFile{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
