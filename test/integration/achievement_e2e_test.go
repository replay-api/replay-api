//go:build integration || e2e
// +build integration e2e

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	shared "github.com/resource-ownership/go-common/pkg/common"
	achievement_entities "github.com/replay-api/replay-api/pkg/domain/achievement/entities"
	achievement_usecases "github.com/replay-api/replay-api/pkg/domain/achievement/usecases"
	db "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
)

// TestE2E_AchievementLifecycle tests the complete achievement lifecycle with real MongoDB
func TestE2E_AchievementLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to test MongoDB
	mongoURI := "mongodb://localhost:27017"
	dbName := "leetgaming_test_achievements_" + uuid.New().String()[:8]

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer func() {
		_ = client.Database(dbName).Drop(ctx)
		_ = client.Disconnect(ctx)
	}()

	// Verify connectivity
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}

	// Initialize repositories
	achievementRepo := db.NewAchievementMongoDBRepository(client, dbName)

	// Initialize query service
	achievementQuery := achievement_usecases.NewAchievementQueryService(achievementRepo)

	t.Run("GetAllAchievements_Empty", func(t *testing.T) {
		achievements, err := achievementQuery.GetAllAchievements(ctx)
		require.NoError(t, err)
		assert.Empty(t, achievements)
	})

	// Seed test achievements
	testAchievements := []achievement_entities.Achievement{
		{
			ID:          uuid.New(),
			Code:        "TEST_FIRST_BLOOD",
			Name:        "First Blood",
			Description: "Get your first kill",
			Category:    achievement_entities.CategoryCombat,
			Rarity:      achievement_entities.RarityCommon,
			XPReward:    50,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Code:        "TEST_HEADHUNTER",
			Name:        "Headhunter",
			Description: "Get 10 headshot kills",
			Category:    achievement_entities.CategoryCombat,
			Rarity:      achievement_entities.RarityUncommon,
			XPReward:    100,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Code:        "TEST_MATCHES_10",
			Name:        "Getting Started",
			Description: "Complete 10 matches",
			Category:    achievement_entities.CategoryProgress,
			Rarity:      achievement_entities.RarityCommon,
			XPReward:    75,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, ach := range testAchievements {
		err := achievementRepo.CreateAchievement(ctx, &ach)
		require.NoError(t, err, "Failed to seed achievement: %s", ach.Code)
	}

	t.Run("GetAllAchievements_WithData", func(t *testing.T) {
		achievements, err := achievementQuery.GetAllAchievements(ctx)
		require.NoError(t, err)
		assert.Len(t, achievements, 3)
	})

	t.Run("GetAchievementsByCategory", func(t *testing.T) {
		combatAchievements, err := achievementQuery.GetAchievementsByCategory(ctx, achievement_entities.CategoryCombat)
		require.NoError(t, err)
		assert.Len(t, combatAchievements, 2)

		progressAchievements, err := achievementQuery.GetAchievementsByCategory(ctx, achievement_entities.CategoryProgress)
		require.NoError(t, err)
		assert.Len(t, progressAchievements, 1)
	})

	t.Run("GetAchievementByID", func(t *testing.T) {
		achievement, err := achievementQuery.GetAchievementByID(ctx, testAchievements[0].ID)
		require.NoError(t, err)
		require.NotNil(t, achievement)
		assert.Equal(t, "First Blood", achievement.Name)
		assert.Equal(t, achievement_entities.RarityCommon, achievement.Rarity)
	})

	// Test player achievements
	playerID := uuid.New()
	resourceOwner := shared.ResourceOwner{
		UserID:   playerID,
		TenantID: uuid.New(),
		ClientID: uuid.New(),
	}

	// Create player achievements (tracking progress)
	playerAchievements := []achievement_entities.PlayerAchievement{
		*achievement_entities.NewPlayerAchievement(playerID, testAchievements[0].ID, 1, resourceOwner),
		*achievement_entities.NewPlayerAchievement(playerID, testAchievements[1].ID, 10, resourceOwner),
		*achievement_entities.NewPlayerAchievement(playerID, testAchievements[2].ID, 10, resourceOwner),
	}

	// Mark first one as unlocked
	playerAchievements[0].IncrementProgress(1)

	// Mark second one as partially complete
	playerAchievements[1].IncrementProgress(5)

	for _, pa := range playerAchievements {
		err := achievementRepo.CreatePlayerAchievement(ctx, &pa)
		require.NoError(t, err)
	}

	t.Run("GetPlayerAchievements", func(t *testing.T) {
		achievements, err := achievementQuery.GetPlayerAchievements(ctx, playerID)
		require.NoError(t, err)
		assert.Len(t, achievements, 3)

		// Check first one is unlocked
		var firstBlood *achievement_entities.PlayerAchievement
		for _, a := range achievements {
			if a.AchievementID == testAchievements[0].ID {
				firstBlood = &a
				break
			}
		}
		require.NotNil(t, firstBlood)
		assert.True(t, firstBlood.IsUnlocked())
		assert.NotNil(t, firstBlood.Achievement)
		assert.Equal(t, "First Blood", firstBlood.Achievement.Name)
	})

	t.Run("GetPlayerAchievementSummary", func(t *testing.T) {
		summary, err := achievementQuery.GetPlayerAchievementSummary(ctx, playerID)
		require.NoError(t, err)

		assert.Equal(t, playerID, summary.PlayerID)
		assert.Equal(t, 3, summary.TotalAchievements)
		assert.Equal(t, 1, summary.UnlockedCount) // Only first one is unlocked
		assert.Equal(t, 50, summary.TotalXP)      // XP from first achievement

		// Check category breakdown
		assert.NotNil(t, summary.ByCategory)
		combatProgress := summary.ByCategory[achievement_entities.CategoryCombat]
		assert.Equal(t, 2, combatProgress.Total)
		assert.Equal(t, 1, combatProgress.Unlocked)
	})

	t.Run("GetRecentUnlocks", func(t *testing.T) {
		unlocks, err := achievementQuery.GetRecentUnlocks(ctx, playerID, 10)
		require.NoError(t, err)
		assert.Len(t, unlocks, 1) // Only one achievement is unlocked
	})

	t.Run("AchievementProgress_IncrementToUnlock", func(t *testing.T) {
		// Get the partially completed achievement
		pa, err := achievementRepo.GetPlayerAchievementByCode(ctx, playerID, "TEST_HEADHUNTER")
		require.NoError(t, err)
		require.NotNil(t, pa)

		assert.False(t, pa.IsUnlocked())
		assert.Equal(t, 5, pa.Progress)
		assert.Equal(t, float64(50), pa.GetProgressPercentage()) // 5/10 = 50%

		// Increment to complete
		pa.IncrementProgress(5)
		assert.True(t, pa.IsUnlocked())
		assert.NotNil(t, pa.UnlockedAt)

		err = achievementRepo.UpdatePlayerAchievement(ctx, pa)
		require.NoError(t, err)

		// Verify update persisted
		updated, err := achievementRepo.GetPlayerAchievementByID(ctx, pa.ID)
		require.NoError(t, err)
		assert.True(t, updated.IsUnlocked())
		assert.Equal(t, 10, updated.Progress)
	})

	t.Run("GetRecentUnlocks_AfterNewUnlock", func(t *testing.T) {
		unlocks, err := achievementQuery.GetRecentUnlocks(ctx, playerID, 10)
		require.NoError(t, err)
		assert.Len(t, unlocks, 2) // Now two achievements are unlocked
	})

	t.Run("UnlockedCount", func(t *testing.T) {
		count, err := achievementRepo.GetUnlockedCount(ctx, playerID)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Log("Achievement E2E lifecycle test passed ✅")
}

// TestE2E_AchievementCategories tests achievement category filtering
func TestE2E_AchievementCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to test MongoDB
	mongoURI := "mongodb://localhost:27017"
	dbName := "leetgaming_test_achievements_cat_" + uuid.New().String()[:8]

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer func() {
		_ = client.Database(dbName).Drop(ctx)
		_ = client.Disconnect(ctx)
	}()

	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}

	achievementRepo := db.NewAchievementMongoDBRepository(client, dbName)

	// Seed achievements from all categories
	categories := []achievement_entities.AchievementCategory{
		achievement_entities.CategoryCombat,
		achievement_entities.CategoryObjective,
		achievement_entities.CategoryTeamwork,
		achievement_entities.CategoryProgress,
		achievement_entities.CategoryCompetitive,
		achievement_entities.CategorySocial,
		achievement_entities.CategoryMilestone,
	}

	for _, cat := range categories {
		ach := achievement_entities.Achievement{
			ID:          uuid.New(),
			Code:        "TEST_" + string(cat),
			Name:        "Test " + string(cat),
			Description: "Test achievement for " + string(cat),
			Category:    cat,
			Rarity:      achievement_entities.RarityCommon,
			XPReward:    50,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := achievementRepo.CreateAchievement(ctx, &ach)
		require.NoError(t, err)
	}

	// Test each category filter
	for _, cat := range categories {
		t.Run("Category_"+string(cat), func(t *testing.T) {
			achievements, err := achievementRepo.GetAchievementsByCategory(ctx, cat)
			require.NoError(t, err)
			assert.Len(t, achievements, 1)
			assert.Equal(t, cat, achievements[0].Category)
		})
	}

	t.Log("Achievement categories E2E test passed ✅")
}



