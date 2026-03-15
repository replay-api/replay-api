package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
)

// GameImportController handles HTTP requests for game discovery import operations.
type GameImportController struct {
	importHandler oracle_in.GameImportCommandHandler
}

// NewGameImportController creates a new controller resolving deps from DI container.
func NewGameImportController(c container.Container) *GameImportController {
	var importHandler oracle_in.GameImportCommandHandler

	if err := c.Resolve(&importHandler); err != nil {
		slog.Warn("GameImportCommandHandler not available", "error", err)
	}

	return &GameImportController{
		importHandler: importHandler,
	}
}

// ImportDiscoveredMatchHandler handles POST /oracle/import
// Accepts an ImportDiscoveredMatchCommand JSON body and runs the full import pipeline.
func (c *GameImportController) ImportDiscoveredMatchHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.importHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		ctx := r.Context()

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

		var cmd oracle_in.ImportDiscoveredMatchCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			slog.ErrorContext(ctx, "invalid request body for import", "error", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, "invalid command parameters", http.StatusBadRequest)
			return
		}

		if err := c.importHandler.ImportDiscoveredMatch(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "failed to import discovered match",
				"external_match_id", cmd.ExternalMatch.ExternalMatchID,
				"game_id", cmd.ExternalMatch.GameID,
				"error", err,
			)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "imported",
			"external_match_id": cmd.ExternalMatch.ExternalMatchID,
			"game_id":           cmd.ExternalMatch.GameID,
		})
	}
}

// ImportFromYouTubeVODHandler handles POST /oracle/import/youtube
// Triggers OCR processing of a YouTube VOD for score extraction.
func (c *GameImportController) ImportFromYouTubeVODHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.importHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		ctx := r.Context()

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var cmd oracle_in.ImportFromYouTubeVODCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			slog.ErrorContext(ctx, "invalid request body for youtube import", "error", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, "invalid command parameters", http.StatusBadRequest)
			return
		}

		if err := c.importHandler.ImportFromYouTubeVOD(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "failed to import youtube VOD",
				"video_url", cmd.VideoURL,
				"game_id", cmd.GameID,
				"error", err,
			)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "processing",
			"video_url": cmd.VideoURL,
			"game_id":   cmd.GameID,
		})
	}
}

// BridgeOracleToMatchResultHandler handles POST /oracle/results/{id}/bridge
// Bridges a finalized OracleResult into a platform MatchResult.
func (c *GameImportController) BridgeOracleToMatchResultHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.importHandler == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		ctx := r.Context()

		vars := mux.Vars(r)
		resultID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "invalid oracle result id", http.StatusBadRequest)
			return
		}

		// Optional body with match_id
		var body struct {
			MatchID *string `json:"match_id,omitempty"`
		}
		json.NewDecoder(r.Body).Decode(&body) // body is optional

		var matchID *uuid.UUID
		if body.MatchID != nil {
			parsed, err := uuid.Parse(*body.MatchID)
			if err != nil {
				http.Error(w, "invalid match_id", http.StatusBadRequest)
				return
			}
			matchID = &parsed
		}

		cmd := oracle_in.BridgeOracleToMatchResultCommand{
			OracleResultID: resultID,
			MatchID:        matchID,
		}

		if err := cmd.Validate(); err != nil {
			http.Error(w, "invalid request parameters", http.StatusBadRequest)
			return
		}

		if err := c.importHandler.BridgeOracleToMatchResult(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "failed to bridge oracle result to match result",
				"oracle_result_id", resultID,
				"error", err,
			)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":           "bridged",
			"oracle_result_id": resultID,
		})
	}
}
