package query_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	achievement_entities "github.com/replay-api/replay-api/pkg/domain/achievement/entities"
	achievement_in "github.com/replay-api/replay-api/pkg/domain/achievement/ports/in"
)

// AchievementQueryController handles achievement query HTTP requests
type AchievementQueryController struct {
	container        container.Container
	achievementQuery achievement_in.AchievementQuery
}

// NewAchievementQueryController creates a new AchievementQueryController
func NewAchievementQueryController(c container.Container) *AchievementQueryController {
	ctrl := &AchievementQueryController{container: c}

	if err := c.Resolve(&ctrl.achievementQuery); err != nil {
		slog.Warn("AchievementQuery not available", "error", err)
	}

	return ctrl
}

// GetPlayerAchievementsHandler handles GET /achievements/player/:player_id
func (ctrl *AchievementQueryController) GetPlayerAchievementsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.achievementQuery == nil {
			http.Error(w, `{"error":"achievement service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		playerIDStr := vars["player_id"]

		playerID, err := uuid.Parse(playerIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid player_id"}`, http.StatusBadRequest)
			return
		}

		achievements, err := ctrl.achievementQuery.GetPlayerAchievements(ctx, playerID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get player achievements", "error", err)
			http.Error(w, `{"error":"failed to get achievements"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"achievements": achievements,
			"count":        len(achievements),
			"player_id":    playerID,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetPlayerAchievementSummaryHandler handles GET /achievements/player/:player_id/summary
func (ctrl *AchievementQueryController) GetPlayerAchievementSummaryHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.achievementQuery == nil {
			http.Error(w, `{"error":"achievement service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		playerIDStr := vars["player_id"]

		playerID, err := uuid.Parse(playerIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid player_id"}`, http.StatusBadRequest)
			return
		}

		summary, err := ctrl.achievementQuery.GetPlayerAchievementSummary(ctx, playerID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get player achievement summary", "error", err)
			http.Error(w, `{"error":"failed to get achievement summary"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(summary)
	}
}

// GetAllAchievementsHandler handles GET /achievements
func (ctrl *AchievementQueryController) GetAllAchievementsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.achievementQuery == nil {
			http.Error(w, `{"error":"achievement service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Optional category filter
		category := r.URL.Query().Get("category")
		
		var achievements []achievement_entities.Achievement
		var err error

		if category != "" {
			achievements, err = ctrl.achievementQuery.GetAchievementsByCategory(ctx, achievement_entities.AchievementCategory(category))
		} else {
			achievements, err = ctrl.achievementQuery.GetAllAchievements(ctx)
		}

		if err != nil {
			slog.ErrorContext(ctx, "Failed to get achievements", "error", err)
			http.Error(w, `{"error":"failed to get achievements"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"achievements": achievements,
			"count":        len(achievements),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetAchievementByIDHandler handles GET /achievements/:id
func (ctrl *AchievementQueryController) GetAchievementByIDHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.achievementQuery == nil {
			http.Error(w, `{"error":"achievement service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		idStr := vars["id"]

		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, `{"error":"invalid achievement id"}`, http.StatusBadRequest)
			return
		}

		achievement, err := ctrl.achievementQuery.GetAchievementByID(ctx, id)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get achievement", "error", err)
			http.Error(w, `{"error":"failed to get achievement"}`, http.StatusInternalServerError)
			return
		}

		if achievement == nil {
			http.Error(w, `{"error":"achievement not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(achievement)
	}
}

// GetRecentUnlocksHandler handles GET /achievements/player/:player_id/recent
func (ctrl *AchievementQueryController) GetRecentUnlocksHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.achievementQuery == nil {
			http.Error(w, `{"error":"achievement service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		vars := mux.Vars(r)
		playerIDStr := vars["player_id"]

		playerID, err := uuid.Parse(playerIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid player_id"}`, http.StatusBadRequest)
			return
		}

		// Parse limit from query params, default to 10
		limit := 10
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
				limit = parsed
			}
		}

		unlocks, err := ctrl.achievementQuery.GetRecentUnlocks(ctx, playerID, limit)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get recent unlocks", "error", err)
			http.Error(w, `{"error":"failed to get recent unlocks"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"unlocks":   unlocks,
			"count":     len(unlocks),
			"player_id": playerID,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// GetMyAchievementsHandler handles GET /me/achievements (authenticated user's achievements)
func (ctrl *AchievementQueryController) GetMyAchievementsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if ctrl.achievementQuery == nil {
			http.Error(w, `{"error":"achievement service unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		// Check authentication
		authenticated, ok := ctx.Value(common.AuthenticatedKey).(bool)
		if !ok || !authenticated {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}

		resourceOwner := common.GetResourceOwner(ctx)
		if resourceOwner.UserID == uuid.Nil {
			http.Error(w, `{"error":"valid user required"}`, http.StatusUnauthorized)
			return
		}

		// Get player ID from user ID (assuming they're the same for now)
		playerID := resourceOwner.UserID

		achievements, err := ctrl.achievementQuery.GetPlayerAchievements(ctx, playerID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get player achievements", "error", err)
			http.Error(w, `{"error":"failed to get achievements"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"achievements": achievements,
			"count":        len(achievements),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}



