package matchmaking_entities

import (
	"time"

	"github.com/google/uuid"
)

// MatchmakingPool represents a pool of players waiting to be matched
type MatchmakingPool struct {
	ID             uuid.UUID              `json:"id" bson:"_id"`
	GameID         string                 `json:"game_id" bson:"game_id"`
	GameMode       string                 `json:"game_mode" bson:"game_mode"`
	Region         string                 `json:"region" bson:"region"`
	ActiveSessions []uuid.UUID            `json:"active_sessions" bson:"active_sessions"`
	PoolStats      PoolStatistics         `json:"pool_stats" bson:"pool_stats"`
	CreatedAt      time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" bson:"updated_at"`
}

// GetID returns the matchmaking pool ID
func (p MatchmakingPool) GetID() uuid.UUID {
	return p.ID
}

// PoolStatistics tracks real-time pool metrics
type PoolStatistics struct {
	TotalPlayers      int                         `json:"total_players" bson:"total_players"`
	AverageWaitTime   int                         `json:"average_wait_time_seconds" bson:"average_wait_time_seconds"`
	PlayersByTier     map[MatchmakingTier]int     `json:"players_by_tier" bson:"players_by_tier"`
	PlayersBySkill    map[string]int              `json:"players_by_skill" bson:"players_by_skill"` // MMR ranges
	EstimatedMatchTime int                        `json:"estimated_match_time_seconds" bson:"estimated_match_time_seconds"`
	MatchesLast24h    int                         `json:"matches_last_24h" bson:"matches_last_24h"`
}

// PoolSnapshot represents a point-in-time view of the pool
type PoolSnapshot struct {
	PoolID        uuid.UUID      `json:"pool_id"`
	GameID        string         `json:"game_id"`
	GameMode      string         `json:"game_mode"`
	Region        string         `json:"region"`
	Stats         PoolStatistics `json:"stats"`
	Timestamp     time.Time      `json:"timestamp"`
	QueueHealth   string         `json:"queue_health"` // "healthy", "slow", "degraded"
}

// GetQueueHealth calculates queue health based on metrics
func (p *MatchmakingPool) GetQueueHealth() string {
	if p.PoolStats.TotalPlayers >= 50 && p.PoolStats.AverageWaitTime < 120 {
		return "healthy"
	} else if p.PoolStats.TotalPlayers >= 20 && p.PoolStats.AverageWaitTime < 300 {
		return "moderate"
	}
	return "slow"
}
