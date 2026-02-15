package entities

import (
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type Badge struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	GameID            string        `json:"game_id" bson:"game_id"`
	MatchID           uuid.UUID     `json:"match_id" bson:"match_id"`
	PlayerID          uuid.UUID     `json:"player_id" bson:"player_id"`
	Name              string        `json:"name" bson:"name"`
	Events            []interface{} `json:"events" bson:"events"`
	Description       string        `json:"description" bson:"description"`
	ImageURL          string        `json:"image_url" bson:"image_url"`
}

func (b Badge) GetID() uuid.UUID {
	return b.ID
}
