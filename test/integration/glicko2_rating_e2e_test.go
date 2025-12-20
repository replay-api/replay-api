//go:build integration || e2e
// +build integration e2e

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_services "github.com/replay-api/replay-api/pkg/domain/matchmaking/services"
	db "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
)

// TestE2E_Glicko2RatingLifecycle tests the complete Glicko-2 rating lifecycle with real MongoDB
// NO MOCKS - Production-grade integration test
func TestE2E_Glicko2RatingLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Setup: Connect to test MongoDB
	mongoURI := getMongoTestURIForRating()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err, "Failed to connect to MongoDB")
	defer func() { _ = client.Disconnect(ctx) }()

	// Create test database
	dbName := "replay_test_rating_" + uuid.New().String()[:8]
	database := client.Database(dbName)
	defer func() {
		_ = database.Drop(ctx)
	}()

	// Initialize repository and service
	ratingRepo := db.NewPlayerRatingMongoDBRepository(database)
	ratingService := matchmaking_services.NewGlicko2RatingService(ratingRepo)

	// Create test context with resource owner
	userID := uuid.New()
	groupID := uuid.New()
	tenantID := common.TeamPROTenantID
	ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.GroupIDKey, groupID)

	gameID := common.CS2.ID

	t.Run("NewPlayer_GetsDefaultRating", func(t *testing.T) {
		playerID := uuid.New()

		rating, err := ratingService.GetPlayerRating(ctx, playerID, gameID)
		require.NoError(t, err)
		require.NotNil(t, rating)

		assert.Equal(t, playerID, rating.PlayerID)
		assert.Equal(t, gameID, rating.GameID)
		assert.Equal(t, matchmaking_entities.DefaultRating, rating.Rating)
		assert.Equal(t, matchmaking_entities.DefaultRatingDeviation, rating.RatingDeviation)
		assert.Equal(t, matchmaking_entities.DefaultVolatility, rating.Volatility)
		assert.True(t, rating.IsProvisional())
		assert.Equal(t, 0, rating.MatchesPlayed)

		t.Log("✓ New player gets default rating 1500")
	})

	t.Run("ExistingPlayer_RatingPersists", func(t *testing.T) {
		playerID := uuid.New()

		// First call creates rating
		rating1, err := ratingService.GetPlayerRating(ctx, playerID, gameID)
		require.NoError(t, err)
		ratingID := rating1.ID

		// Second call should return same rating
		rating2, err := ratingService.GetPlayerRating(ctx, playerID, gameID)
		require.NoError(t, err)

		assert.Equal(t, ratingID, rating2.ID)
		assert.Equal(t, rating1.Rating, rating2.Rating)

		t.Log("✓ Rating persists in MongoDB across calls")
	})

	t.Run("MatchResult_UpdatesWinnerAndLoserRatings", func(t *testing.T) {
		// Create 4 players (2v2 match)
		winner1 := uuid.New()
		winner2 := uuid.New()
		loser1 := uuid.New()
		loser2 := uuid.New()

		// Initialize ratings for all players
		for _, playerID := range []uuid.UUID{winner1, winner2, loser1, loser2} {
			_, err := ratingService.GetPlayerRating(ctx, playerID, gameID)
			require.NoError(t, err)
		}

		// Update ratings after match
		cmd := matchmaking_in.UpdateRatingsCommand{
			MatchID:         uuid.New(),
			GameID:          gameID,
			WinnerPlayerIDs: []uuid.UUID{winner1, winner2},
			LoserPlayerIDs:  []uuid.UUID{loser1, loser2},
		}

		err := ratingService.UpdateRatingsAfterMatch(ctx, cmd)
		require.NoError(t, err)

		// Verify winner ratings increased
		winner1Rating, err := ratingService.GetPlayerRating(ctx, winner1, gameID)
		require.NoError(t, err)
		assert.Greater(t, winner1Rating.Rating, matchmaking_entities.DefaultRating)
		assert.Equal(t, 1, winner1Rating.Wins)
		assert.Equal(t, 1, winner1Rating.MatchesPlayed)
		assert.Equal(t, 1, winner1Rating.WinStreak)
		assert.NotNil(t, winner1Rating.LastMatchAt)

		// Verify loser ratings decreased
		loser1Rating, err := ratingService.GetPlayerRating(ctx, loser1, gameID)
		require.NoError(t, err)
		assert.Less(t, loser1Rating.Rating, matchmaking_entities.DefaultRating)
		assert.Equal(t, 1, loser1Rating.Losses)
		assert.Equal(t, 0, loser1Rating.WinStreak)

		t.Logf("✓ Winner rating: %.2f (+%.2f), Loser rating: %.2f (%.2f)",
			winner1Rating.Rating,
			winner1Rating.Rating-matchmaking_entities.DefaultRating,
			loser1Rating.Rating,
			loser1Rating.Rating-matchmaking_entities.DefaultRating,
		)
	})

	t.Run("RatingHistory_TracksChanges", func(t *testing.T) {
		playerID := uuid.New()
		opponentID := uuid.New()

		// Initialize players
		_, err := ratingService.GetPlayerRating(ctx, playerID, gameID)
		require.NoError(t, err)
		_, err = ratingService.GetPlayerRating(ctx, opponentID, gameID)
		require.NoError(t, err)

		// Play 5 matches
		for i := 0; i < 5; i++ {
			cmd := matchmaking_in.UpdateRatingsCommand{
				MatchID:         uuid.New(),
				GameID:          gameID,
				WinnerPlayerIDs: []uuid.UUID{playerID},
				LoserPlayerIDs:  []uuid.UUID{opponentID},
			}
			err := ratingService.UpdateRatingsAfterMatch(ctx, cmd)
			require.NoError(t, err)
		}

		rating, err := ratingService.GetPlayerRating(ctx, playerID, gameID)
		require.NoError(t, err)

		assert.Equal(t, 5, rating.MatchesPlayed)
		assert.Equal(t, 5, rating.Wins)
		assert.Equal(t, 5, rating.WinStreak)
		assert.Len(t, rating.RatingHistory, 5)
		assert.Greater(t, rating.Rating, matchmaking_entities.DefaultRating+100)

		t.Logf("✓ After 5 wins: Rating %.2f, Peak %.2f, History %d entries",
			rating.Rating, rating.PeakRating, len(rating.RatingHistory))
	})

	t.Run("PeakRating_TracksHighest", func(t *testing.T) {
		playerID := uuid.New()
		opponentID := uuid.New()

		// Initialize
		_, _ = ratingService.GetPlayerRating(ctx, playerID, gameID)
		_, _ = ratingService.GetPlayerRating(ctx, opponentID, gameID)

		// Win 3 matches to increase rating
		for i := 0; i < 3; i++ {
			cmd := matchmaking_in.UpdateRatingsCommand{
				MatchID:         uuid.New(),
				GameID:          gameID,
				WinnerPlayerIDs: []uuid.UUID{playerID},
				LoserPlayerIDs:  []uuid.UUID{opponentID},
			}
			_ = ratingService.UpdateRatingsAfterMatch(ctx, cmd)
		}

		rating, _ := ratingService.GetPlayerRating(ctx, playerID, gameID)
		peakAfterWins := rating.PeakRating

		// Now lose 3 matches
		for i := 0; i < 3; i++ {
			cmd := matchmaking_in.UpdateRatingsCommand{
				MatchID:         uuid.New(),
				GameID:          gameID,
				WinnerPlayerIDs: []uuid.UUID{opponentID},
				LoserPlayerIDs:  []uuid.UUID{playerID},
			}
			_ = ratingService.UpdateRatingsAfterMatch(ctx, cmd)
		}

		rating, _ = ratingService.GetPlayerRating(ctx, playerID, gameID)

		// Peak should remain at highest point
		assert.Equal(t, peakAfterWins, rating.PeakRating)
		assert.Less(t, rating.Rating, rating.PeakRating)

		t.Logf("✓ Current: %.2f, Peak: %.2f (maintained after losses)", rating.Rating, rating.PeakRating)
	})

	t.Run("Leaderboard_ReturnsTopPlayers", func(t *testing.T) {
		// Create players with different ratings
		players := make([]uuid.UUID, 5)
		for i := range players {
			players[i] = uuid.New()
			_, _ = ratingService.GetPlayerRating(ctx, players[i], gameID)
		}

		// Make player[0] the best by winning against all others
		for i := 1; i < len(players); i++ {
			cmd := matchmaking_in.UpdateRatingsCommand{
				MatchID:         uuid.New(),
				GameID:          gameID,
				WinnerPlayerIDs: []uuid.UUID{players[0]},
				LoserPlayerIDs:  []uuid.UUID{players[i]},
			}
			_ = ratingService.UpdateRatingsAfterMatch(ctx, cmd)
		}

		leaderboard, err := ratingService.GetLeaderboard(ctx, gameID, 10)
		require.NoError(t, err)
		require.NotEmpty(t, leaderboard)

		// First place should be our winner
		assert.Equal(t, players[0], leaderboard[0].PlayerID)

		t.Logf("✓ Leaderboard has %d players, top rated: %.2f", len(leaderboard), leaderboard[0].Rating)
	})

	t.Run("InactivityDecay_IncreasesRD", func(t *testing.T) {
		playerID := uuid.New()

		// Create player
		rating, _ := ratingService.GetPlayerRating(ctx, playerID, gameID)
		initialRD := rating.RatingDeviation

		// Simulate 30 days of inactivity
		pastTime := time.Now().AddDate(0, 0, -30)
		rating.LastMatchAt = &pastTime

		// Next call should apply decay
		rating.ApplyInactivityDecay(30)

		assert.Greater(t, rating.RatingDeviation, initialRD)
		t.Logf("✓ RD increased from %.2f to %.2f after 30 days inactivity", initialRD, rating.RatingDeviation)
	})

	t.Run("MultipleGames_SeparateRatings", func(t *testing.T) {
		playerID := uuid.New()

		// Get rating for CS2
		cs2Rating, err := ratingService.GetPlayerRating(ctx, playerID, common.CS2.ID)
		require.NoError(t, err)

		// Get rating for CSGO (different game)
		csgoRating, err := ratingService.GetPlayerRating(ctx, playerID, common.CSGO.ID)
		require.NoError(t, err)

		// Should be different rating records
		assert.NotEqual(t, cs2Rating.ID, csgoRating.ID)
		assert.Equal(t, common.CS2.ID, cs2Rating.GameID)
		assert.Equal(t, common.CSGO.ID, csgoRating.GameID)

		t.Log("✓ Player has separate ratings for different games")
	})
}

// getMongoTestURIForRating returns the MongoDB connection string for testing
func getMongoTestURIForRating() string {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		uri = os.Getenv("MONGODB_URI")
	}
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	return uri
}

