package entities

import (
	"time"

	"github.com/google/uuid"
)

type Round struct {
	ID uuid.UUID `json:"id" bson:"_id"`
	// GameID      string    `json:"game_id" bson:"game_id"`
	// MatchID     uuid.UUID `json:"match_id" bson:"match_id"`
	// PlayerID    uuid.UUID `json:"player_id" bson:"player_id"`
	// Name        string    `json:"name" bson:"name"`
	// Events      []string  `json:"events" bson:"events"`
	// Description string    `json:"description" bson:"description"`
	ImageURL  string    `json:"image_url" bson:"image_url"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
