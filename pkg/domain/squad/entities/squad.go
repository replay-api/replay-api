package squad_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
)

type Squad struct {
	ID            uuid.UUID                              `json:"id" bson:"_id"`
	GroupID       uuid.UUID                              `json:"group_id" bson:"group_id"`
	GameID        common.GameIDKey                       `json:"game_id" bson:"game_id"`
	Name          string                                 `json:"name" bson:"name"`
	Symbol        string                                 `json:"symbol" bson:"symbol"`
	Description   string                                 `json:"description" bson:"description"`
	LogoURI       string                                 `json:"logo_uri" bson:"logo_uri"`
	Profiles      map[string]squad_value_objects.Profile `json:"profiles" bson:"profiles"`
	ResourceOwner common.ResourceOwner                   `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time                              `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time                              `json:"updated_at" bson:"updated_at"`
}

func NewSquad(groupID uuid.UUID, gameID common.GameIDKey, name, symbol, description string, profiles map[string]squad_value_objects.Profile, resourceOwner common.ResourceOwner) Squad {
	return Squad{
		ID:            uuid.New(),
		GroupID:       groupID,
		GameID:        gameID,
		Name:          name,
		Symbol:        symbol,
		Description:   description,
		Profiles:      profiles,
		ResourceOwner: resourceOwner,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// func (e *Squad) GetID() uuid.UUID {
// 	return e.ID
// }

func (e Squad) GetID() uuid.UUID {
	return e.ID
}
