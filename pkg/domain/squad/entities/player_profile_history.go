package squad_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type PlayerProfileHistory struct {
	common.BaseEntity
	PlayerID  uuid.UUID           `json:"player_profile"`
	Changes   PlayerHistoryAction `json:"changes"`
	CreatedAt time.Time           `json:"created_at"`
}

func (e PlayerProfileHistory) GetID() uuid.UUID {
	return e.ID
}

func NewPlayerProfileHistory(playerID uuid.UUID, changes PlayerHistoryAction) *PlayerProfileHistory {
	return &PlayerProfileHistory{
		BaseEntity: common.BaseEntity{
			ID: uuid.New(),
		},
		PlayerID:  playerID,
		Changes:   changes,
		CreatedAt: time.Now(),
	}
}
