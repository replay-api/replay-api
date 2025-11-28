package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
)

// MatchmakingSessionRepository handles persistence of matchmaking sessions
type MatchmakingSessionRepository interface {
	// Save creates or updates a matchmaking session
	Save(ctx context.Context, session *matchmaking_entities.MatchmakingSession) error

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingSession, error)

	// GetByPlayerID retrieves active sessions for a player
	GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*matchmaking_entities.MatchmakingSession, error)

	// GetActiveSessions retrieves all active sessions (queued or searching)
	GetActiveSessions(ctx context.Context, filters SessionFilters) ([]*matchmaking_entities.MatchmakingSession, error)

	// UpdateStatus updates the session status
	UpdateStatus(ctx context.Context, id uuid.UUID, status matchmaking_entities.SessionStatus) error

	// Delete removes a session
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteExpired removes expired sessions
	DeleteExpired(ctx context.Context) (int64, error)
}

// SessionFilters defines filters for querying sessions
type SessionFilters struct {
	GameID   string
	GameMode string
	Region   string
	Tier     *matchmaking_entities.MatchmakingTier
	Status   *matchmaking_entities.SessionStatus
	MinMMR   *int
	MaxMMR   *int
	Limit    int
	Offset   int
}

// MatchmakingPoolRepository handles persistence of matchmaking pools
type MatchmakingPoolRepository interface {
	// Save creates or updates a pool
	Save(ctx context.Context, pool *matchmaking_entities.MatchmakingPool) error

	// GetByID retrieves a pool by ID
	GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingPool, error)

	// GetByGameModeRegion retrieves a pool by game, mode, and region
	GetByGameModeRegion(ctx context.Context, gameID, gameMode, region string) (*matchmaking_entities.MatchmakingPool, error)

	// UpdateStats updates pool statistics
	UpdateStats(ctx context.Context, poolID uuid.UUID, stats matchmaking_entities.PoolStatistics) error

	// GetAllActive retrieves all active pools
	GetAllActive(ctx context.Context) ([]*matchmaking_entities.MatchmakingPool, error)
}
