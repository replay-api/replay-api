package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/gorilla/mux"
	controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
)

// ReplayFileQueryController handles GET requests for replay files
type ReplayFileQueryController struct {
	controllers.DefaultSearchController[replay_entity.ReplayFile]
	replayFileReader replay_in.ReplayFileReader
}

// Default search fields for replays - used when no search_fields param is provided
// These must match fields defined in the search schema
var replayDefaultSearchFields = []string{"Header"}

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
// Supports query parameters:
//   - q: Text search term (searches in fields specified by search_fields)
//   - search_fields: Comma-separated list of fields to search (default: Header)
//   - player_id, squad_id, status, visibility: Exact match filters
//   - limit, offset: Pagination
func (c *ReplayFileQueryController) ListReplayFilesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["game_id"]

	// Get query parameters
	query := r.URL.Query()
	searchTerm := query.Get("q")
	searchFieldsParam := query.Get("search_fields")
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

	slog.Info("ListReplayFilesHandler", "game_id", gameID, "q", searchTerm, "search_fields", searchFieldsParam, "limit", limit, "offset", offset)

	// Build value parameters for search
	var valueParams []shared.SearchableValue

	// Always filter by game_id if provided
	if gameID != "" {
		valueParams = append(valueParams, shared.SearchableValue{
			Field:    "GameID",
			Values:   []interface{}{gameID},
			Operator: shared.EqualsOperator,
		})
	}

	// Filter by player_id if provided
	if playerID != "" {
		valueParams = append(valueParams, shared.SearchableValue{
			Field:    "ResourceOwner",
			Values:   []interface{}{playerID},
			Operator: shared.EqualsOperator,
		})
	}

	// Filter by squad_id if provided
	if squadID != "" {
		valueParams = append(valueParams, shared.SearchableValue{
			Field:    "ResourceOwner",
			Values:   []interface{}{squadID},
			Operator: shared.EqualsOperator,
		})
	}

	// Filter by status if provided
	if status != "" {
		valueParams = append(valueParams, shared.SearchableValue{
			Field:    "Status",
			Values:   []interface{}{status},
			Operator: shared.EqualsOperator,
		})
	}

	// Filter by visibility if provided (maps to VisibilityLevel in schema)
	if visibility != "" {
		valueParams = append(valueParams, shared.SearchableValue{
			Field:    "VisibilityLevel",
			Values:   []interface{}{visibility},
			Operator: shared.EqualsOperator,
		})
	}

	// Add search term if provided - use search_fields from query or defaults
	if searchTerm != "" {
		searchFields := replayDefaultSearchFields
		if searchFieldsParam != "" {
			// Use fields from query param (SDK sends these based on schema)
			searchFields = strings.Split(searchFieldsParam, ",")
			for i := range searchFields {
				searchFields[i] = strings.TrimSpace(searchFields[i])
			}
		}

		// Add text search for each specified field (OR logic handled by search engine)
		for _, field := range searchFields {
			if field != "" {
				valueParams = append(valueParams, shared.SearchableValue{
					Field:    field,
					Values:   []interface{}{searchTerm},
					Operator: shared.ContainsOperator,
				})
			}
		}
	}

	// Build the search using the helper function
	resultOptions := shared.SearchResultOptions{
		Skip:  uint(offset),
		Limit: uint(limit),
	}

	search := shared.NewSearchByValues(r.Context(), valueParams, resultOptions, shared.UserAudienceIDKey)

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
