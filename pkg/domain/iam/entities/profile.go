package iam_entities

import (
	"context"
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type Profile struct {
	ID              uuid.UUID                  `json:"id" bson:"_id"`
	RIDSource       RIDSourceKey               `json:"rid_source" bson:"rid_source"`
	SourceKey       string                     `json:"source_key" bson:"source_key"` // ie. steam id, google@, etc
	Details         interface{}                `json:"details" bson:"details"`       // TODO: deprecate. GET /profile/:id/details => mux para steam,google,squad,player
	Links           map[string]string          `json:"links" bson:"links"`
	VisibilityLevel common.IntendedAudienceKey `json:"visibility_level" bson:"visibility_level"`
	VisbilityType   common.VisibilityTypeKey   `json:"visibility_type" bson:"visibility_type"`
	ResourceOwner   common.ResourceOwner       `json:"resource_owner" bson:"resource_owner"`
	CreatedAt       time.Time                  `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at" bson:"updated_at"`
}

func NewProfile(userID uuid.UUID, groupID uuid.UUID, ridSource RIDSourceKey, sourceKey string, details interface{}, resourceOwner common.ResourceOwner) *Profile {
	resourceOwner.UserID = userID
	resourceOwner.GroupID = groupID

	return &Profile{
		ID:            uuid.New(),
		RIDSource:     ridSource,
		SourceKey:     sourceKey,
		Details:       details,
		ResourceOwner: resourceOwner,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (p *Profile) GetContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, common.GroupIDKey, p.ResourceOwner.GroupID)
	ctx = context.WithValue(ctx, common.UserIDKey, p.ResourceOwner.UserID)

	return ctx
}

func (p *Profile) GetResourceOwner(ctx context.Context) common.ResourceOwner {
	ctx = context.WithValue(ctx, common.GroupIDKey, p.ResourceOwner.GroupID)
	ctx = context.WithValue(ctx, common.UserIDKey, p.ResourceOwner.UserID)

	return common.GetResourceOwner(ctx)
}

// func (p *Profile) GetID() uuid.UUID {
// 	return p.ID
// }

func (p Profile) GetID() uuid.UUID {
	return p.ID
}
