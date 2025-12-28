package achievement_in

import (
	"context"

	"github.com/google/uuid"
	achievement_entities "github.com/replay-api/replay-api/pkg/domain/achievement/entities"
)

// UnlockAchievementCommand represents a request to unlock an achievement
type UnlockAchievementCommand struct {
	PlayerID      uuid.UUID
	AchievementID uuid.UUID
}

// UpdateProgressCommand represents a request to update achievement progress
type UpdateProgressCommand struct {
	PlayerID        uuid.UUID
	AchievementCode string
	IncrementBy     int
}

// AchievementCommand defines write operations for achievements
type AchievementCommand interface {
	// UnlockAchievement explicitly unlocks an achievement for a player
	UnlockAchievement(ctx context.Context, cmd UnlockAchievementCommand) (*achievement_entities.PlayerAchievement, error)

	// UpdateProgress updates progress for an achievement
	UpdateProgress(ctx context.Context, cmd UpdateProgressCommand) (*achievement_entities.PlayerAchievement, error)

	// InitializePlayerAchievements creates achievement tracking for a new player
	InitializePlayerAchievements(ctx context.Context, playerID uuid.UUID) error

	// CheckAndUnlock checks if criteria are met and unlocks applicable achievements
	CheckAndUnlock(ctx context.Context, playerID uuid.UUID, event AchievementEvent) ([]achievement_entities.PlayerAchievement, error)
}

// AchievementEvent represents an event that might trigger achievement progress
type AchievementEvent struct {
	Type     string                 `json:"type"`
	PlayerID uuid.UUID              `json:"player_id"`
	Data     map[string]interface{} `json:"data"`
}

// Event types for achievement triggers
const (
	EventMatchCompleted     = "match_completed"
	EventMatchWon           = "match_won"
	EventKill               = "kill"
	EventHeadshotKill       = "headshot_kill"
	EventAce                = "ace"
	EventClutch             = "clutch"
	EventBombPlanted        = "bomb_planted"
	EventBombDefused        = "bomb_defused"
	EventMVP                = "mvp"
	EventTournamentWon      = "tournament_won"
	EventPrizeWon           = "prize_won"
	EventSquadJoined        = "squad_joined"
	EventSquadMatchCompleted = "squad_match_completed"
)



