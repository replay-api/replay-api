package iam_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type RIDSourceKey string

const (
	RIDSource_Steam  RIDSourceKey = "steam"
	RIDSource_Google RIDSourceKey = "google"
	RIDSource_Email  RIDSourceKey = "email"
	RIDSource_Guest  RIDSourceKey = "guest"
)

const (
	DefaultTokenAudience = shared.UserAudienceIDKey
)

type RIDToken struct {
	ID               uuid.UUID                  `json:"-" bson:"_id"`
	Key              uuid.UUID                  `json:"-" bson:"key"` // deprecated TODO: delete
	Source           RIDSourceKey               `json:"-" bson:"source"`
	ResourceOwner    shared.ResourceOwner       `json:"-" bson:"resource_owner"`
	IntendedAudience shared.IntendedAudienceKey `json:"-" bson:"intended_audience"`
	GrantType        string                     `json:"-" bson:"grant_type"`
	ExpiresAt        time.Time                  `json:"-" bson:"expires_at"`
	RevokedAt        *time.Time                 `json:"-" bson:"revoked_at,omitempty"`
	CreatedAt        time.Time                  `json:"-" bson:"created_at"`
	UpdatedAt        time.Time                  `json:"-" bson:"updated_at"`
}

func (t RIDToken) GetID() uuid.UUID {
	return t.ID
}

func (t RIDToken) IsExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}

func (t RIDToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t RIDToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}
