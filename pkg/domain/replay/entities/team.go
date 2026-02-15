package entities

import (
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type Team struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	NetworkID         string           `json:"network_id" bson:"network_id"`
	NetworkTeamID     string           `json:"network_team_id" bson:"network_team_id"`
	TeamHashID        string           `json:"team_hash_id" bson:"team_hash_id"` // network_id + network_player_id (asc,concat.,sha256)
	Name              string           `json:"name" bson:"name"`
	ShortName         string           `json:"short_name" bson:"short_name"`
	CurrentDisplayName string          `json:"display_name" bson:"display_name"`
	NameHistory       []string         `json:"name_history" bson:"name_history"`
	Players           []PlayerMetadata `json:"players" bson:"players"`
}

func (e Team) GetID() uuid.UUID {
	return e.ID
}
