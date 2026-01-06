package cmd_controllers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
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

		reqContext := context.WithValue(r.Context(), shared.GameIDParamKey, r.FormValue("game_id"))

		slog.InfoContext(reqContext, "Receiving file", string(shared.GameIDParamKey), r.FormValue("game_id"))

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

// 		reqContext := context.WithValue(r.Context(), shared.GameIDParamKey, r.FormValue("game_id"))

// 		var replayFileMetadataReader replay_in.ReplayFileMetadataReader
// 		err := ctlr.container.Resolve(&replayFileMetadataReader)
// 		if err != nil {
// 			slog.ErrorContext(reqContext, "Failed to resolve replayFileMetadataReader", "err", err)
// 			w.WriteHeader(http.StatusServiceUnavailable)
// 			return
// 		}

// 		var params []shared.SearchAggregation

// 		// for key, values := range r.URL.Query() {
// 		// 	params = append(params, shared.SearchAggregation{
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

		valueParams := []shared.SearchableValue{
			{Field: "ID", Values: []interface{}{idUUID}, Operator: shared.EqualsOperator},
		}
		if gameID != "" {
			valueParams = append(valueParams, shared.SearchableValue{Field: "GameID", Values: []interface{}{gameID}, Operator: shared.EqualsOperator})
		}

		search := shared.NewSearchByValues(r.Context(), valueParams, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)
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

// requireReplayOwnership verifies the user owns the replay before allowing modifications
// Returns the replay if owned, nil if not owned (writes error response)
func (ctlr *FileController) requireReplayOwnership(w http.ResponseWriter, r *http.Request, replayID uuid.UUID) *replay_entity.ReplayFile {
	ctx := r.Context()
	
	// Check authentication
	authenticated, ok := ctx.Value(shared.AuthenticatedKey).(bool)
	if !ok || !authenticated {
		slog.WarnContext(ctx, "replay modification attempted without authentication", "replay_id", replayID)
		http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return nil
	}
	
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		slog.WarnContext(ctx, "replay modification attempted without valid user ID", "replay_id", replayID)
		http.Error(w, `{"error":"valid user authentication required"}`, http.StatusUnauthorized)
		return nil
	}
	
	// Fetch the replay to check ownership
	var replayFileReader replay_in.ReplayFileReader
	if err := ctlr.container.Resolve(&replayFileReader); err != nil {
		slog.ErrorContext(ctx, "failed to resolve ReplayFileReader", "err", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return nil
	}
	
	valueParams := []shared.SearchableValue{
		{Field: "ID", Values: []interface{}{replayID}, Operator: shared.EqualsOperator},
	}
	search := shared.NewSearchByValues(ctx, valueParams, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)
	results, err := replayFileReader.Search(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error fetching replay for ownership check", "err", err, "replay_id", replayID)
		http.Error(w, "error fetching replay", http.StatusInternalServerError)
		return nil
	}
	
	if len(results) == 0 {
		slog.WarnContext(ctx, "replay not found for ownership check", "replay_id", replayID)
		http.Error(w, "replay not found", http.StatusNotFound)
		return nil
	}
	
	replay := &results[0]
	
	// SECURITY: Verify ownership - user must own the replay to modify it
	// Allow admins to bypass this check
	isAdmin := shared.IsAdmin(ctx)
	if !isAdmin && replay.ResourceOwner.UserID != resourceOwner.UserID {
		slog.WarnContext(ctx, "unauthorized replay modification attempt",
			"replay_id", replayID,
			"replay_owner", replay.ResourceOwner.UserID,
			"requesting_user", resourceOwner.UserID,
		)
		http.Error(w, `{"error":"you do not have permission to modify this replay"}`, http.StatusForbidden)
		return nil
	}
	
	return replay
}

// DownloadReplayFile handles GET /games/{game_id}/replays/{id}/download
// SECURITY: For private replays, user must be owner or have share token
func (ctlr *FileController) DownloadReplayFile(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := r.URL.Query()
		replayIDStr := vars.Get("id")
		shareToken := vars.Get("token") // Optional share token for private replays

		if replayIDStr == "" {
			http.Error(w, `{"error":"replay_id is required"}`, http.StatusBadRequest)
			return
		}

		replayID, err := uuid.Parse(replayIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid replay_id"}`, http.StatusBadRequest)
			return
		}

		// Resolve dependencies
		var replayFileReader replay_in.ReplayFileReader
		if err := ctlr.container.Resolve(&replayFileReader); err != nil {
			slog.ErrorContext(ctx, "failed to resolve ReplayFileReader", "err", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		var replayContentReader replay_in.ReplayContentReader
		if err := ctlr.container.Resolve(&replayContentReader); err != nil {
			slog.ErrorContext(ctx, "failed to resolve ReplayContentReader", "err", err)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		// Get replay metadata
		valueParams := []shared.SearchableValue{
			{Field: "ID", Values: []interface{}{replayID}, Operator: shared.EqualsOperator},
		}
		search := shared.NewSearchByValues(ctx, valueParams, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)
		results, err := replayFileReader.Search(ctx, search)
		if err != nil {
			slog.ErrorContext(ctx, "error fetching replay metadata", "err", err, "replay_id", replayID)
			http.Error(w, "error fetching replay", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			http.Error(w, `{"error":"replay not found"}`, http.StatusNotFound)
			return
		}

		replay := &results[0]

		// Check access: owner, share token, or admin
		resourceOwner := shared.GetResourceOwner(ctx)
		isOwner := replay.ResourceOwner.UserID == resourceOwner.UserID
		isAdmin := shared.IsAdmin(ctx)

		// Check share token if provided
		hasValidShareToken := false
		if shareToken != "" {
			var shareTokenReader replay_in.ShareTokenReader
			if err := ctlr.container.Resolve(&shareTokenReader); err == nil {
				tokenID, parseErr := uuid.Parse(shareToken)
				if parseErr == nil {
					token, tokenErr := shareTokenReader.FindByToken(ctx, tokenID)
					if tokenErr == nil && token != nil && token.ResourceID == replayID && token.IsValid() {
						hasValidShareToken = true
					}
				}
			}
		}

		// SECURITY: Verify access - user must be owner, admin, or have valid share token
		if !isOwner && !isAdmin && !hasValidShareToken {
			slog.WarnContext(ctx, "unauthorized replay download attempt",
				"replay_id", replayID,
				"replay_owner", replay.ResourceOwner.UserID,
				"requesting_user", resourceOwner.UserID,
			)
			http.Error(w, `{"error":"access denied - you do not have permission to download this replay"}`, http.StatusForbidden)
			return
		}

		// Stream replay content
		content, err := replayContentReader.GetByID(ctx, replayID)
		if err != nil {
			slog.ErrorContext(ctx, "error fetching replay content", "err", err, "replay_id", replayID)
			http.Error(w, "error fetching replay content", http.StatusInternalServerError)
			return
		}
		defer content.Close()

		// Set headers for file download
		fileName := replayID.String() + ".dem"
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		w.Header().Set("Cache-Control", "private, max-age=3600")

		// Stream the content
		written, err := io.Copy(w, content)
		if err != nil {
			slog.ErrorContext(ctx, "error streaming replay content", "err", err, "replay_id", replayID, "bytes_written", written)
			return
		}

		slog.InfoContext(ctx, "replay downloaded",
			"replay_id", replayID,
			"user_id", resourceOwner.UserID,
			"bytes", written,
		)
	}
}

// DeleteReplayFile handles DELETE /games/{game_id}/replays/{id}
// SECURITY: Only the replay owner can delete their replay
func (ctlr *FileController) DeleteReplayFile(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		replayIDStr := vars.Get("id")
		
		if replayIDStr == "" {
			http.Error(w, "replay_id is required", http.StatusBadRequest)
			return
		}
		
		replayID, err := uuid.Parse(replayIDStr)
		if err != nil {
			http.Error(w, "invalid replay_id", http.StatusBadRequest)
			return
		}
		
		// SECURITY: Verify ownership before deletion
		replay := ctlr.requireReplayOwnership(w, r, replayID)
		if replay == nil {
			return // Response already written
		}
		
		// TODO: Implement actual deletion from storage and database
		// For now, just log that deletion was requested
		slog.InfoContext(r.Context(), "Replay deletion requested",
			"replay_id", replayID,
			"owner_id", replay.ResourceOwner.UserID,
		)
		
		http.Error(w, "Delete endpoint not yet implemented - requires S3/MinIO integration", http.StatusNotImplemented)
	}
}

// UpdateReplayMetadata handles PUT /games/{game_id}/replays/{id}
// SECURITY: Only the replay owner can update their replay metadata
func (ctlr *FileController) UpdateReplayMetadata(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		replayIDStr := vars.Get("id")
		
		if replayIDStr == "" {
			http.Error(w, "replay_id is required", http.StatusBadRequest)
			return
		}
		
		replayID, err := uuid.Parse(replayIDStr)
		if err != nil {
			http.Error(w, "invalid replay_id", http.StatusBadRequest)
			return
		}
		
		// SECURITY: Verify ownership before update
		replay := ctlr.requireReplayOwnership(w, r, replayID)
		if replay == nil {
			return // Response already written
		}
		
		// TODO: Parse update request body and update metadata
		// Allowed fields: title, description, visibility_type, tags
		// Do NOT allow changing: owner, game_id, match data
		slog.InfoContext(r.Context(), "Replay update requested",
			"replay_id", replayID,
			"owner_id", replay.ResourceOwner.UserID,
		)
		
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

		valueParams := []shared.SearchableValue{
			{Field: "ID", Values: []interface{}{idUUID}, Operator: shared.EqualsOperator},
		}

		search := shared.NewSearchByValues(r.Context(), valueParams, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)
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

		valueParams := []shared.SearchableValue{
			{Field: "replay_file_id", Values: []interface{}{idUUID}, Operator: shared.EqualsOperator},
		}

		search := shared.NewSearchByValues(r.Context(), valueParams, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)
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

		valueParams := []shared.SearchableValue{
			{Field: "replay_file_id", Values: []interface{}{idUUID}, Operator: shared.EqualsOperator},
		}

		search := shared.NewSearchByValues(r.Context(), valueParams, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)
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

		valueParams := []shared.SearchableValue{
			{Field: "replay_file_id", Values: []interface{}{idUUID}, Operator: shared.EqualsOperator},
		}

		search := shared.NewSearchByValues(r.Context(), valueParams, shared.SearchResultOptions{Limit: 1}, shared.UserAudienceIDKey)
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
