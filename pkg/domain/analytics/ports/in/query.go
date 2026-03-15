package analytics_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
)

// GetViewStatisticsQuery fetches pre-aggregated statistics for an entity
type GetViewStatisticsQuery struct {
	EntityID   uuid.UUID                       `json:"entity_id"`
	EntityType analytics_entities.EntityTypeKey `json:"entity_type"`
	Period     string                          `json:"period"` // 7d, 30d, 90d
}

// Validate validates the GetViewStatisticsQuery
func (q *GetViewStatisticsQuery) Validate() error {
	if q.EntityID == uuid.Nil {
		return errors.New("entity_id is required")
	}
	if q.EntityType == "" {
		return errors.New("entity_type is required")
	}
	return nil
}

// GetViewInsightsQuery fetches viewer insights (who viewed) for an entity
type GetViewInsightsQuery struct {
	EntityID   uuid.UUID                       `json:"entity_id"`
	EntityType analytics_entities.EntityTypeKey `json:"entity_type"`
	OwnerID    uuid.UUID                       `json:"owner_id"`
	Skip       uint                            `json:"skip"`
	Limit      uint                            `json:"limit"`
	SortBy     string                          `json:"sort_by"` // "recent" or "frequent"
}

// Validate validates the GetViewInsightsQuery
func (q *GetViewInsightsQuery) Validate() error {
	if q.EntityID == uuid.Nil {
		return errors.New("entity_id is required")
	}
	if q.EntityType == "" {
		return errors.New("entity_type is required")
	}
	if q.OwnerID == uuid.Nil {
		return errors.New("owner_id is required")
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	return nil
}

// GetMyAnalyticsQuery fetches aggregated view stats across all entities owned by a user
type GetMyAnalyticsQuery struct {
	UserID     uuid.UUID                        `json:"user_id"`
	EntityType *analytics_entities.EntityTypeKey `json:"entity_type,omitempty"`
	Period     string                           `json:"period"` // 7d, 30d, 90d
}

// Validate validates the GetMyAnalyticsQuery
func (q *GetMyAnalyticsQuery) Validate() error {
	if q.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	return nil
}

// ViewStatisticsQueryHandler reads view statistics
type ViewStatisticsQueryHandler interface {
	GetViewStatistics(ctx context.Context, query GetViewStatisticsQuery) (*analytics_entities.ViewStatistics, error)
}

// ViewInsightsQueryHandler reads viewer insights
type ViewInsightsQueryHandler interface {
	GetViewInsights(ctx context.Context, query GetViewInsightsQuery) ([]analytics_entities.ViewerInsight, int64, error)
}

// MyAnalyticsQueryHandler reads aggregated user analytics
type MyAnalyticsQueryHandler interface {
	GetMyAnalytics(ctx context.Context, query GetMyAnalyticsQuery) ([]analytics_entities.ViewStatistics, error)
}
