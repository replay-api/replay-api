package entities

import (
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type CSMatchStats struct {
	MatchID     uuid.UUID      `json:"match_id" bson:"match_id"`
	GameState   CSGameState    `json:"game_state" bson:"game_state"`
	Rules       CSGameRules    `json:"rules" bson:"rules"`
	RoundsStats []CSRoundStats `json:"rounds_stats"`
	// ResourceOwner shared.ResourceOwner `json:"resource_owner"`
	Header *CSReplayFileHeader `json:"replay_file_header" bson:"replay_file_header"`
}

func NewCSMatchStats(matchID uuid.UUID, resourceOwner shared.ResourceOwner, roundCount int) CSMatchStats {
	return CSMatchStats{
		MatchID:     matchID,
		RoundsStats: make([]CSRoundStats, roundCount),
		// ResourceOwner: resourceOwner,
	}
}
