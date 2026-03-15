package oracle_vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// OracleStatus Tests
// =============================================================================

func TestOracleStatus_IsValid(t *testing.T) {
	validStatuses := []OracleStatus{
		OracleStatusPending, OracleStatusConsensusReached, OracleStatusPublishing,
		OracleStatusPublished, OracleStatusFinalized, OracleStatusDisputed, OracleStatusFailed,
	}

	for _, s := range validStatuses {
		assert.True(t, s.IsValid(), "expected %s to be valid", s)
	}
	assert.False(t, OracleStatus("invalid").IsValid())
	assert.False(t, OracleStatus("").IsValid())
}

func TestOracleStatus_AllowedTransitions(t *testing.T) {
	tests := []struct {
		from    OracleStatus
		to      OracleStatus
		allowed bool
	}{
		// From pending
		{OracleStatusPending, OracleStatusConsensusReached, true},
		{OracleStatusPending, OracleStatusFailed, true},
		{OracleStatusPending, OracleStatusPublished, false},
		// From consensus_reached
		{OracleStatusConsensusReached, OracleStatusPublishing, true},
		{OracleStatusConsensusReached, OracleStatusFailed, true},
		{OracleStatusConsensusReached, OracleStatusPending, false},
		// From publishing
		{OracleStatusPublishing, OracleStatusPublished, true},
		{OracleStatusPublishing, OracleStatusFailed, true},
		{OracleStatusPublishing, OracleStatusPending, false},
		// From published
		{OracleStatusPublished, OracleStatusFinalized, true},
		{OracleStatusPublished, OracleStatusDisputed, true},
		{OracleStatusPublished, OracleStatusPending, false},
		// From finalized (terminal)
		{OracleStatusFinalized, OracleStatusPending, false},
		{OracleStatusFinalized, OracleStatusDisputed, false},
		// From disputed
		{OracleStatusDisputed, OracleStatusPending, true},
		{OracleStatusDisputed, OracleStatusFailed, true},
		{OracleStatusDisputed, OracleStatusPublished, false},
		// From failed
		{OracleStatusFailed, OracleStatusPending, true},
		{OracleStatusFailed, OracleStatusPublished, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			assert.Equal(t, tt.allowed, tt.from.CanTransitionTo(tt.to))
		})
	}
}

func TestOracleStatus_IsTerminal(t *testing.T) {
	assert.True(t, OracleStatusFinalized.IsTerminal())

	nonTerminal := []OracleStatus{
		OracleStatusPending, OracleStatusConsensusReached, OracleStatusPublishing,
		OracleStatusPublished, OracleStatusDisputed, OracleStatusFailed,
	}
	for _, s := range nonTerminal {
		assert.False(t, s.IsTerminal(), "%s should not be terminal", s)
	}
}

func TestOracleStatus_IsPublishable(t *testing.T) {
	assert.True(t, OracleStatusConsensusReached.IsPublishable())
	assert.False(t, OracleStatusPending.IsPublishable())
	assert.False(t, OracleStatusPublished.IsPublishable())
}

func TestOracleStatus_IsDisputable(t *testing.T) {
	assert.True(t, OracleStatusPublished.IsDisputable())
	assert.False(t, OracleStatusPending.IsDisputable())
	assert.False(t, OracleStatusFinalized.IsDisputable())
}

func TestOracleStatus_ValidateTransition(t *testing.T) {
	assert.NoError(t, OracleStatusPending.ValidateTransition(OracleStatusConsensusReached))
	assert.Error(t, OracleStatusPending.ValidateTransition(OracleStatusPublished))
	assert.Error(t, OracleStatusPending.ValidateTransition(OracleStatus("invalid")))
}

// =============================================================================
// OracleSourceType Tests
// =============================================================================

func TestOracleSourceType_Validate(t *testing.T) {
	validSources := []OracleSourceType{
		OracleSourcePandaScore, OracleSourceSteamWebAPI, OracleSourceFACEIT,
		OracleSourceSportsDataIO, OracleSourceGRID, OracleSourceAbios,
		OracleSourceOCRStream, OracleSourceOCRUpload,
	}

	for _, s := range validSources {
		assert.NoError(t, s.Validate(), "expected %s to be valid", s)
	}
	assert.Error(t, OracleSourceType("invalid").Validate())
}

func TestSourceConfidenceWeights(t *testing.T) {
	// Steam Web API should have highest weight
	assert.Equal(t, 0.95, SourceConfidenceWeights[OracleSourceSteamWebAPI])
	// OCR upload should have lowest
	assert.Equal(t, 0.50, SourceConfidenceWeights[OracleSourceOCRUpload])
	// PandaScore mid-range
	assert.Equal(t, 0.85, SourceConfidenceWeights[OracleSourcePandaScore])

	// All source types should have a weight
	validSources := []OracleSourceType{
		OracleSourcePandaScore, OracleSourceSteamWebAPI, OracleSourceFACEIT,
		OracleSourceSportsDataIO, OracleSourceGRID, OracleSourceAbios,
		OracleSourceOCRStream, OracleSourceOCRUpload,
	}
	for _, s := range validSources {
		_, exists := SourceConfidenceWeights[s]
		assert.True(t, exists, "missing weight for %s", s)
	}
}

func TestOracleSourceType_Methods(t *testing.T) {
	// ConfidenceWeight
	assert.Equal(t, 0.95, OracleSourceSteamWebAPI.ConfidenceWeight())
	assert.Equal(t, 0.0, OracleSourceType("invalid").ConfidenceWeight())

	// IsAutomated
	assert.True(t, OracleSourcePandaScore.IsAutomated())
	assert.True(t, OracleSourceOCRStream.IsAutomated())
	assert.False(t, OracleSourceOCRUpload.IsAutomated())

	// IsOCR
	assert.True(t, OracleSourceOCRStream.IsOCR())
	assert.True(t, OracleSourceOCRUpload.IsOCR())
	assert.False(t, OracleSourcePandaScore.IsOCR())

	// IsValid
	assert.True(t, OracleSourceSteamWebAPI.IsValid())
	assert.False(t, OracleSourceType("invalid").IsValid())

	// String
	assert.Equal(t, "pandascore", OracleSourcePandaScore.String())
}

// =============================================================================
// ChainID Tests
// =============================================================================

func TestChainID_CAIP2(t *testing.T) {
	assert.Equal(t, "eip155:137", ChainIDPolygon.CAIP2())
	assert.Equal(t, "eip155:80002", ChainIDPolygonAmoy.CAIP2())
	assert.Equal(t, "solana:mainnet", ChainIDSolanaMainnet.CAIP2())
	assert.Equal(t, "solana:devnet", ChainIDSolanaDevnet.CAIP2())
	assert.Equal(t, "unknown", ChainID(999).CAIP2())
}

func TestChainID_IsEVM(t *testing.T) {
	assert.True(t, ChainIDPolygon.IsEVM())
	assert.True(t, ChainIDPolygonAmoy.IsEVM())
	assert.False(t, ChainIDSolanaMainnet.IsEVM())
	assert.False(t, ChainIDSolanaDevnet.IsEVM())
}

func TestChainID_IsSolana(t *testing.T) {
	assert.True(t, ChainIDSolanaMainnet.IsSolana())
	assert.True(t, ChainIDSolanaDevnet.IsSolana())
	assert.False(t, ChainIDPolygon.IsSolana())
}

func TestChainID_IsMainnet(t *testing.T) {
	assert.True(t, ChainIDPolygon.IsMainnet())
	assert.True(t, ChainIDSolanaMainnet.IsMainnet())
	assert.False(t, ChainIDPolygonAmoy.IsMainnet())
	assert.False(t, ChainIDSolanaDevnet.IsMainnet())
}

// =============================================================================
// ConsensusPolicy Tests
// =============================================================================

func TestStrictPolicy(t *testing.T) {
	p := StrictPolicy()
	assert.Equal(t, 3, p.MinSources)
	assert.Equal(t, 0.90, p.MinConfidence)
	assert.Equal(t, 0.60, p.WinnerWeight)
	assert.Equal(t, 0.30, p.SeriesScoreWeight)
	assert.Equal(t, 0.10, p.GameScoreWeight)
}

func TestStandardPolicy(t *testing.T) {
	p := StandardPolicy()
	assert.Equal(t, 3, p.MinSources)
	assert.Equal(t, 0.75, p.MinConfidence)
	assert.Equal(t, 0.60, p.WinnerWeight)
}

func TestRelaxedPolicy(t *testing.T) {
	p := RelaxedPolicy()
	assert.Equal(t, 2, p.MinSources)
	assert.Equal(t, 0.60, p.MinConfidence)
}

func TestOCROnlyPolicy(t *testing.T) {
	p := OCROnlyPolicy()
	assert.Equal(t, 1, p.MinSources)
	assert.Equal(t, 0.50, p.MinConfidence)
	assert.Equal(t, 0.60, p.WinnerWeight)
	assert.Equal(t, 0.30, p.SeriesScoreWeight)
	assert.Equal(t, 0.10, p.GameScoreWeight)
}

func TestConsensusPolicy_WeightsSumToOne(t *testing.T) {
	policies := []ConsensusPolicy{StrictPolicy(), StandardPolicy(), RelaxedPolicy(), OCROnlyPolicy()}
	for _, p := range policies {
		sum := p.WinnerWeight + p.SeriesScoreWeight + p.GameScoreWeight
		assert.InDelta(t, 1.0, sum, 0.001, "policy weights should sum to 1.0")
	}
}
