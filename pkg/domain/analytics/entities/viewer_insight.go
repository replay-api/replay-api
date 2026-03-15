package analytics_entities

import (
	"time"

	"github.com/google/uuid"
)

// ViewerInsight represents aggregated data about a specific viewer of an entity
type ViewerInsight struct {
	ID             uuid.UUID     `json:"id" bson:"_id"`
	EntityID       uuid.UUID     `json:"entity_id" bson:"entity_id"`
	EntityType     EntityTypeKey `json:"entity_type" bson:"entity_type"`
	ViewerID       uuid.UUID     `json:"viewer_id" bson:"viewer_id"`
	ViewerNickname string        `json:"viewer_nickname" bson:"viewer_nickname"`
	ViewerAvatar   string        `json:"viewer_avatar" bson:"viewer_avatar"`
	ViewerGameID   string        `json:"viewer_game_id" bson:"viewer_game_id"`
	ViewCount      int           `json:"view_count" bson:"view_count"`
	FirstViewedAt  time.Time     `json:"first_viewed_at" bson:"first_viewed_at"`
	LastViewedAt   time.Time     `json:"last_viewed_at" bson:"last_viewed_at"`
	IsAnonymous    bool          `json:"is_anonymous" bson:"is_anonymous"`
	CreatedAt      time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" bson:"updated_at"`
}

func (v ViewerInsight) GetID() uuid.UUID {
	return v.ID
}

// NewViewerInsight creates a new insight record for a viewer-entity pair
func NewViewerInsight(
	entityID uuid.UUID,
	entityType EntityTypeKey,
	viewerID uuid.UUID,
	viewerNickname string,
	viewerAvatar string,
	viewerGameID string,
	isAnonymous bool,
) *ViewerInsight {
	now := time.Now()
	return &ViewerInsight{
		ID:             uuid.New(),
		EntityID:       entityID,
		EntityType:     entityType,
		ViewerID:       viewerID,
		ViewerNickname: viewerNickname,
		ViewerAvatar:   viewerAvatar,
		ViewerGameID:   viewerGameID,
		ViewCount:      1,
		FirstViewedAt:  now,
		LastViewedAt:   now,
		IsAnonymous:    isAnonymous,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// IncrementView bumps the view count and updates last viewed time
func (v *ViewerInsight) IncrementView() {
	v.ViewCount++
	v.LastViewedAt = time.Now()
	v.UpdatedAt = time.Now()
}
