package squad_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type PlayerProfileHistory struct {
	common.BaseEntity
	PlayerID  uuid.UUID           `json:"player_id"`
	Changes   PlayerHistoryAction `json:"changes"`
	CreatedAt time.Time           `json:"created_at"`
}

func (e PlayerProfileHistory) GetID() uuid.UUID {
	return e.ID
}

func NewPlayerProfileHistory(playerID uuid.UUID, changes PlayerHistoryAction, rxn common.ResourceOwner) *PlayerProfileHistory {
	return &PlayerProfileHistory{
		BaseEntity: common.NewRestrictedEntity(rxn),
		PlayerID:   playerID,
		Changes:    changes,
		CreatedAt:  time.Now(),
	}
}
