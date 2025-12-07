package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
)

type FileController struct {
	container container.Container
}

func NewFileController(container container.Container) *FileController {
	return &FileController{container: container}
}

func (ctlr *FileController) UploadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // todo: PARAMETRIZAR
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// r.Body = http.MaxBytesReader(w, r.Body, 32<<57)
		_ = r.ParseMultipartForm(32 << 50)

		reqContext := context.WithValue(r.Context(), common.GameIDParamKey, r.FormValue("game_id"))

		slog.InfoContext(reqContext, "Receiving file", string(common.GameIDParamKey), r.FormValue("game_id"))

		file, _, err := r.FormFile("file")
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to get file", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		var uploadAndProcessReplayFileCommand replay_in.UploadAndProcessReplayFileCommand
		err = ctlr.container.Resolve(&uploadAndProcessReplayFileCommand)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to resolve uploadAndProcessReplayFileCommand", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		match, err := uploadAndProcessReplayFileCommand.Exec(reqContext, file)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to upload and process file", "err", err)
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			}
			return
		}

		match.Events = nil

		err = json.NewEncoder(w).Encode(match)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to encode response", "err", err, "match", match)
			w.WriteHeader(http.StatusBadGateway)
		}

		w.Header().Set("Location", r.URL.Path+"/"+match.ID.String())
		w.WriteHeader(http.StatusCreated)
	}
}

// func (ctlr *FileController) ReplayMetadataFilterHandler(apiContext context.Context) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "localhost:3000")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

// 		reqContext := context.WithValue(r.Context(), common.GameIDParamKey, r.FormValue("game_id"))

// 		var replayFileMetadataReader replay_in.ReplayFileMetadataReader
// 		err := ctlr.container.Resolve(&replayFileMetadataReader)
// 		if err != nil {
// 			slog.ErrorContext(reqContext, "Failed to resolve replayFileMetadataReader", "err", err)
// 			w.WriteHeader(http.StatusServiceUnavailable)
// 			return
// 		}

// 		var params []common.SearchAggregation

// 		// for key, values := range r.URL.Query() {
// 		// 	params = append(params, common.SearchAggregation{
// 		// 		Key:    key,
// 		// 		Values: values,
// 		// 	})
// 		// }

// 		// replayFiles, err := replayFileMetadataReader.Filter(reqContext, r.URL.Query())
// 		// if err != nil {
// 		// 	slog.ErrorContext(reqContext, "Failed to get replay files", "err", err)
// 		// 	w.WriteHeader(http.StatusInternalServerError)
// 		// 	return
// 		// }

// 		// err = json.NewEncoder(w).Encode(replayFiles)
// 		// if err != nil {
// 		// 	slog.ErrorContext(reqContext, "Failed to encode response", err, "replayFiles", replayFiles)
// 		// 	w.WriteHeader(http.StatusBadGateway)
// 		// }

// 		// w.Header().Set("Location", r.URL.Path)
// 		// w.WriteHeader(http.StatusOK)
// 	}
// }

// GetReplayMetadata handles GET /games/{game_id}/replays/{id}
func (ctlr *FileController) GetReplayMetadata(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		replayID := vars.Get("id")
		gameID := vars.Get("game_id")

		if replayID == "" || gameID == "" {
			slog.Error("GetReplayMetadata: missing replay_id or game_id")
			http.Error(w, "replay_id and game_id are required", http.StatusBadRequest)
			return
		}

		var replayFileReader replay_in.ReplayFileReader
		err := ctlr.container.Resolve(&replayFileReader)
		if err != nil {
			slog.Error("GetReplayMetadata: failed to resolve ReplayFileReader", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Build search using strongly typed aggregations
		idUUID, err := uuid.Parse(replayID)
		if err != nil {
			slog.Error("GetReplayMetadata: invalid replay_id UUID", "replay_id", replayID, "err", err)
			http.Error(w, "invalid replay_id", http.StatusBadRequest)
			return
		}

		valueParams := []common.SearchableValue{
			{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator},
		}
		if gameID != "" {
			valueParams = append(valueParams, common.SearchableValue{Field: "GameID", Values: []interface{}{gameID}, Operator: common.EqualsOperator})
		}

		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := replayFileReader.Search(r.Context(), search)
		if err != nil {
			slog.Error("GetReplayMetadata: error searching replay", "err", err)
			http.Error(w, "error fetching replay", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			slog.Warn("GetReplayMetadata: replay not found", "replay_id", replayID)
			http.Error(w, "replay not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(results[0])
	}
}

// DownloadReplayFile handles GET /games/{game_id}/replays/{id}/download
func (ctlr *FileController) DownloadReplayFile(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Implementation would stream the file from storage
		// This is a placeholder for the download endpoint
		http.Error(w, "Download endpoint not yet implemented - requires S3/MinIO integration", http.StatusNotImplemented)
	}
}

// DeleteReplayFile handles DELETE /games/{game_id}/replays/{id}
func (ctlr *FileController) DeleteReplayFile(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Implementation would delete from both storage and database
		// This is a placeholder for the delete endpoint
		http.Error(w, "Delete endpoint not yet implemented - requires S3/MinIO integration", http.StatusNotImplemented)
	}
}

// UpdateReplayMetadata handles PUT /games/{game_id}/replays/{id}
func (ctlr *FileController) UpdateReplayMetadata(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Implementation would update replay metadata (title, description, visibility, etc.)
		// This is a placeholder for the update endpoint
		http.Error(w, "Update endpoint not yet implemented", http.StatusNotImplemented)
	}
}

// GetReplayProcessingStatus handles GET /games/{game_id}/replays/{id}/status
func (ctlr *FileController) GetReplayProcessingStatus(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		replayID := vars.Get("id")

		if replayID == "" {
			http.Error(w, "replay_id is required", http.StatusBadRequest)
			return
		}

		var replayFileReader replay_in.ReplayFileReader
		err := ctlr.container.Resolve(&replayFileReader)
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(replayID)
		if err != nil {
			http.Error(w, "invalid replay_id", http.StatusBadRequest)
			return
		}

		valueParams := []common.SearchableValue{
			{Field: "ID", Values: []interface{}{idUUID}, Operator: common.EqualsOperator},
		}

		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := replayFileReader.Search(r.Context(), search)
		if err != nil {
			http.Error(w, "error fetching replay", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			http.Error(w, "replay not found", http.StatusNotFound)
			return
		}

		replay := results[0]
		status := map[string]interface{}{
			"id":              replay.ID,
			"status":          replay.Status,
			"processing_pct":  100, // Completed if found
			"created_at":      replay.CreatedAt,
			"updated_at":      replay.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(status)
	}
}

// GetReplayEvents handles GET /games/{game_id}/replays/{id}/events
func (ctlr *FileController) GetReplayEvents(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		replayID := vars.Get("id")
		eventType := r.URL.Query().Get("type") // Optional: kill, plant, defuse, etc.

		if replayID == "" {
			http.Error(w, "replay_id is required", http.StatusBadRequest)
			return
		}

		var matchReader replay_in.MatchReader
		err := ctlr.container.Resolve(&matchReader)
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(replayID)
		if err != nil {
			http.Error(w, "invalid replay_id", http.StatusBadRequest)
			return
		}

		valueParams := []common.SearchableValue{
			{Field: "replay_file_id", Values: []interface{}{idUUID}, Operator: common.EqualsOperator},
		}

		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := matchReader.Search(r.Context(), search)
		if err != nil {
			http.Error(w, "error fetching match", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			http.Error(w, "match not found for replay", http.StatusNotFound)
			return
		}

		match := results[0]
		events := match.Events

		// Filter by event type if specified
		if eventType != "" && len(events) > 0 {
			var filtered []*replay_entity.GameEvent
			for _, evt := range events {
				if evt != nil && string(evt.Type) == eventType {
					filtered = append(filtered, evt)
				}
			}
			events = filtered
		}

		response := map[string]interface{}{
			"replay_id":    replayID,
			"match_id":     match.ID,
			"events":       events,
			"total_events": len(events),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetReplayScoreboard handles GET /games/{game_id}/replays/{id}/scoreboard
func (ctlr *FileController) GetReplayScoreboard(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		replayID := vars.Get("id")

		if replayID == "" {
			http.Error(w, "replay_id is required", http.StatusBadRequest)
			return
		}

		var matchReader replay_in.MatchReader
		err := ctlr.container.Resolve(&matchReader)
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(replayID)
		if err != nil {
			http.Error(w, "invalid replay_id", http.StatusBadRequest)
			return
		}

		valueParams := []common.SearchableValue{
			{Field: "replay_file_id", Values: []interface{}{idUUID}, Operator: common.EqualsOperator},
		}

		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := matchReader.Search(r.Context(), search)
		if err != nil {
			http.Error(w, "error fetching match", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			http.Error(w, "match not found for replay", http.StatusNotFound)
			return
		}

		match := results[0]
		
		response := map[string]interface{}{
			"replay_id":   replayID,
			"match_id":    match.ID,
			"scoreboard":  match.Scoreboard,
			"teams":       match.Scoreboard.TeamScoreboards,
			"mvp":         match.Scoreboard.MatchMVP,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetReplayTimeline handles GET /games/{game_id}/replays/{id}/timeline
func (ctlr *FileController) GetReplayTimeline(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		replayID := vars.Get("id")

		if replayID == "" {
			http.Error(w, "replay_id is required", http.StatusBadRequest)
			return
		}

		var matchReader replay_in.MatchReader
		err := ctlr.container.Resolve(&matchReader)
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		idUUID, err := uuid.Parse(replayID)
		if err != nil {
			http.Error(w, "invalid replay_id", http.StatusBadRequest)
			return
		}

		valueParams := []common.SearchableValue{
			{Field: "replay_file_id", Values: []interface{}{idUUID}, Operator: common.EqualsOperator},
		}

		search := common.NewSearchByValues(r.Context(), valueParams, common.SearchResultOptions{Limit: 1}, common.UserAudienceIDKey)
		results, err := matchReader.Search(r.Context(), search)
		if err != nil {
			http.Error(w, "error fetching match", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			http.Error(w, "match not found for replay", http.StatusNotFound)
			return
		}

		match := results[0]

		// Build timeline from rounds and events
		timeline := make([]map[string]interface{}, 0)
		totalRounds := 0
		
		// Add round data to timeline
		for i, team := range match.Scoreboard.TeamScoreboards {
			for _, round := range team.Rounds {
				timeline = append(timeline, map[string]interface{}{
					"round":        round.RoundNumber,
					"team":         i,
					"team_name":    team.Team.Name,
					"winner":       round.WinnerTeamID,
					"round_mvp":    round.RoundMVPPlayerID,
					"type":         "round_end",
				})
			}
			if len(team.Rounds) > totalRounds {
				totalRounds = len(team.Rounds)
			}
		}

		// Calculate final score from team scoreboards
		finalScore := ""
		if len(match.Scoreboard.TeamScoreboards) >= 2 {
			finalScore = string(rune(match.Scoreboard.TeamScoreboards[0].TeamScore)) + "-" + string(rune(match.Scoreboard.TeamScoreboards[1].TeamScore))
		}

		response := map[string]interface{}{
			"replay_id":     replayID,
			"match_id":      match.ID,
			"timeline":      timeline,
			"total_rounds":  totalRounds,
			"final_score":   finalScore,
			"scoreboard":    match.Scoreboard,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}
