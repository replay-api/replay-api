package entities

import (
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type PlayerMetadata struct {
	ID              shared.PlayerIDType          `json:"id" bson:"_id"`
	VisibilityLevel shared.IntendedAudienceKey   `json:"visibility_level" bson:"visibility_level"`
	VisibilityType  shared.VisibilityTypeKey     `json:"visibility_type" bson:"visibility_type"`
	ResourceOwner   shared.ResourceOwner         `json:"-" bson:"resource_owner"`
	CreatedAt       time.Time                    `json:"-" bson:"created_at"`
	UpdatedAt       *time.Time                   `json:"-" bson:"updated_at"`
	GameID          replay_common.GameIDKey      `json:"game_id" bson:"game_id"`
	UserID          *uuid.UUID                   `json:"-" bson:"user_id"`
	NetworkUserID   string                       `json:"-" bson:"network_user_id"`
	NetworkID       replay_common.NetworkIDKey   `json:"network_id" bson:"network_id"`
	Name            string                       `json:"name" bson:"name"`
	NameHistory     []string                     `json:"-" bson:"name_history"`
	ClanName        string                       `json:"clan_name" bson:"clan_name"`
	AvatarURI       string                       `json:"avatar_uri" bson:"avatar_uri"`
	NetworkClanID   string                       `json:"network_clan_id" bson:"network_clan_id"`
	VerifiedAt      *time.Time                   `json:"verified_at" bson:"verified_at"`
	ShareTokens     []ShareToken                 `json:"-" bson:"share_tokens"`
}

func NewPlayerMetadata(currentName string, networkUserID string, networkID replay_common.NetworkIDKey, clanName string, res shared.ResourceOwner) *PlayerMetadata {
	entity := shared.NewUnrestrictedEntity(res)
	return &PlayerMetadata{
		ID:              shared.PlayerIDType(entity.ID),
		VisibilityLevel: entity.VisibilityLevel,
		VisibilityType:  entity.VisibilityType,
		ResourceOwner:   res,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       nil,
		GameID:          replay_common.CS2.ID,
		UserID:          nil,
		NetworkID:       networkID,
		NetworkUserID:   networkUserID,
		Name:            currentName,
		ClanName:        clanName,
		NameHistory:     []string{},
		NetworkClanID:   "",
		AvatarURI:       "",
		VerifiedAt:      nil,
		ShareTokens:     []ShareToken{},
	}
}

func (e PlayerMetadata) GetID() uuid.UUID {
	return uuid.UUID(e.ID)
}
