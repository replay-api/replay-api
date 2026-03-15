package analytics_out

import (
	"context"
	"time"

	"github.com/google/uuid"
	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// EntityViewReader reads individual view records
type EntityViewReader interface {
	GetLastViewByViewer(ctx context.Context, entityID uuid.UUID, viewerID uuid.UUID, since time.Time) (*analytics_entities.EntityView, error)
	CountByEntity(ctx context.Context, entityID uuid.UUID, entityType analytics_entities.EntityTypeKey) (int64, error)
}

// ViewStatisticsReader reads pre-aggregated view statistics
type ViewStatisticsReader interface {
	GetByEntity(ctx context.Context, entityID uuid.UUID, entityType analytics_entities.EntityTypeKey) (*analytics_entities.ViewStatistics, error)
	GetByOwner(ctx context.Context, ownerID uuid.UUID, entityType *analytics_entities.EntityTypeKey) ([]analytics_entities.ViewStatistics, error)
}

// ViewerInsightReader reads viewer insight records
type ViewerInsightReader interface {
	GetByEntity(ctx context.Context, entityID uuid.UUID, entityType analytics_entities.EntityTypeKey, search shared.Search) ([]analytics_entities.ViewerInsight, int64, error)
}

// ViewPrivacyReader reads view privacy settings
type ViewPrivacyReader interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*analytics_entities.ViewPrivacySettings, error)
}
