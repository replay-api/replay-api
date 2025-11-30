package squad_entities

import (
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type SquadOpenPosition struct {
	common.BaseEntity
	SquadID          uuid.UUID `json:"squad_id" bson:"squad_id"`
	PositionID       uuid.UUID `json:"position_id" bson:"position_id"`
	Description      string    `json:"description" bson:"description"`
	Requirements     []string  `json:"requirements" bson:"requirements"`
	Roles            []string  `json:"roles" bson:"roles"`
	Responsibilities []string  `json:"responsibilities" bson:"responsibilities"`
	Benefits         []string  `json:"benefits" bson:"benefits"`
}
