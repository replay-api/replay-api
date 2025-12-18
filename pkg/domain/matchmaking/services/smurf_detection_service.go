package matchmaking_services

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
)

// SmurfDetectionService implements e-sports level smurf detection
type SmurfDetectionService struct {
	smurfRepo     SmurfProfileRepository
	matchRepo     MatchStatsRepository
	ratingRepo    PlayerRatingRepository
	thresholds    matchmaking_entities.SmurfDetectionThresholds
}

// SmurfProfileRepository defines persistence operations
type SmurfProfileRepository interface {
	Create(ctx context.Context, profile *matchmaking_entities.SmurfProfile) error
	Update(ctx context.Context, profile *matchmaking_entities.SmurfProfile) error
	GetByPlayerID(ctx context.Context, playerID uuid.UUID) (*matchmaking_entities.SmurfProfile, error)
	GetFlaggedProfiles(ctx context.Context, limit, offset int) ([]matchmaking_entities.SmurfProfile, error)
}

// MatchStatsRepository provides match statistics
type MatchStatsRepository interface {
	GetPlayerMatchStats(ctx context.Context, playerID uuid.UUID, limit int) ([]MatchStatsSummary, error)
	GetPlayerMatchCount(ctx context.Context, playerID uuid.UUID) (int, error)
}

// PlayerRatingRepository provides rating data
type PlayerRatingRepository interface {
	GetPlayerRating(ctx context.Context, playerID uuid.UUID) (*matchmaking_entities.PlayerRating, error)
	GetRatingHistory(ctx context.Context, playerID uuid.UUID, limit int) ([]RatingSnapshot, error)
}

// MatchStatsSummary contains summarized match stats for analysis
type MatchStatsSummary struct {
	MatchID       uuid.UUID
	PlayedAt      time.Time
	Won           bool
	Kills         int
	Deaths        int
	Assists       int
	Headshots     int
	TotalShots    int
	UtilityDamage float64
	FlashAssists  int
	EntryFrags    int
	ClutchWins    int
	RatingChange  float64
}

// RatingSnapshot represents a historical rating point
type RatingSnapshot struct {
	Rating    float64
	Deviation float64
	Timestamp time.Time
}

// NewSmurfDetectionService creates a new smurf detection service
func NewSmurfDetectionService(
	smurfRepo SmurfProfileRepository,
	matchRepo MatchStatsRepository,
	ratingRepo PlayerRatingRepository,
	thresholds matchmaking_entities.SmurfDetectionThresholds,
) *SmurfDetectionService {
	return &SmurfDetectionService{
		smurfRepo:  smurfRepo,
		matchRepo:  matchRepo,
		ratingRepo: ratingRepo,
		thresholds: thresholds,
	}
}

// AnalyzePlayer performs comprehensive smurf analysis on a player
func (s *SmurfDetectionService) AnalyzePlayer(ctx context.Context, playerID uuid.UUID) (*matchmaking_entities.SmurfProfile, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	// Get or create smurf profile
	profile, err := s.smurfRepo.GetByPlayerID(ctx, playerID)
	if err != nil || profile == nil {
		profile = matchmaking_entities.NewSmurfProfile(playerID, resourceOwner)
	}

	// Get match statistics
	matchStats, err := s.matchRepo.GetPlayerMatchStats(ctx, playerID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get match stats: %w", err)
	}

	matchCount, err := s.matchRepo.GetPlayerMatchCount(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match count: %w", err)
	}

	// Get rating history
	ratingHistory, err := s.ratingRepo.GetRatingHistory(ctx, playerID, 50)
	if err != nil {
		slog.WarnContext(ctx, "Failed to get rating history", "error", err)
	}

	// Get current rating
	currentRating, err := s.ratingRepo.GetPlayerRating(ctx, playerID)
	if err != nil {
		slog.WarnContext(ctx, "Failed to get current rating", "error", err)
	}

	// Clear previous indicators for fresh analysis
	profile.Indicators = make([]matchmaking_entities.SmurfIndicator, 0)
	profile.MatchesAnalyzed = len(matchStats)
	profile.LastAnalyzedAt = time.Now().UTC()

	// Run detection algorithms
	s.analyzeWinRate(profile, matchStats, matchCount)
	s.analyzeRatingProgression(profile, ratingHistory, currentRating)
	s.analyzePerformanceMetrics(profile, matchStats)
	s.analyzeGameKnowledge(profile, matchStats)
	s.analyzeMechanicalSkill(profile, matchStats)
	s.analyzeNewAccountBehavior(profile, matchStats, matchCount)

	// Calculate performance analysis
	profile.PerformanceAnalysis = s.calculatePerformanceAnalysis(matchStats, ratingHistory, currentRating)

	// Save updated profile
	if profile.ID == uuid.Nil {
		err = s.smurfRepo.Create(ctx, profile)
	} else {
		err = s.smurfRepo.Update(ctx, profile)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save smurf profile: %w", err)
	}

	slog.InfoContext(ctx, "Smurf analysis completed",
		"player_id", playerID,
		"status", profile.Status,
		"score", profile.OverallScore,
		"indicators", len(profile.Indicators),
	)

	return profile, nil
}

// analyzeWinRate checks for abnormal win rates
func (s *SmurfDetectionService) analyzeWinRate(profile *matchmaking_entities.SmurfProfile, stats []MatchStatsSummary, totalMatches int) {
	if len(stats) < 10 {
		return
	}

	// Calculate win rates for different periods
	var wins, first10Wins, last10Wins int
	for i, stat := range stats {
		if stat.Won {
			wins++
			if i < 10 {
				first10Wins++
			}
			if i >= len(stats)-10 {
				last10Wins++
			}
		}
	}

	overallWinRate := float64(wins) / float64(len(stats))
	first10WinRate := float64(first10Wins) / 10.0
	last10WinRate := float64(last10Wins) / 10.0

	// New account with high win rate
	if totalMatches <= s.thresholds.NewAccountMatchThreshold && overallWinRate >= s.thresholds.NewAccountHighWinRateThreshold {
		profile.AddIndicator(matchmaking_entities.SmurfIndicator{
			Type:        matchmaking_entities.IndicatorNewAccountHighSkill,
			Confidence:  calculateConfidence(overallWinRate, s.thresholds.NewAccountHighWinRateThreshold, 1.0),
			Weight:      2.0,
			Description: fmt.Sprintf("New account with %.0f%% win rate over %d matches", overallWinRate*100, len(stats)),
			Evidence: map[string]interface{}{
				"win_rate":     overallWinRate,
				"match_count":  len(stats),
				"total_matches": totalMatches,
			},
			DetectedAt: time.Now().UTC(),
		})
	}

	// Abnormally high win rate overall
	if overallWinRate >= s.thresholds.SuspiciousWinRateThreshold && len(stats) >= 20 {
		profile.AddIndicator(matchmaking_entities.SmurfIndicator{
			Type:        matchmaking_entities.IndicatorAbnormalWinRate,
			Confidence:  calculateConfidence(overallWinRate, s.thresholds.SuspiciousWinRateThreshold, 1.0),
			Weight:      2.5,
			Description: fmt.Sprintf("Sustained %.0f%% win rate over %d matches", overallWinRate*100, len(stats)),
			Evidence: map[string]interface{}{
				"win_rate":    overallWinRate,
				"match_count": len(stats),
			},
			DetectedAt: time.Now().UTC(),
		})
	}

	// Inconsistent performance (high first games, then lower - "throwing")
	if first10WinRate >= 0.8 && last10WinRate < 0.5 {
		profile.AddIndicator(matchmaking_entities.SmurfIndicator{
			Type:        matchmaking_entities.IndicatorInconsistentPerformance,
			Confidence:  0.6,
			Weight:      1.5,
			Description: "Significant win rate drop from first to recent matches",
			Evidence: map[string]interface{}{
				"first_10_win_rate": first10WinRate,
				"last_10_win_rate":  last10WinRate,
			},
			DetectedAt: time.Now().UTC(),
		})
	}
}

// analyzeRatingProgression checks for abnormal rating gains
func (s *SmurfDetectionService) analyzeRatingProgression(profile *matchmaking_entities.SmurfProfile, history []RatingSnapshot, current *matchmaking_entities.PlayerRating) {
	if len(history) < 5 || current == nil {
		return
	}

	// Calculate rating gain rate
	oldest := history[len(history)-1]
	newest := history[0]
	timeDiff := newest.Timestamp.Sub(oldest.Timestamp).Hours() / 24 // Days
	ratingGain := newest.Rating - oldest.Rating

	if timeDiff > 0 {
		ratingGainPerDay := ratingGain / timeDiff

		// Rapid rating gain
		if ratingGainPerDay >= s.thresholds.RapidRatingGainThreshold {
			profile.AddIndicator(matchmaking_entities.SmurfIndicator{
				Type:        matchmaking_entities.IndicatorRapidSkillProgression,
				Confidence:  calculateConfidence(ratingGainPerDay, s.thresholds.RapidRatingGainThreshold, s.thresholds.RapidRatingGainThreshold*2),
				Weight:      2.0,
				Description: fmt.Sprintf("Rating gain of %.1f points/day over %.0f days", ratingGainPerDay, timeDiff),
				Evidence: map[string]interface{}{
					"rating_gain_per_day": ratingGainPerDay,
					"total_gain":          ratingGain,
					"period_days":         timeDiff,
					"initial_rating":      oldest.Rating,
					"current_rating":      newest.Rating,
				},
				DetectedAt: time.Now().UTC(),
			})
		}
	}

	// Check for low initial placement with high current rating
	if oldest.Rating < 1200 && current.Rating > 1800 {
		profile.AddIndicator(matchmaking_entities.SmurfIndicator{
			Type:        matchmaking_entities.IndicatorStatisticalAnomaly,
			Confidence:  0.7,
			Weight:      1.8,
			Description: fmt.Sprintf("Large rating gap: started at %.0f, now at %.0f", oldest.Rating, current.Rating),
			Evidence: map[string]interface{}{
				"initial_rating": oldest.Rating,
				"current_rating": current.Rating,
				"gain":           current.Rating - oldest.Rating,
			},
			DetectedAt: time.Now().UTC(),
		})
	}
}

// analyzePerformanceMetrics checks for expert-level performance
func (s *SmurfDetectionService) analyzePerformanceMetrics(profile *matchmaking_entities.SmurfProfile, stats []MatchStatsSummary) {
	if len(stats) < 10 {
		return
	}

	// Calculate headshot rate
	var totalHeadshots, totalShots int
	for _, stat := range stats {
		totalHeadshots += stat.Headshots
		totalShots += stat.TotalShots
	}

	if totalShots > 0 {
		headshotRate := float64(totalHeadshots) / float64(totalShots)

		// High headshot rate (pro level is typically 50%+)
		if headshotRate >= s.thresholds.HighHeadshotRateThreshold {
			profile.AddIndicator(matchmaking_entities.SmurfIndicator{
				Type:        matchmaking_entities.IndicatorHeadshotRateAnomaly,
				Confidence:  calculateConfidence(headshotRate, s.thresholds.HighHeadshotRateThreshold, 0.70),
				Weight:      1.8,
				Description: fmt.Sprintf("Professional-level %.0f%% headshot rate", headshotRate*100),
				Evidence: map[string]interface{}{
					"headshot_rate":   headshotRate,
					"total_headshots": totalHeadshots,
					"total_shots":     totalShots,
				},
				DetectedAt: time.Now().UTC(),
			})
		}
	}

	// Calculate KDA variance (consistent high performance is suspicious)
	var kdas []float64
	for _, stat := range stats {
		kda := float64(stat.Kills+stat.Assists) / math.Max(1, float64(stat.Deaths))
		kdas = append(kdas, kda)
	}

	if len(kdas) > 0 {
		mean := calculateMean(kdas)
		variance := calculateVariance(kdas, mean)

		// Low variance + high mean = consistently high performance
		if variance < 0.5 && mean > 2.0 {
			profile.AddIndicator(matchmaking_entities.SmurfIndicator{
				Type:        matchmaking_entities.IndicatorStatisticalAnomaly,
				Confidence:  0.65,
				Weight:      1.5,
				Description: fmt.Sprintf("Consistently high KDA (%.2f avg, %.2f variance)", mean, variance),
				Evidence: map[string]interface{}{
					"kda_mean":     mean,
					"kda_variance": variance,
				},
				DetectedAt: time.Now().UTC(),
			})
		}
	}
}

// analyzeGameKnowledge checks for expert-level game knowledge
func (s *SmurfDetectionService) analyzeGameKnowledge(profile *matchmaking_entities.SmurfProfile, stats []MatchStatsSummary) {
	if len(stats) < 10 {
		return
	}

	// Calculate utility usage score (flashes, utility damage)
	var totalFlashAssists, totalUtilityDamage float64
	var clutchWins, entryFrags int

	for _, stat := range stats {
		totalFlashAssists += float64(stat.FlashAssists)
		totalUtilityDamage += stat.UtilityDamage
		clutchWins += stat.ClutchWins
		entryFrags += stat.EntryFrags
	}

	avgFlashAssists := totalFlashAssists / float64(len(stats))
	avgUtilityDamage := totalUtilityDamage / float64(len(stats))
	avgClutchWins := float64(clutchWins) / float64(len(stats))
	avgEntryFrags := float64(entryFrags) / float64(len(stats))

	// Expert-level utility usage for new player
	if avgFlashAssists >= 2.0 && avgUtilityDamage >= 50 {
		profile.AddIndicator(matchmaking_entities.SmurfIndicator{
			Type:        matchmaking_entities.IndicatorCrosshairPlacement,
			Confidence:  0.6,
			Weight:      1.3,
			Description: "Expert-level utility usage for account age",
			Evidence: map[string]interface{}{
				"avg_flash_assists":  avgFlashAssists,
				"avg_utility_damage": avgUtilityDamage,
			},
			DetectedAt: time.Now().UTC(),
		})
	}

	// High clutch and entry performance
	if avgClutchWins >= 0.5 && avgEntryFrags >= 1.5 {
		profile.AddIndicator(matchmaking_entities.SmurfIndicator{
			Type:        matchmaking_entities.IndicatorPreFirePattern,
			Confidence:  0.55,
			Weight:      1.4,
			Description: "High clutch and entry frag rates indicating experience",
			Evidence: map[string]interface{}{
				"avg_clutch_wins": avgClutchWins,
				"avg_entry_frags": avgEntryFrags,
			},
			DetectedAt: time.Now().UTC(),
		})
	}
}

// analyzeMechanicalSkill checks for expert-level mechanics
func (s *SmurfDetectionService) analyzeMechanicalSkill(profile *matchmaking_entities.SmurfProfile, stats []MatchStatsSummary) {
	// This would integrate with replay analysis for detailed mechanical analysis
	// For now, we use available statistics as proxies
	if len(stats) < 10 {
		return
	}

	// Entry frag success rate as proxy for crosshair placement
	var entrySuccesses, totalEntries int
	for _, stat := range stats {
		if stat.EntryFrags > 0 {
			entrySuccesses += stat.EntryFrags
		}
		totalEntries++
	}

	if totalEntries > 0 {
		entryRate := float64(entrySuccesses) / float64(totalEntries)
		if entryRate >= 1.5 { // Average 1.5+ entry frags per match
			profile.AddIndicator(matchmaking_entities.SmurfIndicator{
				Type:        matchmaking_entities.IndicatorMovementOptimization,
				Confidence:  0.5,
				Weight:      1.2,
				Description: "High entry frag rate suggesting expert movement",
				Evidence: map[string]interface{}{
					"avg_entry_frags_per_match": entryRate,
				},
				DetectedAt: time.Now().UTC(),
			})
		}
	}
}

// analyzeNewAccountBehavior checks for new account suspicious patterns
func (s *SmurfDetectionService) analyzeNewAccountBehavior(profile *matchmaking_entities.SmurfProfile, stats []MatchStatsSummary, totalMatches int) {
	if totalMatches > s.thresholds.NewAccountMatchThreshold {
		return // Not a new account
	}

	// Check if performance is too good for account age
	if len(stats) >= 5 {
		var wins int
		var totalKDA float64
		for _, stat := range stats {
			if stat.Won {
				wins++
			}
			totalKDA += float64(stat.Kills+stat.Assists) / math.Max(1, float64(stat.Deaths))
		}

		winRate := float64(wins) / float64(len(stats))
		avgKDA := totalKDA / float64(len(stats))

		// New account with exceptional stats
		if winRate >= 0.8 && avgKDA >= 2.5 {
			profile.AddIndicator(matchmaking_entities.SmurfIndicator{
				Type:        matchmaking_entities.IndicatorNewAccountHighSkill,
				Confidence:  0.75,
				Weight:      2.2,
				Description: fmt.Sprintf("New account dominance: %.0f%% WR, %.2f KDA in %d matches", winRate*100, avgKDA, len(stats)),
				Evidence: map[string]interface{}{
					"win_rate":      winRate,
					"avg_kda":       avgKDA,
					"match_count":   len(stats),
					"total_matches": totalMatches,
				},
				DetectedAt: time.Now().UTC(),
			})
		}
	}
}

// calculatePerformanceAnalysis builds the complete performance analysis
func (s *SmurfDetectionService) calculatePerformanceAnalysis(stats []MatchStatsSummary, history []RatingSnapshot, current *matchmaking_entities.PlayerRating) matchmaking_entities.PerformanceAnalysis {
	analysis := matchmaking_entities.PerformanceAnalysis{}

	if len(stats) == 0 {
		return analysis
	}

	// Win rates
	var wins, first10Wins, last10Wins int
	for i, stat := range stats {
		if stat.Won {
			wins++
			if i < 10 {
				first10Wins++
			}
			if i >= len(stats)-10 && i >= 0 {
				last10Wins++
			}
		}
	}

	analysis.OverallWinRate = float64(wins) / float64(len(stats))
	if len(stats) >= 10 {
		analysis.First10GamesWinRate = float64(first10Wins) / 10.0
		analysis.Last10GamesWinRate = float64(last10Wins) / float64(min(10, len(stats)))
	}

	// Rating analysis
	if len(history) > 0 {
		analysis.InitialRating = history[len(history)-1].Rating
		analysis.CurrentRating = history[0].Rating

		timeDiff := history[0].Timestamp.Sub(history[len(history)-1].Timestamp).Hours()
		if timeDiff > 0 {
			analysis.RatingGainRate = (analysis.CurrentRating - analysis.InitialRating) / (timeDiff / 24)
		}
	}

	// KDA and headshot analysis
	var kdas []float64
	var headshotRates []float64
	var totalHeadshots, totalShots int

	for _, stat := range stats {
		kda := float64(stat.Kills+stat.Assists) / math.Max(1, float64(stat.Deaths))
		kdas = append(kdas, kda)

		totalHeadshots += stat.Headshots
		totalShots += stat.TotalShots
		if stat.TotalShots > 0 {
			headshotRates = append(headshotRates, float64(stat.Headshots)/float64(stat.TotalShots))
		}
	}

	if len(kdas) > 0 {
		mean := calculateMean(kdas)
		analysis.KDAVariance = calculateVariance(kdas, mean)
	}

	if totalShots > 0 {
		analysis.HeadshotRateAvg = float64(totalHeadshots) / float64(totalShots)
	}
	if len(headshotRates) > 0 {
		mean := calculateMean(headshotRates)
		analysis.HeadshotRateVariance = calculateVariance(headshotRates, mean)
	}

	// Determine expected vs actual rank
	if current != nil {
		analysis.ActualRank = string(current.GetRank())
		// Calculate expected rank based on performance metrics
		expectedRating := s.estimateSkillFromStats(stats)
		analysis.ExpectedRank = s.getRankFromRating(expectedRating)
		analysis.RankDiscrepancy = expectedRating - current.Rating
	}

	return analysis
}

// estimateSkillFromStats estimates true skill level from performance stats
func (s *SmurfDetectionService) estimateSkillFromStats(stats []MatchStatsSummary) float64 {
	if len(stats) == 0 {
		return 1000 // Default
	}

	// Weighted skill estimation based on multiple factors
	var totalKDA, totalHeadshotRate float64
	var wins int
	var headshotMatches int

	for _, stat := range stats {
		kda := float64(stat.Kills+stat.Assists) / math.Max(1, float64(stat.Deaths))
		totalKDA += kda
		if stat.Won {
			wins++
		}
		if stat.TotalShots > 0 {
			totalHeadshotRate += float64(stat.Headshots) / float64(stat.TotalShots)
			headshotMatches++
		}
	}

	avgKDA := totalKDA / float64(len(stats))
	winRate := float64(wins) / float64(len(stats))
	avgHeadshotRate := 0.0
	if headshotMatches > 0 {
		avgHeadshotRate = totalHeadshotRate / float64(headshotMatches)
	}

	// Skill formula (simplified)
	// Base: 1000, KDA contribution, Win rate contribution, Headshot contribution
	estimatedRating := 1000.0
	estimatedRating += (avgKDA - 1.0) * 200   // KDA above 1.0 adds rating
	estimatedRating += (winRate - 0.5) * 400  // Win rate above 50% adds rating
	estimatedRating += avgHeadshotRate * 300  // Headshot percentage contribution

	return math.Max(0, math.Min(3000, estimatedRating))
}

// getRankFromRating converts rating to rank name
func (s *SmurfDetectionService) getRankFromRating(rating float64) string {
	switch {
	case rating >= 2400:
		return "Global Elite"
	case rating >= 2100:
		return "Supreme"
	case rating >= 1800:
		return "LEM"
	case rating >= 1500:
		return "DMG"
	case rating >= 1200:
		return "MG2"
	case rating >= 900:
		return "MG1"
	case rating >= 600:
		return "GN4"
	default:
		return "Silver"
	}
}

// Helper functions
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateVariance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values))
}

func calculateConfidence(value, threshold, maxValue float64) float64 {
	if value <= threshold {
		return 0
	}
	if value >= maxValue {
		return 1.0
	}
	return (value - threshold) / (maxValue - threshold)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

