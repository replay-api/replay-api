package achievement_in

import (
	"context"

	"github.com/google/uuid"
	achievement_entities "github.com/replay-api/replay-api/pkg/domain/achievement/entities"
)

// AchievementQuery defines read operations for achievements
type AchievementQuery interface {
	// GetPlayerAchievements retrieves all achievements for a player
	GetPlayerAchievements(ctx context.Context, playerID uuid.UUID) ([]achievement_entities.PlayerAchievement, error)

	// GetPlayerAchievementSummary retrieves achievement summary for a player
	GetPlayerAchievementSummary(ctx context.Context, playerID uuid.UUID) (*achievement_entities.PlayerAchievementSummary, error)

	// GetAllAchievements retrieves all available achievements
	GetAllAchievements(ctx context.Context) ([]achievement_entities.Achievement, error)

	// GetAchievementByID retrieves a specific achievement by ID
	GetAchievementByID(ctx context.Context, id uuid.UUID) (*achievement_entities.Achievement, error)

	// GetAchievementsByCategory retrieves achievements by category
	GetAchievementsByCategory(ctx context.Context, category achievement_entities.AchievementCategory) ([]achievement_entities.Achievement, error)

	// GetRecentUnlocks retrieves recent achievement unlocks for a player
	GetRecentUnlocks(ctx context.Context, playerID uuid.UUID, limit int) ([]achievement_entities.PlayerAchievement, error)
}

