package iam_entities

import (
	"context"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type ProfileLinkType string

const (
	ProfileLinkTypeSteam     ProfileLinkType = "steam"
	ProfileLinkTypeSquad     ProfileLinkType = "instagram"
	ProfileLinkTypeGoogle    ProfileLinkType = "google"
	ProfileLinkTypeTwitch    ProfileLinkType = "twitch"
	ProfileLinkTypeTwitter   ProfileLinkType = "twitter"
	ProfileLinkTypeInstagram ProfileLinkType = "linkedin"
	ProfileLinkTypeFacebook  ProfileLinkType = "facebook"
	ProfileLinkTypeYoutube   ProfileLinkType = "youtube"
	ProfileLinkTypeDiscord   ProfileLinkType = "discord"
	ProfileLinkTypeWebsite   ProfileLinkType = "website"
	ProfileLinkTypeOther     ProfileLinkType = "user-defined"
)

type ProfileType string

const (
	ProfileTypeSteam  ProfileType = "account" // steam, google
	ProfileTypeSquad  ProfileType = "squad"
	ProfileTypePlayer ProfileType = "player"
)

type Profile struct {
	ID              uuid.UUID                  `json:"id" bson:"_id"`
	RIDSource       RIDSourceKey               `json:"rid_source" bson:"rid_source"`
	SourceKey       string                     `json:"source_key" bson:"source_key"` // ie. steam id, google@, etc
	Details         interface{}                `json:"details" bson:"details"`       // TODO: deprecate. GET /profile/:id/details => mux para steam,google,squad,player
	Links           map[ProfileLinkType]string `json:"links" bson:"links"`
	VisibilityLevel shared.IntendedAudienceKey `json:"visibility_level" bson:"visibility_level"`
	VisibilityType  shared.VisibilityTypeKey   `json:"visibility_type" bson:"visibility_type"`
	Type            ProfileType                `json:"type" bson:"type"` // ie. steam, google, team/squad, player
	ResourceOwner   shared.ResourceOwner       `json:"resource_owner" bson:"resource_owner"`
	CreatedAt       time.Time                  `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at" bson:"updated_at"`
}

func NewProfile(userID uuid.UUID, groupID uuid.UUID, ridSource RIDSourceKey, sourceKey string, details interface{}, resourceOwner shared.ResourceOwner) *Profile {
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
	ctx = context.WithValue(ctx, shared.GroupIDKey, p.ResourceOwner.GroupID)
	ctx = context.WithValue(ctx, shared.UserIDKey, p.ResourceOwner.UserID)

	return ctx
}

func (p *Profile) GetResourceOwner(ctx context.Context) shared.ResourceOwner {
	ctx = context.WithValue(ctx, shared.GroupIDKey, p.ResourceOwner.GroupID)
	ctx = context.WithValue(ctx, shared.UserIDKey, p.ResourceOwner.UserID)

	return shared.GetResourceOwner(ctx)
}

// func (p *Profile) GetID() uuid.UUID {
// 	return p.ID
// }

func (p Profile) GetID() uuid.UUID {
	return p.ID
}
