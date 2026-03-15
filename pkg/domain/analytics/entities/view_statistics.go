package analytics_entities

import (
	"time"

	"github.com/google/uuid"
)

// TrendKey indicates the direction of a metric trend
type TrendKey string

const (
	TrendUp     TrendKey = "up"
	TrendDown   TrendKey = "down"
	TrendStable TrendKey = "stable"
)

// ViewStatistics holds pre-aggregated view metrics for a single entity
type ViewStatistics struct {
	ID                 uuid.UUID        `json:"id" bson:"_id"`
	EntityID           uuid.UUID        `json:"entity_id" bson:"entity_id"`
	EntityType         EntityTypeKey    `json:"entity_type" bson:"entity_type"`
	TotalViews         int64            `json:"total_views" bson:"total_views"`
	UniqueViewers      int64            `json:"unique_viewers" bson:"unique_viewers"`
	AuthenticatedViews int64            `json:"authenticated_views" bson:"authenticated_views"`
	AnonymousViews     int64            `json:"anonymous_views" bson:"anonymous_views"`
	ViewsByDay         map[string]int64 `json:"views_by_day" bson:"views_by_day"`
	ViewsByRegion      map[string]int64 `json:"views_by_region" bson:"views_by_region"`
	ViewsByDevice      map[string]int64 `json:"views_by_device" bson:"views_by_device"`
	ViewsByReferrer    map[string]int64 `json:"views_by_referrer" bson:"views_by_referrer"`
	TrendDirection     TrendKey         `json:"trend_direction" bson:"trend_direction"`
	TrendPercentage    float64          `json:"trend_percentage" bson:"trend_percentage"`
	PeakViewsDay       string           `json:"peak_views_day" bson:"peak_views_day"`
	LastComputedAt     time.Time        `json:"last_computed_at" bson:"last_computed_at"`
	CreatedAt          time.Time        `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" bson:"updated_at"`
}

func (v ViewStatistics) GetID() uuid.UUID {
	return v.ID
}

// NewViewStatistics creates a blank statistics document for an entity
func NewViewStatistics(entityID uuid.UUID, entityType EntityTypeKey) *ViewStatistics {
	now := time.Now()
	return &ViewStatistics{
		ID:              uuid.New(),
		EntityID:        entityID,
		EntityType:      entityType,
		TotalViews:      0,
		UniqueViewers:   0,
		ViewsByDay:      make(map[string]int64),
		ViewsByRegion:   make(map[string]int64),
		ViewsByDevice:   make(map[string]int64),
		ViewsByReferrer: make(map[string]int64),
		TrendDirection:  TrendStable,
		LastComputedAt:  now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
