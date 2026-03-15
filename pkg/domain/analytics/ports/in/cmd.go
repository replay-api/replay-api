package analytics_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
)

// RecordViewCommand is the command to record an entity view
type RecordViewCommand struct {
	EntityID     uuid.UUID                         `json:"entity_id"`
	EntityType   analytics_entities.EntityTypeKey   `json:"entity_type"`
	ViewerID     *uuid.UUID                        `json:"viewer_id,omitempty"`
	SessionID    string                            `json:"session_id"`
	ReferrerType analytics_entities.ReferrerTypeKey `json:"referrer_type"`
	DeviceType   analytics_entities.DeviceTypeKey   `json:"device_type"`
	GeoRegion    string                            `json:"geo_region"`
}

// Validate validates the RecordViewCommand
func (c *RecordViewCommand) Validate() error {
	if c.EntityID == uuid.Nil {
		return errors.New("entity_id is required")
	}
	if c.EntityType == "" {
		return errors.New("entity_type is required")
	}
	switch c.EntityType {
	case analytics_entities.EntityTypePlayer,
		analytics_entities.EntityTypeTeam,
		analytics_entities.EntityTypeMatch,
		analytics_entities.EntityTypeReplay:
		// valid
	default:
		return errors.New("entity_type must be one of: player, team, match, replay")
	}
	return nil
}

// RecordViewCommandHandler records a view event
type RecordViewCommandHandler interface {
	Exec(ctx context.Context, cmd RecordViewCommand) error
}

// UpdateViewPrivacyCommand updates view privacy settings
type UpdateViewPrivacyCommand struct {
	UserID                    uuid.UUID `json:"user_id"`
	ShowProfileViews          *bool     `json:"show_profile_views,omitempty"`
	AllowViewerIdentification *bool     `json:"allow_viewer_identification,omitempty"`
	AnonymousMode             *bool     `json:"anonymous_mode,omitempty"`
}

// Validate validates the UpdateViewPrivacyCommand
func (c *UpdateViewPrivacyCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	return nil
}

// UpdateViewPrivacyCommandHandler updates view privacy settings
type UpdateViewPrivacyCommandHandler interface {
	Exec(ctx context.Context, cmd UpdateViewPrivacyCommand) (*analytics_entities.ViewPrivacySettings, error)
}
