package matchmaking_entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestResourceOwner() common.ResourceOwner {
	return common.ResourceOwner{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
	}
}

func TestNewMatchmakingLobby_Success(t *testing.T) {
	creatorID := uuid.New()
	ro := createTestResourceOwner()

	lobby, err := matchmaking_entities.NewMatchmakingLobby(
		ro,
		creatorID,
		"cs2",
		"na-east",
		"premium",
		matchmaking_vo.DistributionRuleWinnerTakesAll,
		5,
		false,
		false,
	)

	require.NoError(t, err)
	assert.NotNil(t, lobby)
	assert.Equal(t, creatorID, lobby.CreatorID)
	assert.Equal(t, "cs2", lobby.GameID)
	assert.Equal(t, "na-east", lobby.Region)
	assert.Equal(t, "premium", lobby.Tier)
	assert.Equal(t, matchmaking_vo.DistributionRuleWinnerTakesAll, lobby.DistributionRule)
	assert.Equal(t, 5, lobby.MaxPlayers)
	assert.Equal(t, matchmaking_entities.LobbyStatusOpen, lobby.Status)
	assert.False(t, lobby.AutoFill)
	assert.False(t, lobby.InviteOnly)
	assert.Len(t, lobby.PlayerSlots, 5)

	// Creator should be in first slot
	assert.NotNil(t, lobby.PlayerSlots[0].PlayerID)
	assert.Equal(t, creatorID, *lobby.PlayerSlots[0].PlayerID)
	assert.Equal(t, 1, lobby.GetPlayerCount())
}

func TestNewMatchmakingLobby_TooFewPlayers(t *testing.T) {
	lobby, err := matchmaking_entities.NewMatchmakingLobby(
		createTestResourceOwner(),
		uuid.New(),
		"cs2",
		"na-east",
		"premium",
		matchmaking_vo.DistributionRuleWinnerTakesAll,
		1, // Invalid - must be at least 2
		false,
		false,
	)

	assert.Error(t, err)
	assert.Nil(t, lobby)
	assert.Contains(t, err.Error(), "at least 2 players")
}

func TestNewMatchmakingLobby_TooManyPlayers(t *testing.T) {
	lobby, err := matchmaking_entities.NewMatchmakingLobby(
		createTestResourceOwner(),
		uuid.New(),
		"cs2",
		"na-east",
		"premium",
		matchmaking_vo.DistributionRuleWinnerTakesAll,
		11, // Invalid - max is 10
		false,
		false,
	)

	assert.Error(t, err)
	assert.Nil(t, lobby)
	assert.Contains(t, err.Error(), "cannot exceed 10 players")
}

func TestNewMatchmakingLobby_InvalidDistributionRule(t *testing.T) {
	lobby, err := matchmaking_entities.NewMatchmakingLobby(
		createTestResourceOwner(),
		uuid.New(),
		"cs2",
		"na-east",
		"premium",
		matchmaking_vo.DistributionRule("invalid_rule"),
		5,
		false,
		false,
	)

	assert.Error(t, err)
	assert.Nil(t, lobby)
	assert.Contains(t, err.Error(), "invalid distribution rule")
}

func TestMatchmakingLobby_AddPlayer_Success(t *testing.T) {
	lobby := createTestLobby(t, 5)
	playerID := uuid.New()

	err := lobby.AddPlayer(playerID, 2000)

	require.NoError(t, err)
	assert.Equal(t, 2, lobby.GetPlayerCount())

	// Find the player
	found := false
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil && *slot.PlayerID == playerID {
			found = true
			assert.NotNil(t, slot.MMR)
			assert.Equal(t, 2000, *slot.MMR)
			assert.False(t, slot.IsReady)
			break
		}
	}
	assert.True(t, found, "Player should be in lobby")
}

func TestMatchmakingLobby_AddPlayer_AlreadyInLobby(t *testing.T) {
	lobby := createTestLobby(t, 5)
	playerID := uuid.New()

	err := lobby.AddPlayer(playerID, 2000)
	require.NoError(t, err)

	// Try to add same player again
	err = lobby.AddPlayer(playerID, 2000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "player already in lobby")
}

func TestMatchmakingLobby_AddPlayer_LobbyFull(t *testing.T) {
	lobby := createTestLobby(t, 2) // Creator takes first slot

	// Add second player
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)
	assert.True(t, lobby.IsFull())

	// Try to add third player
	err = lobby.AddPlayer(uuid.New(), 2000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lobby is full")
}

func TestMatchmakingLobby_AddPlayer_LobbyNotOpen(t *testing.T) {
	lobby := createTestLobby(t, 5)

	// Add player and start ready check
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)
	err = lobby.StartReadyCheck()
	require.NoError(t, err)

	// Try to add another player during ready check
	err = lobby.AddPlayer(uuid.New(), 2000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lobby is not open")
}

func TestMatchmakingLobby_RemovePlayer_Success(t *testing.T) {
	lobby := createTestLobby(t, 5)
	playerID := uuid.New()

	err := lobby.AddPlayer(playerID, 2000)
	require.NoError(t, err)
	assert.Equal(t, 2, lobby.GetPlayerCount())

	err = lobby.RemovePlayer(playerID)
	require.NoError(t, err)
	assert.Equal(t, 1, lobby.GetPlayerCount())
}

func TestMatchmakingLobby_RemovePlayer_NotInLobby(t *testing.T) {
	lobby := createTestLobby(t, 5)

	err := lobby.RemovePlayer(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "player not in lobby")
}

func TestMatchmakingLobby_RemovePlayer_CreatorLeaves(t *testing.T) {
	creatorID := uuid.New()
	lobby, err := matchmaking_entities.NewMatchmakingLobby(
		createTestResourceOwner(),
		creatorID,
		"cs2",
		"na-east",
		"premium",
		matchmaking_vo.DistributionRuleWinnerTakesAll,
		5,
		false,
		false,
	)
	require.NoError(t, err)

	err = lobby.RemovePlayer(creatorID)
	require.NoError(t, err)
	assert.Equal(t, matchmaking_entities.LobbyStatusCancelled, lobby.Status)
	assert.Contains(t, lobby.CancelReason, "creator left")
}

func TestMatchmakingLobby_RemovePlayer_FromStartedMatch(t *testing.T) {
	lobby := createTestLobby(t, 2)
	playerID := uuid.New()

	err := lobby.AddPlayer(playerID, 2000)
	require.NoError(t, err)

	// Ready up all players and start match
	for i := range lobby.PlayerSlots {
		if lobby.PlayerSlots[i].PlayerID != nil {
			_ = lobby.SetPlayerReady(*lobby.PlayerSlots[i].PlayerID, true)
		}
	}
	err = lobby.StartReadyCheck()
	require.NoError(t, err)
	err = lobby.StartMatch(uuid.New())
	require.NoError(t, err)
	err = lobby.MarkMatchStarted()
	require.NoError(t, err)

	err = lobby.RemovePlayer(playerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove player from started match")
}

func TestMatchmakingLobby_SetPlayerReady_Success(t *testing.T) {
	lobby := createTestLobby(t, 5)
	playerID := uuid.New()

	err := lobby.AddPlayer(playerID, 2000)
	require.NoError(t, err)

	err = lobby.SetPlayerReady(playerID, true)
	require.NoError(t, err)

	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil && *slot.PlayerID == playerID {
			assert.True(t, slot.IsReady)
			break
		}
	}
}

func TestMatchmakingLobby_SetPlayerReady_NotInLobby(t *testing.T) {
	lobby := createTestLobby(t, 5)

	err := lobby.SetPlayerReady(uuid.New(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "player not in lobby")
}

func TestMatchmakingLobby_StartReadyCheck_Success(t *testing.T) {
	lobby := createTestLobby(t, 5)
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)

	err = lobby.StartReadyCheck()
	require.NoError(t, err)

	assert.Equal(t, matchmaking_entities.LobbyStatusReadyCheck, lobby.Status)
	assert.NotNil(t, lobby.ReadyCheckStart)
	assert.NotNil(t, lobby.ReadyCheckEnd)
}

func TestMatchmakingLobby_StartReadyCheck_NotEnoughPlayers(t *testing.T) {
	lobby := createTestLobby(t, 5)
	// Only creator is present

	err := lobby.StartReadyCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 players")
}

func TestMatchmakingLobby_StartReadyCheck_NotOpen(t *testing.T) {
	lobby := createTestLobby(t, 5)
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)

	err = lobby.StartReadyCheck()
	require.NoError(t, err)

	// Try to start again
	err = lobby.StartReadyCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lobby is not open")
}

func TestMatchmakingLobby_CheckReadyStatus(t *testing.T) {
	lobby := createTestLobby(t, 3)
	player2 := uuid.New()
	player3 := uuid.New()

	err := lobby.AddPlayer(player2, 2000)
	require.NoError(t, err)
	err = lobby.AddPlayer(player3, 2000)
	require.NoError(t, err)

	// No one is ready yet
	allReady, notReady := lobby.CheckReadyStatus()
	assert.False(t, allReady)
	assert.Len(t, notReady, 3)

	// Mark all players ready
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil {
			_ = lobby.SetPlayerReady(*slot.PlayerID, true)
		}
	}

	allReady, notReady = lobby.CheckReadyStatus()
	assert.True(t, allReady)
	assert.Len(t, notReady, 0)
}

func TestMatchmakingLobby_StartMatch_Success(t *testing.T) {
	lobby := createTestLobby(t, 2)
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)

	// Ready up all players
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil {
			_ = lobby.SetPlayerReady(*slot.PlayerID, true)
		}
	}

	err = lobby.StartReadyCheck()
	require.NoError(t, err)

	matchID := uuid.New()
	err = lobby.StartMatch(matchID)
	require.NoError(t, err)

	assert.Equal(t, matchmaking_entities.LobbyStatusStarting, lobby.Status)
	assert.NotNil(t, lobby.MatchID)
	assert.Equal(t, matchID, *lobby.MatchID)
}

func TestMatchmakingLobby_StartMatch_NotAllReady(t *testing.T) {
	lobby := createTestLobby(t, 2)
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)

	// Only ready one player (creator)
	_ = lobby.SetPlayerReady(lobby.CreatorID, true)

	err = lobby.StartReadyCheck()
	require.NoError(t, err)

	err = lobby.StartMatch(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not all players are ready")
}

func TestMatchmakingLobby_MarkMatchStarted_Success(t *testing.T) {
	lobby := createTestLobby(t, 2)
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)

	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil {
			_ = lobby.SetPlayerReady(*slot.PlayerID, true)
		}
	}

	err = lobby.StartReadyCheck()
	require.NoError(t, err)
	err = lobby.StartMatch(uuid.New())
	require.NoError(t, err)
	err = lobby.MarkMatchStarted()
	require.NoError(t, err)

	assert.Equal(t, matchmaking_entities.LobbyStatusStarted, lobby.Status)
}

func TestMatchmakingLobby_MarkMatchStarted_NotStarting(t *testing.T) {
	lobby := createTestLobby(t, 2)

	err := lobby.MarkMatchStarted()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lobby is not starting")
}

func TestMatchmakingLobby_Cancel_Success(t *testing.T) {
	lobby := createTestLobby(t, 5)

	err := lobby.Cancel("test reason")
	require.NoError(t, err)

	assert.Equal(t, matchmaking_entities.LobbyStatusCancelled, lobby.Status)
	assert.Equal(t, "test reason", lobby.CancelReason)
}

func TestMatchmakingLobby_Cancel_AlreadyStarted(t *testing.T) {
	lobby := createTestLobby(t, 2)
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)

	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil {
			_ = lobby.SetPlayerReady(*slot.PlayerID, true)
		}
	}

	err = lobby.StartReadyCheck()
	require.NoError(t, err)
	err = lobby.StartMatch(uuid.New())
	require.NoError(t, err)
	err = lobby.MarkMatchStarted()
	require.NoError(t, err)

	err = lobby.Cancel("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel started match")
}

func TestMatchmakingLobby_GetPlayerIDs(t *testing.T) {
	lobby := createTestLobby(t, 5)
	player2 := uuid.New()
	player3 := uuid.New()

	err := lobby.AddPlayer(player2, 2000)
	require.NoError(t, err)
	err = lobby.AddPlayer(player3, 2000)
	require.NoError(t, err)

	playerIDs := lobby.GetPlayerIDs()
	assert.Len(t, playerIDs, 3)
	assert.Contains(t, playerIDs, lobby.CreatorID)
	assert.Contains(t, playerIDs, player2)
	assert.Contains(t, playerIDs, player3)
}

func TestMatchmakingLobby_IsFull(t *testing.T) {
	lobby := createTestLobby(t, 2)

	assert.False(t, lobby.IsFull())

	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)

	assert.True(t, lobby.IsFull())
}

func TestMatchmakingLobby_CanStart(t *testing.T) {
	tests := []struct {
		name        string
		maxPlayers  int
		addPlayers  int
		autoFill    bool
		setStatus   matchmaking_entities.LobbyStatus
		expectStart bool
	}{
		{
			name:        "can start - full lobby no autofill",
			maxPlayers:  2,
			addPlayers:  1,
			autoFill:    false,
			setStatus:   matchmaking_entities.LobbyStatusOpen,
			expectStart: true,
		},
		{
			name:        "can start - partial lobby no autofill",
			maxPlayers:  5,
			addPlayers:  1,
			autoFill:    false,
			setStatus:   matchmaking_entities.LobbyStatusOpen,
			expectStart: true,
		},
		{
			name:        "cannot start - partial with autofill",
			maxPlayers:  5,
			addPlayers:  1,
			autoFill:    true,
			setStatus:   matchmaking_entities.LobbyStatusOpen,
			expectStart: false,
		},
		{
			name:        "can start - full with autofill",
			maxPlayers:  2,
			addPlayers:  1,
			autoFill:    true,
			setStatus:   matchmaking_entities.LobbyStatusOpen,
			expectStart: true,
		},
		{
			name:        "cannot start - only 1 player",
			maxPlayers:  5,
			addPlayers:  0,
			autoFill:    false,
			setStatus:   matchmaking_entities.LobbyStatusOpen,
			expectStart: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lobby, err := matchmaking_entities.NewMatchmakingLobby(
				createTestResourceOwner(),
				uuid.New(),
				"cs2",
				"na-east",
				"premium",
				matchmaking_vo.DistributionRuleWinnerTakesAll,
				tt.maxPlayers,
				tt.autoFill,
				false,
			)
			require.NoError(t, err)

			for i := 0; i < tt.addPlayers; i++ {
				_ = lobby.AddPlayer(uuid.New(), 2000)
			}

			assert.Equal(t, tt.expectStart, lobby.CanStart())
		})
	}
}

func TestMatchmakingLobby_IsReadyCheckExpired(t *testing.T) {
	lobby := createTestLobby(t, 2)
	err := lobby.AddPlayer(uuid.New(), 2000)
	require.NoError(t, err)

	// Not in ready check - should be false
	assert.False(t, lobby.IsReadyCheckExpired())

	// Start ready check with a very short timeout
	lobby.ReadyTimeout = 1 * time.Millisecond
	err = lobby.StartReadyCheck()
	require.NoError(t, err)

	// Wait for expiry
	time.Sleep(5 * time.Millisecond)
	assert.True(t, lobby.IsReadyCheckExpired())
}

func TestMatchmakingLobby_Validate(t *testing.T) {
	tests := []struct {
		name        string
		modify      func(*matchmaking_entities.MatchmakingLobby)
		expectError string
	}{
		{
			name:   "valid lobby",
			modify: func(l *matchmaking_entities.MatchmakingLobby) {},
		},
		{
			name: "invalid max players - too few",
			modify: func(l *matchmaking_entities.MatchmakingLobby) {
				l.MaxPlayers = 1
			},
			expectError: "max players must be between 2 and 10",
		},
		{
			name: "invalid max players - too many",
			modify: func(l *matchmaking_entities.MatchmakingLobby) {
				l.MaxPlayers = 11
			},
			expectError: "max players must be between 2 and 10",
		},
		{
			name: "slots count mismatch",
			modify: func(l *matchmaking_entities.MatchmakingLobby) {
				l.PlayerSlots = l.PlayerSlots[:2]
			},
			expectError: "player slots count",
		},
		{
			name: "invalid distribution rule",
			modify: func(l *matchmaking_entities.MatchmakingLobby) {
				l.DistributionRule = matchmaking_vo.DistributionRule("invalid")
			},
			expectError: "invalid distribution rule",
		},
		{
			name: "ready check with too few players",
			modify: func(l *matchmaking_entities.MatchmakingLobby) {
				l.Status = matchmaking_entities.LobbyStatusReadyCheck
				// Remove all but creator
				for i := range l.PlayerSlots {
					if i > 0 {
						l.PlayerSlots[i].PlayerID = nil
					}
				}
			},
			expectError: "ready check requires at least 2 players",
		},
		{
			name: "started without match ID",
			modify: func(l *matchmaking_entities.MatchmakingLobby) {
				l.Status = matchmaking_entities.LobbyStatusStarted
				l.MatchID = nil
			},
			expectError: "started lobby must have match ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lobby := createTestLobby(t, 5)
			tt.modify(lobby)

			err := lobby.Validate()
			if tt.expectError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			}
		})
	}
}

func TestMatchmakingLobby_GetID(t *testing.T) {
	lobby := createTestLobby(t, 5)
	assert.Equal(t, lobby.ID, lobby.GetID())
}

func TestLobbyStatus_Constants(t *testing.T) {
	assert.Equal(t, matchmaking_entities.LobbyStatus("open"), matchmaking_entities.LobbyStatusOpen)
	assert.Equal(t, matchmaking_entities.LobbyStatus("ready_check"), matchmaking_entities.LobbyStatusReadyCheck)
	assert.Equal(t, matchmaking_entities.LobbyStatus("starting"), matchmaking_entities.LobbyStatusStarting)
	assert.Equal(t, matchmaking_entities.LobbyStatus("started"), matchmaking_entities.LobbyStatusStarted)
	assert.Equal(t, matchmaking_entities.LobbyStatus("cancelled"), matchmaking_entities.LobbyStatusCancelled)
}

// Helper function to create a test lobby
func createTestLobby(t *testing.T, maxPlayers int) *matchmaking_entities.MatchmakingLobby {
	lobby, err := matchmaking_entities.NewMatchmakingLobby(
		createTestResourceOwner(),
		uuid.New(),
		"cs2",
		"na-east",
		"premium",
		matchmaking_vo.DistributionRuleWinnerTakesAll,
		maxPlayers,
		false,
		false,
	)
	require.NoError(t, err)
	return lobby
}
