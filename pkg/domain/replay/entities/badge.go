package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type Badge struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	GameID        string               `json:"game_id" bson:"game_id"`
	MatchID       uuid.UUID            `json:"match_id" bson:"match_id"`
	PlayerID      uuid.UUID            `json:"player_id" bson:"player_id"`
	Name          string               `json:"name" bson:"name"`
	Events        []interface{}        `json:"events" bson:"events"`
	Description   string               `json:"description" bson:"description"`
	ImageURL      string               `json:"image_url" bson:"image_url"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func (b Badge) GetID() uuid.UUID {
	return b.ID
}
