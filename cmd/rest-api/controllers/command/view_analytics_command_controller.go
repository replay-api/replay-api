package cmd_controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_in "github.com/replay-api/replay-api/pkg/domain/analytics/ports/in"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type ViewAnalyticsCommandController struct {
	container            container.Container
	recordViewUseCase    analytics_in.RecordViewCommandHandler
	updatePrivacyUseCase analytics_in.UpdateViewPrivacyCommandHandler
}

func NewViewAnalyticsCommandController(c container.Container) *ViewAnalyticsCommandController {
	var recordViewUseCase analytics_in.RecordViewCommandHandler
	if err := c.Resolve(&recordViewUseCase); err != nil {
		slog.Warn("RecordViewCommandHandler not available", "error", err)
	}

	var updatePrivacyUseCase analytics_in.UpdateViewPrivacyCommandHandler
	if err := c.Resolve(&updatePrivacyUseCase); err != nil {
		slog.Warn("UpdateViewPrivacyCommandHandler not available", "error", err)
	}

	return &ViewAnalyticsCommandController{
		container:            c,
		recordViewUseCase:    recordViewUseCase,
		updatePrivacyUseCase: updatePrivacyUseCase,
	}
}

// RecordViewHandler handles POST /{entityType}/{id}/views
func (ctrl *ViewAnalyticsCommandController) RecordViewHandler(entityType analytics_entities.EntityTypeKey) func(context.Context) http.HandlerFunc {
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

			if entityID == "" {
				http.Error(w, "entity id is required", http.StatusBadRequest)
				return
			}

			entityUUID, err := uuid.Parse(entityID)
			if err != nil {
				http.Error(w, "invalid entity id format", http.StatusBadRequest)
				return
			}

			var reqBody struct {
				SessionID    string `json:"session_id"`
				ReferrerType string `json:"referrer_type"`
			}
			_ = json.NewDecoder(r.Body).Decode(&reqBody)

			// Extract viewer identity from context
			var viewerID *uuid.UUID
			isAuthenticated, _ := r.Context().Value(shared.AuthenticatedKey).(bool)
			if isAuthenticated {
				if uid, ok := r.Context().Value(shared.UserIDKey).(uuid.UUID); ok && uid != uuid.Nil {
					viewerID = &uid
				}
			}

			// Determine device type from User-Agent
			deviceType := parseDeviceType(r.UserAgent())

			// Determine referrer type
			referrerType := analytics_entities.ReferrerTypeKey(reqBody.ReferrerType)
			if referrerType == "" {
				referrerType = parseReferrerType(r.Referer())
			}

			// Geo region from header (set by CDN/proxy)
			geoRegion := r.Header.Get("CF-IPCountry")
			if geoRegion == "" {
				geoRegion = r.Header.Get("X-Geo-Region")
			}

			cmd := analytics_in.RecordViewCommand{
				EntityID:     entityUUID,
				EntityType:   entityType,
				ViewerID:     viewerID,
				SessionID:    reqBody.SessionID,
				ReferrerType: referrerType,
				DeviceType:   deviceType,
				GeoRegion:    geoRegion,
			}

			if err := ctrl.recordViewUseCase.Exec(r.Context(), cmd); err != nil {
				slog.ErrorContext(r.Context(), "failed to record view", "err", err, "entity_id", entityUUID)
				// Still return 202 — view recording should not block the caller
			}

			w.WriteHeader(http.StatusAccepted)
		}
	}
}

// UpdateViewPrivacyHandler handles PUT /me/settings/view-privacy
func (ctrl *ViewAnalyticsCommandController) UpdateViewPrivacyHandler(apiContext context.Context) http.HandlerFunc {
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

		var cmd analytics_in.UpdateViewPrivacyCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		cmd.UserID = userID

		settings, err := ctrl.updatePrivacyUseCase.Exec(r.Context(), cmd)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to update view privacy", "err", err, "user_id", userID)
			http.Error(w, "failed to update settings", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(settings)
	}
}

func parseDeviceType(userAgent string) analytics_entities.DeviceTypeKey {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone"):
		return analytics_entities.DeviceTypeMobile
	case strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad"):
		return analytics_entities.DeviceTypeTablet
	case strings.Contains(ua, "mozilla") || strings.Contains(ua, "chrome") || strings.Contains(ua, "safari"):
		return analytics_entities.DeviceTypeDesktop
	default:
		return analytics_entities.DeviceTypeUnknown
	}
}

func parseReferrerType(referer string) analytics_entities.ReferrerTypeKey {
	if referer == "" {
		return analytics_entities.ReferrerTypeDirect
	}
	ref := strings.ToLower(referer)
	switch {
	case strings.Contains(ref, "google") || strings.Contains(ref, "bing") || strings.Contains(ref, "duckduckgo"):
		return analytics_entities.ReferrerTypeSearch
	default:
		return analytics_entities.ReferrerTypeLink
	}
}
