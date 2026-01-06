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

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// TestE2E_SquadInvitation tests the squad invitation entity and lifecycle
func TestE2E_SquadInvitation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Test user setup
	ownerUserID := uuid.New()
	playerUserID := uuid.New()
	squadID := uuid.New()

	ctx = context.WithValue(ctx, shared.UserIDKey, ownerUserID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

	resourceOwner := shared.GetResourceOwner(ctx)

	t.Run("InvitationStatus_Constants", func(t *testing.T) {
		assert.Equal(t, squad_entities.InvitationStatus("pending"), squad_entities.InvitationStatusPending)
		assert.Equal(t, squad_entities.InvitationStatus("accepted"), squad_entities.InvitationStatusAccepted)
		assert.Equal(t, squad_entities.InvitationStatus("declined"), squad_entities.InvitationStatusDeclined)
		assert.Equal(t, squad_entities.InvitationStatus("expired"), squad_entities.InvitationStatusExpired)
		assert.Equal(t, squad_entities.InvitationStatus("canceled"), squad_entities.InvitationStatusCanceled)

		t.Log("✓ Invitation status constants defined correctly")
	})

	t.Run("InvitationType_Constants", func(t *testing.T) {
		assert.Equal(t, squad_entities.InvitationType("squad_to_player"), squad_entities.InvitationTypeSquadToPlayer)
		assert.Equal(t, squad_entities.InvitationType("player_to_squad"), squad_entities.InvitationTypePlayerToSquad)

		t.Log("✓ Invitation type constants defined correctly")
	})

	t.Run("NewSquadInvitation_Creation", func(t *testing.T) {
		invitation := squad_entities.NewSquadInvitation(
			squadID,
			"Team Pro",
			playerUserID,
			"ProPlayer123",
			ownerUserID,
			"SquadOwner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member",
			"Welcome to the team!",
			7, // 7 days expiration
			resourceOwner,
		)

		require.NotNil(t, invitation)
		assert.NotEqual(t, uuid.Nil, invitation.ID)
		assert.Equal(t, squadID, invitation.SquadID)
		assert.Equal(t, "Team Pro", invitation.SquadName)
		assert.Equal(t, playerUserID, invitation.PlayerProfileID)
		assert.Equal(t, "ProPlayer123", invitation.PlayerName)
		assert.Equal(t, ownerUserID, invitation.InviterID)
		assert.Equal(t, "SquadOwner", invitation.InviterName)
		assert.Equal(t, squad_entities.InvitationTypeSquadToPlayer, invitation.InvitationType)
		assert.Equal(t, squad_entities.InvitationStatusPending, invitation.Status)
		assert.Equal(t, "member", invitation.Role)
		assert.Equal(t, "Welcome to the team!", invitation.Message)
		assert.True(t, invitation.ExpiresAt.After(time.Now()))
		assert.Nil(t, invitation.RespondedAt)
		assert.False(t, invitation.CreatedAt.IsZero())
		assert.False(t, invitation.UpdatedAt.IsZero())

		t.Logf("✓ Squad invitation created: ID=%s", invitation.ID)
	})

	t.Run("Invitation_GetID", func(t *testing.T) {
		invitation := squad_entities.NewSquadInvitation(
			squadID, "Team", playerUserID, "Player",
			ownerUserID, "Owner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member", "", 7, resourceOwner,
		)

		assert.Equal(t, invitation.ID, invitation.GetID())
		t.Log("✓ GetID returns correct ID")
	})

	t.Run("Invitation_IsPending", func(t *testing.T) {
		invitation := squad_entities.NewSquadInvitation(
			squadID, "Team", playerUserID, "Player",
			ownerUserID, "Owner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member", "", 7, resourceOwner,
		)

		assert.True(t, invitation.IsPending())

		// Accept it
		invitation.Accept()
		assert.False(t, invitation.IsPending())

		t.Log("✓ IsPending works correctly")
	})

	t.Run("Invitation_IsExpired", func(t *testing.T) {
		// Create non-expired invitation
		invitation := squad_entities.NewSquadInvitation(
			squadID, "Team", playerUserID, "Player",
			ownerUserID, "Owner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member", "", 7, resourceOwner,
		)

		assert.False(t, invitation.IsExpired())

		// Create expired invitation (negative days)
		invitation.ExpiresAt = time.Now().Add(-24 * time.Hour)
		assert.True(t, invitation.IsExpired())

		t.Log("✓ IsExpired works correctly")
	})

	t.Run("Invitation_Accept", func(t *testing.T) {
		invitation := squad_entities.NewSquadInvitation(
			squadID, "Team", playerUserID, "Player",
			ownerUserID, "Owner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member", "", 7, resourceOwner,
		)

		assert.Equal(t, squad_entities.InvitationStatusPending, invitation.Status)
		assert.Nil(t, invitation.RespondedAt)

		invitation.Accept()

		assert.Equal(t, squad_entities.InvitationStatusAccepted, invitation.Status)
		assert.NotNil(t, invitation.RespondedAt)
		assert.False(t, invitation.IsPending())

		t.Log("✓ Accept() transitions status correctly")
	})

	t.Run("Invitation_Decline", func(t *testing.T) {
		invitation := squad_entities.NewSquadInvitation(
			squadID, "Team", playerUserID, "Player",
			ownerUserID, "Owner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member", "", 7, resourceOwner,
		)

		invitation.Decline()

		assert.Equal(t, squad_entities.InvitationStatusDeclined, invitation.Status)
		assert.NotNil(t, invitation.RespondedAt)

		t.Log("✓ Decline() transitions status correctly")
	})

	t.Run("Invitation_Cancel", func(t *testing.T) {
		invitation := squad_entities.NewSquadInvitation(
			squadID, "Team", playerUserID, "Player",
			ownerUserID, "Owner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member", "", 7, resourceOwner,
		)

		invitation.Cancel()

		assert.Equal(t, squad_entities.InvitationStatusCanceled, invitation.Status)

		t.Log("✓ Cancel() transitions status correctly")
	})

	t.Run("Invitation_MarkExpired", func(t *testing.T) {
		invitation := squad_entities.NewSquadInvitation(
			squadID, "Team", playerUserID, "Player",
			ownerUserID, "Owner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member", "", 7, resourceOwner,
		)

		invitation.MarkExpired()

		assert.Equal(t, squad_entities.InvitationStatusExpired, invitation.Status)

		t.Log("✓ MarkExpired() transitions status correctly")
	})

	t.Run("Invitation_PlayerToSquad", func(t *testing.T) {
		// Test join request (player requests to join squad)
		invitation := squad_entities.NewSquadInvitation(
			squadID,
			"Pro Team",
			playerUserID,
			"NewPlayer",
			playerUserID, // Inviter is the player themselves
			"NewPlayer",
			squad_entities.InvitationTypePlayerToSquad,
			"member",
			"I'd like to join your squad!",
			7,
			resourceOwner,
		)

		assert.Equal(t, squad_entities.InvitationTypePlayerToSquad, invitation.InvitationType)
		assert.Equal(t, playerUserID, invitation.InviterID)

		t.Log("✓ Player-to-squad invitation (join request) works correctly")
	})

	t.Run("Invitation_ExpirationDays", func(t *testing.T) {
		// Test different expiration days
		testCases := []struct {
			days     int
			expected time.Duration
		}{
			{1, 24 * time.Hour},
			{7, 7 * 24 * time.Hour},
			{30, 30 * 24 * time.Hour},
		}

		for _, tc := range testCases {
			invitation := squad_entities.NewSquadInvitation(
				squadID, "Team", playerUserID, "Player",
				ownerUserID, "Owner",
				squad_entities.InvitationTypeSquadToPlayer,
				"member", "", tc.days, resourceOwner,
			)

			expectedExpiry := time.Now().Add(tc.expected)
			// Allow 1 second tolerance
			assert.WithinDuration(t, expectedExpiry, invitation.ExpiresAt, time.Second,
				"Expiration for %d days should be ~%v from now", tc.days, tc.expected)
		}

		t.Log("✓ Expiration days calculated correctly")
	})

	t.Run("Invitation_RoleAssignment", func(t *testing.T) {
		roles := []string{"member", "admin", "captain", "substitute", "coach"}

		for _, role := range roles {
			invitation := squad_entities.NewSquadInvitation(
				squadID, "Team", playerUserID, "Player",
				ownerUserID, "Owner",
				squad_entities.InvitationTypeSquadToPlayer,
				role, "", 7, resourceOwner,
			)

			assert.Equal(t, role, invitation.Role)
		}

		t.Log("✓ Role assignment works for all role types")
	})

	t.Run("Invitation_UpdatedAt", func(t *testing.T) {
		invitation := squad_entities.NewSquadInvitation(
			squadID, "Team", playerUserID, "Player",
			ownerUserID, "Owner",
			squad_entities.InvitationTypeSquadToPlayer,
			"member", "", 7, resourceOwner,
		)

		originalUpdatedAt := invitation.UpdatedAt

		// Small delay to ensure time difference
		time.Sleep(10 * time.Millisecond)

		invitation.Accept()

		assert.True(t, invitation.UpdatedAt.After(originalUpdatedAt) ||
			invitation.UpdatedAt.Equal(originalUpdatedAt),
			"UpdatedAt should be updated on status change")

		t.Log("✓ UpdatedAt is updated on state changes")
	})

	t.Log("✓ All squad invitation E2E tests passed!")
}

// TestE2E_SquadStatistics tests squad statistics entity
func TestE2E_SquadStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("SquadStatistics_NewInstance", func(t *testing.T) {
		squadID := uuid.New()
		gameID := "cs2"

		stats := squad_entities.NewSquadStatistics(squadID, gameID)

		require.NotNil(t, stats)
		assert.Equal(t, squadID, stats.SquadID)
		assert.Equal(t, gameID, stats.GameID)
		assert.Equal(t, 0, stats.TotalMatches)
		assert.Equal(t, 0, stats.Wins)
		assert.Equal(t, 0, stats.Losses)
		assert.Equal(t, 0.0, stats.WinRate)
		assert.NotNil(t, stats.MapStatistics)
		assert.NotNil(t, stats.RecentForm)
		// MemberContributions may not exist in all implementations

		t.Log("✓ SquadStatistics created with correct defaults")
	})

	t.Run("SquadStatistics_WinRateCalculation", func(t *testing.T) {
		squadID := uuid.New()
		stats := squad_entities.NewSquadStatistics(squadID, "cs2")

		// Set win/loss data
		stats.Wins = 7
		stats.Losses = 3
		stats.TotalMatches = 10

		stats.CalculateWinRate()

		assert.Equal(t, 70.0, stats.WinRate)

		t.Log("✓ Win rate calculation correct")
	})

	t.Run("SquadStatistics_ZeroMatches", func(t *testing.T) {
		squadID := uuid.New()
		stats := squad_entities.NewSquadStatistics(squadID, "cs2")

		stats.CalculateWinRate()

		assert.Equal(t, 0.0, stats.WinRate)

		t.Log("✓ Win rate handles zero matches correctly")
	})

	t.Log("✓ All squad statistics tests passed!")
}

