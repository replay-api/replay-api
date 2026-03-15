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
	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_in "github.com/replay-api/replay-api/pkg/domain/analytics/ports/in"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type ViewAnalyticsQueryController struct {
	container          container.Container
	statsHandler       analytics_in.ViewStatisticsQueryHandler
	insightsHandler    analytics_in.ViewInsightsQueryHandler
	myAnalyticsHandler analytics_in.MyAnalyticsQueryHandler
}

func NewViewAnalyticsQueryController(c container.Container) *ViewAnalyticsQueryController {
	var statsHandler analytics_in.ViewStatisticsQueryHandler
	if err := c.Resolve(&statsHandler); err != nil {
		slog.Warn("ViewStatisticsQueryHandler not available", "error", err)
	}

	var insightsHandler analytics_in.ViewInsightsQueryHandler
	if err := c.Resolve(&insightsHandler); err != nil {
		slog.Warn("ViewInsightsQueryHandler not available", "error", err)
	}

	var myAnalyticsHandler analytics_in.MyAnalyticsQueryHandler
	if err := c.Resolve(&myAnalyticsHandler); err != nil {
		slog.Warn("MyAnalyticsQueryHandler not available", "error", err)
	}

	return &ViewAnalyticsQueryController{
		container:          c,
		statsHandler:       statsHandler,
		insightsHandler:    insightsHandler,
		myAnalyticsHandler: myAnalyticsHandler,
	}
}

// GetViewStatisticsHandler handles GET /{entityType}/{id}/views/statistics
func (ctrl *ViewAnalyticsQueryController) GetViewStatisticsHandler(entityType analytics_entities.EntityTypeKey) func(context.Context) http.HandlerFunc {
	return func(apiContext context.Context) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			entityID := vars["id"]
			if entityID == "" {
				entityID = vars["match_id"]
			}
			if entityID == "" {
				entityID = vars["replay_id"]
			}

			entityUUID, err := uuid.Parse(entityID)
			if err != nil {
				http.Error(w, "invalid entity id format", http.StatusBadRequest)
				return
			}

			period := r.URL.Query().Get("period")
			if period == "" {
				period = "30d"
			}

			query := analytics_in.GetViewStatisticsQuery{
				EntityID:   entityUUID,
				EntityType: entityType,
				Period:     period,
			}

			stats, err := ctrl.statsHandler.GetViewStatistics(r.Context(), query)
			if err != nil {
				slog.ErrorContext(r.Context(), "failed to get view statistics", "err", err)
				http.Error(w, "failed to get statistics", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(stats)
		}
	}
}

// GetViewInsightsHandler handles GET /{entityType}/{id}/views/insights
func (ctrl *ViewAnalyticsQueryController) GetViewInsightsHandler(entityType analytics_entities.EntityTypeKey) func(context.Context) http.HandlerFunc {
	return func(apiContext context.Context) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			isAuthenticated, _ := r.Context().Value(shared.AuthenticatedKey).(bool)
			if !isAuthenticated {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			userID, _ := r.Context().Value(shared.UserIDKey).(uuid.UUID)
			if userID == uuid.Nil {
				http.Error(w, "user identity required", http.StatusUnauthorized)
				return
			}

			vars := mux.Vars(r)
			entityID := vars["id"]
			if entityID == "" {
				entityID = vars["match_id"]
			}
			if entityID == "" {
				entityID = vars["replay_id"]
			}

			entityUUID, err := uuid.Parse(entityID)
			if err != nil {
				http.Error(w, "invalid entity id format", http.StatusBadRequest)
				return
			}

			sortBy := r.URL.Query().Get("sort")
			if sortBy == "" {
				sortBy = "recent"
			}

			limit := uint(20)
			if l := r.URL.Query().Get("limit"); l != "" {
				if parsed, err := strconv.ParseUint(l, 10, 32); err == nil {
					limit = uint(parsed)
				}
			}

			skip := uint(0)
			if s := r.URL.Query().Get("offset"); s != "" {
				if parsed, err := strconv.ParseUint(s, 10, 32); err == nil {
					skip = uint(parsed)
				}
			}

			query := analytics_in.GetViewInsightsQuery{
				EntityID:   entityUUID,
				EntityType: entityType,
				OwnerID:    userID,
				Skip:       skip,
				Limit:      limit,
				SortBy:     sortBy,
			}

			insights, total, err := ctrl.insightsHandler.GetViewInsights(r.Context(), query)
			if err != nil {
				slog.ErrorContext(r.Context(), "failed to get view insights", "err", err)
				http.Error(w, "failed to get insights", http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"data":  insights,
				"total": total,
				"skip":  skip,
				"limit": limit,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}
	}
}

// GetMyAnalyticsHandler handles GET /me/analytics/views
func (ctrl *ViewAnalyticsQueryController) GetMyAnalyticsHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isAuthenticated, _ := r.Context().Value(shared.AuthenticatedKey).(bool)
		if !isAuthenticated {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		userID, _ := r.Context().Value(shared.UserIDKey).(uuid.UUID)
		if userID == uuid.Nil {
			http.Error(w, "user identity required", http.StatusUnauthorized)
			return
		}

		period := r.URL.Query().Get("period")
		if period == "" {
			period = "30d"
		}

		var entityType *analytics_entities.EntityTypeKey
		if et := r.URL.Query().Get("entity_type"); et != "" {
			t := analytics_entities.EntityTypeKey(et)
			entityType = &t
		}

		query := analytics_in.GetMyAnalyticsQuery{
			UserID:     userID,
			EntityType: entityType,
			Period:     period,
		}

		stats, err := ctrl.myAnalyticsHandler.GetMyAnalytics(r.Context(), query)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get my analytics", "err", err)
			http.Error(w, "failed to get analytics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(stats)
	}
}

// GetViewPrivacyHandler handles GET /me/settings/view-privacy
func (ctrl *ViewAnalyticsQueryController) GetViewPrivacyHandler(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isAuthenticated, _ := r.Context().Value(shared.AuthenticatedKey).(bool)
		if !isAuthenticated {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		userID, _ := r.Context().Value(shared.UserIDKey).(uuid.UUID)

		// Return defaults — view privacy settings will be created on first update
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":                     userID,
			"show_profile_views":          true,
			"allow_viewer_identification": true,
			"anonymous_mode":              false,
		})
	}
}
