package squad_in

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type SquadReader interface {
	shared.Searchable[squad_entities.Squad]
}

type PlayerProfileReader interface {
	shared.Searchable[squad_entities.PlayerProfile]
}

// PlayerStatisticsReader defines the interface for reading player statistics
type PlayerStatisticsReader interface {
	// GetPlayerStatistics retrieves aggregated statistics for a player
	GetPlayerStatistics(ctx context.Context, playerID uuid.UUID, gameID *replay_common.GameIDKey) (*squad_entities.PlayerStatistics, error)
}
