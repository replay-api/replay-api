package squad_entities

import (
	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type PlayerIDKey uuid.UUID

type PlayerHistoryAction string

const (
	PlayerHistoryActionCreate                 PlayerHistoryAction = "Create"
	PlayerHistoryActionUpdate                 PlayerHistoryAction = "Update"
	PlayerHistoryActionDelete                 PlayerHistoryAction = "Delete"
	PlayerHistoryActionVisibilityRestricted   PlayerHistoryAction = "VisibilityRestricted"
	PlayerHistoryActionVisibilityUnrestricted PlayerHistoryAction = "VisibilityUnrestricted"
)

type Player struct {
	common.BaseEntity
	GameID      common.GameIDKey `json:"game_id" bson:"game_id"`
	Nickname    string           `json:"nickname" bson:"nickname"`
	Avatar      string           `json:"avatar" bson:"avatar"`
	Description string           `json:"description" bson:"description"`
	LogoURI     string           `json:"logo_uri" bson:"logo_uri"`
}

func NewPlayer(gameID common.GameIDKey, nickname, avatar, description, logoURI string, rxn common.ResourceOwner) Player {
	return Player{
		BaseEntity:  common.NewUnrestrictedEntity(rxn),
		GameID:      gameID,
		Nickname:    nickname,
		Avatar:      avatar,
		Description: description,
		LogoURI:     logoURI,
	}
}
