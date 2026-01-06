package squad_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type PlayerProfileHistory struct {
	shared.BaseEntity
	PlayerID  uuid.UUID           `json:"player_id" bson:"player_id"`
	Changes   PlayerHistoryAction `json:"changes" bson:"changes"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
}

func (e PlayerProfileHistory) GetID() uuid.UUID {
	return e.ID
}

func NewPlayerProfileHistory(playerID uuid.UUID, changes PlayerHistoryAction, rxn shared.ResourceOwner) *PlayerProfileHistory {
	return &PlayerProfileHistory{
		BaseEntity: shared.NewRestrictedEntity(rxn),
		PlayerID:   playerID,
		Changes:    changes,
		CreatedAt:  time.Now(),
	}
}
