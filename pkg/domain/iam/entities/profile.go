package iam_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type Profile struct {
	ID            uuid.UUID              `json:"id" bson:"_id"`
	RIDSource     RIDSourceKey           `json:"rid_source" bson:"rid_source"`
	SourceKey     string                 `json:"source_key" bson:"source_key"` // ie. steam id, google@, etc
	Details       map[string]interface{} `json:"details" bson:"details"`
	ResourceOwner common.ResourceOwner   `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" bson:"updated_at"`
}

func (p Profile) GetID() uuid.UUID {
	return p.ID
}
