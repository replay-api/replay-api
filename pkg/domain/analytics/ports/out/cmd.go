package analytics_out

import (
	"context"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
)

// EntityViewWriter persists individual view records
type EntityViewWriter interface {
	Create(ctx context.Context, view *analytics_entities.EntityView) (*analytics_entities.EntityView, error)
}

// ViewStatisticsWriter persists pre-aggregated view statistics
type ViewStatisticsWriter interface {
	Upsert(ctx context.Context, stats *analytics_entities.ViewStatistics) error
}

// ViewerInsightWriter persists viewer insight records
type ViewerInsightWriter interface {
	Upsert(ctx context.Context, insight *analytics_entities.ViewerInsight) error
}

// ViewPrivacyWriter persists view privacy settings
type ViewPrivacyWriter interface {
	Upsert(ctx context.Context, settings *analytics_entities.ViewPrivacySettings) error
}

// ViewEventPublisher publishes view events to the message broker
type ViewEventPublisher interface {
	PublishEntityViewed(ctx context.Context, view *analytics_entities.EntityView) error
}
