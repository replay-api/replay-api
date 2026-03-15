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
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_metadata "github.com/replay-api/replay-api/pkg/domain/replay/services/metadata"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mocks specific to game import tests
// ============================================================================

// --- Mock OracleCommandHandler ---

type mockGameImportOracleCommandHandler struct {
	createExternalResult *oracle_entities.OracleResult
	createExternalErr    error
	ingestErr            error
	triggerErr           error
	ingestCalls          int
	triggerCalls         int
}

func newMockOracleCommandHandler() *mockGameImportOracleCommandHandler {
	result := &oracle_entities.OracleResult{
		BaseEntity: shared.BaseEntity{ID: uuid.New()},
		Status:     oracle_vo.OracleStatusPending,
	}
	return &mockGameImportOracleCommandHandler{
		createExternalResult: result,
	}
}

func (m *mockGameImportOracleCommandHandler) IngestExternalScore(ctx context.Context, cmd oracle_in.IngestExternalScoreCommand) error {
	m.ingestCalls++
	return m.ingestErr
}

func (m *mockGameImportOracleCommandHandler) CreateExternalMatchOracle(ctx context.Context, cmd oracle_in.CreateExternalMatchOracleCommand) (*oracle_entities.OracleResult, error) {
	if m.createExternalErr != nil {
		return nil, m.createExternalErr
	}
	return m.createExternalResult, nil
}

func (m *mockGameImportOracleCommandHandler) PublishToChain(ctx context.Context, cmd oracle_in.PublishToChainCommand) error {
	return nil
}

func (m *mockGameImportOracleCommandHandler) HandleDisputeEscalation(ctx context.Context, cmd oracle_in.HandleDisputeCommand) error {
	return nil
}

func (m *mockGameImportOracleCommandHandler) TriggerIngestionFromAllProviders(ctx context.Context, cmd oracle_in.TriggerIngestionCommand) error {
	m.triggerCalls++
	return m.triggerErr
}

// --- Mock OCRStreamConfigRepository ---

type mockOCRStreamConfigRepo struct {
	saved      []*oracle_entities.OCRStreamConfig
	byExtID    map[string]*oracle_entities.OCRStreamConfig
	pending    []*oracle_entities.OCRStreamConfig
	saveErr    error
}

func newMockStreamConfigRepo() *mockOCRStreamConfigRepo {
	return &mockOCRStreamConfigRepo{
		saved:   make([]*oracle_entities.OCRStreamConfig, 0),
		byExtID: make(map[string]*oracle_entities.OCRStreamConfig),
	}
}

func (m *mockOCRStreamConfigRepo) Save(ctx context.Context, config *oracle_entities.OCRStreamConfig) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, config)
	return nil
}

func (m *mockOCRStreamConfigRepo) FindByID(ctx context.Context, id uuid.UUID) (*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockOCRStreamConfigRepo) FindByExternalMatchID(ctx context.Context, externalMatchID string) (*oracle_entities.OCRStreamConfig, error) {
	if c, ok := m.byExtID[externalMatchID]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockOCRStreamConfigRepo) FindByVideoID(ctx context.Context, videoID string) (*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockOCRStreamConfigRepo) FindByStatus(ctx context.Context, status oracle_entities.OCRStreamStatus, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockOCRStreamConfigRepo) FindPending(ctx context.Context, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	return m.pending, nil
}

func (m *mockOCRStreamConfigRepo) FindByGameID(ctx context.Context, gameID replay_common.GameIDKey, limit int) ([]*oracle_entities.OCRStreamConfig, error) {
	return nil, nil
}

func (m *mockOCRStreamConfigRepo) Update(ctx context.Context, config *oracle_entities.OCRStreamConfig) error {
	return nil
}

func (m *mockOCRStreamConfigRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// --- Mock MatchMetadataWriter ---

type mockMatchWriter struct {
	created        []replay_entity.Match
	updated        []replay_entity.Match
	confirmations  []replay_entity.SourceConfirmation
	upsertBySlug   map[string]*replay_entity.Match // slug → existing match (nil = create)
	err            error
}

func (m *mockMatchWriter) Create(ctx context.Context, match replay_entity.Match) error {
	if m.err != nil {
		return m.err
	}
	m.created = append(m.created, match)
	return nil
}

func (m *mockMatchWriter) CreateMany(ctx context.Context, matches []replay_entity.Match) error {
	return nil
}

func (m *mockMatchWriter) Update(ctx context.Context, match replay_entity.Match) error {
	m.updated = append(m.updated, match)
	return nil
}

func (m *mockMatchWriter) FindOneAndUpsertBySlug(ctx context.Context, slug string, match replay_entity.Match) (*replay_entity.Match, bool, error) {
	if m.err != nil {
		return nil, false, m.err
	}
	if existing, ok := m.upsertBySlug[slug]; ok && existing != nil {
		return existing, false, nil // existing match found
	}
	// Created new — return the candidate match with its ID
	m.created = append(m.created, match)
	return &match, true, nil
}

func (m *mockMatchWriter) AppendSourceConfirmation(ctx context.Context, matchID uuid.UUID, confirmation replay_entity.SourceConfirmation, needsReview bool, conflictDetails string) error {
	m.confirmations = append(m.confirmations, confirmation)
	return nil
}

// --- Mock MatchMetadataReader ---

type mockMatchReader struct {
	bySlug  map[string]*replay_entity.Match
	byExtID map[string]*replay_entity.Match
	byID    map[uuid.UUID]*replay_entity.Match
}

func newMockMatchReader() *mockMatchReader {
	return &mockMatchReader{
		bySlug:  make(map[string]*replay_entity.Match),
		byExtID: make(map[string]*replay_entity.Match),
		byID:    make(map[uuid.UUID]*replay_entity.Match),
	}
}

func (m *mockMatchReader) FindBySlug(ctx context.Context, slug string) (*replay_entity.Match, error) {
	if match, ok := m.bySlug[slug]; ok {
		return match, nil
	}
	return nil, nil
}

func (m *mockMatchReader) FindByExternalMatchID(ctx context.Context, externalMatchID string) (*replay_entity.Match, error) {
	if match, ok := m.byExtID[externalMatchID]; ok {
		return match, nil
	}
	return nil, nil
}

func (m *mockMatchReader) GetByID(ctx context.Context, id uuid.UUID) (*replay_entity.Match, error) {
	if match, ok := m.byID[id]; ok {
		return match, nil
	}
	return nil, nil
}

func (m *mockMatchReader) Search(ctx context.Context, s shared.Search) ([]replay_entity.Match, error) {
	return nil, nil
}

func (m *mockMatchReader) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	return nil, nil
}

// --- Mock MatchResultRepository ---

type mockMatchResultRepo struct {
	saved   []*scores_entities.MatchResult
	saveErr error
}

func (m *mockMatchResultRepo) Save(ctx context.Context, result *scores_entities.MatchResult) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, result)
	return nil
}

func (m *mockMatchResultRepo) FindByID(ctx context.Context, id uuid.UUID) (*scores_entities.MatchResult, error) {
	return nil, nil
}

func (m *mockMatchResultRepo) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*scores_entities.MatchResult, error) {
	return nil, nil
}

func (m *mockMatchResultRepo) FindByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]*scores_entities.MatchResult, error) {
	return nil, nil
}

func (m *mockMatchResultRepo) FindByMatchmakingSessionID(ctx context.Context, sessionID uuid.UUID) (*scores_entities.MatchResult, error) {
	return nil, nil
}

func (m *mockMatchResultRepo) FindByStatus(ctx context.Context, status scores_vo.ResultStatus, limit int, offset int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockMatchResultRepo) FindByPlayerID(ctx context.Context, playerID uuid.UUID, limit int, offset int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockMatchResultRepo) FindByTeamID(ctx context.Context, teamID uuid.UUID, limit int, offset int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

func (m *mockMatchResultRepo) Update(ctx context.Context, result *scores_entities.MatchResult) error {
	return nil
}

func (m *mockMatchResultRepo) Count(ctx context.Context, filter scores_out.MatchResultFilter) (int64, error) {
	return 0, nil
}

func (m *mockMatchResultRepo) Search(ctx context.Context, filter scores_out.MatchResultFilter, limit int, offset int) ([]*scores_entities.MatchResult, int64, error) {
	return nil, 0, nil
}

// ============================================================================
// Helper to create the handler with mocks
// ============================================================================

type gameImportTestHarness struct {
	handler          oracle_in.GameImportCommandHandler
	oracleCmd        *mockGameImportOracleCommandHandler
	oracleResultRepo *mockOracleResultRepository
	streamConfigRepo *mockOCRStreamConfigRepo
	matchWriter      *mockMatchWriter
	matchReader      *mockMatchReader
	matchResultRepo  *mockMatchResultRepo
	eventPub         *mockOracleEventPublisher
}

func newGameImportTestHarness() *gameImportTestHarness {
	oracleCmd := newMockOracleCommandHandler()
	oracleResultRepo := newMockRepository()
	streamConfigRepo := newMockStreamConfigRepo()
	matchWriter := &mockMatchWriter{upsertBySlug: make(map[string]*replay_entity.Match)}
	matchReader := newMockMatchReader()
	matchResultRepo := &mockMatchResultRepo{}
	eventPub := newMockEventPublisher()

	reconciliationService := replay_metadata.NewMatchReconciliationService(matchReader, matchWriter)

	handler := NewGameImportCommandHandler(
		oracleCmd,
		oracleResultRepo,
		streamConfigRepo,
		reconciliationService,
		matchResultRepo,
		eventPub,
	)

	return &gameImportTestHarness{
		handler:          handler,
		oracleCmd:        oracleCmd,
		oracleResultRepo: oracleResultRepo,
		streamConfigRepo: streamConfigRepo,
		matchWriter:      matchWriter,
		matchReader:      matchReader,
		matchResultRepo:  matchResultRepo,
		eventPub:         eventPub,
	}
}

// testSystemContext creates a context with system-level resource ownership
// for tests that exercise code paths calling shared.GetResourceOwner.
func testSystemContext() context.Context {
	ctx := context.WithValue(context.Background(), shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.UserIDKey, replay_common.TeamPROAppClientID) // system user
	return ctx
}

// ============================================================================
// Tests — ImportDiscoveredMatch
// ============================================================================

func TestImportDiscoveredMatch_Success(t *testing.T) {
	h := newGameImportTestHarness()
	teamA := uuid.New()
	teamB := uuid.New()

	cmd := oracle_in.ImportDiscoveredMatchCommand{
		ExternalMatch: oracle_out.ExternalMatch{
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
		},
		TriggerOCR:       false,
		TriggerAPIIngest: true,
	}

	err := h.handler.ImportDiscoveredMatch(testSystemContext(), cmd)
	require.NoError(t, err)

	// Should have created a match
	assert.Len(t, h.matchWriter.created, 1)
	assert.Equal(t, replay_common.CS2_GAME_ID, h.matchWriter.created[0].GameID)
	assert.Equal(t, "ps-match-001", h.matchWriter.created[0].ExternalMatchID)

	// Should have called IngestExternalScore
	assert.Equal(t, 1, h.oracleCmd.ingestCalls)

	// Should have triggered multi-provider ingestion
	assert.Equal(t, 1, h.oracleCmd.triggerCalls)
}

func TestImportDiscoveredMatch_ValidationError(t *testing.T) {
	h := newGameImportTestHarness()

	// Missing external_match_id
	cmd := oracle_in.ImportDiscoveredMatchCommand{
		ExternalMatch: oracle_out.ExternalMatch{
			GameID: replay_common.CS2_GAME_ID,
		},
	}

	err := h.handler.ImportDiscoveredMatch(testSystemContext(), cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "external_match_id")
}

func TestImportDiscoveredMatch_SkipsDuplicate(t *testing.T) {
	h := newGameImportTestHarness()

	// Pre-populate existing oracle result
	extID := "existing-match"
	existingResult := &oracle_entities.OracleResult{
		BaseEntity:      shared.BaseEntity{ID: uuid.New()},
		ExternalMatchID: &extID,
		Status:          oracle_vo.OracleStatusConsensusReached,
	}
	h.oracleResultRepo.byExtID[extID] = existingResult

	cmd := oracle_in.ImportDiscoveredMatchCommand{
		ExternalMatch: oracle_out.ExternalMatch{
			ExternalMatchID: extID,
			GameID:          replay_common.CS2_GAME_ID,
			Provider:        oracle_vo.OracleSourcePandaScore,
		},
		TriggerAPIIngest: true,
	}

	err := h.handler.ImportDiscoveredMatch(testSystemContext(), cmd)
	require.NoError(t, err)

	// Should NOT have created a match or called oracle commands (dedup)
	assert.Empty(t, h.matchWriter.created)
	assert.Equal(t, 0, h.oracleCmd.ingestCalls)
}

func TestImportDiscoveredMatch_WithVODs(t *testing.T) {
	h := newGameImportTestHarness()

	cmd := oracle_in.ImportDiscoveredMatchCommand{
		ExternalMatch: oracle_out.ExternalMatch{
			ExternalMatchID: "vod-match",
			GameID:          replay_common.CS2_GAME_ID,
			Provider:        oracle_vo.OracleSourcePandaScore,
			TeamAName:       "Astralis",
			TeamBName:       "NAVI",
			Status:          "finished",
			VODURLs:         []string{"https://youtube.com/watch?v=abc", "https://youtube.com/watch?v=def"},
		},
		TriggerOCR:       true,
		TriggerAPIIngest: false,
	}

	err := h.handler.ImportDiscoveredMatch(testSystemContext(), cmd)
	require.NoError(t, err)

	// Should have created 2 OCR stream configs (one per VOD)
	assert.Len(t, h.streamConfigRepo.saved, 2)

	// Check CS2 scoreboard region was set
	for _, cfg := range h.streamConfigRepo.saved {
		assert.NotNil(t, cfg.ScoreboardRegion, "CS2 matches should have scoreboard region")
		assert.Equal(t, 350, cfg.ScoreboardRegion.X)
		assert.Equal(t, 0, cfg.ScoreboardRegion.Y)
		assert.Equal(t, 530, cfg.ScoreboardRegion.Width)
		assert.Equal(t, 80, cfg.ScoreboardRegion.Height)
		assert.Equal(t, "Astralis", cfg.TeamAHint)
		assert.Equal(t, "NAVI", cfg.TeamBHint)
	}
}

func TestImportDiscoveredMatch_WithStreamURL(t *testing.T) {
	h := newGameImportTestHarness()

	cmd := oracle_in.ImportDiscoveredMatchCommand{
		ExternalMatch: oracle_out.ExternalMatch{
			ExternalMatchID: "stream-match",
			GameID:          "cs2",
			Provider:        oracle_vo.OracleSourcePandaScore,
			TeamAName:       "G2",
			TeamBName:       "Liquid",
			Status:          "finished",
			StreamURL:       "https://twitch.tv/blast",
		},
		TriggerOCR:       true,
		TriggerAPIIngest: false,
	}

	err := h.handler.ImportDiscoveredMatch(testSystemContext(), cmd)
	require.NoError(t, err)

	// Should have created 1 OCR stream config from stream URL
	assert.Len(t, h.streamConfigRepo.saved, 1)
	assert.Equal(t, "G2", h.streamConfigRepo.saved[0].TeamAHint)
}

func TestImportDiscoveredMatch_NoOCRWhenDisabled(t *testing.T) {
	h := newGameImportTestHarness()

	cmd := oracle_in.ImportDiscoveredMatchCommand{
		ExternalMatch: oracle_out.ExternalMatch{
			ExternalMatchID: "no-ocr-match",
			GameID:          replay_common.CS2_GAME_ID,
			Provider:        oracle_vo.OracleSourcePandaScore,
			VODURLs:         []string{"https://youtube.com/watch?v=xyz"},
		},
		TriggerOCR:       false,
		TriggerAPIIngest: true,
	}

	err := h.handler.ImportDiscoveredMatch(testSystemContext(), cmd)
	require.NoError(t, err)

	// No OCR configs should be created
	assert.Empty(t, h.streamConfigRepo.saved)
}

func TestImportDiscoveredMatch_OracleCreateError(t *testing.T) {
	h := newGameImportTestHarness()
	h.oracleCmd.createExternalErr = fmt.Errorf("oracle creation failed")

	cmd := oracle_in.ImportDiscoveredMatchCommand{
		ExternalMatch: oracle_out.ExternalMatch{
			ExternalMatchID: "fail-match",
			GameID:          replay_common.CS2_GAME_ID,
			Provider:        oracle_vo.OracleSourcePandaScore,
		},
		TriggerAPIIngest: true,
	}

	err := h.handler.ImportDiscoveredMatch(testSystemContext(), cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oracle")
}

// ============================================================================
// Tests — ImportFromYouTubeVOD
// ============================================================================

func TestImportFromYouTubeVOD_Success(t *testing.T) {
	h := newGameImportTestHarness()

	cmd := oracle_in.ImportFromYouTubeVODCommand{
		VideoURL:        "https://youtube.com/watch?v=test123",
		GameID:          replay_common.CS2_GAME_ID,
		ExternalMatchID: "yt-match-1",
		TeamAHint:       "Vitality",
		TeamBHint:       "FaZe",
	}

	err := h.handler.ImportFromYouTubeVOD(testSystemContext(), cmd)
	require.NoError(t, err)

	require.Len(t, h.streamConfigRepo.saved, 1)
	saved := h.streamConfigRepo.saved[0]
	assert.Equal(t, "Vitality", saved.TeamAHint)
	assert.Equal(t, "FaZe", saved.TeamBHint)
	// CS2 default scoreboard region
	assert.NotNil(t, saved.ScoreboardRegion)
	assert.Equal(t, 350, saved.ScoreboardRegion.X)
}

func TestImportFromYouTubeVOD_CustomScoreboardRegion(t *testing.T) {
	h := newGameImportTestHarness()

	customRegion := &oracle_entities.ScoreboardRegion{
		X: 100, Y: 50, Width: 400, Height: 60,
	}

	cmd := oracle_in.ImportFromYouTubeVODCommand{
		VideoURL:         "https://youtube.com/watch?v=custom",
		GameID:           replay_common.CS2_GAME_ID,
		ExternalMatchID:  "custom-region-match",
		ScoreboardRegion: customRegion,
	}

	err := h.handler.ImportFromYouTubeVOD(testSystemContext(), cmd)
	require.NoError(t, err)

	require.Len(t, h.streamConfigRepo.saved, 1)
	assert.Equal(t, 100, h.streamConfigRepo.saved[0].ScoreboardRegion.X)
	assert.Equal(t, 400, h.streamConfigRepo.saved[0].ScoreboardRegion.Width)
}

func TestImportFromYouTubeVOD_SkipsDuplicate(t *testing.T) {
	h := newGameImportTestHarness()

	// Pre-populate existing stream config
	existing := &oracle_entities.OCRStreamConfig{}
	h.streamConfigRepo.byExtID["dup-match"] = existing

	cmd := oracle_in.ImportFromYouTubeVODCommand{
		VideoURL:        "https://youtube.com/watch?v=dup",
		GameID:          replay_common.CS2_GAME_ID,
		ExternalMatchID: "dup-match",
	}

	err := h.handler.ImportFromYouTubeVOD(testSystemContext(), cmd)
	require.NoError(t, err)

	// Should not save a new one
	assert.Empty(t, h.streamConfigRepo.saved)
}

func TestImportFromYouTubeVOD_ValidationErrors(t *testing.T) {
	h := newGameImportTestHarness()

	tests := []struct {
		name string
		cmd  oracle_in.ImportFromYouTubeVODCommand
		want string
	}{
		{
			name: "missing video_url",
			cmd: oracle_in.ImportFromYouTubeVODCommand{
				GameID:          replay_common.CS2_GAME_ID,
				ExternalMatchID: "match-1",
			},
			want: "video_url",
		},
		{
			name: "missing game_id",
			cmd: oracle_in.ImportFromYouTubeVODCommand{
				VideoURL:        "https://youtube.com/watch?v=test",
				ExternalMatchID: "match-1",
			},
			want: "game_id",
		},
		{
			name: "missing external_match_id",
			cmd: oracle_in.ImportFromYouTubeVODCommand{
				VideoURL: "https://youtube.com/watch?v=test",
				GameID:   replay_common.CS2_GAME_ID,
			},
			want: "external_match_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.handler.ImportFromYouTubeVOD(testSystemContext(), tt.cmd)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

// ============================================================================
// Tests — BridgeOracleToMatchResult
// ============================================================================

func TestBridgeOracleToMatchResult_Success(t *testing.T) {
	h := newGameImportTestHarness()

	teamA := uuid.New()
	teamB := uuid.New()
	matchID := uuid.New()

	oracleResult := &oracle_entities.OracleResult{
		BaseEntity: shared.BaseEntity{ID: uuid.New(), CreatedAt: time.Now().UTC()},
		GameID: replay_common.CS2_GAME_ID,
		Status: oracle_vo.OracleStatusConsensusReached,
		ConsensusResult: &oracle_entities.ConsensusOutcome{
			ConfidenceLevel: 90,
			TeamScores: []oracle_entities.ConsensusTeamScore{
				{TeamID: teamA, Score: 2},
				{TeamID: teamB, Score: 1},
			},
			SeriesFormat: "bo3",
			GameOutcomes: []oracle_entities.GameConsensusOutcome{
				{MapName: "de_mirage"},
			},
		},
		MatchID:   &matchID,
	}
	h.oracleResultRepo.byID[oracleResult.GetID()] = oracleResult

	cmd := oracle_in.BridgeOracleToMatchResultCommand{
		OracleResultID: oracleResult.GetID(),
	}

	err := h.handler.BridgeOracleToMatchResult(testSystemContext(), cmd)
	require.NoError(t, err)

	require.Len(t, h.matchResultRepo.saved, 1)
	saved := h.matchResultRepo.saved[0]
	assert.Equal(t, replay_common.CS2_GAME_ID, saved.GameID)
	// High confidence should auto-verify
	assert.Equal(t, scores_vo.ResultStatusVerified, saved.Status)
}

func TestBridgeOracleToMatchResult_LowConfidenceNotAutoVerified(t *testing.T) {
	h := newGameImportTestHarness()

	teamA := uuid.New()
	teamB := uuid.New()

	oracleResult := &oracle_entities.OracleResult{
		BaseEntity: shared.BaseEntity{ID: uuid.New(), CreatedAt: time.Now().UTC()},
		GameID: replay_common.CS2_GAME_ID,
		Status: oracle_vo.OracleStatusConsensusReached,
		ConsensusResult: &oracle_entities.ConsensusOutcome{
			ConfidenceLevel: 50, // Below 80 threshold
			TeamScores: []oracle_entities.ConsensusTeamScore{
				{TeamID: teamA, Score: 2},
				{TeamID: teamB, Score: 1},
			},
			SeriesFormat: "bo3",
		},
	}
	h.oracleResultRepo.byID[oracleResult.GetID()] = oracleResult

	cmd := oracle_in.BridgeOracleToMatchResultCommand{
		OracleResultID: oracleResult.GetID(),
		MatchID:        ptrUUID(uuid.New()),
	}

	err := h.handler.BridgeOracleToMatchResult(testSystemContext(), cmd)
	require.NoError(t, err)

	require.Len(t, h.matchResultRepo.saved, 1)
	// Should NOT be auto-verified (low confidence)
	assert.NotEqual(t, scores_vo.ResultStatusVerified, h.matchResultRepo.saved[0].Status)
}

func TestBridgeOracleToMatchResult_RejectsNonConsensusState(t *testing.T) {
	h := newGameImportTestHarness()

	oracleResult := &oracle_entities.OracleResult{
		BaseEntity: shared.BaseEntity{ID: uuid.New(), CreatedAt: time.Now().UTC()},
		GameID: replay_common.CS2_GAME_ID,
		Status: oracle_vo.OracleStatusPending, // Not in consensus state
		ConsensusResult: &oracle_entities.ConsensusOutcome{
			ConfidenceLevel: 90,
		},
	}
	h.oracleResultRepo.byID[oracleResult.GetID()] = oracleResult

	cmd := oracle_in.BridgeOracleToMatchResultCommand{
		OracleResultID: oracleResult.GetID(),
	}

	err := h.handler.BridgeOracleToMatchResult(testSystemContext(), cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "consensus")
}

func TestBridgeOracleToMatchResult_RejectsNilConsensus(t *testing.T) {
	h := newGameImportTestHarness()

	oracleResult := &oracle_entities.OracleResult{
		BaseEntity:      shared.BaseEntity{ID: uuid.New(), CreatedAt: time.Now().UTC()},
		GameID:          replay_common.CS2_GAME_ID,
		Status:          oracle_vo.OracleStatusConsensusReached,
		ConsensusResult: nil, // No consensus outcome
	}
	h.oracleResultRepo.byID[oracleResult.GetID()] = oracleResult

	cmd := oracle_in.BridgeOracleToMatchResultCommand{
		OracleResultID: oracleResult.GetID(),
	}

	err := h.handler.BridgeOracleToMatchResult(testSystemContext(), cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "consensus")
}

func TestBridgeOracleToMatchResult_NotFound(t *testing.T) {
	h := newGameImportTestHarness()

	cmd := oracle_in.BridgeOracleToMatchResultCommand{
		OracleResultID: uuid.New(), // Does not exist
	}

	err := h.handler.BridgeOracleToMatchResult(testSystemContext(), cmd)
	assert.Error(t, err)
}

func TestBridgeOracleToMatchResult_ValidationError(t *testing.T) {
	h := newGameImportTestHarness()

	cmd := oracle_in.BridgeOracleToMatchResultCommand{
		OracleResultID: uuid.Nil, // Invalid
	}

	err := h.handler.BridgeOracleToMatchResult(testSystemContext(), cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oracle_result_id")
}

// ============================================================================
// Tests — Command Validation
// ============================================================================

func TestImportDiscoveredMatchCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     oracle_in.ImportDiscoveredMatchCommand
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid",
			cmd: oracle_in.ImportDiscoveredMatchCommand{
				ExternalMatch: oracle_out.ExternalMatch{
					ExternalMatchID: "match-1",
					GameID:          replay_common.CS2_GAME_ID,
				},
			},
			wantErr: false,
		},
		{
			name: "missing external_match_id",
			cmd: oracle_in.ImportDiscoveredMatchCommand{
				ExternalMatch: oracle_out.ExternalMatch{
					GameID: replay_common.CS2_GAME_ID,
				},
			},
			wantErr: true,
			errMsg:  "external_match_id",
		},
		{
			name: "missing game_id",
			cmd: oracle_in.ImportDiscoveredMatchCommand{
				ExternalMatch: oracle_out.ExternalMatch{
					ExternalMatchID: "match-1",
				},
			},
			wantErr: true,
			errMsg:  "game_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBridgeOracleToMatchResultCommand_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cmd := oracle_in.BridgeOracleToMatchResultCommand{
			OracleResultID: uuid.New(),
		}
		assert.NoError(t, cmd.Validate())
	})

	t.Run("nil uuid", func(t *testing.T) {
		cmd := oracle_in.BridgeOracleToMatchResultCommand{
			OracleResultID: uuid.Nil,
		}
		err := cmd.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "oracle_result_id")
	})
}
