package oracle_services

import (
	"testing"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSubmission(source oracle_vo.OracleSourceType, teamA, teamB uuid.UUID, scoreA, scoreB int) oracle_entities.ScoreSubmission {
	var winner *uuid.UUID
	if scoreA > scoreB {
		winner = &teamA
	} else if scoreB > scoreA {
		winner = &teamB
	}
	return oracle_entities.ScoreSubmission{
		ID:              uuid.New(),
		SourceType:      source,
		ProviderMatchID: uuid.New().String(),
		WinnerTeamID:    winner,
		IsDraw:          scoreA == scoreB,
		TeamAID:         teamA,
		TeamBID:         teamB,
		TeamAScore:      scoreA,
		TeamBScore:      scoreB,
		RoundsPlayed:    scoreA + scoreB,
		SourceHash:      "test-hash",
	}
}

func newTestSubmissionWithGames(source oracle_vo.OracleSourceType, teamA, teamB uuid.UUID, scoreA, scoreB int, games []oracle_entities.SubmissionGameDetail) oracle_entities.ScoreSubmission {
	sub := newTestSubmission(source, teamA, teamB, scoreA, scoreB)
	sub.GameDetails = games
	return sub
}

func TestEvaluateConsensus_AllAgree(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	engine := NewConsensusEngine(nil)
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceFACEIT, teamA, teamB, 16, 12),
	}
	outcome, err := engine.EvaluateConsensus(submissions, oracle_vo.StandardPolicy())
	require.NoError(t, err)
	require.NotNil(t, outcome)
	assert.Equal(t, &teamA, outcome.WinnerTeamID)
	assert.False(t, outcome.IsDraw)
	assert.Equal(t, 3, outcome.SourceCount)
	assert.Equal(t, 3, outcome.ConfidenceLevel)
	assert.InDelta(t, 1.0, outcome.AgreementRatio, 0.01)
	require.Len(t, outcome.TeamScores, 2)
	assert.Equal(t, 16, outcome.TeamScores[0].Score)
	assert.Equal(t, 12, outcome.TeamScores[1].Score)
	assert.Len(t, outcome.CrossValidation, 3)
	for _, cv := range outcome.CrossValidation {
		assert.True(t, cv.WinnerAgree)
		assert.True(t, cv.ScoreAgree)
	}
	assert.Empty(t, outcome.DisagreementNotes)
	assert.NotEmpty(t, outcome.SourceHash)
}

func TestEvaluateConsensus_AllAgreeDraw(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	engine := NewConsensusEngine(nil)
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 15, 15),
		newTestSubmission(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 15, 15),
		newTestSubmission(oracle_vo.OracleSourceFACEIT, teamA, teamB, 15, 15),
	}
	outcome, err := engine.EvaluateConsensus(submissions, oracle_vo.StandardPolicy())
	require.NoError(t, err)
	assert.Nil(t, outcome.WinnerTeamID)
	assert.True(t, outcome.IsDraw)
}

func TestEvaluateConsensus_WinnerDisagreement(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	engine := NewConsensusEngine(nil)
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceFACEIT, teamA, teamB, 12, 16),
	}
	outcome, err := engine.EvaluateConsensus(submissions, oracle_vo.StandardPolicy())
	if err == nil {
		assert.NotNil(t, outcome)
		assert.Equal(t, &teamA, outcome.WinnerTeamID)
		assert.Less(t, outcome.AgreementRatio, 1.0)
	}
}

func TestEvaluateConsensus_ScoreDisagreement_WinnerAgrees(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	engine := NewConsensusEngine(nil)
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceFACEIT, teamA, teamB, 16, 14),
	}
	outcome, err := engine.EvaluateConsensus(submissions, oracle_vo.StandardPolicy())
	require.NoError(t, err)
	assert.Equal(t, &teamA, outcome.WinnerTeamID)
	assert.GreaterOrEqual(t, outcome.AgreementRatio, 0.75)
	hasDisagreement := false
	for _, cv := range outcome.CrossValidation {
		if !cv.ScoreAgree {
			hasDisagreement = true
			break
		}
	}
	assert.True(t, hasDisagreement)
}

func TestEvaluateConsensus_InsufficientSources(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	engine := NewConsensusEngine(nil)
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
	}
	outcome, err := engine.EvaluateConsensus(submissions, oracle_vo.StandardPolicy())
	assert.Error(t, err)
	assert.Nil(t, outcome)
	assert.Contains(t, err.Error(), "insufficient sources")
}

func TestEvaluateConsensus_RelaxedPolicy_TwoSources(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	engine := NewConsensusEngine(nil)
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 16, 12),
	}
	outcome, err := engine.EvaluateConsensus(submissions, oracle_vo.RelaxedPolicy())
	require.NoError(t, err)
	assert.Equal(t, 2, outcome.SourceCount)
	assert.Equal(t, &teamA, outcome.WinnerTeamID)
}

func TestEvaluateConsensus_WithGameDetails(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	engine := NewConsensusEngine(nil)
	games := []oracle_entities.SubmissionGameDetail{
		{Position: 1, MapName: "Inferno", TeamAScore: 16, TeamBScore: 10, TeamAWon: true},
		{Position: 2, MapName: "Mirage", TeamAScore: 12, TeamBScore: 16, TeamAWon: false},
		{Position: 3, MapName: "Nuke", TeamAScore: 16, TeamBScore: 14, TeamAWon: true},
	}
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmissionWithGames(oracle_vo.OracleSourcePandaScore, teamA, teamB, 2, 1, games),
		newTestSubmissionWithGames(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 2, 1, games),
		newTestSubmissionWithGames(oracle_vo.OracleSourceFACEIT, teamA, teamB, 2, 1, games),
	}
	outcome, err := engine.EvaluateConsensus(submissions, oracle_vo.StandardPolicy())
	require.NoError(t, err)
	assert.Equal(t, "bo3", outcome.SeriesFormat)
	assert.Equal(t, 3, outcome.GamesPlayed)
	require.Len(t, outcome.GameOutcomes, 3)
	assert.Equal(t, "Inferno", outcome.GameOutcomes[0].MapName)
	assert.Equal(t, "Mirage", outcome.GameOutcomes[1].MapName)
	assert.Equal(t, "Nuke", outcome.GameOutcomes[2].MapName)
}

func TestEvaluateConsensus_WithReliabilityTracker(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	tracker := NewProviderReliabilityTracker()
	for i := 0; i < 10; i++ {
		tracker.RecordInaccurate(oracle_vo.OracleSourceFACEIT)
	}
	for i := 0; i < 3; i++ {
		tracker.RecordAccurate(oracle_vo.OracleSourceFACEIT, 100)
	}
	engine := NewConsensusEngine(tracker)
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceFACEIT, teamA, teamB, 12, 16),
	}
	outcome, err := engine.EvaluateConsensus(submissions, oracle_vo.StandardPolicy())
	require.NoError(t, err)
	assert.Equal(t, &teamA, outcome.WinnerTeamID)
}

func TestComputeConfidenceLevel(t *testing.T) {
	tests := []struct {
		name     string
		ratio    float64
		expected int
	}{
		{"high >= 0.95", 0.95, 3},
		{"high = 1.0", 1.0, 3},
		{"medium >= 0.80", 0.80, 2},
		{"medium = 0.89", 0.89, 2},
		{"low >= 0.60", 0.60, 1},
		{"low = 0.79", 0.79, 1},
		{"none < 0.60", 0.59, 0},
		{"none = 0.0", 0.0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, computeConfidenceLevel(tt.ratio))
		})
	}
}

func TestInferSeriesFormat(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	tests := []struct {
		name           string
		gameCount      int
		expectedFormat string
	}{
		{"bo1", 1, "bo1"},
		{"bo3", 2, "bo3"},
		{"bo3 full", 3, "bo3"},
		{"bo5", 4, "bo5"},
		{"bo5 full", 5, "bo5"},
		{"no games", 0, "bo1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			games := make([]oracle_entities.SubmissionGameDetail, tt.gameCount)
			for i := range games {
				games[i] = oracle_entities.SubmissionGameDetail{Position: i + 1, MapName: "Map"}
			}
			submissions := []oracle_entities.ScoreSubmission{
				newTestSubmissionWithGames(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12, games),
			}
			format, count := inferSeriesFormat(submissions)
			assert.Equal(t, tt.expectedFormat, format)
			assert.Equal(t, tt.gameCount, count)
		})
	}
}

func TestDetectAnomalies_ScoreOutOfRange(t *testing.T) {
	engine := NewConsensusEngine(nil)
	teamA := uuid.New()
	teamB := uuid.New()
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 30, 12),
	}
	anomalies := engine.detectAnomalies(submissions)
	found := false
	for _, a := range anomalies {
		if a.Type == AnomalyScoreOutOfRange {
			found = true
		}
	}
	assert.True(t, found, "expected score out of range anomaly")
}

func TestDetectAnomalies_ImpossibleCombo(t *testing.T) {
	engine := NewConsensusEngine(nil)
	teamA := uuid.New()
	teamB := uuid.New()
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 16),
	}
	anomalies := engine.detectAnomalies(submissions)
	found := false
	for _, a := range anomalies {
		if a.Type == AnomalyImpossibleCombo {
			found = true
		}
	}
	assert.True(t, found, "expected impossible combo anomaly")
}

func TestDetectAnomalies_WinnerScoreMismatch(t *testing.T) {
	engine := NewConsensusEngine(nil)
	teamA := uuid.New()
	teamB := uuid.New()
	sub := newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 10, 16)
	sub.WinnerTeamID = &teamA
	sub.IsDraw = false
	anomalies := engine.detectAnomalies([]oracle_entities.ScoreSubmission{sub})
	found := false
	for _, a := range anomalies {
		if a.Type == AnomalyWinnerScoreMismatch {
			found = true
		}
	}
	assert.True(t, found, "expected winner score mismatch anomaly")
}

func TestDetectAnomalies_Outlier(t *testing.T) {
	engine := NewConsensusEngine(nil)
	teamA := uuid.New()
	teamB := uuid.New()
	// Need 5+ sources for outlier detection to trigger (outlier inflates std with small N)
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceFACEIT, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceSportsDataIO, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceGRID, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceAbios, teamA, teamB, 2, 200), // Extreme outlier
	}
	anomalies := engine.detectAnomalies(submissions)
	found := false
	for _, a := range anomalies {
		if a.Type == AnomalyOutlierScore {
			found = true
		}
	}
	assert.True(t, found, "expected outlier score anomaly")
}

func TestDetectAnomalies_NoAnomalies(t *testing.T) {
	engine := NewConsensusEngine(nil)
	teamA := uuid.New()
	teamB := uuid.New()
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceSteamWebAPI, teamA, teamB, 16, 12),
		newTestSubmission(oracle_vo.OracleSourceFACEIT, teamA, teamB, 16, 12),
	}
	anomalies := engine.detectAnomalies(submissions)
	assert.Empty(t, anomalies)
}

func TestReliabilityTracker_RecordAccurate(t *testing.T) {
	tracker := NewProviderReliabilityTracker()
	tracker.RecordAccurate(oracle_vo.OracleSourcePandaScore, 150)
	r := tracker.GetReliability(oracle_vo.OracleSourcePandaScore)
	require.NotNil(t, r)
	assert.Equal(t, int64(1), r.TotalSubmissions)
	assert.Equal(t, int64(1), r.AccurateCount)
	assert.Equal(t, int64(150), r.AverageLatencyMs)
	assert.Equal(t, 1.0, r.AccuracyRate())
}

func TestReliabilityTracker_RecordInaccurate(t *testing.T) {
	tracker := NewProviderReliabilityTracker()
	tracker.RecordAccurate(oracle_vo.OracleSourceFACEIT, 100)
	tracker.RecordInaccurate(oracle_vo.OracleSourceFACEIT)
	r := tracker.GetReliability(oracle_vo.OracleSourceFACEIT)
	require.NotNil(t, r)
	assert.Equal(t, int64(2), r.TotalSubmissions)
	assert.Equal(t, 0.5, r.AccuracyRate())
}

func TestReliabilityTracker_RecordAnomaly(t *testing.T) {
	tracker := NewProviderReliabilityTracker()
	tracker.RecordAnomaly(oracle_vo.OracleSourceOCRUpload)
	r := tracker.GetReliability(oracle_vo.OracleSourceOCRUpload)
	require.NotNil(t, r)
	assert.Equal(t, int64(1), r.AnomalyCount)
}

func TestReliabilityTracker_UnknownSource(t *testing.T) {
	tracker := NewProviderReliabilityTracker()
	r := tracker.GetReliability(oracle_vo.OracleSourcePandaScore)
	assert.Nil(t, r)
}

func TestReliabilityTracker_GetAll(t *testing.T) {
	tracker := NewProviderReliabilityTracker()
	tracker.RecordAccurate(oracle_vo.OracleSourcePandaScore, 100)
	tracker.RecordAccurate(oracle_vo.OracleSourceSteamWebAPI, 200)
	all := tracker.GetAllReliability()
	assert.Len(t, all, 2)
	assert.Contains(t, all, oracle_vo.OracleSourcePandaScore)
	assert.Contains(t, all, oracle_vo.OracleSourceSteamWebAPI)
}

func TestReliability_AccuracyRate_NoSubmissions(t *testing.T) {
	r := &ProviderReliability{}
	assert.Equal(t, 1.0, r.AccuracyRate())
}

func TestReliability_AnomalyRate(t *testing.T) {
	r := &ProviderReliability{TotalSubmissions: 10, AnomalyCount: 3}
	assert.Equal(t, 0.3, r.AnomalyRate())
}

func TestReliability_AnomalyRate_NoSubmissions(t *testing.T) {
	r := &ProviderReliability{}
	assert.Equal(t, 0.0, r.AnomalyRate())
}

func TestMeanStd(t *testing.T) {
	mean, std := meanStd([]float64{16, 16, 16})
	assert.Equal(t, 16.0, mean)
	assert.Equal(t, 0.0, std)
	mean, std = meanStd([]float64{10, 20, 30})
	assert.InDelta(t, 20.0, mean, 0.01)
	assert.Greater(t, std, 0.0)
	mean, std = meanStd([]float64{})
	assert.Equal(t, 0.0, mean)
	assert.Equal(t, 0.0, std)
}

func TestComputeSourceHash(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	submissions := []oracle_entities.ScoreSubmission{
		newTestSubmission(oracle_vo.OracleSourcePandaScore, teamA, teamB, 16, 12),
	}
	hash := computeSourceHash(submissions)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}
