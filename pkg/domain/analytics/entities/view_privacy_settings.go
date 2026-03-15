package analytics_entities

import (
	"time"

	"github.com/google/uuid"
)

// ViewPrivacySettings per-user preferences for view tracking visibility
type ViewPrivacySettings struct {
	ID                        uuid.UUID `json:"id" bson:"_id"`
	UserID                    uuid.UUID `json:"user_id" bson:"user_id"`
	ShowProfileViews          bool      `json:"show_profile_views" bson:"show_profile_views"`
	AllowViewerIdentification bool      `json:"allow_viewer_identification" bson:"allow_viewer_identification"`
	AnonymousMode             bool      `json:"anonymous_mode" bson:"anonymous_mode"`
	CreatedAt                 time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at" bson:"updated_at"`
}

func (v ViewPrivacySettings) GetID() uuid.UUID {
	return v.ID
}

// NewViewPrivacySettings creates default privacy settings for a user (all features enabled)
func NewViewPrivacySettings(userID uuid.UUID) *ViewPrivacySettings {
	now := time.Now()
	return &ViewPrivacySettings{
		ID:                        uuid.New(),
		UserID:                    userID,
		ShowProfileViews:          true,
		AllowViewerIdentification: true,
		AnonymousMode:             false,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
}
