package oracle_services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mocks
// ============================================================================

type mockProvider struct {
	providerID     oracle_vo.OracleSourceType
	supportedGames map[replay_common.GameIDKey]bool
	matches        []oracle_out.ExternalMatch
	fetchErr       error
}

func (m *mockProvider) FetchMatchScore(ctx context.Context, externalMatchID string, gameID replay_common.GameIDKey) (*oracle_entities.ScoreSubmission, error) {
	return nil, nil
}

func (m *mockProvider) ListRecentMatches(ctx context.Context, gameID replay_common.GameIDKey, since time.Time, limit int) ([]oracle_out.ExternalMatch, error) {
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	return m.matches, nil
}

func (m *mockProvider) SupportsGame(gameID replay_common.GameIDKey) bool {
	return m.supportedGames[gameID]
}

func (m *mockProvider) ProviderID() oracle_vo.OracleSourceType {
	return m.providerID
}

func (m *mockProvider) ConfidenceWeight() float64 {
	return 0.8
}

type mockOracleResultRepo struct {
	byExtID map[string]*oracle_entities.OracleResult
	findErr error
}

func newMockOracleResultRepo() *mockOracleResultRepo {
	return &mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)}
}

func (m *mockOracleResultRepo) Save(ctx context.Context, result *oracle_entities.OracleResult) error {
	return nil
}

func (m *mockOracleResultRepo) FindByID(ctx context.Context, id uuid.UUID) (*oracle_entities.OracleResult, error) {
	return nil, nil
}

func (m *mockOracleResultRepo) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*oracle_entities.OracleResult, error) {
	return nil, nil
}

func (m *mockOracleResultRepo) FindByExternalMatchID(ctx context.Context, externalMatchID string) (*oracle_entities.OracleResult, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if r, ok := m.byExtID[externalMatchID]; ok {
		return r, nil
	}
	return nil, nil // nil, nil means not found (no error)
}

func (m *mockOracleResultRepo) FindByStatus(ctx context.Context, status oracle_vo.OracleStatus, limit int, offset int) ([]*oracle_entities.OracleResult, int64, error) {
	return nil, 0, nil
}

func (m *mockOracleResultRepo) FindPendingPublication(ctx context.Context) ([]*oracle_entities.OracleResult, error) {
	return nil, nil
}

func (m *mockOracleResultRepo) FindPublishedBefore(ctx context.Context, before time.Time) ([]*oracle_entities.OracleResult, error) {
	return nil, nil
}

func (m *mockOracleResultRepo) Update(ctx context.Context, result *oracle_entities.OracleResult) error {
	return nil
}

func (m *mockOracleResultRepo) Count(ctx context.Context, filter oracle_out.OracleResultFilter) (int64, error) {
	return 0, nil
}

func (m *mockOracleResultRepo) Search(ctx context.Context, filter oracle_out.OracleResultFilter, limit int, offset int) ([]*oracle_entities.OracleResult, int64, error) {
	return nil, 0, nil
}

type mockStreamConfigRepo struct {
	saved []*oracle_entities.OCRStreamConfig
}

func (m *mockStreamConfigRepo) Save(ctx context.Context, config *oracle_entities.OCRStreamConfig) error {
	m.saved = append(m.saved, config)
	return nil
}

func (m *mockStreamConfigRepo) FindByID(ctx context.Context, id uuid.UUID) (*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockStreamConfigRepo) FindByExternalMatchID(ctx context.Context, externalMatchID string) (*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockStreamConfigRepo) FindByVideoID(ctx context.Context, videoID string) (*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockStreamConfigRepo) FindByStatus(ctx context.Context, status oracle_entities.OCRStreamStatus, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockStreamConfigRepo) FindPending(ctx context.Context, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockStreamConfigRepo) FindByGameID(ctx context.Context, gameID replay_common.GameIDKey, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockStreamConfigRepo) Update(ctx context.Context, config *oracle_entities.OCRStreamConfig) error {
	return nil
}

func (m *mockStreamConfigRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockEventPublisher struct{}

func (m *mockEventPublisher) PublishConsensusReached(ctx context.Context, result *oracle_entities.OracleResult) error {
	return nil
}

func (m *mockEventPublisher) PublishScorePublished(ctx context.Context, result *oracle_entities.OracleResult) error {
	return nil
}

func (m *mockEventPublisher) PublishScoreFinalized(ctx context.Context, result *oracle_entities.OracleResult) error {
	return nil
}

func (m *mockEventPublisher) PublishScoreDisputed(ctx context.Context, result *oracle_entities.OracleResult) error {
	return nil
}

func (m *mockEventPublisher) PublishExternalIngested(ctx context.Context, result *oracle_entities.OracleResult, sub oracle_entities.ScoreSubmission) error {
	return nil
}

// ============================================================================
// Tests — DefaultGameDiscoveryConfig
// ============================================================================

func TestDefaultGameDiscoveryConfig(t *testing.T) {
	cfg := DefaultGameDiscoveryConfig()

	assert.Equal(t, 5*time.Minute, cfg.PollingInterval)
	assert.Equal(t, 24*time.Hour, cfg.LookbackWindow)
	assert.Equal(t, 50, cfg.MaxMatchesPerPoll)
	assert.Contains(t, cfg.SupportedGames, replay_common.CS2_GAME_ID)
}

// ============================================================================
// Tests — NewGameDiscoveryService
// ============================================================================

func TestNewGameDiscoveryService(t *testing.T) {
	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(nil, nil, nil, nil, cfg)

	assert.NotNil(t, svc)
	assert.Equal(t, cfg.PollingInterval, svc.PollingInterval)
	assert.Equal(t, cfg.LookbackWindow, svc.LookbackWindow)
	assert.Equal(t, cfg.MaxMatchesPerPoll, svc.MaxMatchesPerPoll)
}

// ============================================================================
// Tests — DiscoverRecentMatches
// ============================================================================

func TestDiscoverRecentMatches_NoProviders(t *testing.T) {
	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	assert.Empty(t, discoveries)
}

func TestDiscoverRecentMatches_FindsNewMatches(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()

	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{
				ExternalMatchID: "ps-match-001",
				GameID:          replay_common.CS2_GAME_ID,
				Provider:        oracle_vo.OracleSourcePandaScore,
				TeamAName:       "Vitality",
				TeamBName:       "FaZe",
				TeamAID:         teamA,
				TeamBID:         teamB,
				TeamAScore:      2,
				TeamBScore:      1,
				WinnerTeamID:    &teamA,
				Status:          "finished",
				TournamentName:  "PGL Major 2025",
				VODURLs:         []string{"https://youtube.com/watch?v=abc123"},
			},
		},
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	require.Len(t, discoveries, 1)
	assert.True(t, discoveries[0].IsNew)
	assert.True(t, discoveries[0].HasVOD)
	assert.Equal(t, "ps-match-001", discoveries[0].ExternalMatch.ExternalMatchID)
	assert.Equal(t, "Vitality", discoveries[0].ExternalMatch.TeamAName)
	assert.Equal(t, "FaZe", discoveries[0].ExternalMatch.TeamBName)
}

func TestDiscoverRecentMatches_DeduplicatesExisting(t *testing.T) {
	existingResult := &oracle_entities.OracleResult{
		BaseEntity: shared.BaseEntity{ID: uuid.New()},
	}

	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{
				ExternalMatchID: "existing-match",
				GameID:          replay_common.CS2_GAME_ID,
				Provider:        oracle_vo.OracleSourcePandaScore,
				Status:          "finished",
			},
		},
	}

	repo := &mockOracleResultRepo{
		byExtID: map[string]*oracle_entities.OracleResult{
			"existing-match": existingResult,
		},
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		repo,
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	require.Len(t, discoveries, 1)
	assert.False(t, discoveries[0].IsNew, "existing match should not be marked as new")
}

func TestDiscoverRecentMatches_SkipsUnfinishedMatches(t *testing.T) {
	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{
				ExternalMatchID: "running-match",
				GameID:          replay_common.CS2_GAME_ID,
				Provider:        oracle_vo.OracleSourcePandaScore,
				Status:          "running",
			},
			{
				ExternalMatchID: "not-started-match",
				GameID:          replay_common.CS2_GAME_ID,
				Provider:        oracle_vo.OracleSourcePandaScore,
				Status:          "not_started",
			},
		},
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	assert.Empty(t, discoveries, "should skip non-finished matches")
}

func TestDiscoverRecentMatches_SkipsUnsupportedGames(t *testing.T) {
	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{}, // supports nothing
		matches: []oracle_out.ExternalMatch{
			{
				ExternalMatchID: "match-1",
				GameID:          replay_common.CS2_GAME_ID,
				Status:          "finished",
			},
		},
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	assert.Empty(t, discoveries, "provider does not support CS2")
}

func TestDiscoverRecentMatches_ContinuesOnProviderError(t *testing.T) {
	failProvider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		fetchErr:       fmt.Errorf("API rate limited"),
	}

	successProvider := &mockProvider{
		providerID:     oracle_vo.OracleSourceSteamWebAPI,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{
				ExternalMatchID: "steam-match-1",
				GameID:          replay_common.CS2_GAME_ID,
				Provider:        oracle_vo.OracleSourceSteamWebAPI,
				Status:          "finished",
			},
		},
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{failProvider, successProvider},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	require.Len(t, discoveries, 1, "should still return matches from working provider")
	assert.Equal(t, "steam-match-1", discoveries[0].ExternalMatch.ExternalMatchID)
}

func TestDiscoverRecentMatches_MultipleProviders(t *testing.T) {
	provider1 := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{ExternalMatchID: "ps-1", GameID: replay_common.CS2_GAME_ID, Provider: oracle_vo.OracleSourcePandaScore, Status: "finished"},
			{ExternalMatchID: "ps-2", GameID: replay_common.CS2_GAME_ID, Provider: oracle_vo.OracleSourcePandaScore, Status: "finished"},
		},
	}

	provider2 := &mockProvider{
		providerID:     oracle_vo.OracleSourceFACEIT,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{ExternalMatchID: "faceit-1", GameID: replay_common.CS2_GAME_ID, Provider: oracle_vo.OracleSourceFACEIT, Status: "finished"},
		},
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider1, provider2},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	assert.Len(t, discoveries, 3)

	ids := make(map[string]bool)
	for _, d := range discoveries {
		ids[d.ExternalMatch.ExternalMatchID] = true
	}
	assert.True(t, ids["ps-1"])
	assert.True(t, ids["ps-2"])
	assert.True(t, ids["faceit-1"])
}

func TestDiscoverRecentMatches_HasVODAndStreamFlags(t *testing.T) {
	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{
				ExternalMatchID: "with-vod",
				GameID:          replay_common.CS2_GAME_ID,
				Status:          "finished",
				VODURLs:         []string{"https://youtube.com/watch?v=vod1"},
			},
			{
				ExternalMatchID: "with-stream",
				GameID:          replay_common.CS2_GAME_ID,
				Status:          "finished",
				StreamURL:       "https://twitch.tv/blast",
			},
			{
				ExternalMatchID: "no-media",
				GameID:          replay_common.CS2_GAME_ID,
				Status:          "finished",
			},
		},
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	require.Len(t, discoveries, 3)

	byID := make(map[string]DiscoveredMatch)
	for _, d := range discoveries {
		byID[d.ExternalMatch.ExternalMatchID] = d
	}

	assert.True(t, byID["with-vod"].HasVOD)
	assert.False(t, byID["with-vod"].HasStream)
	assert.False(t, byID["with-stream"].HasVOD)
	assert.True(t, byID["with-stream"].HasStream)
	assert.False(t, byID["no-media"].HasVOD)
	assert.False(t, byID["no-media"].HasStream)
}

func TestDiscoverRecentMatches_RepoErrorSkipsMatch(t *testing.T) {
	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{ExternalMatchID: "match-1", GameID: replay_common.CS2_GAME_ID, Status: "finished"},
			{ExternalMatchID: "match-2", GameID: replay_common.CS2_GAME_ID, Status: "finished"},
		},
	}

	repo := &mockOracleResultRepo{
		byExtID: make(map[string]*oracle_entities.OracleResult),
		findErr: fmt.Errorf("db connection lost"),
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		repo,
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	discoveries, err := svc.DiscoverRecentMatches(context.Background())
	require.NoError(t, err)
	assert.Empty(t, discoveries, "should skip matches when repo check fails")
}

// ============================================================================
// Tests — RunDiscoveryLoop
// ============================================================================

func TestRunDiscoveryLoop_CancelsOnContext(t *testing.T) {
	cfg := GameDiscoveryConfig{
		PollingInterval:   50 * time.Millisecond,
		LookbackWindow:    1 * time.Hour,
		MaxMatchesPerPoll: 10,
		SupportedGames:    []replay_common.GameIDKey{replay_common.CS2_GAME_ID},
	}

	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	callCount := 0
	err := svc.RunDiscoveryLoop(ctx, func(ctx context.Context, match DiscoveredMatch) error {
		callCount++
		return nil
	})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRunDiscoveryLoop_CallsCallbackForNewMatches(t *testing.T) {
	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{ExternalMatchID: "loop-match-1", GameID: replay_common.CS2_GAME_ID, Status: "finished"},
			{ExternalMatchID: "loop-match-2", GameID: replay_common.CS2_GAME_ID, Status: "finished"},
		},
	}

	cfg := GameDiscoveryConfig{
		PollingInterval:   1 * time.Second, // long enough that we only get the initial pass
		LookbackWindow:    1 * time.Hour,
		MaxMatchesPerPoll: 10,
		SupportedGames:    []replay_common.GameIDKey{replay_common.CS2_GAME_ID},
	}

	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	discoveredIDs := make([]string, 0)
	_ = svc.RunDiscoveryLoop(ctx, func(ctx context.Context, match DiscoveredMatch) error {
		discoveredIDs = append(discoveredIDs, match.ExternalMatch.ExternalMatchID)
		return nil
	})

	assert.Contains(t, discoveredIDs, "loop-match-1")
	assert.Contains(t, discoveredIDs, "loop-match-2")
}

func TestRunDiscoveryLoop_ContinuesOnCallbackError(t *testing.T) {
	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{ExternalMatchID: "err-match", GameID: replay_common.CS2_GAME_ID, Status: "finished"},
			{ExternalMatchID: "ok-match", GameID: replay_common.CS2_GAME_ID, Status: "finished"},
		},
	}

	cfg := GameDiscoveryConfig{
		PollingInterval:   1 * time.Second,
		LookbackWindow:    1 * time.Hour,
		MaxMatchesPerPoll: 10,
		SupportedGames:    []replay_common.GameIDKey{replay_common.CS2_GAME_ID},
	}

	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		&mockOracleResultRepo{byExtID: make(map[string]*oracle_entities.OracleResult)},
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	processed := make([]string, 0)
	_ = svc.RunDiscoveryLoop(ctx, func(ctx context.Context, match DiscoveredMatch) error {
		if match.ExternalMatch.ExternalMatchID == "err-match" {
			processed = append(processed, match.ExternalMatch.ExternalMatchID)
			return fmt.Errorf("import failed")
		}
		processed = append(processed, match.ExternalMatch.ExternalMatchID)
		return nil
	})

	// Both matches should have been attempted despite the first one failing
	assert.Contains(t, processed, "err-match")
	assert.Contains(t, processed, "ok-match")
}

// ============================================================================
// Tests — discoverAndProcess
// ============================================================================

func TestDiscoverAndProcess_OnlyProcessesNewMatches(t *testing.T) {
	existingResult := &oracle_entities.OracleResult{BaseEntity: shared.BaseEntity{ID: uuid.New()}}

	provider := &mockProvider{
		providerID:     oracle_vo.OracleSourcePandaScore,
		supportedGames: map[replay_common.GameIDKey]bool{replay_common.CS2_GAME_ID: true},
		matches: []oracle_out.ExternalMatch{
			{ExternalMatchID: "new-match", GameID: replay_common.CS2_GAME_ID, Status: "finished"},
			{ExternalMatchID: "old-match", GameID: replay_common.CS2_GAME_ID, Status: "finished"},
		},
	}

	repo := &mockOracleResultRepo{
		byExtID: map[string]*oracle_entities.OracleResult{
			"old-match": existingResult,
		},
	}

	cfg := DefaultGameDiscoveryConfig()
	svc := NewGameDiscoveryService(
		[]oracle_out.ExternalScorePort{provider},
		repo,
		&mockStreamConfigRepo{},
		&mockEventPublisher{},
		cfg,
	)

	processedIDs := make([]string, 0)
	err := svc.discoverAndProcess(context.Background(), func(ctx context.Context, match DiscoveredMatch) error {
		processedIDs = append(processedIDs, match.ExternalMatch.ExternalMatchID)
		return nil
	})

	require.NoError(t, err)
	assert.Contains(t, processedIDs, "new-match")
	assert.NotContains(t, processedIDs, "old-match", "existing matches should not be processed")
}
