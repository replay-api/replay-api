package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type Team struct {
	ID                 uuid.UUID        `json:"id" bson:"_id"`
	NetworkID          string           `json:"network_id" bson:"network_id"`
	NetworkTeamID      string           `json:"network_team_id" bson:"network_team_id"`
	TeamHashID         string           `json:"team_hash_id" bson:"team_hash_id"` // network_id + network_player_id (asc,concat.,sha256)
	Name               string           `json:"name" bson:"name"`
	ShortName          string           `json:"short_name" bson:"short_name"`
	CurrentDisplayName string           `json:"display_name" bson:"display_name"`
	NameHistory        []string         `json:"name_history" bson:"name_history"`
	Players            []PlayerMetadata `json:"players" bson:"players"`

	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func (e Team) GetID() uuid.UUID {
	return e.ID
}
