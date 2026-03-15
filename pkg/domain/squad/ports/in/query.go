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

// PlayerSkillReader defines the interface for reading player skills
type PlayerSkillReader interface {
	// GetPlayerSkills retrieves all skills for a player
	GetPlayerSkills(ctx context.Context, playerID uuid.UUID) ([]squad_entities.PlayerSkill, error)
	// GetSkillProfile retrieves the aggregated skill profile for radar chart display
	GetSkillProfile(ctx context.Context, playerID uuid.UUID) (*squad_entities.SkillProfile, error)
}

// PlayerTraitReader defines the interface for reading player traits
type PlayerTraitReader interface {
	// GetPlayerTraits retrieves all traits for a player
	GetPlayerTraits(ctx context.Context, playerID uuid.UUID) ([]squad_entities.PlayerTrait, error)
}

// PlayerTeamHistoryReader defines the interface for player team history
type PlayerTeamHistoryReader interface {
	// GetPlayerTeamHistory retrieves team history for a player (career timeline)
	GetPlayerTeamHistory(ctx context.Context, playerID uuid.UUID) ([]squad_entities.PlayerTeamHistoryEntry, error)
}

// TeamRosterHistoryReader defines the interface for team roster history
type TeamRosterHistoryReader interface {
	// GetTeamRosterHistory retrieves historical roster for a team
	GetTeamRosterHistory(ctx context.Context, squadID uuid.UUID) ([]squad_entities.TeamRosterHistoryEntry, error)
}
