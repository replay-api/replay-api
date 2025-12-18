package squad_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
)

type SquadReader interface {
	common.Searchable[squad_entities.Squad]
}

type PlayerProfileReader interface {
	common.Searchable[squad_entities.PlayerProfile]
}

// PlayerStatisticsReader defines the interface for reading player statistics
type PlayerStatisticsReader interface {
	// GetPlayerStatistics retrieves aggregated statistics for a player
	GetPlayerStatistics(ctx context.Context, playerID uuid.UUID, gameID *common.GameIDKey) (*squad_entities.PlayerStatistics, error)
}
