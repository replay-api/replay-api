package entities

import (
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type Round struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	GameID            string       `json:"game_id" bson:"game_id"`
	MatchID           uuid.UUID    `json:"match_id" bson:"match_id"`
	Title             string       `json:"title" bson:"title"` // Round 1, Round 2, etc
	Events            []*GameEvent `json:"game_events" bson:"game_events"`
	Description       string       `json:"description" bson:"description"`
	ImageURL          string       `json:"image_url" bson:"image_url"`
}
