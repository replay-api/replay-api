package oracle_usecases

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// GetOracleResult Tests
// ============================================================================

func TestGetOracleResult_HappyPath(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	// Pre-save a result
	ro := shared.GetResourceOwner(ctx)
	matchID := uuid.New()
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, result))

	query := oracle_in.GetOracleResultQuery{OracleResultID: result.ID}
	dto, err := handler.GetOracleResult(ctx, query)
	require.NoError(t, err)
	require.NotNil(t, dto)

	assert.Equal(t, result.ID, dto.ID)
	assert.Equal(t, replay_common.CS2_GAME_ID, dto.GameID)
	assert.Equal(t, oracle_vo.OracleStatusPending, dto.Status)
}

func TestGetOracleResult_NotFound(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	query := oracle_in.GetOracleResultQuery{OracleResultID: uuid.New()}
	dto, err := handler.GetOracleResult(ctx, query)
	require.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetOracleResult_ValidationError(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	query := oracle_in.GetOracleResultQuery{OracleResultID: uuid.Nil}
	dto, err := handler.GetOracleResult(ctx, query)
	require.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "invalid query")
}

// ============================================================================
// GetOracleResultByMatchID Tests
// ============================================================================

func TestGetOracleResultByMatchID_HappyPath(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	ro := shared.GetResourceOwner(ctx)
	matchID := uuid.New()
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, result))

	query := oracle_in.GetOracleResultByMatchIDQuery{MatchID: matchID}
	dto, err := handler.GetOracleResultByMatchID(ctx, query)
	require.NoError(t, err)
	require.NotNil(t, dto)

	assert.Equal(t, result.ID, dto.ID)
	assert.Equal(t, &matchID, dto.MatchID)
}

func TestGetOracleResultByMatchID_NotFound(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	query := oracle_in.GetOracleResultByMatchIDQuery{MatchID: uuid.New()}
	dto, err := handler.GetOracleResultByMatchID(ctx, query)
	require.Error(t, err)
	assert.Nil(t, dto)
}

func TestGetOracleResultByMatchID_ValidationError(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	query := oracle_in.GetOracleResultByMatchIDQuery{MatchID: uuid.Nil}
	_, err := handler.GetOracleResultByMatchID(ctx, query)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query")
}

// ============================================================================
// ListOracleResults Tests
// ============================================================================

func TestListOracleResults_HappyPath(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	// Save multiple results
	ro := shared.GetResourceOwner(ctx)
	for i := 0; i < 3; i++ {
		matchID := uuid.New()
		result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)
		require.NoError(t, repo.Save(ctx, result))
	}

	query := oracle_in.ListOracleResultsQuery{
		Page:     0,
		PageSize: 10,
	}

	listDTO, err := handler.ListOracleResults(ctx, query)
	require.NoError(t, err)
	require.NotNil(t, listDTO)

	assert.Equal(t, 3, len(listDTO.Results))
	assert.Equal(t, int64(3), listDTO.TotalCount)
	assert.Equal(t, 0, listDTO.Page)
	assert.Equal(t, 10, listDTO.PageSize)
}

func TestListOracleResults_WithGameFilter(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, uuid.New(), replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, result))

	gameID := replay_common.CS2_GAME_ID
	query := oracle_in.ListOracleResultsQuery{
		GameID:   &gameID,
		Page:     0,
		PageSize: 10,
	}

	listDTO, err := handler.ListOracleResults(ctx, query)
	require.NoError(t, err)
	require.NotNil(t, listDTO)
}

func TestListOracleResults_ValidationError(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	tests := []struct {
		name  string
		query oracle_in.ListOracleResultsQuery
	}{
		{
			name:  "negative page",
			query: oracle_in.ListOracleResultsQuery{Page: -1, PageSize: 10},
		},
		{
			name:  "zero page size",
			query: oracle_in.ListOracleResultsQuery{Page: 0, PageSize: 0},
		},
		{
			name:  "page size too large",
			query: oracle_in.ListOracleResultsQuery{Page: 0, PageSize: 101},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.ListOracleResults(ctx, tt.query)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid query")
		})
	}
}

// ============================================================================
// GetSubmissionsForResult Tests
// ============================================================================

func TestGetSubmissionsForResult_HappyPath(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	ro := shared.GetResourceOwner(ctx)
	matchID := uuid.New()
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)

	// Add 2 submissions
	sub1 := oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourcePandaScore,
		ProviderMatchID: "panda-1",
		TeamAID:         uuid.New(),
		TeamBID:         uuid.New(),
		TeamAScore:      2,
		TeamBScore:      1,
	}
	sub2 := oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourceFACEIT,
		ProviderMatchID: "faceit-1",
		TeamAID:         uuid.New(),
		TeamBID:         uuid.New(),
		TeamAScore:      2,
		TeamBScore:      1,
	}
	require.NoError(t, result.AddSubmission(sub1))
	require.NoError(t, result.AddSubmission(sub2))
	require.NoError(t, repo.Save(ctx, result))

	query := oracle_in.GetSubmissionsQuery{OracleResultID: result.ID}
	dtos, err := handler.GetSubmissionsForResult(ctx, query)
	require.NoError(t, err)

	assert.Len(t, dtos, 2)
	assert.Equal(t, oracle_vo.OracleSourcePandaScore, dtos[0].SourceType)
	assert.Equal(t, oracle_vo.OracleSourceFACEIT, dtos[1].SourceType)
}

func TestGetSubmissionsForResult_EmptySubmissions(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, uuid.New(), replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, result))

	query := oracle_in.GetSubmissionsQuery{OracleResultID: result.ID}
	dtos, err := handler.GetSubmissionsForResult(ctx, query)
	require.NoError(t, err)
	assert.Len(t, dtos, 0)
}

func TestGetSubmissionsForResult_NotFound(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	query := oracle_in.GetSubmissionsQuery{OracleResultID: uuid.New()}
	_, err := handler.GetSubmissionsForResult(ctx, query)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetSubmissionsForResult_ValidationError(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	query := oracle_in.GetSubmissionsQuery{OracleResultID: uuid.Nil}
	_, err := handler.GetSubmissionsForResult(ctx, query)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query")
}

// ============================================================================
// GetPublicationStatus Tests
// ============================================================================

func TestGetPublicationStatus_HappyPath(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	ro := shared.GetResourceOwner(ctx)
	matchID := uuid.New()
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)

	// Move to published state with a publication
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
	pub := oracle_entities.ChainPublication{
		ChainID:         oracle_vo.ChainIDPolygonAmoy,
		CAIP2:           oracle_vo.ChainIDPolygonAmoy.CAIP2(),
		ContractAddress: "0xcontract",
		TxHash:          "0xtxhash",
		BlockNumber:     999,
		GasUsed:         200000,
		Status:          "confirmed",
		PublishedAt:     time.Now().UTC(),
	}
	require.NoError(t, result.AddPublication(pub))
	require.NoError(t, repo.Save(ctx, result))

	query := oracle_in.GetPublicationStatusQuery{OracleResultID: result.ID}
	dtos, err := handler.GetPublicationStatus(ctx, query)
	require.NoError(t, err)

	require.Len(t, dtos, 1)
	assert.Equal(t, oracle_vo.ChainIDPolygonAmoy, dtos[0].ChainID)
	assert.Equal(t, "0xtxhash", dtos[0].TxHash)
	assert.Equal(t, uint64(999), dtos[0].BlockNumber)
	assert.Equal(t, int64(200000), dtos[0].GasUsed)
	assert.Equal(t, "confirmed", dtos[0].Status)
}

func TestGetPublicationStatus_NoPublications(t *testing.T) {
	ctx := testContext()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, uuid.New(), replay_common.CS2_GAME_ID)
	require.NoError(t, repo.Save(ctx, result))

	query := oracle_in.GetPublicationStatusQuery{OracleResultID: result.ID}
	dtos, err := handler.GetPublicationStatus(ctx, query)
	require.NoError(t, err)
	assert.Len(t, dtos, 0)
}

func TestGetPublicationStatus_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	query := oracle_in.GetPublicationStatusQuery{OracleResultID: uuid.New()}
	_, err := handler.GetPublicationStatus(ctx, query)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetPublicationStatus_ValidationError(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	handler := NewOracleQueryHandler(repo)

	query := oracle_in.GetPublicationStatusQuery{OracleResultID: uuid.Nil}
	_, err := handler.GetPublicationStatus(ctx, query)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query")
}

// ============================================================================
// DTO Mapping Tests
// ============================================================================

func TestMapOracleResultToDTO_WithConsensus(t *testing.T) {
	ctx := testContext()
	ro := shared.GetResourceOwner(ctx)
	matchID := uuid.New()
	result := oracle_entities.NewOracleResult(ro, matchID, replay_common.CS2_GAME_ID)

	winnerID := uuid.New()
	outcome := oracle_entities.ConsensusOutcome{
		WinnerTeamID:    &winnerID,
		IsDraw:          false,
		ConfidenceLevel: 95,
		AgreementRatio:  0.98,
		SourceCount:     4,
		SeriesFormat:    "bo5",
		GamesPlayed:     3,
		TeamScores: []oracle_entities.ConsensusTeamScore{
			{TeamID: uuid.New(), Score: 3},
			{TeamID: uuid.New(), Score: 1},
		},
		SourceHash: "abc123",
		ComputedAt: time.Now().UTC(),
	}
	require.NoError(t, result.SetConsensusResult(outcome))

	dto := oracle_in.MapOracleResultToDTO(result)

	require.NotNil(t, dto.Consensus)
	assert.Equal(t, &winnerID, dto.Consensus.WinnerTeamID)
	assert.Equal(t, 95, dto.Consensus.ConfidenceLevel)
	assert.Equal(t, 0.98, dto.Consensus.AgreementRatio)
	assert.Equal(t, 4, dto.Consensus.SourceCount)
	assert.Equal(t, "bo5", dto.Consensus.SeriesFormat)
	assert.Equal(t, "abc123", dto.Consensus.SourceHash)
}

func TestMapOracleResultToDTO_WithoutConsensus(t *testing.T) {
	ctx := testContext()
	ro := shared.GetResourceOwner(ctx)
	result := oracle_entities.NewOracleResult(ro, uuid.New(), replay_common.CS2_GAME_ID)

	dto := oracle_in.MapOracleResultToDTO(result)

	assert.Nil(t, dto.Consensus)
	assert.Equal(t, oracle_vo.OracleStatusPending, dto.Status)
	assert.Equal(t, 0, dto.SubmissionCount)
}
