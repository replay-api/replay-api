package matchmaking_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// SmurfDetectionStatus indicates the result of smurf analysis
type SmurfDetectionStatus string

const (
	SmurfStatusClean      SmurfDetectionStatus = "CLEAN"
	SmurfStatusSuspicious SmurfDetectionStatus = "SUSPICIOUS"
	SmurfStatusFlagged    SmurfDetectionStatus = "FLAGGED"
	SmurfStatusConfirmed  SmurfDetectionStatus = "CONFIRMED"
	SmurfStatusCleared    SmurfDetectionStatus = "CLEARED"
)

// SmurfIndicatorType categorizes different detection signals
type SmurfIndicatorType string

const (
	// Performance-based indicators
	IndicatorAbnormalWinRate       SmurfIndicatorType = "ABNORMAL_WIN_RATE"
	IndicatorRapidSkillProgression SmurfIndicatorType = "RAPID_SKILL_PROGRESSION"
	IndicatorInconsistentPerformance SmurfIndicatorType = "INCONSISTENT_PERFORMANCE"
	IndicatorStatisticalAnomaly    SmurfIndicatorType = "STATISTICAL_ANOMALY"
	IndicatorHeadshotRateAnomaly   SmurfIndicatorType = "HEADSHOT_RATE_ANOMALY"
	IndicatorReactionTimeAnomaly   SmurfIndicatorType = "REACTION_TIME_ANOMALY"

	// Behavioral indicators
	IndicatorNewAccountHighSkill   SmurfIndicatorType = "NEW_ACCOUNT_HIGH_SKILL"
	IndicatorPlayPatternMatch      SmurfIndicatorType = "PLAY_PATTERN_MATCH"
	IndicatorCrosshairPlacement    SmurfIndicatorType = "CROSSHAIR_PLACEMENT_EXPERT"
	IndicatorPreFirePattern        SmurfIndicatorType = "PREFIRE_PATTERN"
	IndicatorMovementOptimization  SmurfIndicatorType = "MOVEMENT_OPTIMIZATION"

	// Technical indicators
	IndicatorHardwareFingerprint   SmurfIndicatorType = "HARDWARE_FINGERPRINT_MATCH"
	IndicatorIPMatch               SmurfIndicatorType = "IP_MATCH"
	IndicatorDeviceMatch           SmurfIndicatorType = "DEVICE_MATCH"
	IndicatorPlaytimeCorrelation   SmurfIndicatorType = "PLAYTIME_CORRELATION"

	// Account linking indicators
	IndicatorSteamLinkAnomaly      SmurfIndicatorType = "STEAM_LINK_ANOMALY"
	IndicatorFriendListOverlap     SmurfIndicatorType = "FRIEND_LIST_OVERLAP"
	IndicatorNamingPattern         SmurfIndicatorType = "NAMING_PATTERN"
)

// SmurfIndicator represents a single detection signal
type SmurfIndicator struct {
	Type        SmurfIndicatorType `json:"type" bson:"type"`
	Confidence  float64            `json:"confidence" bson:"confidence"` // 0.0 to 1.0
	Weight      float64            `json:"weight" bson:"weight"`         // Impact factor
	Description string             `json:"description" bson:"description"`
	Evidence    map[string]interface{} `json:"evidence" bson:"evidence"`
	DetectedAt  time.Time          `json:"detected_at" bson:"detected_at"`
}

// SmurfProfile represents the complete smurf analysis for a player
type SmurfProfile struct {
	common.BaseEntity
	PlayerID          uuid.UUID            `json:"player_id" bson:"player_id"`
	Status            SmurfDetectionStatus `json:"status" bson:"status"`
	OverallScore      float64              `json:"overall_score" bson:"overall_score"` // 0-100, higher = more likely smurf
	Indicators        []SmurfIndicator     `json:"indicators" bson:"indicators"`

	// Performance Metrics
	PerformanceAnalysis PerformanceAnalysis `json:"performance_analysis" bson:"performance_analysis"`

	// Potential main accounts
	LinkedAccounts    []LinkedAccountMatch `json:"linked_accounts" bson:"linked_accounts"`

	// Historical analysis
	MatchesAnalyzed   int                  `json:"matches_analyzed" bson:"matches_analyzed"`
	FirstAnalyzedAt   time.Time            `json:"first_analyzed_at" bson:"first_analyzed_at"`
	LastAnalyzedAt    time.Time            `json:"last_analyzed_at" bson:"last_analyzed_at"`

	// Admin review
	ReviewedBy        *uuid.UUID           `json:"reviewed_by,omitempty" bson:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time           `json:"reviewed_at,omitempty" bson:"reviewed_at,omitempty"`
	ReviewNotes       string               `json:"review_notes,omitempty" bson:"review_notes,omitempty"`

	// Actions taken
	Restrictions      []SmurfRestriction   `json:"restrictions" bson:"restrictions"`
}

// PerformanceAnalysis contains detailed performance metrics for analysis
type PerformanceAnalysis struct {
	// Win rates at different time periods
	First10GamesWinRate float64 `json:"first_10_games_win_rate" bson:"first_10_games_win_rate"`
	Last10GamesWinRate  float64 `json:"last_10_games_win_rate" bson:"last_10_games_win_rate"`
	OverallWinRate      float64 `json:"overall_win_rate" bson:"overall_win_rate"`

	// Rating progression
	InitialRating       float64 `json:"initial_rating" bson:"initial_rating"`
	CurrentRating       float64 `json:"current_rating" bson:"current_rating"`
	RatingGainRate      float64 `json:"rating_gain_rate" bson:"rating_gain_rate"` // Points per game
	RatingVolatility    float64 `json:"rating_volatility" bson:"rating_volatility"`

	// Performance consistency
	KDAVariance         float64 `json:"kda_variance" bson:"kda_variance"`
	HeadshotRateAvg     float64 `json:"headshot_rate_avg" bson:"headshot_rate_avg"`
	HeadshotRateVariance float64 `json:"headshot_rate_variance" bson:"headshot_rate_variance"`

	// Game knowledge indicators
	UtilityUsageScore   float64 `json:"utility_usage_score" bson:"utility_usage_score"`
	MapKnowledgeScore   float64 `json:"map_knowledge_score" bson:"map_knowledge_score"`
	EconomyDecisionScore float64 `json:"economy_decision_score" bson:"economy_decision_score"`

	// Mechanical skill indicators
	AimAccuracy         float64 `json:"aim_accuracy" bson:"aim_accuracy"`
	ReactionTimeAvg     float64 `json:"reaction_time_avg_ms" bson:"reaction_time_avg_ms"`
	FlickAccuracy       float64 `json:"flick_accuracy" bson:"flick_accuracy"`
	SprayControl        float64 `json:"spray_control" bson:"spray_control"`

	// Comparison to rank average
	PerformanceVsRankAvg float64 `json:"performance_vs_rank_avg" bson:"performance_vs_rank_avg"`

	// Expected vs actual placement
	ExpectedRank        string  `json:"expected_rank" bson:"expected_rank"`
	ActualRank          string  `json:"actual_rank" bson:"actual_rank"`
	RankDiscrepancy     float64 `json:"rank_discrepancy" bson:"rank_discrepancy"`
}

// LinkedAccountMatch represents a potential linked main account
type LinkedAccountMatch struct {
	AccountID       uuid.UUID   `json:"account_id" bson:"account_id"`
	MatchConfidence float64     `json:"match_confidence" bson:"match_confidence"`
	MatchReasons    []string    `json:"match_reasons" bson:"match_reasons"`
	DetectedAt      time.Time   `json:"detected_at" bson:"detected_at"`
}

// SmurfRestriction represents an action taken against a suspected smurf
type SmurfRestriction struct {
	Type        RestrictionType `json:"type" bson:"type"`
	Reason      string          `json:"reason" bson:"reason"`
	StartedAt   time.Time       `json:"started_at" bson:"started_at"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	IssuedBy    uuid.UUID       `json:"issued_by" bson:"issued_by"`
}

// RestrictionType defines types of restrictions for smurfs
type RestrictionType string

const (
	RestrictionPrizePoolExclusion RestrictionType = "PRIZE_POOL_EXCLUSION"
	RestrictionRankedRestriction  RestrictionType = "RANKED_RESTRICTION"
	RestrictionRatingAdjustment   RestrictionType = "RATING_ADJUSTMENT"
	RestrictionAccountLink        RestrictionType = "ACCOUNT_LINK_REQUIRED"
	RestrictionPhoneVerification  RestrictionType = "PHONE_VERIFICATION_REQUIRED"
	RestrictionMatchmakingPenalty RestrictionType = "MATCHMAKING_PENALTY"
)

// NewSmurfProfile creates a new smurf detection profile
func NewSmurfProfile(playerID uuid.UUID, rxn common.ResourceOwner) *SmurfProfile {
	now := time.Now().UTC()
	return &SmurfProfile{
		BaseEntity:      common.NewEntity(rxn),
		PlayerID:        playerID,
		Status:          SmurfStatusClean,
		OverallScore:    0,
		Indicators:      make([]SmurfIndicator, 0),
		FirstAnalyzedAt: now,
		LastAnalyzedAt:  now,
		Restrictions:    make([]SmurfRestriction, 0),
	}
}

// AddIndicator adds a new detection indicator
func (s *SmurfProfile) AddIndicator(indicator SmurfIndicator) {
	s.Indicators = append(s.Indicators, indicator)
	s.recalculateScore()
	s.updateStatus()
}

// recalculateScore updates the overall smurf score based on indicators
func (s *SmurfProfile) recalculateScore() {
	if len(s.Indicators) == 0 {
		s.OverallScore = 0
		return
	}

	var totalWeight float64
	var weightedScore float64

	for _, indicator := range s.Indicators {
		weight := indicator.Weight
		if weight == 0 {
			weight = 1.0
		}
		totalWeight += weight
		weightedScore += indicator.Confidence * weight * 100
	}

	s.OverallScore = weightedScore / totalWeight
}

// updateStatus updates status based on overall score
func (s *SmurfProfile) updateStatus() {
	switch {
	case s.OverallScore >= 80:
		s.Status = SmurfStatusFlagged
	case s.OverallScore >= 50:
		s.Status = SmurfStatusSuspicious
	default:
		s.Status = SmurfStatusClean
	}
}

// IsRestricted checks if the player has any active restrictions
func (s *SmurfProfile) IsRestricted() bool {
	now := time.Now().UTC()
	for _, r := range s.Restrictions {
		if r.ExpiresAt == nil || r.ExpiresAt.After(now) {
			return true
		}
	}
	return false
}

// GetActiveRestrictions returns currently active restrictions
func (s *SmurfProfile) GetActiveRestrictions() []SmurfRestriction {
	now := time.Now().UTC()
	var active []SmurfRestriction
	for _, r := range s.Restrictions {
		if r.ExpiresAt == nil || r.ExpiresAt.After(now) {
			active = append(active, r)
		}
	}
	return active
}

// GetHighConfidenceIndicators returns indicators above a threshold
func (s *SmurfProfile) GetHighConfidenceIndicators(threshold float64) []SmurfIndicator {
	var high []SmurfIndicator
	for _, i := range s.Indicators {
		if i.Confidence >= threshold {
			high = append(high, i)
		}
	}
	return high
}

// CanParticipateInPrizePool checks eligibility for prize pools
func (s *SmurfProfile) CanParticipateInPrizePool() bool {
	// Flagged or confirmed smurfs cannot participate
	if s.Status == SmurfStatusFlagged || s.Status == SmurfStatusConfirmed {
		return false
	}

	// Check for prize pool exclusion restriction
	for _, r := range s.GetActiveRestrictions() {
		if r.Type == RestrictionPrizePoolExclusion {
			return false
		}
	}

	return true
}

// SmurfDetectionThresholds defines configurable thresholds
type SmurfDetectionThresholds struct {
	// Win rate thresholds
	NewAccountHighWinRateThreshold float64 `json:"new_account_high_win_rate" bson:"new_account_high_win_rate"`
	SuspiciousWinRateThreshold     float64 `json:"suspicious_win_rate" bson:"suspicious_win_rate"`

	// Rating progression thresholds
	RapidRatingGainThreshold       float64 `json:"rapid_rating_gain" bson:"rapid_rating_gain"`
	NormalRatingGainMax            float64 `json:"normal_rating_gain_max" bson:"normal_rating_gain_max"`

	// Performance thresholds
	HighHeadshotRateThreshold      float64 `json:"high_headshot_rate" bson:"high_headshot_rate"`
	ExpertKnowledgeThreshold       float64 `json:"expert_knowledge" bson:"expert_knowledge"`

	// Match count for new account
	NewAccountMatchThreshold       int     `json:"new_account_match_threshold" bson:"new_account_match_threshold"`

	// Overall score thresholds
	SuspiciousScoreThreshold       float64 `json:"suspicious_score" bson:"suspicious_score"`
	FlaggedScoreThreshold          float64 `json:"flagged_score" bson:"flagged_score"`
}

// DefaultSmurfThresholds provides sensible defaults
var DefaultSmurfThresholds = SmurfDetectionThresholds{
	NewAccountHighWinRateThreshold: 0.75,
	SuspiciousWinRateThreshold:     0.85,
	RapidRatingGainThreshold:       50.0,
	NormalRatingGainMax:            25.0,
	HighHeadshotRateThreshold:      0.50,
	ExpertKnowledgeThreshold:       0.80,
	NewAccountMatchThreshold:       20,
	SuspiciousScoreThreshold:       50.0,
	FlaggedScoreThreshold:          80.0,
}

