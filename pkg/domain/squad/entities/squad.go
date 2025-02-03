package squad_entities

import (
	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
)

type Squad struct {
	common.BaseEntity
	GameID      common.GameIDKey                                                `json:"game_id" bson:"game_id"`
	Name        string                                                          `json:"name" bson:"name"`
	Symbol      string                                                          `json:"symbol" bson:"symbol"`
	Description string                                                          `json:"description" bson:"description"`
	LogoURI     string                                                          `json:"logo_uri" bson:"logo_uri"`
	BannerURI   string                                                          `json:"banner_uri" bson:"banner_uri"` // TODO: create media collection, for multiple purposes
	Membership  map[common.PlayerIDType]squad_value_objects.SquadMembershipType `json:"membership" bson:"membership"`
}

func NewSquad(groupID uuid.UUID, gameID common.GameIDKey, logorURI, name, symbol, description string, rxn common.ResourceOwner) Squad {
	return Squad{
		BaseEntity:  common.NewUnrestrictedEntity(rxn),
		GameID:      gameID,
		Name:        name,
		Symbol:      symbol,
		Description: description,
	}
}

func (e Squad) GetID() uuid.UUID {
	return e.ID
}
