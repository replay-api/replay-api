package oracle_usecases

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_services "github.com/replay-api/replay-api/pkg/domain/oracle/services"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// --- Mock Repository ---

type mockOracleResultRepository struct {
	saved       []*oracle_entities.OracleResult
	updated     []*oracle_entities.OracleResult
	byID        map[uuid.UUID]*oracle_entities.OracleResult
	byMatchID   map[uuid.UUID]*oracle_entities.OracleResult
	byExtID     map[string]*oracle_entities.OracleResult
	saveErr     error
	findByIDErr error
	updateErr   error
}

func newMockRepository() *mockOracleResultRepository {
	return &mockOracleResultRepository{
		saved:     make([]*oracle_entities.OracleResult, 0),
		updated:   make([]*oracle_entities.OracleResult, 0),
		byID:      make(map[uuid.UUID]*oracle_entities.OracleResult),
		byMatchID: make(map[uuid.UUID]*oracle_entities.OracleResult),
		byExtID:   make(map[string]*oracle_entities.OracleResult),
	}
}

func (m *mockOracleResultRepository) Save(ctx context.Context, result *oracle_entities.OracleResult) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, result)
	m.byID[result.ID] = result
	if result.MatchID != nil {
		m.byMatchID[*result.MatchID] = result
	}
	if result.ExternalMatchID != nil {
		m.byExtID[*result.ExternalMatchID] = result
	}
	return nil
}

func (m *mockOracleResultRepository) FindByID(ctx context.Context, id uuid.UUID) (*oracle_entities.OracleResult, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if r, ok := m.byID[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockOracleResultRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*oracle_entities.OracleResult, error) {
	if r, ok := m.byMatchID[matchID]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockOracleResultRepository) FindByExternalMatchID(ctx context.Context, externalMatchID string) (*oracle_entities.OracleResult, error) {
	if r, ok := m.byExtID[externalMatchID]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockOracleResultRepository) FindByStatus(ctx context.Context, status oracle_vo.OracleStatus, limit int, offset int) ([]*oracle_entities.OracleResult, int64, error) {
	var results []*oracle_entities.OracleResult
	for _, r := range m.byID {
		if r.Status == status {
			results = append(results, r)
		}
	}
	return results, int64(len(results)), nil
}

func (m *mockOracleResultRepository) FindPendingPublication(ctx context.Context) ([]*oracle_entities.OracleResult, error) {
	var results []*oracle_entities.OracleResult
	for _, r := range m.byID {
		if r.Status == oracle_vo.OracleStatusConsensusReached {
			results = append(results, r)
		}
	}
	return results, nil
}

func (m *mockOracleResultRepository) FindPublishedBefore(ctx context.Context, before time.Time) ([]*oracle_entities.OracleResult, error) {
	return nil, nil
}

func (m *mockOracleResultRepository) Update(ctx context.Context, result *oracle_entities.OracleResult) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated = append(m.updated, result)
	m.byID[result.ID] = result
	return nil
}

func (m *mockOracleResultRepository) Count(ctx context.Context, filter oracle_out.OracleResultFilter) (int64, error) {
	return int64(len(m.byID)), nil
}

func (m *mockOracleResultRepository) Search(ctx context.Context, filter oracle_out.OracleResultFilter, limit int, offset int) ([]*oracle_entities.OracleResult, int64, error) {
	var results []*oracle_entities.OracleResult
	for _, r := range m.byID {
		results = append(results, r)
	}
	return results, int64(len(results)), nil
}

// --- Mock Event Publisher ---

type publishedEvent struct {
	eventType string
	result    *oracle_entities.OracleResult
}

type mockOracleEventPublisher struct {
	events    []publishedEvent
	publishErr error
}

func newMockEventPublisher() *mockOracleEventPublisher {
	return &mockOracleEventPublisher{
		events: make([]publishedEvent, 0),
	}
}

func (m *mockOracleEventPublisher) PublishConsensusReached(ctx context.Context, result *oracle_entities.OracleResult) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.events = append(m.events, publishedEvent{eventType: "consensus_reached", result: result})
	return nil
}

func (m *mockOracleEventPublisher) PublishScorePublished(ctx context.Context, result *oracle_entities.OracleResult) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.events = append(m.events, publishedEvent{eventType: "score_published", result: result})
	return nil
}

func (m *mockOracleEventPublisher) PublishScoreFinalized(ctx context.Context, result *oracle_entities.OracleResult) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.events = append(m.events, publishedEvent{eventType: "score_finalized", result: result})
	return nil
}

func (m *mockOracleEventPublisher) PublishScoreDisputed(ctx context.Context, result *oracle_entities.OracleResult) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.events = append(m.events, publishedEvent{eventType: "score_disputed", result: result})
	return nil
}

func (m *mockOracleEventPublisher) PublishExternalIngested(ctx context.Context, result *oracle_entities.OracleResult, sub oracle_entities.ScoreSubmission) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.events = append(m.events, publishedEvent{eventType: "external_ingested", result: result})
	return nil
}

// --- Mock External Score Port ---

type mockExternalScorePort struct {
	providerID       oracle_vo.OracleSourceType
	supportedGames   map[replay_common.GameIDKey]bool
	confidenceWeight float64
	fetchResult      *oracle_entities.ScoreSubmission
	fetchErr         error
}

func newMockProvider(providerID oracle_vo.OracleSourceType, confidence float64) *mockExternalScorePort {
	return &mockExternalScorePort{
		providerID:       providerID,
		supportedGames:   map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		confidenceWeight: confidence,
	}
}

func (m *mockExternalScorePort) FetchMatchScore(ctx context.Context, externalMatchID string, gameID replay_common.GameIDKey) (*oracle_entities.ScoreSubmission, error) {
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	if m.fetchResult != nil {
		return m.fetchResult, nil
	}
	// Return a default submission
	teamA := uuid.New()
	teamB := uuid.New()
	winner := teamA
	return &oracle_entities.ScoreSubmission{
		SourceType:      m.providerID,
		ProviderMatchID: externalMatchID,
		WinnerTeamID:    &winner,
		TeamAID:         teamA,
		TeamBID:         teamB,
		TeamAScore:      2,
		TeamBScore:      1,
		RoundsPlayed:    3,
	}, nil
}

func (m *mockExternalScorePort) SupportsGame(gameID replay_common.GameIDKey) bool {
	return m.supportedGames[gameID]
}

func (m *mockExternalScorePort) ProviderID() oracle_vo.OracleSourceType {
	return m.providerID
}

func (m *mockExternalScorePort) ConfidenceWeight() float64 {
	return m.confidenceWeight
}

func (m *mockExternalScorePort) ListRecentMatches(ctx context.Context, gameID replay_common.GameIDKey, since time.Time, limit int) ([]oracle_out.ExternalMatch, error) {
	return nil, nil
}

// --- Mock Chain Score Gateway ---

type mockChainScoreGateway struct {
	publications    []*oracle_entities.ChainPublication
	supportedChains []oracle_vo.ChainID
	publishErr      error
	publishResult   *oracle_entities.ChainPublication
}

func newMockChainGateway() *mockChainScoreGateway {
	return &mockChainScoreGateway{
		publications:    make([]*oracle_entities.ChainPublication, 0),
		supportedChains: []oracle_vo.ChainID{oracle_vo.ChainIDPolygonAmoy},
	}
}

func (m *mockChainScoreGateway) PublishScore(ctx context.Context, chainID oracle_vo.ChainID, result *oracle_entities.OracleResult) (*oracle_entities.ChainPublication, error) {
	if m.publishErr != nil {
		return nil, m.publishErr
	}
	if m.publishResult != nil {
		return m.publishResult, nil
	}
	pub := &oracle_entities.ChainPublication{
		ChainID:         chainID,
		CAIP2:           chainID.CAIP2(),
		ContractAddress: "0x1234567890abcdef1234567890abcdef12345678",
		TxHash:          fmt.Sprintf("0xtx_%s_%d", result.ID.String()[:8], chainID),
		BlockNumber:     12345,
		GasUsed:         150000,
		Status:          "confirmed",
		PublishedAt:     time.Now().UTC(),
	}
	m.publications = append(m.publications, pub)
	return pub, nil
}

func (m *mockChainScoreGateway) GetPublishedScore(ctx context.Context, chainID oracle_vo.ChainID, matchID uuid.UUID) (*oracle_out.OnChainScore, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockChainScoreGateway) IsScoreFinalized(ctx context.Context, chainID oracle_vo.ChainID, matchID uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockChainScoreGateway) SupportedChains() []oracle_vo.ChainID {
	return m.supportedChains
}

// ============================================================================
// Test Helpers
// ============================================================================

// testContext creates a context with valid resource ownership for testing
func testContext() context.Context {
	ctx := context.Background()
	tenantID := uuid.New()
	clientID := uuid.New()
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, clientID)
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	return ctx
}

// newTestHandler creates a command handler with mock dependencies and returns all mocks
func newTestHandler(providers []oracle_out.ExternalScorePort) (oracle_in.OracleCommandHandler, *mockOracleResultRepository, *mockOracleEventPublisher, *mockChainScoreGateway) {
	repo := newMockRepository()
	eventPub := newMockEventPublisher()
	chainGw := newMockChainGateway()
	tracker := oracle_services.NewProviderReliabilityTracker()
	engine := oracle_services.NewConsensusEngine(tracker)
	policy := oracle_vo.StandardPolicy()

	handler := NewOracleCommandHandler(repo, eventPub, providers, engine, chainGw, policy)
	return handler, repo, eventPub, chainGw
}

// makeIngestCommand creates a valid IngestExternalScoreCommand
func makeIngestCommand(matchID *uuid.UUID, extMatchID *string) oracle_in.IngestExternalScoreCommand {
	teamA := uuid.New()
	teamB := uuid.New()
	winner := teamA

	return oracle_in.IngestExternalScoreCommand{
		MatchID:         matchID,
		ExternalMatchID: extMatchID,
		GameID:          replay_common.CS2_GAME_ID,
		SourceType:      oracle_vo.OracleSourcePandaScore,
		ProviderMatchID: "panda-match-001",
		WinnerTeamID:    &winner,
		TeamAID:         teamA,
		TeamBID:         teamB,
		TeamAScore:      2,
		TeamBScore:      1,
		RoundsPlayed:    3,
	}
}

// ============================================================================
// IngestExternalScore Tests
// ============================================================================

func TestIngestExternalScore_HappyPath_NewResult(t *testing.T) {
	ctx := testContext()
	handler, repo, eventPub, _ := newTestHandler(nil)

	matchID := uuid.New()
	cmd := makeIngestCommand(&matchID, nil)

	err := handler.IngestExternalScore(ctx, cmd)
	require.NoError(t, err)

	// Should have saved a new oracle result
	assert.Len(t, repo.saved, 1, "should save new oracle result")
	assert.Equal(t, oracle_vo.OracleStatusPending, repo.saved[0].Status)
	assert.Equal(t, 1, repo.saved[0].GetSubmissionCount())

	// Should have updated it
	assert.Len(t, repo.updated, 1, "should update oracle result")

	// Should have published ingestion event
	assert.Len(t, eventPub.events, 1)
	assert.Equal(t, "external_ingested", eventPub.events[0].eventType)
}

func TestIngestExternalScore_HappyPath_ExistingResult(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)

	// Pre-save an oracle result
	matchID := uuid.New()
	ro := shared.GetResourceOwner(ctx)
	existing := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, existing))

	cmd := makeIngestCommand(&matchID, nil)
	err := handler.IngestExternalScore(ctx, cmd)
	require.NoError(t, err)

	// Should NOT save a new one, only update existing
	assert.Len(t, repo.saved, 1, "should reuse existing oracle result")
	assert.Len(t, repo.updated, 1)
	assert.Equal(t, 1, repo.updated[0].GetSubmissionCount())
}

func TestIngestExternalScore_ValidationError(t *testing.T) {
	ctx := testContext()
	handler, _, _, _ := newTestHandler(nil)

	// Missing all IDs
	cmd := oracle_in.IngestExternalScoreCommand{
		GameID:          replay_common.CS2_GAME_ID,
		SourceType:      oracle_vo.OracleSourcePandaScore,
		ProviderMatchID: "match-1",
		TeamAID:         uuid.New(),
		TeamBID:         uuid.New(),
	}

	err := handler.IngestExternalScore(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command")
}

func TestIngestExternalScore_DuplicateSubmission(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)

	matchID := uuid.New()
	cmd := makeIngestCommand(&matchID, nil)

	// First ingestion should succeed
	err := handler.IngestExternalScore(ctx, cmd)
	require.NoError(t, err)

	// Same source + provider match ID → duplicate
	cmd2 := makeIngestCommand(&matchID, nil)
	cmd2.TeamAID = cmd.TeamAID
	cmd2.TeamBID = cmd.TeamBID
	err = handler.IngestExternalScore(ctx, cmd2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate submission")

	// Should still only have 1 submission
	assert.Equal(t, 1, repo.byMatchID[matchID].GetSubmissionCount())
}

func TestIngestExternalScore_ByExternalMatchID(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)

	extID := "ext-match-999"
	cmd := makeIngestCommand(nil, &extID)

	err := handler.IngestExternalScore(ctx, cmd)
	require.NoError(t, err)

	// Should be findable by external ID
	assert.Len(t, repo.saved, 1)
	assert.NotNil(t, repo.saved[0].ExternalMatchID)
	assert.Equal(t, extID, *repo.saved[0].ExternalMatchID)
}

func TestIngestExternalScore_ByOracleResultID(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)

	// Pre-save an oracle result
	matchID := uuid.New()
	ro := shared.GetResourceOwner(ctx)
	existing := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, existing))

	cmd := makeIngestCommand(nil, nil)
	cmd.OracleResultID = &existing.ID

	err := handler.IngestExternalScore(ctx, cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, repo.byID[existing.ID].GetSubmissionCount())
}

func TestIngestExternalScore_TriggersConsensus(t *testing.T) {
	ctx := testContext()
	handler, repo, eventPub, _ := newTestHandler(nil)

	matchID := uuid.New()
	teamA := uuid.New()
	teamB := uuid.New()
	winner := teamA

	// Ingest 3 submissions from different sources (min for StandardPolicy)
	sources := []oracle_vo.OracleSourceType{
		oracle_vo.OracleSourcePandaScore,
		oracle_vo.OracleSourceSteamWebAPI,
		oracle_vo.OracleSourceFACEIT,
	}

	for i, source := range sources {
		cmd := oracle_in.IngestExternalScoreCommand{
			MatchID:         &matchID,
			GameID:          replay_common.CS2_GAME_ID,
			SourceType:      source,
			ProviderMatchID: fmt.Sprintf("provider-match-%d", i),
			WinnerTeamID:    &winner,
			TeamAID:         teamA,
			TeamBID:         teamB,
			TeamAScore:      2,
			TeamBScore:      1,
			RoundsPlayed:    3,
		}
		err := handler.IngestExternalScore(ctx, cmd)
		require.NoError(t, err, "submission %d from %s should succeed", i, source)
	}

	// After 3 submissions, consensus should have been triggered
	result := repo.byMatchID[matchID]
	require.NotNil(t, result)
	assert.Equal(t, oracle_vo.OracleStatusConsensusReached, result.Status, "should reach consensus after 3 agreeing submissions")
	assert.NotNil(t, result.ConsensusResult)
	assert.Greater(t, result.ConsensusResult.AgreementRatio, 0.0)

	// Should have published both ingestion events and consensus event
	hasConsensus := false
	for _, e := range eventPub.events {
		if e.eventType == "consensus_reached" {
			hasConsensus = true
		}
	}
	assert.True(t, hasConsensus, "should publish consensus_reached event")
}

func TestIngestExternalScore_NoIdentifier(t *testing.T) {
	ctx := testContext()
	handler, _, _, _ := newTestHandler(nil)

	// All identifiers nil → validation should fail
	cmd := oracle_in.IngestExternalScoreCommand{
		GameID:          replay_common.CS2_GAME_ID,
		SourceType:      oracle_vo.OracleSourcePandaScore,
		ProviderMatchID: "match-1",
		TeamAID:         uuid.New(),
		TeamBID:         uuid.New(),
	}

	err := handler.IngestExternalScore(ctx, cmd)
	require.Error(t, err)
}

func TestIngestExternalScore_UpdateError(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)
	repo.updateErr = fmt.Errorf("db write failed")

	matchID := uuid.New()
	cmd := makeIngestCommand(&matchID, nil)

	err := handler.IngestExternalScore(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update oracle result")
}

// ============================================================================
// CreateExternalMatchOracle Tests
// ============================================================================

func TestCreateExternalMatchOracle_HappyPath(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)

	cmd := oracle_in.CreateExternalMatchOracleCommand{
		ExternalMatchID: "ext-match-abc",
		GameID:          replay_common.CS2_GAME_ID,
	}

	result, err := handler.CreateExternalMatchOracle(ctx, cmd)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, oracle_vo.OracleStatusPending, result.Status)
	assert.NotNil(t, result.ExternalMatchID)
	assert.Equal(t, "ext-match-abc", *result.ExternalMatchID)
	assert.Len(t, repo.saved, 1)
}

func TestCreateExternalMatchOracle_Idempotent(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)

	cmd := oracle_in.CreateExternalMatchOracleCommand{
		ExternalMatchID: "ext-match-abc",
		GameID:          replay_common.CS2_GAME_ID,
	}

	// First call
	result1, err := handler.CreateExternalMatchOracle(ctx, cmd)
	require.NoError(t, err)

	// Second call should return same result
	result2, err := handler.CreateExternalMatchOracle(ctx, cmd)
	require.NoError(t, err)

	assert.Equal(t, result1.ID, result2.ID, "second call should return existing result")
	assert.Len(t, repo.saved, 1, "should only save once")
}

func TestCreateExternalMatchOracle_ValidationError(t *testing.T) {
	ctx := testContext()
	handler, _, _, _ := newTestHandler(nil)

	cmd := oracle_in.CreateExternalMatchOracleCommand{
		ExternalMatchID: "", // Missing
		GameID:          replay_common.CS2_GAME_ID,
	}

	result, err := handler.CreateExternalMatchOracle(ctx, cmd)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid command")
}

func TestCreateExternalMatchOracle_SaveError(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)
	repo.saveErr = fmt.Errorf("db error")

	cmd := oracle_in.CreateExternalMatchOracleCommand{
		ExternalMatchID: "ext-match-abc",
		GameID:          replay_common.CS2_GAME_ID,
	}

	result, err := handler.CreateExternalMatchOracle(ctx, cmd)
	require.Error(t, err)
	assert.Nil(t, result)
}

// ============================================================================
// PublishToChain Tests
// ============================================================================

func TestPublishToChain_HappyPath(t *testing.T) {
	ctx := testContext()
	handler, repo, eventPub, chainGw := newTestHandler(nil)

	// Create a result in ConsensusReached state
	matchID := uuid.New()
	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)

	// Manually set consensus
	outcome := oracle_entities.ConsensusOutcome{
		WinnerTeamID:    ptrUUID(uuid.New()),
		ConfidenceLevel: 90,
		AgreementRatio:  0.95,
		SourceCount:     3,
		SeriesFormat:    "bo3",
		GamesPlayed:     2,
		TeamScores: []oracle_entities.ConsensusTeamScore{
			{TeamID: uuid.New(), Score: 2},
			{TeamID: uuid.New(), Score: 1},
		},
		ComputedAt: time.Now().UTC(),
	}
	require.NoError(t, result.SetConsensusResult(outcome))
	require.NoError(t, repo.Save(ctx, result))

	cmd := oracle_in.PublishToChainCommand{
		OracleResultID: result.ID,
	}

	err := handler.PublishToChain(ctx, cmd)
	require.NoError(t, err)

	// Should have published to the mock chain
	assert.Len(t, chainGw.publications, 1)
	assert.Equal(t, oracle_vo.ChainIDPolygonAmoy, chainGw.publications[0].ChainID)

	// Should have published event
	hasPublished := false
	for _, e := range eventPub.events {
		if e.eventType == "score_published" {
			hasPublished = true
		}
	}
	assert.True(t, hasPublished)

	// Result should be updated
	require.True(t, len(repo.updated) > 0)
}

func TestPublishToChain_NotPublishable(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)

	// Create a result still in Pending state
	matchID := uuid.New()
	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, result))

	cmd := oracle_in.PublishToChainCommand{
		OracleResultID: result.ID,
	}

	err := handler.PublishToChain(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not publishable")
}

func TestPublishToChain_NotFound(t *testing.T) {
	ctx := testContext()
	handler, _, _, _ := newTestHandler(nil)

	cmd := oracle_in.PublishToChainCommand{
		OracleResultID: uuid.New(), // Non-existent
	}

	err := handler.PublishToChain(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPublishToChain_ValidationError(t *testing.T) {
	ctx := testContext()
	handler, _, _, _ := newTestHandler(nil)

	cmd := oracle_in.PublishToChainCommand{
		OracleResultID: uuid.Nil, // Invalid
	}

	err := handler.PublishToChain(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command")
}

func TestPublishToChain_SpecificChains(t *testing.T) {
	ctx := testContext()
	handler, repo, _, chainGw := newTestHandler(nil)

	// Create a result in ConsensusReached state
	matchID := uuid.New()
	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	outcome := oracle_entities.ConsensusOutcome{
		WinnerTeamID:    ptrUUID(uuid.New()),
		ConfidenceLevel: 90,
		AgreementRatio:  0.95,
		SourceCount:     3,
		SeriesFormat:    "bo3",
		GamesPlayed:     2,
		TeamScores: []oracle_entities.ConsensusTeamScore{
			{TeamID: uuid.New(), Score: 2},
			{TeamID: uuid.New(), Score: 1},
		},
		ComputedAt: time.Now().UTC(),
	}
	require.NoError(t, result.SetConsensusResult(outcome))
	require.NoError(t, repo.Save(ctx, result))

	cmd := oracle_in.PublishToChainCommand{
		OracleResultID: result.ID,
		ChainIDs:       []oracle_vo.ChainID{oracle_vo.ChainIDPolygonAmoy, oracle_vo.ChainIDSolanaDevnet},
	}

	err := handler.PublishToChain(ctx, cmd)
	require.NoError(t, err)

	// Should publish to both chains
	assert.Len(t, chainGw.publications, 2)
}

func TestPublishToChain_ChainError(t *testing.T) {
	ctx := testContext()
	handler, repo, _, chainGw := newTestHandler(nil)
	chainGw.publishErr = fmt.Errorf("chain unavailable")

	matchID := uuid.New()
	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	outcome := oracle_entities.ConsensusOutcome{
		WinnerTeamID:    ptrUUID(uuid.New()),
		ConfidenceLevel: 90,
		AgreementRatio:  0.95,
		SourceCount:     3,
		SeriesFormat:    "bo3",
		GamesPlayed:     2,
		TeamScores: []oracle_entities.ConsensusTeamScore{
			{TeamID: uuid.New(), Score: 2},
			{TeamID: uuid.New(), Score: 1},
		},
		ComputedAt: time.Now().UTC(),
	}
	require.NoError(t, result.SetConsensusResult(outcome))
	require.NoError(t, repo.Save(ctx, result))

	cmd := oracle_in.PublishToChainCommand{
		OracleResultID: result.ID,
	}

	// Should NOT fail the overall operation — chain errors are logged but not fatal
	err := handler.PublishToChain(ctx, cmd)
	require.NoError(t, err)
}

// ============================================================================
// HandleDisputeEscalation Tests
// ============================================================================

func TestHandleDisputeEscalation_HappyPath(t *testing.T) {
	ctx := testContext()
	handler, repo, eventPub, _ := newTestHandler(nil)

	// Create a published result
	matchID := uuid.New()
	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	outcome := oracle_entities.ConsensusOutcome{
		WinnerTeamID:    ptrUUID(uuid.New()),
		ConfidenceLevel: 90,
		AgreementRatio:  0.95,
		SourceCount:     3,
		SeriesFormat:    "bo3",
		GamesPlayed:     2,
		TeamScores: []oracle_entities.ConsensusTeamScore{
			{TeamID: uuid.New(), Score: 2},
			{TeamID: uuid.New(), Score: 1},
		},
		ComputedAt: time.Now().UTC(),
	}
	require.NoError(t, result.SetConsensusResult(outcome))
	require.NoError(t, result.MarkPublishing())
	now := time.Now().UTC()
	pub := oracle_entities.ChainPublication{
		ChainID:         oracle_vo.ChainIDPolygonAmoy,
		CAIP2:           oracle_vo.ChainIDPolygonAmoy.CAIP2(),
		ContractAddress: "0x123",
		TxHash:          "0xabc",
		BlockNumber:     100,
		Status:          "confirmed",
		PublishedAt:     now,
	}
	require.NoError(t, result.AddPublication(pub))
	require.NoError(t, repo.Save(ctx, result))

	disputerID := uuid.New()
	cmd := oracle_in.HandleDisputeCommand{
		OracleResultID: result.ID,
		Reason:         "score is incorrect",
		DisputedBy:     disputerID,
	}

	err := handler.HandleDisputeEscalation(ctx, cmd)
	require.NoError(t, err)

	// Should be disputed
	updated := repo.byID[result.ID]
	assert.Equal(t, oracle_vo.OracleStatusDisputed, updated.Status)
	assert.NotNil(t, updated.DisputeReason)
	assert.Equal(t, "score is incorrect", *updated.DisputeReason)

	// Should publish event
	hasDisputed := false
	for _, e := range eventPub.events {
		if e.eventType == "score_disputed" {
			hasDisputed = true
		}
	}
	assert.True(t, hasDisputed)
}

func TestHandleDisputeEscalation_InvalidState(t *testing.T) {
	ctx := testContext()
	handler, repo, _, _ := newTestHandler(nil)

	// Create a pending result (cannot be disputed)
	matchID := uuid.New()
	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, result))

	cmd := oracle_in.HandleDisputeCommand{
		OracleResultID: result.ID,
		Reason:         "wrong score",
		DisputedBy:     uuid.New(),
	}

	err := handler.HandleDisputeEscalation(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to dispute")
}

func TestHandleDisputeEscalation_ValidationError(t *testing.T) {
	ctx := testContext()
	handler, _, _, _ := newTestHandler(nil)

	cmd := oracle_in.HandleDisputeCommand{
		OracleResultID: uuid.Nil, // Invalid
		Reason:         "wrong",
		DisputedBy:     uuid.New(),
	}

	err := handler.HandleDisputeEscalation(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command")
}

func TestHandleDisputeEscalation_NotFound(t *testing.T) {
	ctx := testContext()
	handler, _, _, _ := newTestHandler(nil)

	cmd := oracle_in.HandleDisputeCommand{
		OracleResultID: uuid.New(), // Non-existent
		Reason:         "wrong score",
		DisputedBy:     uuid.New(),
	}

	err := handler.HandleDisputeEscalation(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ============================================================================
// TriggerIngestionFromAllProviders Tests
// ============================================================================

func TestTriggerIngestion_HappyPath_ExternalMatch(t *testing.T) {
	ctx := testContext()

	// Set up providers with consistent data
	teamA := uuid.New()
	teamB := uuid.New()
	winner := teamA

	provider1 := newMockProvider(oracle_vo.OracleSourcePandaScore, 0.95)
	provider1.fetchResult = &oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourcePandaScore,
		ProviderMatchID: "ext-match-42",
		WinnerTeamID:    &winner,
		TeamAID:         teamA,
		TeamBID:         teamB,
		TeamAScore:      2,
		TeamBScore:      1,
		RoundsPlayed:    3,
	}

	provider2 := newMockProvider(oracle_vo.OracleSourceFACEIT, 0.85)
	provider2.fetchResult = &oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourceFACEIT,
		ProviderMatchID: "ext-match-42",
		WinnerTeamID:    &winner,
		TeamAID:         teamA,
		TeamBID:         teamB,
		TeamAScore:      2,
		TeamBScore:      1,
		RoundsPlayed:    3,
	}

	provider3 := newMockProvider(oracle_vo.OracleSourceSteamWebAPI, 0.90)
	provider3.fetchResult = &oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourceSteamWebAPI,
		ProviderMatchID: "ext-match-42",
		WinnerTeamID:    &winner,
		TeamAID:         teamA,
		TeamBID:         teamB,
		TeamAScore:      2,
		TeamBScore:      1,
		RoundsPlayed:    3,
	}

	providers := []oracle_out.ExternalScorePort{provider1, provider2, provider3}
	handler, repo, _, _ := newTestHandler(providers)

	extID := "ext-match-42"
	cmd := oracle_in.TriggerIngestionCommand{
		ExternalMatchID: &extID,
		GameID:          replay_common.CS2_GAME_ID,
	}

	err := handler.TriggerIngestionFromAllProviders(ctx, cmd)
	require.NoError(t, err)

	// Should have created oracle result and ingested from all 3 providers
	require.Len(t, repo.saved, 1)
	result := repo.saved[0]
	assert.Equal(t, 3, result.GetSubmissionCount())

	// With 3 agreeing sources, consensus should be reached
	assert.Equal(t, oracle_vo.OracleStatusConsensusReached, result.Status)
	assert.NotNil(t, result.ConsensusResult)
}

func TestTriggerIngestion_HappyPath_MatchID(t *testing.T) {
	ctx := testContext()

	provider := newMockProvider(oracle_vo.OracleSourcePandaScore, 0.95)
	providers := []oracle_out.ExternalScorePort{provider}
	handler, repo, _, _ := newTestHandler(providers)

	matchID := uuid.New()
	cmd := oracle_in.TriggerIngestionCommand{
		MatchID: &matchID,
		GameID:  replay_common.CS2_GAME_ID,
	}

	// This will create a new result but has no external match ID → will return early
	err := handler.TriggerIngestionFromAllProviders(ctx, cmd)
	require.NoError(t, err)

	// Result created but no ingestion (no external match ID to query providers)
	assert.Len(t, repo.saved, 1)
	assert.Equal(t, 0, repo.saved[0].GetSubmissionCount())
}

func TestTriggerIngestion_SkipsUnsupportedGame(t *testing.T) {
	ctx := testContext()

	provider := newMockProvider(oracle_vo.OracleSourcePandaScore, 0.95)
	// Only supports CS2, not Valorant
	providers := []oracle_out.ExternalScorePort{provider}
	handler, repo, _, _ := newTestHandler(providers)

	extID := "ext-match-42"
	cmd := oracle_in.TriggerIngestionCommand{
		ExternalMatchID: &extID,
		GameID:          replay_common.VLRNT_GAME_ID, // Provider doesn't support this
	}

	err := handler.TriggerIngestionFromAllProviders(ctx, cmd)
	require.NoError(t, err)

	assert.Len(t, repo.saved, 1)
	assert.Equal(t, 0, repo.saved[0].GetSubmissionCount(), "should skip unsupported game")
}

func TestTriggerIngestion_SkipsAlreadyIngested(t *testing.T) {
	ctx := testContext()

	provider := newMockProvider(oracle_vo.OracleSourcePandaScore, 0.95)
	providers := []oracle_out.ExternalScorePort{provider}
	handler, repo, _, _ := newTestHandler(providers)

	// Pre-create oracle result with a PandaScore submission
	extID := "ext-match-42"
	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewExternalOracleResult(ro, extID, replay_common.CS2_GAME_ID)
	sub := oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourcePandaScore,
		ProviderMatchID: "ext-match-42",
		TeamAID:         uuid.New(),
		TeamBID:         uuid.New(),
		TeamAScore:      2,
		TeamBScore:      1,
	}
	require.NoError(t, result.AddSubmission(sub))
	require.NoError(t, repo.Save(ctx, result))

	cmd := oracle_in.TriggerIngestionCommand{
		ExternalMatchID: &extID,
		GameID:          replay_common.CS2_GAME_ID,
	}

	err := handler.TriggerIngestionFromAllProviders(ctx, cmd)
	require.NoError(t, err)

	// Should not have added another PandaScore submission
	assert.Equal(t, 1, repo.byExtID[extID].GetSubmissionCount(), "should skip already-ingested provider")
}

func TestTriggerIngestion_ProviderFetchError(t *testing.T) {
	ctx := testContext()

	provider := newMockProvider(oracle_vo.OracleSourcePandaScore, 0.95)
	provider.fetchErr = fmt.Errorf("provider API error")
	providers := []oracle_out.ExternalScorePort{provider}
	handler, repo, _, _ := newTestHandler(providers)

	extID := "ext-match-42"
	cmd := oracle_in.TriggerIngestionCommand{
		ExternalMatchID: &extID,
		GameID:          replay_common.CS2_GAME_ID,
	}

	// Should NOT fail overall — error is logged and skipped
	err := handler.TriggerIngestionFromAllProviders(ctx, cmd)
	require.NoError(t, err)

	assert.Len(t, repo.saved, 1)
	assert.Equal(t, 0, repo.saved[0].GetSubmissionCount())
}

func TestTriggerIngestion_ValidationError(t *testing.T) {
	ctx := testContext()
	handler, _, _, _ := newTestHandler(nil)

	cmd := oracle_in.TriggerIngestionCommand{
		// Missing both match IDs
		GameID: replay_common.CS2_GAME_ID,
	}

	err := handler.TriggerIngestionFromAllProviders(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command")
}

// ============================================================================
// Integration-style Tests
// ============================================================================

func TestFullLifecycle_IngestToPublish(t *testing.T) {
	ctx := testContext()

	// Setup 3 agreeing providers
	teamA := uuid.New()
	teamB := uuid.New()
	winner := teamA

	makeProvider := func(source oracle_vo.OracleSourceType) *mockExternalScorePort {
		p := newMockProvider(source, 0.90)
		p.fetchResult = &oracle_entities.ScoreSubmission{
			SourceType:      source,
			ProviderMatchID: "lifecycle-match-1",
			WinnerTeamID:    &winner,
			TeamAID:         teamA,
			TeamBID:         teamB,
			TeamAScore:      2,
			TeamBScore:      0,
			RoundsPlayed:    2,
		}
		return p
	}

	providers := []oracle_out.ExternalScorePort{
		makeProvider(oracle_vo.OracleSourcePandaScore),
		makeProvider(oracle_vo.OracleSourceSteamWebAPI),
		makeProvider(oracle_vo.OracleSourceFACEIT),
	}

	handler, repo, eventPub, chainGw := newTestHandler(providers)

	// Step 1: Trigger ingestion → should create result, ingest all 3, reach consensus
	extID := "lifecycle-match-1"
	triggerCmd := oracle_in.TriggerIngestionCommand{
		ExternalMatchID: &extID,
		GameID:          replay_common.CS2_GAME_ID,
	}

	err := handler.TriggerIngestionFromAllProviders(ctx, triggerCmd)
	require.NoError(t, err)

	require.Len(t, repo.saved, 1)
	result := repo.saved[0]
	assert.Equal(t, oracle_vo.OracleStatusConsensusReached, result.Status, "should reach consensus")
	assert.Equal(t, 3, result.GetSubmissionCount())

	// Step 2: Publish to chain
	publishCmd := oracle_in.PublishToChainCommand{
		OracleResultID: result.ID,
	}

	err = handler.PublishToChain(ctx, publishCmd)
	require.NoError(t, err)

	assert.Len(t, chainGw.publications, 1)

	// Step 3: Verify events published in order
	eventTypes := make([]string, len(eventPub.events))
	for i, e := range eventPub.events {
		eventTypes[i] = e.eventType
	}

	// Should have: external_ingested x3, consensus_reached, score_published
	assert.Contains(t, eventTypes, "external_ingested")
	assert.Contains(t, eventTypes, "consensus_reached")
	assert.Contains(t, eventTypes, "score_published")
}

func TestFullLifecycle_IngestPublishDispute(t *testing.T) {
	ctx := testContext()

	teamA := uuid.New()
	teamB := uuid.New()
	winner := teamA

	// Manually ingest 3 submissions to reach consensus
	handler, repo, _, _ := newTestHandler(nil)

	matchID := uuid.New()
	sources := []oracle_vo.OracleSourceType{
		oracle_vo.OracleSourcePandaScore,
		oracle_vo.OracleSourceSteamWebAPI,
		oracle_vo.OracleSourceFACEIT,
	}

	for i, source := range sources {
		cmd := oracle_in.IngestExternalScoreCommand{
			MatchID:         &matchID,
			GameID:          replay_common.CS2_GAME_ID,
			SourceType:      source,
			ProviderMatchID: fmt.Sprintf("match-%d", i),
			WinnerTeamID:    &winner,
			TeamAID:         teamA,
			TeamBID:         teamB,
			TeamAScore:      2,
			TeamBScore:      0,
			RoundsPlayed:    2,
		}
		err := handler.IngestExternalScore(ctx, cmd)
		require.NoError(t, err)
	}

	result := repo.byMatchID[matchID]
	require.Equal(t, oracle_vo.OracleStatusConsensusReached, result.Status)

	// Publish
	publishCmd := oracle_in.PublishToChainCommand{
		OracleResultID: result.ID,
	}
	err := handler.PublishToChain(ctx, publishCmd)
	require.NoError(t, err)

	// Dispute
	disputeCmd := oracle_in.HandleDisputeCommand{
		OracleResultID: result.ID,
		Reason:         "incorrect score",
		DisputedBy:     uuid.New(),
	}
	err = handler.HandleDisputeEscalation(ctx, disputeCmd)
	require.NoError(t, err)

	assert.Equal(t, oracle_vo.OracleStatusDisputed, repo.byID[result.ID].Status)
}

// ============================================================================
// Helpers
// ============================================================================

func ptrUUID(u uuid.UUID) *uuid.UUID {
	return &u
}
