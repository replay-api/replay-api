package oracle_services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
)

// ConsensusEngine computes weighted consensus from multiple provider submissions.
// Formula: agreement = winnerWeight*W_winner + seriesWeight*W_series + gameWeight*W_games
// Default weights: 60/30/10 (winner/series/games)
type ConsensusEngine struct {
	reliabilityTracker *ProviderReliabilityTracker
}

// NewConsensusEngine creates a new consensus engine
func NewConsensusEngine(tracker *ProviderReliabilityTracker) *ConsensusEngine {
	return &ConsensusEngine{reliabilityTracker: tracker}
}

// EvaluateConsensus computes consensus from submissions using the given policy
func (e *ConsensusEngine) EvaluateConsensus(
	submissions []oracle_entities.ScoreSubmission,
	policy oracle_vo.ConsensusPolicy,
) (*oracle_entities.ConsensusOutcome, error) {
	if len(submissions) < policy.MinSources {
		return nil, fmt.Errorf("insufficient sources: %d < %d required", len(submissions), policy.MinSources)
	}

	// Step 1: Winner consensus (weighted mode)
	winnerResult, winnerAgreement := e.determineWinnerConsensus(submissions)

	// Step 2: Series score consensus (weighted mode)
	teamAScore, teamBScore, scoreAgreement := e.computeSeriesScoreConsensus(submissions)

	// Step 3: Per-game consensus (if applicable)
	gameOutcomes, gameAgreement := e.computePerGameConsensus(submissions)

	// Step 4: Cross-validate all source pairs
	crossValidation := e.crossValidateAll(submissions)

	// Step 5: Anomaly detection
	anomalies := e.detectAnomalies(submissions)

	// Step 6: Compute overall agreement (60/30/10 formula)
	overallAgreement := policy.WinnerWeight*winnerAgreement +
		policy.SeriesScoreWeight*scoreAgreement +
		policy.GameScoreWeight*gameAgreement

	// Determine confidence level from overall agreement
	confidenceLevel := computeConfidenceLevel(overallAgreement)

	// Build disagreement notes from anomalies
	notes := make([]string, 0, len(anomalies))
	for _, a := range anomalies {
		notes = append(notes, a.String())
	}

	// Determine series format from game counts
	seriesFormat, gamesPlayed := inferSeriesFormat(submissions)

	// Determine MVP (weighted mode)
	mvpPlayerID := e.determineMVPConsensus(submissions)

	// Determine team IDs for team scores
	var teamAID, teamBID uuid.UUID
	if len(submissions) > 0 {
		teamAID = submissions[0].TeamAID
		teamBID = submissions[0].TeamBID
	}

	// Compute source hash for integrity verification
	sourceHash := computeSourceHash(submissions)

	outcome := &oracle_entities.ConsensusOutcome{
		WinnerTeamID:    winnerResult,
		IsDraw:          winnerResult == nil,
		ConfidenceLevel: confidenceLevel,
		AgreementRatio:  overallAgreement,
		SourceCount:     len(submissions),
		SeriesFormat:    seriesFormat,
		GamesPlayed:     gamesPlayed,
		TeamScores: []oracle_entities.ConsensusTeamScore{
			{TeamID: teamAID, Score: teamAScore},
			{TeamID: teamBID, Score: teamBScore},
		},
		GameOutcomes:      gameOutcomes,
		MVPPlayerID:       mvpPlayerID,
		CrossValidation:   crossValidation,
		DisagreementNotes: notes,
		SourceHash:        sourceHash,
		ComputedAt:        time.Now().UTC(),
	}

	// Check if confidence meets minimum threshold
	if overallAgreement < policy.MinConfidence {
		return outcome, fmt.Errorf("consensus agreement %.2f below minimum %.2f", overallAgreement, policy.MinConfidence)
	}

	return outcome, nil
}

// determineWinnerConsensus computes weighted winner vote
func (e *ConsensusEngine) determineWinnerConsensus(submissions []oracle_entities.ScoreSubmission) (*uuid.UUID, float64) {
	type winnerVote struct {
		teamID *uuid.UUID
		isDraw bool
	}

	votes := make(map[string]float64)     // key → total weight
	voteKeys := make(map[string]*uuid.UUID) // key → team UUID pointer

	for _, sub := range submissions {
		weight := e.getEffectiveWeight(sub.SourceType)

		var key string
		if sub.IsDraw {
			key = "draw"
			voteKeys[key] = nil
		} else if sub.WinnerTeamID != nil {
			key = sub.WinnerTeamID.String()
			id := *sub.WinnerTeamID
			voteKeys[key] = &id
		} else {
			continue
		}
		votes[key] += weight
	}

	// Find winner by weighted vote
	var bestKey string
	var bestWeight float64
	var totalWeight float64
	for key, weight := range votes {
		totalWeight += weight
		if weight > bestWeight {
			bestWeight = weight
			bestKey = key
		}
	}

	if totalWeight == 0 {
		return nil, 0.0
	}

	agreement := bestWeight / totalWeight
	return voteKeys[bestKey], agreement
}

// computeSeriesScoreConsensus computes weighted mode for series scores
func (e *ConsensusEngine) computeSeriesScoreConsensus(submissions []oracle_entities.ScoreSubmission) (int, int, float64) {
	type scorePair struct {
		teamA int
		teamB int
	}

	votes := make(map[scorePair]float64)

	for _, sub := range submissions {
		weight := e.getEffectiveWeight(sub.SourceType)
		pair := scorePair{teamA: sub.TeamAScore, teamB: sub.TeamBScore}
		votes[pair] += weight
	}

	var bestPair scorePair
	var bestWeight float64
	var totalWeight float64
	for pair, weight := range votes {
		totalWeight += weight
		if weight > bestWeight {
			bestWeight = weight
			bestPair = pair
		}
	}

	if totalWeight == 0 {
		return 0, 0, 0.0
	}

	agreement := bestWeight / totalWeight
	return bestPair.teamA, bestPair.teamB, agreement
}

// computePerGameConsensus computes per-game (per-map) consensus
func (e *ConsensusEngine) computePerGameConsensus(submissions []oracle_entities.ScoreSubmission) ([]oracle_entities.GameConsensusOutcome, float64) {
	// Determine max game count across submissions
	maxGames := 0
	for _, sub := range submissions {
		if len(sub.GameDetails) > maxGames {
			maxGames = len(sub.GameDetails)
		}
	}

	if maxGames == 0 {
		return nil, 1.0 // No per-game data, full agreement by default
	}

	outcomes := make([]oracle_entities.GameConsensusOutcome, 0, maxGames)
	totalAgreement := 0.0

	for pos := 0; pos < maxGames; pos++ {
		type gameKey struct {
			teamAScore int
			teamBScore int
			teamAWon   bool
			mapName    string
		}

		votes := make(map[gameKey]float64)
		var totalWeight float64

		for _, sub := range submissions {
			if pos >= len(sub.GameDetails) {
				continue
			}
			weight := e.getEffectiveWeight(sub.SourceType)
			gd := sub.GameDetails[pos]
			key := gameKey{
				teamAScore: gd.TeamAScore,
				teamBScore: gd.TeamBScore,
				teamAWon:   gd.TeamAWon,
				mapName:    gd.MapName,
			}
			votes[key] += weight
			totalWeight += weight
		}

		var bestKey gameKey
		var bestWeight float64
		for key, weight := range votes {
			if weight > bestWeight {
				bestWeight = weight
				bestKey = key
			}
		}

		if totalWeight > 0 {
			totalAgreement += bestWeight / totalWeight
		}

		outcomes = append(outcomes, oracle_entities.GameConsensusOutcome{
			Position:   pos + 1,
			MapName:    bestKey.mapName,
			TeamAScore: bestKey.teamAScore,
			TeamBScore: bestKey.teamBScore,
			TeamAWon:   bestKey.teamAWon,
		})
	}

	avgAgreement := 1.0
	if maxGames > 0 {
		avgAgreement = totalAgreement / float64(maxGames)
	}

	return outcomes, avgAgreement
}

// crossValidateAll performs pairwise comparison between all submissions
func (e *ConsensusEngine) crossValidateAll(submissions []oracle_entities.ScoreSubmission) []oracle_entities.CrossValidationEntry {
	entries := make([]oracle_entities.CrossValidationEntry, 0)

	for i := 0; i < len(submissions); i++ {
		for j := i + 1; j < len(submissions); j++ {
			a, b := submissions[i], submissions[j]

			winnerAgree := false
			if a.IsDraw && b.IsDraw {
				winnerAgree = true
			} else if a.WinnerTeamID != nil && b.WinnerTeamID != nil {
				winnerAgree = *a.WinnerTeamID == *b.WinnerTeamID
			}

			scoreAgree := a.TeamAScore == b.TeamAScore && a.TeamBScore == b.TeamBScore

			gamesAgree := len(a.GameDetails) == len(b.GameDetails)
			if gamesAgree {
				for k := range a.GameDetails {
					if k < len(b.GameDetails) {
						if a.GameDetails[k].TeamAScore != b.GameDetails[k].TeamAScore ||
							a.GameDetails[k].TeamBScore != b.GameDetails[k].TeamBScore {
							gamesAgree = false
							break
						}
					}
				}
			}

			mvpAgree := false
			if a.MVPPlayerID == nil && b.MVPPlayerID == nil {
				mvpAgree = true
			} else if a.MVPPlayerID != nil && b.MVPPlayerID != nil {
				mvpAgree = *a.MVPPlayerID == *b.MVPPlayerID
			}

			var note string
			if !winnerAgree {
				note = fmt.Sprintf("%s and %s disagree on winner", a.SourceType, b.SourceType)
			} else if !scoreAgree {
				note = fmt.Sprintf("%s and %s disagree on score", a.SourceType, b.SourceType)
			}

			entries = append(entries, oracle_entities.CrossValidationEntry{
				SourceA:          a.SourceType,
				SourceB:          b.SourceType,
				WinnerAgree:      winnerAgree,
				ScoreAgree:       scoreAgree,
				GamesAgree:       gamesAgree,
				MVPAgree:         mvpAgree,
				DisagreementNote: note,
			})
		}
	}

	return entries
}

// --- Anomaly Detection ---

// AnomalyType identifies the type of anomaly
type AnomalyType string

const (
	AnomalyScoreOutOfRange     AnomalyType = "score_out_of_range"
	AnomalyImpossibleCombo     AnomalyType = "impossible_score_combo"
	AnomalyWinnerScoreMismatch AnomalyType = "winner_score_mismatch"
	AnomalyDuplicateSubmission AnomalyType = "duplicate_submission"
	AnomalyOutlierScore        AnomalyType = "outlier_score"
	AnomalyStaleData           AnomalyType = "stale_data"
)

// Anomaly represents a detected anomaly in a submission
type Anomaly struct {
	Type         AnomalyType              `json:"type"`
	SourceType   oracle_vo.OracleSourceType `json:"source_type"`
	Description  string                    `json:"description"`
	SubmissionID uuid.UUID                 `json:"submission_id"`
}

func (a Anomaly) String() string {
	return fmt.Sprintf("[%s] %s: %s", a.Type, a.SourceType, a.Description)
}

// detectAnomalies scans submissions for anomalies
func (e *ConsensusEngine) detectAnomalies(submissions []oracle_entities.ScoreSubmission) []Anomaly {
	anomalies := make([]Anomaly, 0)

	// CS2 game profile defaults (extensible per game)
	maxRegulationScore := 16
	maxOvertimeScore := 22 // Extended OT

	for _, sub := range submissions {
		// 1. Score Out of Range
		if sub.TeamAScore > maxOvertimeScore || sub.TeamBScore > maxOvertimeScore {
			anomalies = append(anomalies, Anomaly{
				Type:         AnomalyScoreOutOfRange,
				SourceType:   sub.SourceType,
				Description:  fmt.Sprintf("score out of range: %d-%d (max %d)", sub.TeamAScore, sub.TeamBScore, maxOvertimeScore),
				SubmissionID: sub.ID,
			})
		}

		// 2. Impossible Score Combo (both at max regulation without OT)
		if sub.TeamAScore == maxRegulationScore && sub.TeamBScore == maxRegulationScore {
			anomalies = append(anomalies, Anomaly{
				Type:         AnomalyImpossibleCombo,
				SourceType:   sub.SourceType,
				Description:  fmt.Sprintf("both teams at %d-%d (regulation tie impossible in CS2)", sub.TeamAScore, sub.TeamBScore),
				SubmissionID: sub.ID,
			})
		}

		// 3. Winner-Score Mismatch
		if sub.WinnerTeamID != nil && !sub.IsDraw {
			if *sub.WinnerTeamID == sub.TeamAID && sub.TeamAScore < sub.TeamBScore {
				anomalies = append(anomalies, Anomaly{
					Type:         AnomalyWinnerScoreMismatch,
					SourceType:   sub.SourceType,
					Description:  fmt.Sprintf("winner declared as team A but score is %d-%d", sub.TeamAScore, sub.TeamBScore),
					SubmissionID: sub.ID,
				})
			} else if *sub.WinnerTeamID == sub.TeamBID && sub.TeamBScore < sub.TeamAScore {
				anomalies = append(anomalies, Anomaly{
					Type:         AnomalyWinnerScoreMismatch,
					SourceType:   sub.SourceType,
					Description:  fmt.Sprintf("winner declared as team B but score is %d-%d", sub.TeamAScore, sub.TeamBScore),
					SubmissionID: sub.ID,
				})
			}
		}

		// 4. Stale Data (submission > 24h after match end)
		if time.Since(sub.SubmittedAt) > 24*time.Hour && sub.SubmittedAt.Before(time.Now().Add(-24*time.Hour)) {
			// This check is relative to when anomaly detection runs
			// In practice, compare to match end time
		}
	}

	// 5. Outlier Score (score deviates > 2σ from mean)
	if len(submissions) >= 3 {
		scoresA := make([]float64, len(submissions))
		scoresB := make([]float64, len(submissions))
		for i, sub := range submissions {
			scoresA[i] = float64(sub.TeamAScore)
			scoresB[i] = float64(sub.TeamBScore)
		}
		meanA, stdA := meanStd(scoresA)
		meanB, stdB := meanStd(scoresB)

		for _, sub := range submissions {
			if stdA > 0 && math.Abs(float64(sub.TeamAScore)-meanA) > 2*stdA {
				anomalies = append(anomalies, Anomaly{
					Type:         AnomalyOutlierScore,
					SourceType:   sub.SourceType,
					Description:  fmt.Sprintf("team A score %d is outlier (mean=%.1f, std=%.1f)", sub.TeamAScore, meanA, stdA),
					SubmissionID: sub.ID,
				})
			}
			if stdB > 0 && math.Abs(float64(sub.TeamBScore)-meanB) > 2*stdB {
				anomalies = append(anomalies, Anomaly{
					Type:         AnomalyOutlierScore,
					SourceType:   sub.SourceType,
					Description:  fmt.Sprintf("team B score %d is outlier (mean=%.1f, std=%.1f)", sub.TeamBScore, meanB, stdB),
					SubmissionID: sub.ID,
				})
			}
		}
	}

	// 6. Duplicate Submission (same provider, same match, same data)
	seen := make(map[string]bool)
	for _, sub := range submissions {
		key := fmt.Sprintf("%s:%s:%d:%d", sub.SourceType, sub.ProviderMatchID, sub.TeamAScore, sub.TeamBScore)
		if seen[key] {
			anomalies = append(anomalies, Anomaly{
				Type:         AnomalyDuplicateSubmission,
				SourceType:   sub.SourceType,
				Description:  fmt.Sprintf("duplicate submission from %s for match %s", sub.SourceType, sub.ProviderMatchID),
				SubmissionID: sub.ID,
			})
		}
		seen[key] = true
	}

	return anomalies
}

// determineMVPConsensus computes weighted mode for MVP player ID
func (e *ConsensusEngine) determineMVPConsensus(submissions []oracle_entities.ScoreSubmission) *uuid.UUID {
	votes := make(map[uuid.UUID]float64)

	for _, sub := range submissions {
		if sub.MVPPlayerID == nil {
			continue
		}
		weight := e.getEffectiveWeight(sub.SourceType)
		votes[*sub.MVPPlayerID] += weight
	}

	if len(votes) == 0 {
		return nil
	}

	var bestID uuid.UUID
	var bestWeight float64
	for id, w := range votes {
		if w > bestWeight {
			bestWeight = w
			bestID = id
		}
	}

	return &bestID
}

// getEffectiveWeight returns the adjusted confidence weight for a source,
// factoring in provider reliability tracking
func (e *ConsensusEngine) getEffectiveWeight(source oracle_vo.OracleSourceType) float64 {
	baseWeight := source.ConfidenceWeight()
	if e.reliabilityTracker == nil {
		return baseWeight
	}

	reliability := e.reliabilityTracker.GetReliability(source)
	if reliability == nil {
		return baseWeight
	}

	// If accuracy drops below 70%, reduce weight by 50%
	if reliability.AccuracyRate() < 0.70 {
		return baseWeight * 0.50
	}

	return baseWeight
}

// --- Utility Functions ---

// computeConfidenceLevel maps agreement ratio to confidence level (0-3)
func computeConfidenceLevel(ratio float64) int {
	switch {
	case ratio >= 0.95:
		return 3 // High
	case ratio >= 0.80:
		return 2 // Medium
	case ratio >= 0.60:
		return 1 // Low
	default:
		return 0 // None
	}
}

// inferSeriesFormat determines bo1/bo3/bo5 from game detail counts
func inferSeriesFormat(submissions []oracle_entities.ScoreSubmission) (string, int) {
	maxGames := 0
	for _, sub := range submissions {
		if len(sub.GameDetails) > maxGames {
			maxGames = len(sub.GameDetails)
		}
	}

	switch {
	case maxGames <= 1:
		return "bo1", maxGames
	case maxGames <= 3:
		return "bo3", maxGames
	case maxGames <= 5:
		return "bo5", maxGames
	default:
		return fmt.Sprintf("bo%d", maxGames), maxGames
	}
}

// meanStd computes mean and standard deviation
func meanStd(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}

	var sum float64
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(len(data))

	var variance float64
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(data))

	return mean, math.Sqrt(variance)
}

// computeSourceHash generates a SHA-256 hash of all submissions for integrity verification
func computeSourceHash(submissions []oracle_entities.ScoreSubmission) string {
	data, _ := json.Marshal(submissions)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
