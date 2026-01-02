package achievement_usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	achievement_entities "github.com/replay-api/replay-api/pkg/domain/achievement/entities"
	achievement_in "github.com/replay-api/replay-api/pkg/domain/achievement/ports/in"
	achievement_out "github.com/replay-api/replay-api/pkg/domain/achievement/ports/out"
)

// AchievementQueryService implements AchievementQuery
type AchievementQueryService struct {
	repo achievement_out.AchievementRepository
}

// NewAchievementQueryService creates a new AchievementQueryService
func NewAchievementQueryService(repo achievement_out.AchievementRepository) achievement_in.AchievementQuery {
	return &AchievementQueryService{repo: repo}
}

// GetPlayerAchievements retrieves all achievements for a player
func (s *AchievementQueryService) GetPlayerAchievements(ctx context.Context, playerID uuid.UUID) ([]achievement_entities.PlayerAchievement, error) {
	slog.InfoContext(ctx, "Getting player achievements", "player_id", playerID)

	achievements, err := s.repo.GetPlayerAchievements(ctx, playerID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get player achievements", "error", err)
		return nil, err
	}

	// Enrich with achievement details
	allAchievements, err := s.repo.GetAllAchievements(ctx)
	if err == nil {
		achievementMap := make(map[uuid.UUID]*achievement_entities.Achievement)
		for i := range allAchievements {
			achievementMap[allAchievements[i].ID] = &allAchievements[i]
		}
		for i := range achievements {
			if ach, ok := achievementMap[achievements[i].AchievementID]; ok {
				achievements[i].Achievement = ach
			}
		}
	}

	return achievements, nil
}

// GetPlayerAchievementSummary retrieves achievement summary for a player
func (s *AchievementQueryService) GetPlayerAchievementSummary(ctx context.Context, playerID uuid.UUID) (*achievement_entities.PlayerAchievementSummary, error) {
	slog.InfoContext(ctx, "Getting player achievement summary", "player_id", playerID)

	// Get all achievements for totals
	allAchievements, err := s.repo.GetAllAchievements(ctx)
	if err != nil {
		return nil, err
	}

	// Get player achievements
	playerAchievements, err := s.repo.GetPlayerAchievements(ctx, playerID)
	if err != nil {
		return nil, err
	}

	// Build achievement lookup
	achievementMap := make(map[uuid.UUID]*achievement_entities.Achievement)
	for i := range allAchievements {
		achievementMap[allAchievements[i].ID] = &allAchievements[i]
	}

	// Calculate stats
	unlockedCount := 0
	totalXP := 0
	byCategory := make(map[achievement_entities.AchievementCategory]achievement_entities.AchievementCategoryProgress)

	// Initialize categories from all achievements
	for _, ach := range allAchievements {
		cat := byCategory[ach.Category]
		cat.Total++
		byCategory[ach.Category] = cat
	}

	// Count unlocked and XP
	for _, pa := range playerAchievements {
		if pa.IsUnlocked() {
			unlockedCount++
			if ach, ok := achievementMap[pa.AchievementID]; ok {
				totalXP += ach.XPReward
				cat := byCategory[ach.Category]
				cat.Unlocked++
				byCategory[ach.Category] = cat
			}
		}
	}

	// Get recent unlocks
	recentUnlocks, _ := s.repo.GetRecentUnlocks(ctx, playerID, 5)
	for i := range recentUnlocks {
		if ach, ok := achievementMap[recentUnlocks[i].AchievementID]; ok {
			recentUnlocks[i].Achievement = ach
		}
	}

	return &achievement_entities.PlayerAchievementSummary{
		PlayerID:          playerID,
		TotalAchievements: len(allAchievements),
		UnlockedCount:     unlockedCount,
		TotalXP:           totalXP,
		RecentUnlocks:     recentUnlocks,
		ByCategory:        byCategory,
	}, nil
}

// GetAllAchievements retrieves all available achievements
func (s *AchievementQueryService) GetAllAchievements(ctx context.Context) ([]achievement_entities.Achievement, error) {
	return s.repo.GetAllAchievements(ctx)
}

// GetAchievementByID retrieves a specific achievement by ID
func (s *AchievementQueryService) GetAchievementByID(ctx context.Context, id uuid.UUID) (*achievement_entities.Achievement, error) {
	return s.repo.GetAchievementByID(ctx, id)
}

// GetAchievementsByCategory retrieves achievements by category
func (s *AchievementQueryService) GetAchievementsByCategory(ctx context.Context, category achievement_entities.AchievementCategory) ([]achievement_entities.Achievement, error) {
	return s.repo.GetAchievementsByCategory(ctx, category)
}

// GetRecentUnlocks retrieves recent achievement unlocks for a player
func (s *AchievementQueryService) GetRecentUnlocks(ctx context.Context, playerID uuid.UUID, limit int) ([]achievement_entities.PlayerAchievement, error) {
	// Check if caller can view these achievements (own or admin)
	// TODO: Implement permission checking for private achievements
	// Currently allows public viewing - may filter sensitive data in future
	resourceOwner := common.GetResourceOwner(ctx)
	_ = resourceOwner // Avoid unused variable warning

	return s.repo.GetRecentUnlocks(ctx, playerID, limit)
}

