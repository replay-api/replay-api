package iam_entities

import (
	"context"

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
	shared.BaseEntity `json:",inline" bson:",inline"`
	RIDSource         RIDSourceKey               `json:"rid_source" bson:"rid_source"`
	SourceKey         string                     `json:"source_key" bson:"source_key"` // ie. steam id, google@, etc
	Details           interface{}                `json:"details" bson:"details"`       // TODO: deprecate. GET /profile/:id/details => mux para steam,google,squad,player
	Links             map[ProfileLinkType]string `json:"links" bson:"links"`
	Type              ProfileType                `json:"type" bson:"type"` // ie. steam, google, team/squad, player
}

func NewProfile(userID uuid.UUID, groupID uuid.UUID, ridSource RIDSourceKey, sourceKey string, details interface{}, resourceOwner shared.ResourceOwner) *Profile {
	resourceOwner.UserID = userID
	resourceOwner.GroupID = groupID
	entity := shared.NewEntity(resourceOwner)
	return &Profile{
		BaseEntity: entity,
		RIDSource:  ridSource,
		SourceKey:  sourceKey,
		Details:    details,
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
