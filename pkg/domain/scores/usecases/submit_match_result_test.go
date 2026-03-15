package scores_usecases

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
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Mock Implementations
// ============================================================

// --- Mock Repository ---

type mockRepository struct {
	results    map[uuid.UUID]*scores_entities.MatchResult
	matchIndex map[uuid.UUID]*scores_entities.MatchResult
	savedCount int
	updateErr  error
	saveErr    error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		results:    make(map[uuid.UUID]*scores_entities.MatchResult),
		matchIndex: make(map[uuid.UUID]*scores_entities.MatchResult),
	}
}

func (m *mockRepository) Save(_ context.Context, result *scores_entities.MatchResult) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.results[result.ID] = result
	m.matchIndex[result.MatchID] = result
	m.savedCount++
	return nil
}

func (m *mockRepository) FindByID(_ context.Context, id uuid.UUID) (*scores_entities.MatchResult, error) {
	r, ok := m.results[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockRepository) FindByMatchID(_ context.Context, matchID uuid.UUID) (*scores_entities.MatchResult, error) {
	r, ok := m.matchIndex[matchID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockRepository) FindByTournamentID(_ context.Context, _ uuid.UUID) ([]*scores_entities.MatchResult, error) {
	return nil, nil
}

func (m *mockRepository) FindByMatchmakingSessionID(_ context.Context, _ uuid.UUID) (*scores_entities.MatchResult, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockRepository) FindByStatus(_ context.Context, _ scores_vo.ResultStatus, _ int, _ int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockRepository) FindByPlayerID(_ context.Context, _ uuid.UUID, _ int, _ int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockRepository) FindByTeamID(_ context.Context, _ uuid.UUID, _ int, _ int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockRepository) Update(_ context.Context, result *scores_entities.MatchResult) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.results[result.ID] = result
	return nil
}

func (m *mockRepository) Count(_ context.Context, _ scores_out.MatchResultFilter) (int64, error) {
	return 0, nil
}

func (m *mockRepository) Search(_ context.Context, _ scores_out.MatchResultFilter, _ int, _ int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

// --- Mock Event Publisher ---

type mockEventPublisher struct {
	submittedEvents   int
	verifiedEvents    int
	disputedEvents    int
	conciliatedEvents int
	finalizedEvents   int
	cancelledEvents   int
}

func (m *mockEventPublisher) PublishMatchResultSubmitted(_ context.Context, _ *scores_entities.MatchResult) error {
	m.submittedEvents++
	return nil
}

func (m *mockEventPublisher) PublishMatchResultVerified(_ context.Context, _ *scores_entities.MatchResult) error {
	m.verifiedEvents++
	return nil
}

func (m *mockEventPublisher) PublishMatchResultDisputed(_ context.Context, _ *scores_entities.MatchResult) error {
	m.disputedEvents++
	return nil
}

func (m *mockEventPublisher) PublishMatchResultConciliated(_ context.Context, _ *scores_entities.MatchResult) error {
	m.conciliatedEvents++
	return nil
}

func (m *mockEventPublisher) PublishMatchResultFinalized(_ context.Context, _ *scores_entities.MatchResult) error {
	m.finalizedEvents++
	return nil
}

func (m *mockEventPublisher) PublishMatchResultCancelled(_ context.Context, _ *scores_entities.MatchResult) error {
	m.cancelledEvents++
	return nil
}

// --- Mock Prize Distribution Gateway ---

type mockPrizeGateway struct {
	tournamentCalled  bool
	matchmakingCalled bool
	returnID          *uuid.UUID
	returnErr         error
}

func (m *mockPrizeGateway) TriggerTournamentPrizeDistribution(_ context.Context, _ uuid.UUID, _ []scores_entities.RankedResult) (*uuid.UUID, error) {
	m.tournamentCalled = true
	return m.returnID, m.returnErr
}

func (m *mockPrizeGateway) TriggerMatchmakingPrizeDistribution(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _ *uuid.UUID) (*uuid.UUID, error) {
	m.matchmakingCalled = true
	return m.returnID, m.returnErr
}

// --- Mock Authorization ---

type mockAuthorization struct {
	isPlatformAdmin   bool
	isOrganizerResult bool
	isParticipant     bool
	organizerErr      error
	participantErr    error
}

func (m *mockAuthorization) IsTournamentOrganizer(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	return m.isOrganizerResult, m.organizerErr
}

func (m *mockAuthorization) IsMatchParticipant(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	return m.isParticipant, m.participantErr
}

func (m *mockAuthorization) IsPlatformAdmin(_ context.Context) bool {
	return m.isPlatformAdmin
}

// ============================================================
// Test Helpers
// ============================================================

func testContext(userID, tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	return ctx
}

func testAdminContext(userID, tenantID uuid.UUID) context.Context {
	ctx := testContext(userID, tenantID)
	ctx = context.WithValue(ctx, shared.AudienceKey, shared.TenantAudienceIDKey)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	return ctx
}

func testTeamResults() []scores_entities.TeamResult {
	return []scores_entities.TeamResult{
		{TeamID: uuid.New(), TeamName: "Team Alpha", Score: 16, Players: []uuid.UUID{uuid.New(), uuid.New()}},
		{TeamID: uuid.New(), TeamName: "Team Beta", Score: 12, Players: []uuid.UUID{uuid.New(), uuid.New()}},
	}
}

func testPlayerResults(teams []scores_entities.TeamResult) []scores_entities.PlayerResult {
	var results []scores_entities.PlayerResult
	for _, team := range teams {
		for _, pid := range team.Players {
			results = append(results, scores_entities.PlayerResult{
				PlayerID: pid,
				TeamID:   team.TeamID,
				Score:    100,
				Kills:    20,
				Deaths:   15,
				Assists:  5,
				Rating:   1.15,
			})
		}
	}
	if len(results) > 0 {
		results[0].IsMVP = true
	}
	return results
}

func newTestHandler() (*matchResultCommandHandler, *mockRepository, *mockEventPublisher, *mockPrizeGateway, *mockAuthorization) {
	repo := newMockRepository()
	events := &mockEventPublisher{}
	prizes := &mockPrizeGateway{}
	auth := &mockAuthorization{isPlatformAdmin: true}

	handler := &matchResultCommandHandler{
		repository:               repo,
		eventPublisher:           events,
		prizeDistributionGateway: prizes,
		authorization:            auth,
	}
	return handler, repo, events, prizes, auth
}

func submitAndVerify(t *testing.T, handler *matchResultCommandHandler, ctx context.Context) *scores_entities.MatchResult {
	t.Helper()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	// ScoreSourceTournamentAdmin auto-verifies on submission, so no separate verify call needed
	result, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})
	require.NoError(t, err)
	require.Equal(t, scores_vo.ResultStatusVerified, result.Status, "tournament_admin source should auto-verify")

	return result
}

// ============================================================
// Tests: SubmitMatchResult
// ============================================================

func TestSubmitMatchResult_HappyPath(t *testing.T) {
	handler, repo, events, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testAdminContext(userID, tenantID)

	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, scores_vo.ResultStatusVerified, result.Status)
	assert.Equal(t, 1, repo.savedCount)
	assert.Equal(t, 1, events.submittedEvents)
	assert.Equal(t, 1, events.verifiedEvents)
}

func TestSubmitMatchResult_Idempotency(t *testing.T) {
	handler, _, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testAdminContext(userID, tenantID)

	teams := testTeamResults()
	players := testPlayerResults(teams)
	matchID := uuid.New()

	_, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       matchID,
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})
	require.NoError(t, err)

	_, err = handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       matchID,
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestSubmitMatchResult_InvalidCommand(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	_, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		GameID: "cs2",
		Source: scores_vo.ScoreSourceTournamentAdmin,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command")
}

func TestSubmitMatchResult_RBAC_NonAdminBlocked(t *testing.T) {
	handler, _, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = false
	auth.isOrganizerResult = false
	userID := uuid.New()
	ctx := testContext(userID, uuid.New())

	teams := testTeamResults()
	players := testPlayerResults(teams)

	_, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceExternalAPI,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestSubmitMatchResult_RBAC_OrganizerAllowed(t *testing.T) {
	handler, _, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = false
	auth.isOrganizerResult = true
	userID := uuid.New()
	ctx := testContext(userID, uuid.New())

	teams := testTeamResults()
	players := testPlayerResults(teams)
	tournamentID := uuid.New()

	result, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TournamentID:  &tournamentID,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSubmitMatchResult_FuturePlayedAtRejected(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	teams := testTeamResults()
	players := testPlayerResults(teams)

	_, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(2 * time.Hour),
		Duration:      45 * time.Minute,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "future")
}

func TestSubmitMatchResult_NegativeScoreRejected(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	teams := testTeamResults()
	teams[0].Score = -1
	players := testPlayerResults(teams)

	_, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative")
}

// ============================================================
// Tests: SubmitMatchResultFromReplay
// ============================================================

func TestSubmitMatchResultFromReplay_AutoVerified(t *testing.T) {
	handler, _, events, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := handler.SubmitMatchResultFromReplay(ctx, scores_in.SubmitReplayResultCommand{
		MatchID:       uuid.New(),
		ReplayID:      uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      50 * time.Minute,
	})

	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusVerified, result.Status)
	assert.Equal(t, 1, events.submittedEvents)
	assert.Equal(t, 1, events.verifiedEvents)
}

// ============================================================
// Tests: VerifyMatchResult
// ============================================================

func TestVerifyMatchResult_HappyPath(t *testing.T) {
	handler, repo, events, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testAdminContext(userID, tenantID)

	teams := testTeamResults()
	players := testPlayerResults(teams)
	ro := shared.NewResourceOwner(tenantID, uuid.Nil, uuid.Nil, userID)
	result, err := scores_entities.NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceConsensus, userID,
		teams, players, time.Now().Add(-1*time.Hour), 45*time.Minute,
	)
	require.NoError(t, err)
	repo.results[result.ID] = result

	err = handler.VerifyMatchResult(ctx, scores_in.VerifyMatchResultCommand{
		MatchResultID:      result.ID,
		VerificationMethod: scores_vo.VerificationMethodManual,
	})

	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusVerified, result.Status)
	assert.Equal(t, 1, events.verifiedEvents)
}

func TestVerifyMatchResult_RBAC_NonAdminBlocked(t *testing.T) {
	handler, repo, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = false
	auth.isOrganizerResult = false
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testContext(userID, tenantID)

	teams := testTeamResults()
	players := testPlayerResults(teams)
	ro := shared.NewResourceOwner(tenantID, uuid.Nil, uuid.Nil, userID)
	result, _ := scores_entities.NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceConsensus, userID,
		teams, players, time.Now().Add(-1*time.Hour), 45*time.Minute,
	)
	repo.results[result.ID] = result

	err := handler.VerifyMatchResult(ctx, scores_in.VerifyMatchResultCommand{
		MatchResultID:      result.ID,
		VerificationMethod: scores_vo.VerificationMethodManual,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

// ============================================================
// Tests: DisputeMatchResult
// ============================================================

func TestDisputeMatchResult_HappyPath(t *testing.T) {
	handler, _, events, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testAdminContext(userID, tenantID)

	result := submitAndVerify(t, handler, ctx)

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Wrong scores reported",
	})

	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusDisputed, result.Status)
	assert.Equal(t, 1, result.DisputeCount)
	assert.Equal(t, 1, events.disputedEvents)
}

func TestDisputeMatchResult_RBAC_ParticipantAllowed(t *testing.T) {
	handler, _, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = false
	auth.isParticipant = true
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testContext(userID, tenantID)

	auth.isPlatformAdmin = true
	result := submitAndVerify(t, handler, ctx)
	auth.isPlatformAdmin = false

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Scores are incorrect",
	})
	require.NoError(t, err)
}

func TestDisputeMatchResult_RBAC_NonParticipantBlocked(t *testing.T) {
	handler, _, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testAdminContext(userID, tenantID)

	result := submitAndVerify(t, handler, ctx)

	auth.isPlatformAdmin = false
	auth.isParticipant = false

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "I'm not even in this match",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestDisputeMatchResult_EmptyReason(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reason")
}

func TestDisputeMatchResult_MaxDisputeLimit(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	for i := 0; i < scores_entities.MaxDisputeCount; i++ {
		err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
			MatchResultID: result.ID,
			Reason:        fmt.Sprintf("Dispute round %d", i+1),
		})
		require.NoError(t, err, "dispute %d should succeed", i+1)

		err = handler.ConciliateMatchResult(ctx, scores_in.ConciliateMatchResultCommand{
			MatchResultID: result.ID,
			Notes:         fmt.Sprintf("Resolved round %d", i+1),
		})
		require.NoError(t, err, "conciliate %d should succeed", i+1)
	}

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "One too many",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum dispute limit")
}

func TestDisputeMatchResult_CannotDisputeSubmitted(t *testing.T) {
	handler, repo, _, _, _ := newTestHandler()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testAdminContext(userID, tenantID)

	teams := testTeamResults()
	players := testPlayerResults(teams)
	ro := shared.NewResourceOwner(tenantID, uuid.Nil, uuid.Nil, userID)
	result, err := scores_entities.NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceConsensus, userID,
		teams, players, time.Now().Add(-1*time.Hour), 45*time.Minute,
	)
	require.NoError(t, err)
	repo.results[result.ID] = result

	err = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Too early",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot dispute")
}

// ============================================================
// Tests: ConciliateMatchResult
// ============================================================

func TestConciliateMatchResult_HappyPath(t *testing.T) {
	handler, _, events, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Wrong scores",
	})
	require.NoError(t, err)

	err = handler.ConciliateMatchResult(ctx, scores_in.ConciliateMatchResultCommand{
		MatchResultID: result.ID,
		Notes:         "Review confirms scores are correct",
	})

	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusConciliated, result.Status)
	assert.Equal(t, 1, events.conciliatedEvents)
}

func TestConciliateMatchResult_WithScoreAdjustment(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	_ = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Wrong scores",
	})

	adjustedTeams := []scores_entities.TeamResult{
		{TeamID: result.TeamResults[0].TeamID, TeamName: result.TeamResults[0].TeamName, Score: 14, Players: result.TeamResults[0].Players},
		{TeamID: result.TeamResults[1].TeamID, TeamName: result.TeamResults[1].TeamName, Score: 16, Players: result.TeamResults[1].Players},
	}

	err := handler.ConciliateMatchResult(ctx, scores_in.ConciliateMatchResultCommand{
		MatchResultID:       result.ID,
		Notes:               "Corrected scores after review",
		AdjustedTeamResults: adjustedTeams,
	})

	require.NoError(t, err)
	assert.True(t, result.WasAdjusted())
	assert.Equal(t, result.TeamResults[0].TeamID, *result.WinnerTeamID)
}

func TestConciliateMatchResult_RBAC_NonAdminBlocked(t *testing.T) {
	handler, _, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)
	_ = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Wrong scores",
	})

	auth.isPlatformAdmin = false
	auth.isOrganizerResult = false

	err := handler.ConciliateMatchResult(ctx, scores_in.ConciliateMatchResultCommand{
		MatchResultID: result.ID,
		Notes:         "I shouldn't be allowed to do this",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

// ============================================================
// Tests: FinalizeMatchResult
// ============================================================

func TestFinalizeMatchResult_HappyPath_AdminForceFinalize(t *testing.T) {
	handler, _, events, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	err := handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})

	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusFinalized, result.Status)
	assert.NotNil(t, result.FinalizedAt)
	assert.Equal(t, 1, events.finalizedEvents)
}

func TestFinalizeMatchResult_DisputeWindowGuard_NonAdmin(t *testing.T) {
	handler, repo, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testAdminContext(userID, tenantID)

	result := submitAndVerify(t, handler, ctx)

	auth.isPlatformAdmin = false
	auth.isOrganizerResult = true
	tournamentID := uuid.New()
	result.TournamentID = &tournamentID
	repo.results[result.ID] = result

	err := handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispute window")
}

func TestFinalizeMatchResult_DisputeWindowPassed(t *testing.T) {
	handler, repo, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := testAdminContext(userID, tenantID)

	result := submitAndVerify(t, handler, ctx)

	past := time.Now().Add(-73 * time.Hour)
	result.VerifiedAt = &past
	repo.results[result.ID] = result

	auth.isPlatformAdmin = false
	auth.isOrganizerResult = true
	tournamentID := uuid.New()
	result.TournamentID = &tournamentID

	err := handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusFinalized, result.Status)
}

func TestFinalizeMatchResult_TriggersTournamentPrizeDistribution(t *testing.T) {
	handler, _, _, prizes, auth := newTestHandler()
	auth.isPlatformAdmin = true
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)
	tournamentID := uuid.New()
	result.TournamentID = &tournamentID

	prizeID := uuid.New()
	prizes.returnID = &prizeID

	err := handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})

	require.NoError(t, err)
	assert.True(t, prizes.tournamentCalled)
	assert.NotNil(t, result.PrizeDistributionID)
	assert.Equal(t, prizeID, *result.PrizeDistributionID)
}

func TestFinalizeMatchResult_TriggersMatchmakingPrizeDistribution(t *testing.T) {
	handler, _, _, prizes, auth := newTestHandler()
	auth.isPlatformAdmin = true
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)
	sessionID := uuid.New()
	result.MatchmakingSessionID = &sessionID

	prizeID := uuid.New()
	prizes.returnID = &prizeID

	err := handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})

	require.NoError(t, err)
	assert.True(t, prizes.matchmakingCalled)
}

func TestFinalizeMatchResult_RBAC_NonAdminBlocked(t *testing.T) {
	handler, _, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	auth.isPlatformAdmin = false
	auth.isOrganizerResult = false

	err := handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

// ============================================================
// Tests: CancelMatchResult
// ============================================================

func TestCancelMatchResult_HappyPath(t *testing.T) {
	handler, _, events, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	_ = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Cheating detected",
	})

	err := handler.CancelMatchResult(ctx, scores_in.CancelMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Cheating confirmed after review",
	})

	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusCancelled, result.Status)
	assert.Equal(t, "Cheating confirmed after review", result.CancelReason)
	assert.NotNil(t, result.CancelledBy)
	assert.NotNil(t, result.CancelledAt)
	assert.Equal(t, 1, events.cancelledEvents)
}

func TestCancelMatchResult_PreservesConciliationNotes(t *testing.T) {
	handler, _, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	_ = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Wrong scores",
	})
	_ = handler.ConciliateMatchResult(ctx, scores_in.ConciliateMatchResultCommand{
		MatchResultID: result.ID,
		Notes:         "Adjusted after review",
	})
	_ = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Still wrong",
	})
	_ = handler.CancelMatchResult(ctx, scores_in.CancelMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Cannot be resolved",
	})

	assert.Equal(t, "Adjusted after review", result.ConciliationNotes)
	assert.Equal(t, "Cannot be resolved", result.CancelReason)
}

func TestCancelMatchResult_EmptyReasonRejected(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	err := handler.CancelMatchResult(ctx, scores_in.CancelMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reason")
}

func TestCancelMatchResult_CannotCancelFinalized(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	_ = handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})

	err := handler.CancelMatchResult(ctx, scores_in.CancelMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Too late",
	})
	assert.Error(t, err)
}

// ============================================================
// Tests: Full State Machine Flow
// ============================================================

func TestFullFlow_SubmitToFinalize(t *testing.T) {
	handler, _, events, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusVerified, result.Status)

	err = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Team Beta claims different score",
	})
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusDisputed, result.Status)

	err = handler.ConciliateMatchResult(ctx, scores_in.ConciliateMatchResultCommand{
		MatchResultID: result.ID,
		Notes:         "Reviewed replay footage, scores confirmed",
	})
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusConciliated, result.Status)

	err = handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusFinalized, result.Status)

	assert.Equal(t, 1, events.submittedEvents)
	assert.Equal(t, 1, events.verifiedEvents)
	assert.Equal(t, 1, events.disputedEvents)
	assert.Equal(t, 1, events.conciliatedEvents)
	assert.Equal(t, 1, events.finalizedEvents)
}

func TestFullFlow_SubmitVerifyDisputeCancel(t *testing.T) {
	handler, _, events, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Match involved cheating",
	})
	require.NoError(t, err)

	err = handler.CancelMatchResult(ctx, scores_in.CancelMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Confirmed cheating, result voided",
	})
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusCancelled, result.Status)
	assert.Equal(t, 1, events.cancelledEvents)
}

func TestFullFlow_MultipleDisputeRounds(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	for i := 0; i < scores_entities.MaxDisputeCount; i++ {
		_ = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
			MatchResultID: result.ID,
			Reason:        fmt.Sprintf("Dispute round %d", i+1),
		})
		_ = handler.ConciliateMatchResult(ctx, scores_in.ConciliateMatchResultCommand{
			MatchResultID: result.ID,
			Notes:         fmt.Sprintf("Resolved round %d", i+1),
		})
	}
	assert.Equal(t, scores_entities.MaxDisputeCount, result.DisputeCount)

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Fourth dispute attempt",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum dispute limit")
}

// ============================================================
// Tests: Dispute Window Guard
// ============================================================

func TestDisputeWindowGuard_ConciliatedReferenceTime(t *testing.T) {
	handler, repo, _, _, auth := newTestHandler()
	auth.isPlatformAdmin = true
	ctx := testAdminContext(uuid.New(), uuid.New())

	result := submitAndVerify(t, handler, ctx)

	_ = handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: result.ID,
		Reason:        "Test dispute",
	})
	_ = handler.ConciliateMatchResult(ctx, scores_in.ConciliateMatchResultCommand{
		MatchResultID: result.ID,
		Notes:         "Resolved",
	})

	auth.isPlatformAdmin = false
	auth.isOrganizerResult = true
	tournamentID := uuid.New()
	result.TournamentID = &tournamentID
	repo.results[result.ID] = result

	err := handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispute window")

	past := time.Now().Add(-73 * time.Hour)
	result.ConciliatedAt = &past

	err = handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: result.ID,
	})
	require.NoError(t, err)
}

// ============================================================
// Tests: Edge Cases
// ============================================================

func TestNilAuthorizationBypass(t *testing.T) {
	repo := newMockRepository()
	events := &mockEventPublisher{}

	handler := &matchResultCommandHandler{
		repository:     repo,
		eventPublisher: events,
	}

	ctx := testAdminContext(uuid.New(), uuid.New())
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := handler.SubmitMatchResult(ctx, scores_in.SubmitMatchResultCommand{
		MatchID:       uuid.New(),
		GameID:        "cs2",
		MapName:       "de_dust2",
		Mode:          "competitive",
		Source:        scores_vo.ScoreSourceTournamentAdmin,
		TeamResults:   teams,
		PlayerResults: players,
		PlayedAt:      time.Now().Add(-1 * time.Hour),
		Duration:      45 * time.Minute,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDisputeMatchResult_InvalidID(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	err := handler.DisputeMatchResult(ctx, scores_in.DisputeMatchResultCommand{
		MatchResultID: uuid.Nil,
		Reason:        "test",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "match_result_id")
}

func TestFinalizeMatchResult_NotFound(t *testing.T) {
	handler, _, _, _, _ := newTestHandler()
	ctx := testAdminContext(uuid.New(), uuid.New())

	err := handler.FinalizeMatchResult(ctx, scores_in.FinalizeMatchResultCommand{
		MatchResultID: uuid.New(),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
