package entities

import (
	"time"

	"github.com/google/uuid"
)

type Round struct {
	ID          uuid.UUID    `json:"id" bson:"_id"`
	GameID      string       `json:"game_id" bson:"game_id"`
	MatchID     uuid.UUID    `json:"match_id" bson:"match_id"`
	Title       string       `json:"title" bson:"title"` // Round 1, Round 2, etc
	Events      []*GameEvent `json:"game_events" bson:"game_events"`
	Description string       `json:"description" bson:"description"`
	ImageURL    string       `json:"image_url" bson:"image_url"`
	CreatedAt   time.Time    `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" bson:"updated_at"`
}
