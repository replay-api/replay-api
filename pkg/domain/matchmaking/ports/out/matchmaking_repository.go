package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
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

// PlayerRatingRepository handles persistence of player ratings
type PlayerRatingRepository interface {
	// Save creates or updates a player rating
	Save(ctx context.Context, rating *matchmaking_entities.PlayerRating) error

	// Update updates an existing player rating
	Update(ctx context.Context, rating *matchmaking_entities.PlayerRating) error

	// FindByPlayerAndGame finds a rating by player ID and game ID
	FindByPlayerAndGame(ctx context.Context, playerID uuid.UUID, gameID common.GameIDKey) (*matchmaking_entities.PlayerRating, error)

	// GetByID retrieves a rating by ID
	GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PlayerRating, error)

	// GetTopPlayers retrieves top players by rating for a game
	GetTopPlayers(ctx context.Context, gameID common.GameIDKey, limit int) ([]*matchmaking_entities.PlayerRating, error)

	// GetRankDistribution returns the distribution of ranks for a game
	GetRankDistribution(ctx context.Context, gameID common.GameIDKey) (map[matchmaking_entities.Rank]int, error)

	// Delete removes a player rating
	Delete(ctx context.Context, id uuid.UUID) error
}
