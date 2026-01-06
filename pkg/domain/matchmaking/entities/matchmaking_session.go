package matchmaking_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// MatchmakingTier represents subscription tier for matchmaking
type MatchmakingTier string

const (
	TierFree     MatchmakingTier = "free"
	TierPremium  MatchmakingTier = "premium"
	TierPro      MatchmakingTier = "pro"
	TierElite    MatchmakingTier = "elite"
)

// SessionStatus represents the current state of a matchmaking session
type SessionStatus string

const (
	StatusQueued    SessionStatus = "queued"
	StatusSearching SessionStatus = "searching"
	StatusMatched   SessionStatus = "matched"
	StatusReady     SessionStatus = "ready"
	StatusCancelled SessionStatus = "cancelled"
	StatusExpired   SessionStatus = "expired"
)

// MatchPreferences represents player's match preferences
type MatchPreferences struct {
	GameID            string              `json:"game_id" bson:"game_id"`
	GameMode          string              `json:"game_mode" bson:"game_mode"`
	Region            string              `json:"region" bson:"region"`
	MapPreferences    []string            `json:"map_preferences,omitempty" bson:"map_preferences,omitempty"`
	SkillRange        SkillRange          `json:"skill_range" bson:"skill_range"`
	MaxPing           int                 `json:"max_ping" bson:"max_ping"`
	AllowCrossPlatform bool               `json:"allow_cross_platform" bson:"allow_cross_platform"`
	Tier              MatchmakingTier     `json:"tier" bson:"tier"`
	PriorityBoost     bool                `json:"priority_boost" bson:"priority_boost"` // Premium feature
}

// SkillRange defines acceptable skill level range
type SkillRange struct {
	MinMMR int `json:"min_mmr" bson:"min_mmr"`
	MaxMMR int `json:"max_mmr" bson:"max_mmr"`
}

// MatchmakingSession represents a player's matchmaking session
type MatchmakingSession struct {
	shared.BaseEntity `bson:",inline"`
	PlayerID          uuid.UUID         `json:"player_id" bson:"player_id"`
	SquadID           *uuid.UUID        `json:"squad_id,omitempty" bson:"squad_id,omitempty"`
	Preferences       MatchPreferences  `json:"preferences" bson:"preferences"`
	Status            SessionStatus     `json:"status" bson:"status"`
	PlayerMMR         int               `json:"player_mmr" bson:"player_mmr"`
	QueuedAt          time.Time         `json:"queued_at" bson:"queued_at"`
	MatchedAt         *time.Time        `json:"matched_at,omitempty" bson:"matched_at,omitempty"`
	MatchID           *uuid.UUID        `json:"match_id,omitempty" bson:"match_id,omitempty"`
	EstimatedWait     int               `json:"estimated_wait_seconds" bson:"estimated_wait_seconds"`
	ExpiresAt         time.Time         `json:"expires_at" bson:"expires_at"`
	Metadata          map[string]any    `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func (m *MatchmakingSession) GetID() uuid.UUID {
	return m.ID
}

// IsExpired checks if the session has expired
func (m *MatchmakingSession) IsExpired() bool {
	return time.Now().After(m.ExpiresAt)
}

// CanMatch checks if session is in a matchable state
func (m *MatchmakingSession) CanMatch() bool {
	return m.Status == StatusQueued || m.Status == StatusSearching
}

// GetTierPriority returns numeric priority based on tier
func (m *MatchmakingSession) GetTierPriority() int {
	switch m.Preferences.Tier {
	case TierElite:
		return 4
	case TierPro:
		return 3
	case TierPremium:
		return 2
	default:
		return 1
	}
}
