package oracle_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testResourceOwner() shared.ResourceOwner {
	return shared.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		UserID:   uuid.New(),
	}
}

func testSubmission(sourceType oracle_vo.OracleSourceType, teamAScore, teamBScore int) ScoreSubmission {
	teamA := uuid.New()
	teamB := uuid.New()
	var winner *uuid.UUID
	if teamAScore > teamBScore {
		winner = &teamA
	} else if teamBScore > teamAScore {
		winner = &teamB
	}
	return ScoreSubmission{
		SourceType:      sourceType,
		ProviderMatchID: uuid.New().String(),
		WinnerTeamID:    winner,
		IsDraw:          teamAScore == teamBScore,
		TeamAID:         teamA,
		TeamBID:         teamB,
		TeamAScore:      teamAScore,
		TeamBScore:      teamBScore,
		RoundsPlayed:    teamAScore + teamBScore,
		SourceHash:      "test-hash-" + uuid.New().String()[:8],
	}
}

func testConsensusOutcome() ConsensusOutcome {
	teamA := uuid.New()
	teamB := uuid.New()
	return ConsensusOutcome{
		WinnerTeamID:    &teamA,
		IsDraw:          false,
		ConfidenceLevel: 3,
		AgreementRatio:  0.95,
		SourceCount:     3,
		SeriesFormat:    "bo1",
		GamesPlayed:     1,
		TeamScores: []ConsensusTeamScore{
			{TeamID: teamA, Score: 16},
			{TeamID: teamB, Score: 12},
		},
		SourceHash: "consensus-hash-test",
		ComputedAt: time.Now().UTC(),
	}
}

// =============================================================================
// Factory Method Tests
// =============================================================================

func TestNewOracleResult(t *testing.T) {
	ro := testResourceOwner()
	matchID := uuid.New()
	result := NewOracleResult(ro, matchID, replay_common.GameIDKey("cs2"))

	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, &matchID, result.MatchID)
	assert.Nil(t, result.ExternalMatchID)
	assert.Equal(t, replay_common.GameIDKey("cs2"), result.GameID)
	assert.Equal(t, oracle_vo.OracleStatusPending, result.Status)
	assert.Equal(t, 0, result.ConfidenceLevel)
	assert.Empty(t, result.Submissions)
	assert.Empty(t, result.Publications)
}

func TestNewExternalOracleResult(t *testing.T) {
	ro := testResourceOwner()
	extID := "ext-match-12345"
	result := NewExternalOracleResult(ro, extID, replay_common.GameIDKey("vlrnt"))

	assert.NotNil(t, result)
	assert.Nil(t, result.MatchID)
	assert.NotNil(t, result.ExternalMatchID)
	assert.Equal(t, extID, *result.ExternalMatchID)
	assert.Equal(t, replay_common.GameIDKey("vlrnt"), result.GameID)
	assert.Equal(t, oracle_vo.OracleStatusPending, result.Status)
}

// =============================================================================
// AddSubmission Tests
// =============================================================================

func TestAddSubmission_Success(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	sub := testSubmission(oracle_vo.OracleSourcePandaScore, 16, 12)

	err := result.AddSubmission(sub)
	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Submissions))
	assert.NotEqual(t, uuid.Nil, result.Submissions[0].ID)
	assert.Equal(t, result.ID, result.Submissions[0].OracleResultID)
	assert.False(t, result.Submissions[0].SubmittedAt.IsZero())
}

func TestAddSubmission_MultipleProviders(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourcePandaScore, 16, 12)))
	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourceSteamWebAPI, 16, 12)))
	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourceFACEIT, 16, 12)))
	assert.Equal(t, 3, result.GetSubmissionCount())
}

func TestAddSubmission_DuplicateSource_RejectedBySameProviderMatchID(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	sub := testSubmission(oracle_vo.OracleSourcePandaScore, 16, 12)

	require.NoError(t, result.AddSubmission(sub))
	// Same source type AND provider match ID = duplicate
	err := result.AddSubmission(sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate submission")
}

func TestAddSubmission_SameSourceDifferentMatch_Accepted(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	sub1 := testSubmission(oracle_vo.OracleSourcePandaScore, 16, 12)
	sub2 := testSubmission(oracle_vo.OracleSourcePandaScore, 16, 14) // Different provider match ID

	require.NoError(t, result.AddSubmission(sub1))
	require.NoError(t, result.AddSubmission(sub2))
	assert.Equal(t, 2, result.GetSubmissionCount())
}

func TestAddSubmission_FinalizedResult_Rejected(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusFinalized

	err := result.AddSubmission(testSubmission(oracle_vo.OracleSourcePandaScore, 16, 12))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terminal state")
}

// =============================================================================
// State Machine Tests
// =============================================================================

func TestSetConsensusResult_FromPending(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	outcome := testConsensusOutcome()

	err := result.SetConsensusResult(outcome)
	require.NoError(t, err)
	assert.Equal(t, oracle_vo.OracleStatusConsensusReached, result.Status)
	assert.NotNil(t, result.ConsensusResult)
	assert.Equal(t, 3, result.ConfidenceLevel)
}

func TestSetConsensusResult_InvalidTransition(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusPublished // Can't go from published → consensus_reached

	err := result.SetConsensusResult(testConsensusOutcome())
	assert.Error(t, err)
}

func TestMarkPublishing(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusConsensusReached

	err := result.MarkPublishing()
	require.NoError(t, err)
	assert.Equal(t, oracle_vo.OracleStatusPublishing, result.Status)
}

func TestMarkPublishing_InvalidFromPending(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")

	err := result.MarkPublishing()
	assert.Error(t, err)
}

func TestAddPublication(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusPublishing
	pub := ChainPublication{
		ChainID:         oracle_vo.ChainIDPolygonAmoy,
		CAIP2:           "eip155:80002",
		ContractAddress: "0x1234",
		TxHash:          "0xabcdef",
		Status:          "confirmed",
		PublishedAt:     time.Now().UTC(),
	}

	err := result.AddPublication(pub)
	require.NoError(t, err)
	assert.Equal(t, oracle_vo.OracleStatusPublished, result.Status)
	assert.Equal(t, 1, len(result.Publications))
}

func TestAddPublication_InvalidState(t *testing.T) {
	// Status is pending — can't add publication
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	pub := ChainPublication{ChainID: oracle_vo.ChainIDPolygonAmoy, Status: "confirmed"}

	err := result.AddPublication(pub)
	assert.Error(t, err)
}

func TestFinalize(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusPublished

	err := result.Finalize()
	require.NoError(t, err)
	assert.Equal(t, oracle_vo.OracleStatusFinalized, result.Status)
	assert.NotNil(t, result.FinalizedAt)
}

func TestFinalize_InvalidFromPending(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")

	err := result.Finalize()
	assert.Error(t, err)
}

func TestDispute(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusPublished
	disputedBy := uuid.New()

	err := result.Dispute("score mismatch", disputedBy)
	require.NoError(t, err)
	assert.Equal(t, oracle_vo.OracleStatusDisputed, result.Status)
	assert.Equal(t, "score mismatch", *result.DisputeReason)
	assert.Equal(t, &disputedBy, result.DisputedBy)
	assert.NotNil(t, result.DisputedAt)
}

func TestDispute_InvalidFromPending(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")

	err := result.Dispute("reason", uuid.New())
	assert.Error(t, err)
}

func TestResetForReconsensus(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusDisputed

	err := result.ResetForReconsensus()
	require.NoError(t, err)
	assert.Equal(t, oracle_vo.OracleStatusPending, result.Status)
	assert.Nil(t, result.ConsensusResult)
	assert.Equal(t, 0, result.ConfidenceLevel)
	assert.Empty(t, result.Submissions)
}

func TestResetForReconsensus_InvalidFromPublished(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusPublished

	err := result.ResetForReconsensus()
	assert.Error(t, err)
}

func TestMarkFailed(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")

	err := result.MarkFailed()
	require.NoError(t, err)
	assert.Equal(t, oracle_vo.OracleStatusFailed, result.Status)
}

// =============================================================================
// Full Lifecycle Test
// =============================================================================

func TestFullLifecycle_PendingToFinalized(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")

	// Add submissions
	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourcePandaScore, 16, 12)))
	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourceSteamWebAPI, 16, 12)))
	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourceFACEIT, 16, 12)))
	assert.True(t, result.IsReadyForConsensus(3))

	// Reach consensus
	require.NoError(t, result.SetConsensusResult(testConsensusOutcome()))
	assert.Equal(t, oracle_vo.OracleStatusConsensusReached, result.Status)

	// Publish
	require.NoError(t, result.MarkPublishing())
	require.NoError(t, result.AddPublication(ChainPublication{
		ChainID: oracle_vo.ChainIDPolygonAmoy, TxHash: "0x123", Status: "confirmed",
		PublishedAt: time.Now().UTC(),
	}))

	// Finalize
	require.NoError(t, result.Finalize())
	assert.Equal(t, oracle_vo.OracleStatusFinalized, result.Status)
	assert.NotNil(t, result.FinalizedAt)
}

func TestFullLifecycle_DisputeAndReset(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	// Fast track to published
	result.Status = oracle_vo.OracleStatusPublished

	// Dispute
	require.NoError(t, result.Dispute("wrong score", uuid.New()))
	assert.Equal(t, oracle_vo.OracleStatusDisputed, result.Status)

	// Reset
	require.NoError(t, result.ResetForReconsensus())
	assert.Equal(t, oracle_vo.OracleStatusPending, result.Status)

	// Can add new submissions
	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourcePandaScore, 16, 14)))
	assert.Equal(t, 1, result.GetSubmissionCount())
}

// =============================================================================
// Query Method Tests
// =============================================================================

func TestHasSubmissionFromSource(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourcePandaScore, 16, 12)))

	assert.True(t, result.HasSubmissionFromSource(oracle_vo.OracleSourcePandaScore))
	assert.False(t, result.HasSubmissionFromSource(oracle_vo.OracleSourceSteamWebAPI))
}

func TestIsPublishedOnChain(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusPublishing
	require.NoError(t, result.AddPublication(ChainPublication{
		ChainID: oracle_vo.ChainIDPolygonAmoy, TxHash: "0xabc", Status: "confirmed",
		PublishedAt: time.Now().UTC(),
	}))

	assert.True(t, result.IsPublishedOnChain(oracle_vo.ChainIDPolygonAmoy))
	assert.False(t, result.IsPublishedOnChain(oracle_vo.ChainIDPolygon))
}

func TestGetPublicationForChain(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	result.Status = oracle_vo.OracleStatusPublishing
	require.NoError(t, result.AddPublication(ChainPublication{
		ChainID: oracle_vo.ChainIDPolygonAmoy, TxHash: "0xabc", Status: "confirmed",
		PublishedAt: time.Now().UTC(),
	}))

	pub := result.GetPublicationForChain(oracle_vo.ChainIDPolygonAmoy)
	assert.NotNil(t, pub)
	assert.Equal(t, "0xabc", pub.TxHash)

	pub = result.GetPublicationForChain(oracle_vo.ChainIDPolygon)
	assert.Nil(t, pub)
}

func TestIsReadyForConsensus(t *testing.T) {
	result := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
	assert.False(t, result.IsReadyForConsensus(3))

	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourcePandaScore, 16, 12)))
	assert.False(t, result.IsReadyForConsensus(3))

	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourceSteamWebAPI, 16, 12)))
	assert.False(t, result.IsReadyForConsensus(3))

	require.NoError(t, result.AddSubmission(testSubmission(oracle_vo.OracleSourceFACEIT, 16, 12)))
	assert.True(t, result.IsReadyForConsensus(3))
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		result  *OracleResult
		wantErr bool
	}{
		{
			name:    "valid with match ID",
			result:  NewOracleResult(testResourceOwner(), uuid.New(), "cs2"),
			wantErr: false,
		},
		{
			name:    "valid with external match ID",
			result:  NewExternalOracleResult(testResourceOwner(), "ext-123", "cs2"),
			wantErr: false,
		},
		{
			name: "invalid - no game ID",
			result: func() *OracleResult {
				r := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
				r.GameID = ""
				return r
			}(),
			wantErr: true,
		},
		{
			name: "invalid - no match identifiers",
			result: func() *OracleResult {
				r := NewOracleResult(testResourceOwner(), uuid.New(), "cs2")
				r.MatchID = nil
				r.ExternalMatchID = nil
				return r
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
