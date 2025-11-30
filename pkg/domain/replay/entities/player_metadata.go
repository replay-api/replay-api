package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type PlayerMetadata struct {
	ID            common.PlayerIDType `json:"id" bson:"_id"`
	GameID        common.GameIDKey    `json:"game_id" bson:"game_id"`
	UserID        *uuid.UUID          `json:"-" bson:"user_id"`
	NetworkUserID string              `json:"-" bson:"network_user_id"`
	NetworkID     common.NetworkIDKey `json:"network_id" bson:"network_id"`
	Name          string              `json:"name" bson:"name"`
	NameHistory   []string            `json:"-" bson:"name_history"`
	ClanName      string              `json:"clan_name" bson:"clan_name"`
	AvatarURI     string              `json:"avatar_uri" bson:"avatar_uri"`

	NetworkClanID string     `json:"network_clan_id" bson:"network_clan_id"`
	VerifiedAt    *time.Time `json:"verified_at" bson:"verified_at"`

	ResourceOwner common.ResourceOwner `json:"-" bson:"resource_owner"`
	ShareTokens   []ShareToken         `json:"-" bson:"share_tokens"`
	CreatedAt     time.Time            `json:"-" bson:"created_at"`
	UpdatedAt     *time.Time           `json:"-" bson:"updated_at"`
}

func NewPlayerMetadata(currentName string, networkUserID string, networkID common.NetworkIDKey, clanName string, res common.ResourceOwner) *PlayerMetadata {
	return &PlayerMetadata{
		ID:            common.PlayerIDType(uuid.New()),
		UserID:        nil,
		GameID:        common.CS2.ID,
		NetworkID:     networkID,
		NetworkUserID: networkUserID,
		Name:          currentName,
		ClanName:      clanName,
		NameHistory:   []string{},
		NetworkClanID: "",
		AvatarURI:     "",
		VerifiedAt:    nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     nil,
		ShareTokens:   []ShareToken{},
		ResourceOwner: res,
	}
}

func (e PlayerMetadata) GetID() uuid.UUID {
	return uuid.UUID(e.ID)
}
