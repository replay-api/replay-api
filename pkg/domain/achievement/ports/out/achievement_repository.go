package achievement_out

import (
	"context"

	"github.com/google/uuid"
	achievement_entities "github.com/replay-api/replay-api/pkg/domain/achievement/entities"
)

// AchievementRepository defines data access for achievements
type AchievementRepository interface {
	// Achievement operations
	GetAllAchievements(ctx context.Context) ([]achievement_entities.Achievement, error)
	GetAchievementByID(ctx context.Context, id uuid.UUID) (*achievement_entities.Achievement, error)
	GetAchievementByCode(ctx context.Context, code string) (*achievement_entities.Achievement, error)
	GetAchievementsByCategory(ctx context.Context, category achievement_entities.AchievementCategory) ([]achievement_entities.Achievement, error)
	CreateAchievement(ctx context.Context, achievement *achievement_entities.Achievement) error
	UpdateAchievement(ctx context.Context, achievement *achievement_entities.Achievement) error

	// Player achievement operations
	GetPlayerAchievements(ctx context.Context, playerID uuid.UUID) ([]achievement_entities.PlayerAchievement, error)
	GetPlayerAchievementByCode(ctx context.Context, playerID uuid.UUID, achievementCode string) (*achievement_entities.PlayerAchievement, error)
	GetPlayerAchievementByID(ctx context.Context, id uuid.UUID) (*achievement_entities.PlayerAchievement, error)
	CreatePlayerAchievement(ctx context.Context, pa *achievement_entities.PlayerAchievement) error
	UpdatePlayerAchievement(ctx context.Context, pa *achievement_entities.PlayerAchievement) error
	GetRecentUnlocks(ctx context.Context, playerID uuid.UUID, limit int) ([]achievement_entities.PlayerAchievement, error)
	GetUnlockedCount(ctx context.Context, playerID uuid.UUID) (int, error)
}



