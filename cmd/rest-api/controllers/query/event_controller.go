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
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	"github.com/google/uuid"
)

type EventQueryController struct {
	controllers.DefaultSearchController[replay_entity.GameEvent]
	queryService replay_in.EventReader
}

func NewEventQueryController(c container.Container) *EventQueryController {
	var queryService replay_in.EventReader

	err := c.Resolve(&queryService)

	if err != nil {
		slog.Warn("EventReader not available - event queries will be disabled", "error", err)
	}

	baseController := controllers.NewDefaultSearchController(queryService)

	return &EventQueryController{
		DefaultSearchController: *baseController,
		queryService:            queryService,
	}
}

// GetMatchEventsHandler returns events for a specific match with pagination and filtering
// Query params:
// - limit: max events to return (default: 100, max: 1000)
// - offset: pagination offset (default: 0)
// - event_type: single event type filter (e.g., "kill")
// - event_types: comma-separated event types (e.g., "kill,clutchstart,clutchend")
func (c *EventQueryController) GetMatchEventsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["game_id"]
	matchID := vars["match_id"]

	if gameID == "" || matchID == "" {
		http.Error(w, "game_id and match_id are required", http.StatusBadRequest)
		return
	}

	matchUUID, err := uuid.Parse(matchID)
	if err != nil {
		http.Error(w, "invalid match_id format", http.StatusBadRequest)
		return
	}

	// Parse query params
	limit := 100
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Support both single event_type and multiple event_types (comma-separated)
	var eventTypes []string
	if types := r.URL.Query().Get("event_types"); types != "" {
		// Multiple types: "kill,clutchstart,clutchend"
		for _, t := range strings.Split(types, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				eventTypes = append(eventTypes, trimmed)
			}
		}
	} else if singleType := r.URL.Query().Get("event_type"); singleType != "" {
		// Single type for backward compatibility
		eventTypes = []string{singleType}
	}

	ctx := r.Context()

	if c.queryService == nil {
		http.Error(w, "event service not available", http.StatusServiceUnavailable)
		return
	}

	// Use new method with proper total count for pagination
	events, totalCount, err := c.queryService.GetMatchEventsWithCount(ctx, gameID, matchUUID, limit, offset, eventTypes)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch match events", "error", err, "match_id", matchID)
		http.Error(w, "failed to fetch events", http.StatusInternalServerError)
		return
	}

	// Return response with proper pagination metadata
	response := map[string]interface{}{
		"events":       events,
		"match_id":     matchID,
		"total_events": totalCount,          // Total matching events in database
		"returned":     len(events),         // Number of events in this response
		"limit":        limit,
		"offset":       offset,
		"has_more":     offset+len(events) < int(totalCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
