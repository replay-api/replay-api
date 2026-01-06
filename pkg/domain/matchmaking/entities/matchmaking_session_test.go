package matchmaking_entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	"github.com/stretchr/testify/assert"
)

func TestMatchmakingSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "session not expired - future expiry",
			expiresAt: time.Now().Add(10 * time.Minute),
			expected:  false,
		},
		{
			name:      "session expired - past expiry",
			expiresAt: time.Now().Add(-10 * time.Minute),
			expected:  true,
		},
		{
			name:      "session expired - exactly now (edge case)",
			expiresAt: time.Now().Add(-1 * time.Millisecond),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &matchmaking_entities.MatchmakingSession{
				ExpiresAt: tt.expiresAt,
			}

			result := session.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchmakingSession_CanMatch(t *testing.T) {
	tests := []struct {
		name     string
		status   matchmaking_entities.SessionStatus
		expected bool
	}{
		{
			name:     "can match when queued",
			status:   matchmaking_entities.StatusQueued,
			expected: true,
		},
		{
			name:     "can match when searching",
			status:   matchmaking_entities.StatusSearching,
			expected: true,
		},
		{
			name:     "cannot match when matched",
			status:   matchmaking_entities.StatusMatched,
			expected: false,
		},
		{
			name:     "cannot match when ready",
			status:   matchmaking_entities.StatusReady,
			expected: false,
		},
		{
			name:     "cannot match when cancelled",
			status:   matchmaking_entities.StatusCancelled,
			expected: false,
		},
		{
			name:     "cannot match when expired",
			status:   matchmaking_entities.StatusExpired,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &matchmaking_entities.MatchmakingSession{
				Status: tt.status,
			}

			result := session.CanMatch()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchmakingSession_GetTierPriority(t *testing.T) {
	tests := []struct {
		name     string
		tier     matchmaking_entities.MatchmakingTier
		expected int
	}{
		{
			name:     "elite tier has highest priority",
			tier:     matchmaking_entities.TierElite,
			expected: 4,
		},
		{
			name:     "pro tier has priority 3",
			tier:     matchmaking_entities.TierPro,
			expected: 3,
		},
		{
			name:     "premium tier has priority 2",
			tier:     matchmaking_entities.TierPremium,
			expected: 2,
		},
		{
			name:     "free tier has priority 1",
			tier:     matchmaking_entities.TierFree,
			expected: 1,
		},
		{
			name:     "unknown tier defaults to priority 1",
			tier:     matchmaking_entities.MatchmakingTier("unknown"),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &matchmaking_entities.MatchmakingSession{
				Preferences: matchmaking_entities.MatchPreferences{
					Tier: tt.tier,
				},
			}

			result := session.GetTierPriority()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchmakingSession_GetID(t *testing.T) {
	expectedID := uuid.New()
	resourceOwner := shared.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		UserID:   uuid.New(),
	}
	session := &matchmaking_entities.MatchmakingSession{
		BaseEntity: shared.NewEntity(resourceOwner),
	}
	session.ID = expectedID // Override the generated ID

	assert.Equal(t, expectedID, session.GetID())
}

func TestMatchPreferences_SkillRange(t *testing.T) {
	skillRange := matchmaking_entities.SkillRange{
		MinMMR: 1000,
		MaxMMR: 2000,
	}

	assert.Equal(t, 1000, skillRange.MinMMR)
	assert.Equal(t, 2000, skillRange.MaxMMR)
}

func TestMatchmakingSession_FullSession(t *testing.T) {
	playerID := uuid.New()
	squadID := uuid.New()
	now := time.Now()

	resourceOwner := shared.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		UserID:   playerID,
	}
	session := &matchmaking_entities.MatchmakingSession{
		BaseEntity: shared.NewEntity(resourceOwner),
		PlayerID:   playerID,
		SquadID:    &squadID,
		Preferences: matchmaking_entities.MatchPreferences{
			GameID:   "cs2",
			GameMode: "competitive",
			Region:   "na-east",
			SkillRange: matchmaking_entities.SkillRange{
				MinMMR: 1500,
				MaxMMR: 2500,
			},
			MaxPing:            100,
			AllowCrossPlatform: true,
			Tier:               matchmaking_entities.TierPro,
			PriorityBoost:      true,
		},
		Status:        matchmaking_entities.StatusQueued,
		PlayerMMR:     2000,
		QueuedAt:      now,
		EstimatedWait: 30,
		ExpiresAt:     now.Add(10 * time.Minute),
	}
	session.ID = uuid.New() // Override the generated ID for testing

	assert.True(t, session.CanMatch())
	assert.False(t, session.IsExpired())
	assert.Equal(t, 3, session.GetTierPriority())
	assert.Equal(t, playerID, session.PlayerID)
	assert.Equal(t, squadID, *session.SquadID)
}

func TestSessionStatus_Constants(t *testing.T) {
	assert.Equal(t, matchmaking_entities.SessionStatus("queued"), matchmaking_entities.StatusQueued)
	assert.Equal(t, matchmaking_entities.SessionStatus("searching"), matchmaking_entities.StatusSearching)
	assert.Equal(t, matchmaking_entities.SessionStatus("matched"), matchmaking_entities.StatusMatched)
	assert.Equal(t, matchmaking_entities.SessionStatus("ready"), matchmaking_entities.StatusReady)
	assert.Equal(t, matchmaking_entities.SessionStatus("cancelled"), matchmaking_entities.StatusCancelled)
	assert.Equal(t, matchmaking_entities.SessionStatus("expired"), matchmaking_entities.StatusExpired)
}

func TestMatchmakingTier_Constants(t *testing.T) {
	assert.Equal(t, matchmaking_entities.MatchmakingTier("free"), matchmaking_entities.TierFree)
	assert.Equal(t, matchmaking_entities.MatchmakingTier("premium"), matchmaking_entities.TierPremium)
	assert.Equal(t, matchmaking_entities.MatchmakingTier("pro"), matchmaking_entities.TierPro)
	assert.Equal(t, matchmaking_entities.MatchmakingTier("elite"), matchmaking_entities.TierElite)
}
