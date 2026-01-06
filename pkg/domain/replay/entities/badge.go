package entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
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
	ResourceOwner shared.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func (b Badge) GetID() uuid.UUID {
	return b.ID
}
