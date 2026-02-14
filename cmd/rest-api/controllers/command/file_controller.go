package cmd_controllers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	shared "github.com/resource-ownership/go-common/pkg/common"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type FileController struct {
	container container.Container
}

func NewFileController(container container.Container) *FileController {
	return &FileController{container: container}
}

func (ctlr *FileController) UploadHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers are handled by middleware - don't override them here
		// w.Header().Set("Access-Control-Allow-Methods", "POST")
		// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

		// Parse optional metadata from form data
		var opts *replay_entity.ReplayFileOptions
		title := r.FormValue("title")
		description := r.FormValue("description")
		visibilityStr := r.FormValue("visibility")
		tagsStr := r.FormValue("tags") // comma-separated

		if title != "" || description != "" || visibilityStr != "" || tagsStr != "" {
			opts = &replay_entity.ReplayFileOptions{
				Title:       title,
				Description: description,
			}
			
			// Parse visibility (1=public, 2=restricted, 4=private)
			if visibilityStr != "" {
				var visibility int
				if _, err := json.Number(visibilityStr).Int64(); err == nil {
					visibility = int(json.Number(visibilityStr).String()[0] - '0')
				}
				switch visibility {
				case 1:
					opts.Visibility = shared.PublicVisibilityTypeKey
				case 2:
					opts.Visibility = shared.RestrictedVisibilityTypeKey
				case 4:
					opts.Visibility = shared.PrivateVisibilityTypeKey
				default:
					opts.Visibility = shared.PublicVisibilityTypeKey // default to public
				}
			}
			
			// Parse tags
			if tagsStr != "" {
				opts.Tags = splitAndTrim(tagsStr, ",")
			}
		}

		var uploadAndProcessReplayFileCommand replay_in.UploadAndProcessReplayFileCommand
		err = ctlr.container.Resolve(&uploadAndProcessReplayFileCommand)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to resolve uploadAndProcessReplayFileCommand", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Use ExecWithOptions if we have options, otherwise use Exec
		var replayFile *replay_entity.ReplayFile
		if opts != nil {
			// Check if the command supports options
			if cmdWithOpts, ok := uploadAndProcessReplayFileCommand.(replay_in.UploadAndProcessReplayFileWithOptionsCommand); ok {
				replayFile, err = cmdWithOpts.ExecWithOptions(reqContext, file, opts)
			} else {
				// Fall back to regular Exec
				replayFile, err = uploadAndProcessReplayFileCommand.Exec(reqContext, file)
			}
		} else {
			replayFile, err = uploadAndProcessReplayFileCommand.Exec(reqContext, file)
		}

		if err != nil {
			slog.ErrorContext(reqContext, "Failed to upload and process file", "err", err)
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			}
			return
		}

		err = json.NewEncoder(w).Encode(replayFile)
		if err != nil {
			slog.ErrorContext(reqContext, "Failed to encode response", "err", err, "replayFile", replayFile)
			w.WriteHeader(http.StatusBadGateway)
		}

		w.Header().Set("Location", r.URL.Path+"/"+replayFile.ID.String())
		w.WriteHeader(http.StatusCreated)
	}
}

// splitAndTrim splits a string by separator and trims each element
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, p := range splitString(s, sep) {
		trimmed := trimString(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
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

// GetReplayMetadata handles GET /games/{game_id}/replays/{id} and /games/{game_id}/replay-files/{id}
func (ctlr *FileController) GetReplayMetadata(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try path variables first (preferred)
		vars := mux.Vars(r)
		replayID := vars["id"]
		gameID := vars["game_id"]
		
		// Fallback to query parameters if path variables are empty
		if replayID == "" {
			replayID = r.URL.Query().Get("id")
		}
		if gameID == "" {
			gameID = r.URL.Query().Get("game_id")
		}

		if replayID == "" || gameID == "" {
			slog.Error("GetReplayMetadata: missing replay_id or game_id")
			http.Error(w, "replay_id and game_id are required", http.StatusBadRequest)
			return
		}

		var matchReader replay_in.MatchReader
		err := ctlr.container.Resolve(&matchReader)
		if err != nil {
			slog.Error("GetReplayMetadata: failed to resolve MatchReader", "err", err)
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
		results, err := matchReader.Search(r.Context(), search)
		if err != nil {
			slog.Error("GetReplayMetadata: error searching match", "err", err)
			http.Error(w, "error fetching replay", http.StatusInternalServerError)
			return
		}

		if len(results) == 0 {
			slog.Warn("GetReplayMetadata: match not found", "replay_id", replayID)
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

		// Soft-delete: set status to Deleted
		replay.Status = replay_entity.ReplayFileStatusDeleted

		var replayFileWriter replay_out.ReplayFileMetadataWriter
		if err := ctlr.container.Resolve(&replayFileWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ReplayFileMetadataWriter", "err", err)
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		updatedReplay, err := replayFileWriter.Update(r.Context(), replay)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to soft-delete replay",
				"replay_id", replayID,
				"err", err,
			)
			http.Error(w, `{"error":"failed to delete replay"}`, http.StatusInternalServerError)
			return
		}

		slog.InfoContext(r.Context(), "Replay soft-deleted successfully",
			"replay_id", replayID,
			"owner_id", updatedReplay.ResourceOwner.UserID,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"message":   "Replay deleted successfully",
			"replay_id": replayID,
		})
	}
}

// UpdateReplayMetadataRequest represents the request body for updating replay metadata
type UpdateReplayMetadataRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Visibility  *int     `json:"visibility,omitempty"` // 1=public, 2=restricted, 4=private
}

// UpdateReplayMetadata handles PUT /games/{game_id}/replays/{id}
// SECURITY: Only the replay owner can update their replay metadata
func (ctlr *FileController) UpdateReplayMetadata(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		replayIDStr := vars["id"]
		
		// Fallback to query param
		if replayIDStr == "" {
			replayIDStr = r.URL.Query().Get("id")
		}
		
		if replayIDStr == "" {
			http.Error(w, `{"error":"replay_id is required"}`, http.StatusBadRequest)
			return
		}
		
		replayID, err := uuid.Parse(replayIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid replay_id"}`, http.StatusBadRequest)
			return
		}
		
		// Parse request body
		var updateReq UpdateReplayMetadataRequest
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			slog.ErrorContext(ctx, "failed to parse update request body", "err", err)
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		
		// Build options from request
		opts := &replay_entity.ReplayFileOptions{}
		if updateReq.Title != nil {
			opts.Title = *updateReq.Title
		}
		if updateReq.Description != nil {
			opts.Description = *updateReq.Description
		}
		if updateReq.Tags != nil {
			opts.Tags = updateReq.Tags
		}
		if updateReq.Visibility != nil {
			opts.Visibility = shared.VisibilityTypeKey(*updateReq.Visibility)
		}
		
		// Resolve use case
		var updateCommand replay_in.UpdateReplayMetadataCommand
		if err := ctlr.container.Resolve(&updateCommand); err != nil {
			slog.ErrorContext(ctx, "failed to resolve UpdateReplayMetadataCommand", "err", err)
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		
		// Execute update
		updatedReplay, err := updateCommand.Exec(ctx, replayID, opts)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update replay metadata", "err", err, "replay_id", replayID)
			
			switch err.Error() {
			case "replay not found":
				http.Error(w, `{"error":"replay not found"}`, http.StatusNotFound)
			case "not authorized to update this replay":
				http.Error(w, `{"error":"you do not have permission to modify this replay"}`, http.StatusForbidden)
			case "invalid visibility type":
				http.Error(w, `{"error":"invalid visibility value"}`, http.StatusBadRequest)
			default:
				http.Error(w, `{"error":"failed to update replay"}`, http.StatusInternalServerError)
			}
			return
		}
		
		slog.InfoContext(ctx, "Replay metadata updated",
			"replay_id", replayID,
			"user_id", shared.GetResourceOwner(ctx).UserID,
		)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(updatedReplay)
	}
}

// GetReplayProcessingStatus handles GET /games/{game_id}/replays/{id}/status
func (ctlr *FileController) GetReplayProcessingStatus(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathVars := mux.Vars(r)
		replayID := pathVars["id"]

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
		pathVars := mux.Vars(r)
		replayID := pathVars["id"]
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
			{Field: "ReplayFileID", Values: []interface{}{idUUID}, Operator: shared.EqualsOperator},
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

		var eventReader replay_in.EventReader
		err = ctlr.container.Resolve(&eventReader)
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		// Fetch events for this match
		eventSearch := shared.NewSearchByValues(r.Context(), []shared.SearchableValue{
			{Field: "MatchID", Values: []interface{}{match.ID}, Operator: shared.EqualsOperator},
		}, shared.SearchResultOptions{Limit: 200}, shared.UserAudienceIDKey) // TODO: Add pagination

		eventResults, err := eventReader.Search(r.Context(), eventSearch)
		if err != nil {
			http.Error(w, "error fetching events", http.StatusInternalServerError)
			return
		}

		events := make([]*replay_entity.GameEvent, len(eventResults))
		for i, event := range eventResults {
			events[i] = &event
		}

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

		// Transform events to flatten payload for frontend compatibility
		transformedEvents := make([]map[string]interface{}, len(events))
		for i, event := range events {
			// Start with the base event fields
			eventData, _ := json.Marshal(event)
			json.Unmarshal(eventData, &transformedEvents[i])

			// Flatten payload for specific event types
			if event.Payload != nil {
				switch event.Type {
				case fps_events.Event_KillID:
					payloadData, _ := json.Marshal(event.Payload)
					var payloadMap map[string]interface{}
					json.Unmarshal(payloadData, &payloadMap)
					for k, v := range payloadMap {
						transformedEvents[i][k] = v
					}
					// Remove the payload field since we've flattened it
					delete(transformedEvents[i], "payload")
				default:
					// Keep payload as is for other events
				}
			}
		}

		response := map[string]interface{}{
			"replay_id":    replayID,
			"match_id":     match.ID,
			"events":       transformedEvents,
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
