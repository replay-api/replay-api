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
	PlayerHistoryActionVisibilityPrivate      PlayerHistoryAction = "VisibilityPrivate"
)

type PlayerProfileHistory struct {
	common.BaseEntity
	PlayerID uuid.UUID `json:"player_id" bson:"player_id"`
	Action   PlayerHistoryAction
}

func (e PlayerProfileHistory) GetID() uuid.UUID {
	return e.BaseEntity.ID
}
