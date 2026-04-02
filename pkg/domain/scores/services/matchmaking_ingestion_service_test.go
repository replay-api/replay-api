package scores_services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	"github.com/replay-api/replay-api/pkg/infra/kafka"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Mocks ----

type mockMatchResultRepository struct {
	findByMatchIDResult *scores_entities.MatchResult
	findByMatchIDErr    error
	savedResults        []*scores_entities.MatchResult
}

func (m *mockMatchResultRepository) Save(_ context.Context, result *scores_entities.MatchResult) error {
	m.savedResults = append(m.savedResults, result)
	return nil
}

func (m *mockMatchResultRepository) FindByID(_ context.Context, _ uuid.UUID) (*scores_entities.MatchResult, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockMatchResultRepository) FindByMatchID(_ context.Context, _ uuid.UUID) (*scores_entities.MatchResult, error) {
	return m.findByMatchIDResult, m.findByMatchIDErr
}

func (m *mockMatchResultRepository) FindByTournamentID(_ context.Context, _ uuid.UUID) ([]*scores_entities.MatchResult, error) {
	return nil, nil
}

func (m *mockMatchResultRepository) FindByMatchmakingSessionID(_ context.Context, _ uuid.UUID) (*scores_entities.MatchResult, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockMatchResultRepository) FindByStatus(_ context.Context, _ scores_vo.ResultStatus, _ int, _ int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockMatchResultRepository) FindByPlayerID(_ context.Context, _ uuid.UUID, _ int, _ int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockMatchResultRepository) FindByTeamID(_ context.Context, _ uuid.UUID, _ int, _ int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockMatchResultRepository) Update(_ context.Context, _ *scores_entities.MatchResult) error {
	return nil
}

func (m *mockMatchResultRepository) Count(_ context.Context, _ scores_out.MatchResultFilter) (int64, error) {
	return 0, nil
}

func (m *mockMatchResultRepository) Search(_ context.Context, _ scores_out.MatchResultFilter, _ int, _ int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

type mockCommandHandler struct {
	submitCalled  bool
	submittedCmd  scores_in.SubmitMatchResultCommand
	submitResult  *scores_entities.MatchResult
	submitErr     error
}

func (m *mockCommandHandler) SubmitMatchResult(_ context.Context, cmd scores_in.SubmitMatchResultCommand) (*scores_entities.MatchResult, error) {
	m.submitCalled = true
	m.submittedCmd = cmd
	if m.submitErr != nil {
		return nil, m.submitErr
	}
	if m.submitResult != nil {
		return m.submitResult, nil
	}
	return &scores_entities.MatchResult{
		BaseEntity: shared.BaseEntity{ID: uuid.New()},
		MatchID:    cmd.MatchID,
		Status:     scores_vo.ResultStatusSubmitted,
	}, nil
}

func (m *mockCommandHandler) SubmitMatchResultFromReplay(_ context.Context, _ scores_in.SubmitReplayResultCommand) (*scores_entities.MatchResult, error) {
	return nil, nil
}

func (m *mockCommandHandler) VerifyMatchResult(_ context.Context, _ scores_in.VerifyMatchResultCommand) error {
	return nil
}

func (m *mockCommandHandler) DisputeMatchResult(_ context.Context, _ scores_in.DisputeMatchResultCommand) error {
	return nil
}

func (m *mockCommandHandler) ConciliateMatchResult(_ context.Context, _ scores_in.ConciliateMatchResultCommand) error {
	return nil
}

func (m *mockCommandHandler) FinalizeMatchResult(_ context.Context, _ scores_in.FinalizeMatchResultCommand) error {
	return nil
}

func (m *mockCommandHandler) CancelMatchResult(_ context.Context, _ scores_in.CancelMatchResultCommand) error {
	return nil
}

// ---- Tests ----

func TestProcessMatchEvent_NilEvent(t *testing.T) {
	svc := NewMatchmakingIngestionService(&mockMatchResultRepository{}, &mockCommandHandler{})
	err := svc.ProcessMatchEvent(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil match event")
}

func TestProcessMatchEvent_MissingMatchID(t *testing.T) {
	svc := NewMatchmakingIngestionService(&mockMatchResultRepository{}, &mockCommandHandler{})
	err := svc.ProcessMatchEvent(context.Background(), &kafka.MatchEvent{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing match_id")
}

func TestProcessMatchEvent_NoResult_Skipped(t *testing.T) {
	svc := NewMatchmakingIngestionService(&mockMatchResultRepository{}, &mockCommandHandler{})
	err := svc.ProcessMatchEvent(context.Background(), &kafka.MatchEvent{
		MatchID:   uuid.New(),
		EventType: "MATCH_CREATED",
	})
	assert.NoError(t, err) // No error — silently skipped
}

func TestProcessMatchEvent_Idempotent_ExistingResult(t *testing.T) {
	matchID := uuid.New()
	repo := &mockMatchResultRepository{
		findByMatchIDResult: &scores_entities.MatchResult{
			BaseEntity: shared.BaseEntity{ID: uuid.New()},
			MatchID:    matchID,
			Status:     scores_vo.ResultStatusSubmitted,
		},
	}
	cmdHandler := &mockCommandHandler{}
	svc := NewMatchmakingIngestionService(repo, cmdHandler)

	err := svc.ProcessMatchEvent(context.Background(), &kafka.MatchEvent{
		MatchID:   matchID,
		EventType: "MATCH_COMPLETED",
		Result: &kafka.MatchResult{
			WinnerTeamID: ptrUUID(uuid.New()),
		},
		Teams: []kafka.TeamInfo{
			{TeamID: uuid.New(), Name: "Team A", PlayerIDs: []uuid.UUID{uuid.New()}},
			{TeamID: uuid.New(), Name: "Team B", PlayerIDs: []uuid.UUID{uuid.New()}},
		},
	})

	assert.NoError(t, err)
	assert.False(t, cmdHandler.submitCalled, "should not submit when result already exists")
}

func TestProcessMatchEvent_Success(t *testing.T) {
	matchID := uuid.New()
	lobbyID := uuid.New()
	teamA := uuid.New()
	teamB := uuid.New()
	playerA := uuid.New()
	playerB := uuid.New()

	repo := &mockMatchResultRepository{
		findByMatchIDErr: fmt.Errorf("not found"),
	}
	cmdHandler := &mockCommandHandler{}
	svc := NewMatchmakingIngestionService(repo, cmdHandler)

	completedAt := time.Now().Add(-5 * time.Minute).UnixMilli()

	err := svc.ProcessMatchEvent(context.Background(), &kafka.MatchEvent{
		MatchID:   matchID,
		LobbyID:   lobbyID,
		EventType: "MATCH_COMPLETED",
		GameType:  "cs2",
		Metadata:  map[string]string{"map_name": "de_dust2"},
		Teams: []kafka.TeamInfo{
			{TeamID: teamA, Name: "Team Alpha", PlayerIDs: []uuid.UUID{playerA}},
			{TeamID: teamB, Name: "Team Bravo", PlayerIDs: []uuid.UUID{playerB}},
		},
		Result: &kafka.MatchResult{
			WinnerTeamID: &teamA,
			Scores:       map[string]int{teamA.String(): 16, teamB.String(): 10},
			Duration:     1800,
			CompletedAt:  completedAt,
			PlayerStats: []kafka.PlayerMatchStat{
				{PlayerID: playerA, Kills: 25, Deaths: 10, Assists: 5, Score: 80, MMRChange: 15},
				{PlayerID: playerB, Kills: 10, Deaths: 20, Assists: 3, Score: 40, MMRChange: -15},
			},
		},
	})

	require.NoError(t, err)
	require.True(t, cmdHandler.submitCalled)

	cmd := cmdHandler.submittedCmd
	assert.Equal(t, matchID, cmd.MatchID)
	assert.Equal(t, scores_vo.ScoreSourceMatchmaking, cmd.Source)
	assert.Equal(t, "matchmaking", cmd.Mode)
	assert.Equal(t, "de_dust2", cmd.MapName)
	assert.NotNil(t, cmd.MatchmakingSessionID)
	assert.Equal(t, lobbyID, *cmd.MatchmakingSessionID)

	// Team results
	require.Len(t, cmd.TeamResults, 2)
	assert.Equal(t, teamA, cmd.TeamResults[0].TeamID)
	assert.Equal(t, 16, cmd.TeamResults[0].Score)
	assert.Equal(t, 1, cmd.TeamResults[0].Position) // winner

	// Player results
	require.Len(t, cmd.PlayerResults, 2)
	assert.Equal(t, playerA, cmd.PlayerResults[0].PlayerID)
	assert.Equal(t, 25, cmd.PlayerResults[0].Kills)
	assert.Equal(t, 10, cmd.PlayerResults[0].Deaths)
	assert.Equal(t, 5, cmd.PlayerResults[0].Assists)
	assert.Equal(t, 15, cmd.PlayerResults[0].Stats["mmr_change"])
}

func TestProcessMatchEvent_SubmitError(t *testing.T) {
	repo := &mockMatchResultRepository{
		findByMatchIDErr: fmt.Errorf("not found"),
	}
	cmdHandler := &mockCommandHandler{
		submitErr: fmt.Errorf("command handler error"),
	}
	svc := NewMatchmakingIngestionService(repo, cmdHandler)

	err := svc.ProcessMatchEvent(context.Background(), &kafka.MatchEvent{
		MatchID:   uuid.New(),
		EventType: "MATCH_COMPLETED",
		Teams: []kafka.TeamInfo{
			{TeamID: uuid.New(), Name: "A", PlayerIDs: []uuid.UUID{uuid.New()}},
			{TeamID: uuid.New(), Name: "B", PlayerIDs: []uuid.UUID{uuid.New()}},
		},
		Result: &kafka.MatchResult{
			WinnerTeamID: ptrUUID(uuid.New()),
			Scores:       map[string]int{},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to submit matchmaking result")
}

func TestBuildTeamResults_WinnerFirstPosition(t *testing.T) {
	winnerID := uuid.New()
	loserID := uuid.New()

	svc := NewMatchmakingIngestionService(nil, nil)
	event := &kafka.MatchEvent{
		Teams: []kafka.TeamInfo{
			{TeamID: loserID, Name: "Loser", PlayerIDs: []uuid.UUID{uuid.New()}},
			{TeamID: winnerID, Name: "Winner", PlayerIDs: []uuid.UUID{uuid.New()}},
		},
		Result: &kafka.MatchResult{
			WinnerTeamID: &winnerID,
			Scores:       map[string]int{winnerID.String(): 16, loserID.String(): 8},
		},
	}

	results := svc.buildTeamResults(event)
	require.Len(t, results, 2)

	// Winner should have position 1 regardless of order
	for _, r := range results {
		if r.TeamID == winnerID {
			assert.Equal(t, 1, r.Position)
			assert.Equal(t, 16, r.Score)
		}
	}
}

func TestBuildPlayerResults_Empty(t *testing.T) {
	svc := NewMatchmakingIngestionService(nil, nil)
	event := &kafka.MatchEvent{
		Result: &kafka.MatchResult{},
	}

	results := svc.buildPlayerResults(event)
	assert.Nil(t, results)
}

func TestBuildPlayerResults_MapsTeamID(t *testing.T) {
	teamA := uuid.New()
	playerA := uuid.New()

	svc := NewMatchmakingIngestionService(nil, nil)
	event := &kafka.MatchEvent{
		Teams: []kafka.TeamInfo{
			{TeamID: teamA, Name: "A", PlayerIDs: []uuid.UUID{playerA}},
		},
		Result: &kafka.MatchResult{
			PlayerStats: []kafka.PlayerMatchStat{
				{PlayerID: playerA, Kills: 20, Deaths: 5, Assists: 10, Score: 60, MMRChange: 12},
			},
		},
	}

	results := svc.buildPlayerResults(event)
	require.Len(t, results, 1)
	assert.Equal(t, teamA, results[0].TeamID)
	assert.Equal(t, playerA, results[0].PlayerID)
	assert.Equal(t, 20, results[0].Kills)
	assert.Equal(t, 12, results[0].Stats["mmr_change"])
}

// ---- Helpers ----

func ptrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}
