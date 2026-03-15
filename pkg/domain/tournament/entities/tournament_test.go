package tournament_entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestTournament() *tournament_entities.Tournament {
	now := time.Now()
	resourceOwner := shared.ResourceOwner{
		UserID: uuid.New(),
	}
	tournament, _ := tournament_entities.NewTournament(
		resourceOwner,
		"Test Championship",
		"A test tournament",
		replay_common.GameIDKey("cs2"),
		"5v5",
		"NA",
		tournament_entities.TournamentFormatSingleElimination,
		8, 2,
		wallet_vo.NewAmountFromCents(1000),
		wallet_vo.Currency("USD"),
		now.Add(24*time.Hour),
		now.Add(-1*time.Hour),
		now.Add(12*time.Hour),
		tournament_entities.TournamentRules{BestOf: 3, CheckInRequired: true},
		uuid.New(),
	)
	return tournament
}

func createTournamentWithPlayers(n int) *tournament_entities.Tournament {
	t := createTestTournament()
	t.Status = tournament_entities.TournamentStatusRegistration
	for i := 0; i < n; i++ {
		_ = t.RegisterPlayer(uuid.New(), "Player"+string(rune('A'+i)))
	}
	return t
}

func TestTournament_StateMachine_DraftToRegistration(t *testing.T) {
	tournament := createTestTournament()
	assert.Equal(t, tournament_entities.TournamentStatusDraft, tournament.Status)
	err := tournament.OpenRegistration()
	assert.NoError(t, err)
	assert.Equal(t, tournament_entities.TournamentStatusRegistration, tournament.Status)
}

func TestTournament_StateMachine_RegistrationToReady(t *testing.T) {
	tournament := createTournamentWithPlayers(4)
	assert.Equal(t, tournament_entities.TournamentStatusRegistration, tournament.Status)
	err := tournament.CloseRegistration()
	assert.NoError(t, err)
	assert.Equal(t, tournament_entities.TournamentStatusReady, tournament.Status)
}

func TestTournament_StateMachine_ReadyToInProgress(t *testing.T) {
	tournament := createTournamentWithPlayers(4)
	_ = tournament.CloseRegistration()
	tournament.StartTime = time.Now().Add(-1 * time.Hour)
	err := tournament.Start()
	assert.NoError(t, err)
	assert.Equal(t, tournament_entities.TournamentStatusInProgress, tournament.Status)
}

func TestTournament_StateMachine_InProgressToCompleted(t *testing.T) {
	tournament := createTournamentWithPlayers(4)
	_ = tournament.CloseRegistration()
	tournament.StartTime = time.Now().Add(-1 * time.Hour)
	require.NoError(t, tournament.Start())
	winners := []tournament_entities.TournamentWinner{
		{PlayerID: tournament.Participants[0].PlayerID, Placement: 1, Prize: wallet_vo.NewAmountFromCents(5000)},
		{PlayerID: tournament.Participants[1].PlayerID, Placement: 2, Prize: wallet_vo.NewAmountFromCents(2500)},
	}
	err := tournament.Complete(winners)
	assert.NoError(t, err)
	assert.Equal(t, tournament_entities.TournamentStatusCompleted, tournament.Status)
	assert.NotNil(t, tournament.EndTime)
	assert.Len(t, tournament.Winners, 2)
}

func TestTournament_StateMachine_CancelFromAnyState(t *testing.T) {
	states := []tournament_entities.TournamentStatus{
		tournament_entities.TournamentStatusDraft,
		tournament_entities.TournamentStatusRegistration,
		tournament_entities.TournamentStatusReady,
		tournament_entities.TournamentStatusInProgress,
	}
	for _, status := range states {
		t.Run(string(status), func(t *testing.T) {
			tournament := createTestTournament()
			tournament.Status = status
			err := tournament.Cancel("test cancellation reason")
			assert.NoError(t, err)
			assert.Equal(t, tournament_entities.TournamentStatusCancelled, tournament.Status)
		})
	}
}

func TestTournament_StateMachine_InvalidTransitions(t *testing.T) {
	t.Run("cannot start from draft", func(t *testing.T) {
		tournament := createTestTournament()
		err := tournament.Start()
		assert.Error(t, err)
	})
	t.Run("cannot complete from draft", func(t *testing.T) {
		tournament := createTestTournament()
		err := tournament.Complete(nil)
		assert.Error(t, err)
	})
	t.Run("cannot open registration from in_progress", func(t *testing.T) {
		tournament := createTestTournament()
		tournament.Status = tournament_entities.TournamentStatusInProgress
		err := tournament.OpenRegistration()
		assert.Error(t, err)
	})
	t.Run("cannot cancel from completed", func(t *testing.T) {
		tournament := createTestTournament()
		tournament.Status = tournament_entities.TournamentStatusCompleted
		err := tournament.Cancel("reason")
		assert.Error(t, err)
	})
	// Cancel from cancelled is allowed (idempotent)
	t.Run("cancel from cancelled is idempotent", func(t *testing.T) {
		tournament := createTestTournament()
		tournament.Status = tournament_entities.TournamentStatusCancelled
		err := tournament.Cancel("reason")
		assert.NoError(t, err)
	})
}

func TestTournament_RegisterPlayer_Success(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusRegistration
	playerID := uuid.New()
	err := tournament.RegisterPlayer(playerID, "TestPlayer")
	assert.NoError(t, err)
	assert.Len(t, tournament.Participants, 1)
	assert.Equal(t, playerID, tournament.Participants[0].PlayerID)
	assert.Equal(t, "registered", tournament.Participants[0].Status)
}

func TestTournament_RegisterPlayer_NotInRegistration(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusInProgress
	err := tournament.RegisterPlayer(uuid.New(), "TestPlayer")
	assert.Error(t, err)
}

func TestTournament_RegisterPlayer_MaxReached(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusRegistration
	tournament.MaxParticipants = 2
	_ = tournament.RegisterPlayer(uuid.New(), "P1")
	_ = tournament.RegisterPlayer(uuid.New(), "P2")
	err := tournament.RegisterPlayer(uuid.New(), "P3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "full")
}

func TestTournament_RegisterPlayer_AlreadyRegistered(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusRegistration
	playerID := uuid.New()
	_ = tournament.RegisterPlayer(playerID, "P1")
	err := tournament.RegisterPlayer(playerID, "P1")
	assert.Error(t, err)
}

func TestTournament_UnregisterPlayer_Success(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusRegistration
	playerID := uuid.New()
	_ = tournament.RegisterPlayer(playerID, "P1")
	err := tournament.UnregisterPlayer(playerID)
	assert.NoError(t, err)
	assert.Len(t, tournament.Participants, 0)
}

func TestTournament_UnregisterPlayer_NotRegistered(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusRegistration
	err := tournament.UnregisterPlayer(uuid.New())
	assert.Error(t, err)
}

func TestTournament_IsPlayerRegistered(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusRegistration
	playerID := uuid.New()
	assert.False(t, tournament.IsPlayerRegistered(playerID))
	_ = tournament.RegisterPlayer(playerID, "P1")
	assert.True(t, tournament.IsPlayerRegistered(playerID))
}

func TestTournament_CheckIn_Success(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusRegistration
	playerID := uuid.New()
	_ = tournament.RegisterPlayer(playerID, "P1")
	err := tournament.CheckIn(playerID)
	assert.NoError(t, err)
	for _, p := range tournament.Participants {
		if p.PlayerID == playerID {
			assert.Equal(t, "checked_in", p.Status)
		}
	}
}

func TestTournament_CheckIn_NotRegistered(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusRegistration
	err := tournament.CheckIn(uuid.New())
	assert.Error(t, err)
}

func TestTournament_CheckIn_WrongStatus(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusInProgress
	playerID := uuid.New()
	tournament.Participants = append(tournament.Participants, tournament_entities.TournamentPlayer{
		PlayerID: playerID,
		Status:   "registered",
	})
	err := tournament.CheckIn(playerID)
	assert.Error(t, err)
}

func TestTournament_CheckIn_NotRequired(t *testing.T) {
	tournament := createTestTournament()
	tournament.Rules.CheckInRequired = false
	tournament.Status = tournament_entities.TournamentStatusRegistration
	playerID := uuid.New()
	_ = tournament.RegisterPlayer(playerID, "P1")
	err := tournament.CheckIn(playerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not required")
}

func TestTournament_RecordMatchResult_Success(t *testing.T) {
	tournament := createTournamentWithPlayers(4)
	_ = tournament.CloseRegistration()
	tournament.StartTime = time.Now().Add(-1 * time.Hour)
	require.NoError(t, tournament.Start())
	matchID := uuid.New()
	player1 := tournament.Participants[0].PlayerID
	player2 := tournament.Participants[1].PlayerID
	tournament.Matches = append(tournament.Matches, tournament_entities.TournamentMatch{
		MatchID:   matchID,
		Round:     1,
		Player1ID: player1,
		Player2ID: player2,
		Status:    tournament_entities.MatchStatusScheduled,
	})
	err := tournament.RecordMatchResult(matchID, player1)
	assert.NoError(t, err)
	for _, m := range tournament.Matches {
		if m.MatchID == matchID {
			require.NotNil(t, m.WinnerID)
			assert.Equal(t, player1, *m.WinnerID)
			assert.Equal(t, tournament_entities.MatchStatusCompleted, m.Status)
			assert.NotNil(t, m.CompletedAt)
		}
	}
}

func TestTournament_RecordMatchResult_WrongStatus(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusDraft
	err := tournament.RecordMatchResult(uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "in_progress")
}

func TestTournament_RecordMatchResult_MatchNotFound(t *testing.T) {
	tournament := createTournamentWithPlayers(4)
	_ = tournament.CloseRegistration()
	tournament.StartTime = time.Now().Add(-1 * time.Hour)
	require.NoError(t, tournament.Start())
	err := tournament.RecordMatchResult(uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTournament_RecordMatchResult_WinnerNotParticipant(t *testing.T) {
	tournament := createTournamentWithPlayers(4)
	_ = tournament.CloseRegistration()
	tournament.StartTime = time.Now().Add(-1 * time.Hour)
	require.NoError(t, tournament.Start())
	matchID := uuid.New()
	player1 := tournament.Participants[0].PlayerID
	player2 := tournament.Participants[1].PlayerID
	tournament.Matches = append(tournament.Matches, tournament_entities.TournamentMatch{
		MatchID:   matchID,
		Round:     1,
		Player1ID: player1,
		Player2ID: player2,
		Status:    tournament_entities.MatchStatusScheduled,
	})
	outsider := uuid.New()
	err := tournament.RecordMatchResult(matchID, outsider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of the match players")
}

func TestTournament_AdvanceBracket_Success(t *testing.T) {
	tournament := createTournamentWithPlayers(4)
	_ = tournament.CloseRegistration()
	tournament.StartTime = time.Now().Add(-1 * time.Hour)
	require.NoError(t, tournament.Start())
	p1 := tournament.Participants[0].PlayerID
	p2 := tournament.Participants[1].PlayerID
	p3 := tournament.Participants[2].PlayerID
	p4 := tournament.Participants[3].PlayerID
	match1ID := uuid.New()
	match2ID := uuid.New()
	now := time.Now()
	tournament.Matches = []tournament_entities.TournamentMatch{
		{MatchID: match1ID, Round: 1, Player1ID: p1, Player2ID: p2, WinnerID: &p1, Status: tournament_entities.MatchStatusCompleted, CompletedAt: &now},
		{MatchID: match2ID, Round: 1, Player1ID: p3, Player2ID: p4, WinnerID: &p3, Status: tournament_entities.MatchStatusCompleted, CompletedAt: &now},
	}
	newMatches, err := tournament.AdvanceBracket()
	assert.NoError(t, err)
	require.NotNil(t, newMatches)
	assert.Len(t, newMatches, 1)
	assert.Equal(t, 2, newMatches[0].Round)
	assert.Equal(t, p1, newMatches[0].Player1ID)
	assert.Equal(t, p3, newMatches[0].Player2ID)
}

func TestTournament_AdvanceBracket_IncompleteMatches(t *testing.T) {
	tournament := createTournamentWithPlayers(4)
	_ = tournament.CloseRegistration()
	tournament.StartTime = time.Now().Add(-1 * time.Hour)
	require.NoError(t, tournament.Start())
	p1 := tournament.Participants[0].PlayerID
	p2 := tournament.Participants[1].PlayerID
	p3 := tournament.Participants[2].PlayerID
	p4 := tournament.Participants[3].PlayerID
	now := time.Now()
	tournament.Matches = []tournament_entities.TournamentMatch{
		{MatchID: uuid.New(), Round: 1, Player1ID: p1, Player2ID: p2, WinnerID: &p1, Status: tournament_entities.MatchStatusCompleted, CompletedAt: &now},
		{MatchID: uuid.New(), Round: 1, Player1ID: p3, Player2ID: p4, Status: tournament_entities.MatchStatusScheduled},
	}
	newMatches, err := tournament.AdvanceBracket()
	assert.Error(t, err)
	assert.Nil(t, newMatches)
}

func TestTournament_AdvanceBracket_WrongStatus(t *testing.T) {
	tournament := createTestTournament()
	tournament.Status = tournament_entities.TournamentStatusDraft
	_, err := tournament.AdvanceBracket()
	assert.Error(t, err)
}

func TestTournament_Validate_Success(t *testing.T) {
	tournament := createTestTournament()
	err := tournament.Validate()
	assert.NoError(t, err)
}

func TestTournament_Validate_EmptyName(t *testing.T) {
	tournament := createTestTournament()
	tournament.Name = ""
	err := tournament.Validate()
	assert.Error(t, err)
}

func TestTournament_Validate_NegativeEntryFee(t *testing.T) {
	tournament := createTestTournament()
	tournament.EntryFee = wallet_vo.NewAmountFromCents(-100)
	err := tournament.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "negative")
}

func TestTournament_Validate_MinGreaterThanMax(t *testing.T) {
	tournament := createTestTournament()
	tournament.MinParticipants = 16
	tournament.MaxParticipants = 8
	err := tournament.Validate()
	assert.Error(t, err)
}
