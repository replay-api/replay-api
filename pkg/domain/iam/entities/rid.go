package iam_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type RIDSourceKey string

const (
	RIDSource_Steam  RIDSourceKey = "steam"
	RIDSource_Google RIDSourceKey = "google"
	RIDSource_Email  RIDSourceKey = "email"
)

const (
	DefaultTokenAudience = common.UserAudienceIDKey
)

type RIDToken struct {
	ID               uuid.UUID                  `json:"-" bson:"_id"`
	Key              uuid.UUID                  `json:"-" bson:"key"` // deprecated TODO: delete
	Source           RIDSourceKey               `json:"-" bson:"source"`
	ResourceOwner    common.ResourceOwner       `json:"-" bson:"resource_owner"`
	IntendedAudience common.IntendedAudienceKey `json:"-" bson:"intended_audience"`
	GrantType        string                     `json:"-" bson:"grant_type"`
	ExpiresAt        time.Time                  `json:"-" bson:"expires_at"`
	CreatedAt        time.Time                  `json:"-" bson:"created_at"`
	UpdatedAt        time.Time                  `json:"-" bson:"updated_at"`
}

func (t RIDToken) GetID() uuid.UUID {
	return t.ID
}

func (t RIDToken) IsExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}
