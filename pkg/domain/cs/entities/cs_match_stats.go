package entities

import (
	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type CSMatchStats struct {
	MatchID     uuid.UUID      `json:"match_id" bson:"match_id"`
	GameState   CSGameState    `json:"game_state" bson:"game_state"`
	Rules       CSGameRules    `json:"rules" bson:"rules"`
	RoundsStats []CSRoundStats `json:"rounds_stats"`
	// ResourceOwner common.ResourceOwner `json:"resource_owner"`
	Header *CSReplayFileHeader `json:"replay_file_header" bson:"replay_file_header"`
}

func NewCSMatchStats(matchID uuid.UUID, resourceOwner common.ResourceOwner, roundCount int) CSMatchStats {
	return CSMatchStats{
		MatchID:     matchID,
		RoundsStats: make([]CSRoundStats, roundCount),
		// ResourceOwner: resourceOwner,
	}
}
