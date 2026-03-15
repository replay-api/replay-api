package analytics_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// EntityTypeKey identifies the type of entity being viewed
type EntityTypeKey string

const (
	EntityTypePlayer EntityTypeKey = "player"
	EntityTypeTeam   EntityTypeKey = "team"
	EntityTypeMatch  EntityTypeKey = "match"
	EntityTypeReplay EntityTypeKey = "replay"
)

// ViewerTypeKey identifies the type of viewer
type ViewerTypeKey string

const (
	ViewerTypeAnonymous     ViewerTypeKey = "anonymous"
	ViewerTypeAuthenticated ViewerTypeKey = "authenticated"
	ViewerTypeBot           ViewerTypeKey = "bot"
)

// ReferrerTypeKey categorizes how the viewer arrived
type ReferrerTypeKey string

const (
	ReferrerTypeDirect ReferrerTypeKey = "direct"
	ReferrerTypeSearch ReferrerTypeKey = "search"
	ReferrerTypeLink   ReferrerTypeKey = "link"
	ReferrerTypeEmbed  ReferrerTypeKey = "embed"
)

// DeviceTypeKey categorizes the viewer's device
type DeviceTypeKey string

const (
	DeviceTypeDesktop DeviceTypeKey = "desktop"
	DeviceTypeMobile  DeviceTypeKey = "mobile"
	DeviceTypeTablet  DeviceTypeKey = "tablet"
	DeviceTypeUnknown DeviceTypeKey = "unknown"
)

// EntityView represents a single view event on any entity
type EntityView struct {
	shared.BaseEntity
	EntityID     uuid.UUID       `json:"entity_id" bson:"entity_id"`
	EntityType   EntityTypeKey   `json:"entity_type" bson:"entity_type"`
	ViewerID     *uuid.UUID      `json:"viewer_id,omitempty" bson:"viewer_id,omitempty"`
	ViewerType   ViewerTypeKey   `json:"viewer_type" bson:"viewer_type"`
	SessionID    string          `json:"session_id" bson:"session_id"`
	ReferrerType ReferrerTypeKey `json:"referrer_type" bson:"referrer_type"`
	GeoRegion    string          `json:"geo_region" bson:"geo_region"`
	DeviceType   DeviceTypeKey   `json:"device_type" bson:"device_type"`
	ViewedAt     time.Time       `json:"viewed_at" bson:"viewed_at"`
}

func (v EntityView) GetID() uuid.UUID {
	return v.BaseEntity.ID
}

// NewEntityView creates a new view record
func NewEntityView(
	entityID uuid.UUID,
	entityType EntityTypeKey,
	viewerID *uuid.UUID,
	viewerType ViewerTypeKey,
	sessionID string,
	referrerType ReferrerTypeKey,
	geoRegion string,
	deviceType DeviceTypeKey,
	rxn shared.ResourceOwner,
) *EntityView {
	return &EntityView{
		BaseEntity:   shared.NewUnrestrictedEntity(rxn),
		EntityID:     entityID,
		EntityType:   entityType,
		ViewerID:     viewerID,
		ViewerType:   viewerType,
		SessionID:    sessionID,
		ReferrerType: referrerType,
		GeoRegion:    geoRegion,
		DeviceType:   deviceType,
		ViewedAt:     time.Now(),
	}
}
