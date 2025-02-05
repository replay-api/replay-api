package squad_entities

import (
	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type PlayerHistoryAction string

const (
	PlayerHistoryActionCreate                 PlayerHistoryAction = "Create"
	PlayerHistoryActionUpdate                 PlayerHistoryAction = "Update"
	PlayerHistoryActionDelete                 PlayerHistoryAction = "Delete"
	PlayerHistoryActionVisibilityRestricted   PlayerHistoryAction = "VisibilityRestricted"
	PlayerHistoryActionVisibilityUnrestricted PlayerHistoryAction = "VisibilityUnrestricted"
)

type PlayerProfile struct {
	common.BaseEntity
	GameID      common.GameIDKey `json:"game_id" bson:"game_id"`
	Nickname    string           `json:"nickname" bson:"nickname"`
	Avatar      string           `json:"avatar" bson:"avatar"`
	Roles       []string         `json:"roles" bson:"roles"`
	Description string           `json:"description" bson:"description"`
}

func (e PlayerProfile) GetID() uuid.UUID {
	return e.BaseEntity.ID
}

func NewPlayerProfile(gameID common.GameIDKey, nickname, avatar, description string, visbility common.VisibilityTypeKey, rxn common.ResourceOwner) *PlayerProfile {
	var baseEntity common.BaseEntity

	switch visbility {
	case common.PublicVisibilityTypeKey:
		baseEntity = common.NewUnrestrictedEntity(rxn)
	case common.RestrictedVisibilityTypeKey:
		baseEntity = common.NewRestrictedEntity(rxn)
	case common.PrivateVisibilityTypeKey:
		baseEntity = common.NewPrivateEntity(rxn)
	case common.CustomVisibilityTypeKey:
		baseEntity = common.NewEntity(rxn)
	default:
		baseEntity = common.NewEntity(rxn)
	}

	return &PlayerProfile{
		BaseEntity:  baseEntity,
		GameID:      gameID,
		Nickname:    nickname,
		Avatar:      avatar,
		Description: description,
	}
}
