package matchmaking_entities

import (
	"math"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// PlayerRating represents a player's skill rating using Glicko-2 system
// Glicko-2 is superior to ELO because it accounts for:
// - Rating Deviation (RD): confidence in the rating
// - Volatility (σ): how consistent the player's performance is
type PlayerRating struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	PlayerID      uuid.UUID            `json:"player_id" bson:"player_id"`
	GameID        common.GameIDKey     `json:"game_id" bson:"game_id"`
	Rating        float64              `json:"rating" bson:"rating"`               // μ (mu) - the player's rating (default: 1500)
	RatingDeviation float64            `json:"rating_deviation" bson:"rating_deviation"` // φ (phi) - uncertainty (default: 350)
	Volatility    float64              `json:"volatility" bson:"volatility"`       // σ (sigma) - consistency (default: 0.06)
	MatchesPlayed int                  `json:"matches_played" bson:"matches_played"`
	Wins          int                  `json:"wins" bson:"wins"`
	Losses        int                  `json:"losses" bson:"losses"`
	Draws         int                  `json:"draws" bson:"draws"`
	WinStreak     int                  `json:"win_streak" bson:"win_streak"`
	PeakRating    float64              `json:"peak_rating" bson:"peak_rating"`
	LastMatchAt   *time.Time           `json:"last_match_at" bson:"last_match_at"`
	RatingHistory []RatingChange       `json:"rating_history" bson:"rating_history"`
	ResourceOwner common.ResourceOwner `json:"-" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

// RatingChange tracks historical rating changes
type RatingChange struct {
	MatchID       uuid.UUID `json:"match_id" bson:"match_id"`
	OldRating     float64   `json:"old_rating" bson:"old_rating"`
	NewRating     float64   `json:"new_rating" bson:"new_rating"`
	Change        float64   `json:"change" bson:"change"`
	Result        string    `json:"result" bson:"result"` // "win", "loss", "draw"
	OpponentRating float64  `json:"opponent_rating" bson:"opponent_rating"`
	Timestamp     time.Time `json:"timestamp" bson:"timestamp"`
}

// Glicko-2 Constants
const (
	DefaultRating          = 1500.0
	DefaultRatingDeviation = 350.0
	DefaultVolatility      = 0.06
	ScaleFactor            = 173.7178 // q = ln(10)/400
	Tau                    = 0.5      // System constant constraining volatility
)

// Rank thresholds based on rating
type Rank string

const (
	RankBronze   Rank = "bronze"
	RankSilver   Rank = "silver"
	RankGold     Rank = "gold"
	RankPlatinum Rank = "platinum"
	RankDiamond  Rank = "diamond"
	RankMaster   Rank = "master"
	RankGrandmaster Rank = "grandmaster"
	RankChallenger  Rank = "challenger"
)

// NewPlayerRating creates a new player rating with default values
func NewPlayerRating(playerID uuid.UUID, gameID common.GameIDKey, resourceOwner common.ResourceOwner) *PlayerRating {
	now := time.Now()
	return &PlayerRating{
		ID:              uuid.New(),
		PlayerID:        playerID,
		GameID:          gameID,
		Rating:          DefaultRating,
		RatingDeviation: DefaultRatingDeviation,
		Volatility:      DefaultVolatility,
		MatchesPlayed:   0,
		Wins:            0,
		Losses:          0,
		Draws:           0,
		WinStreak:       0,
		PeakRating:      DefaultRating,
		RatingHistory:   make([]RatingChange, 0),
		ResourceOwner:   resourceOwner,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (r *PlayerRating) GetID() uuid.UUID {
	return r.ID
}

// GetMMR returns the integer MMR value for matchmaking
func (r *PlayerRating) GetMMR() int {
	return int(math.Round(r.Rating))
}

// GetRank returns the player's rank based on rating
func (r *PlayerRating) GetRank() Rank {
	switch {
	case r.Rating >= 2800:
		return RankChallenger
	case r.Rating >= 2500:
		return RankGrandmaster
	case r.Rating >= 2200:
		return RankMaster
	case r.Rating >= 1900:
		return RankDiamond
	case r.Rating >= 1600:
		return RankPlatinum
	case r.Rating >= 1400:
		return RankGold
	case r.Rating >= 1200:
		return RankSilver
	default:
		return RankBronze
	}
}

// GetConfidence returns confidence percentage in the rating (100% = fully confident)
func (r *PlayerRating) GetConfidence() float64 {
	// Lower RD = higher confidence
	// RD of 50 = ~100% confidence, RD of 350 = ~15% confidence
	confidence := 100.0 * (1.0 - (r.RatingDeviation / DefaultRatingDeviation))
	return math.Max(0, math.Min(100, confidence))
}

// IsProvisional returns true if player hasn't completed placement matches
func (r *PlayerRating) IsProvisional() bool {
	return r.MatchesPlayed < 10
}

// GetWinRate returns the player's win rate as a percentage
func (r *PlayerRating) GetWinRate() float64 {
	total := r.Wins + r.Losses + r.Draws
	if total == 0 {
		return 0
	}
	return float64(r.Wins) / float64(total) * 100
}

// UpdateRatingDeviation applies rating period decay
// RD increases over time to reflect uncertainty
func (r *PlayerRating) ApplyInactivityDecay(daysSinceLastMatch int) {
	if daysSinceLastMatch <= 0 {
		return
	}

	// RD increases by ~25 points per month of inactivity
	decayPerDay := 25.0 / 30.0
	newRD := math.Sqrt(r.RatingDeviation*r.RatingDeviation + decayPerDay*decayPerDay*float64(daysSinceLastMatch))
	
	// Cap RD at default (350)
	r.RatingDeviation = math.Min(newRD, DefaultRatingDeviation)
	r.UpdatedAt = time.Now()
}

// toGlicko2Scale converts rating to Glicko-2 scale (μ' = (μ-1500)/173.7178)
//
//nolint:unused // Reserved for Glicko-2 algorithm implementation
func (r *PlayerRating) toGlicko2Scale() (mu, phi float64) {
	mu = (r.Rating - DefaultRating) / ScaleFactor
	phi = r.RatingDeviation / ScaleFactor
	return mu, phi
}

// fromGlicko2Scale converts Glicko-2 scale back to normal rating
//
//nolint:unused // Reserved for Glicko-2 algorithm implementation
func fromGlicko2Scale(mu, phi float64) (rating, rd float64) {
	rating = mu*ScaleFactor + DefaultRating
	rd = phi * ScaleFactor
	return rating, rd
}

// g(φ) function for Glicko-2
//
//nolint:unused // Called by expectedScore which is used by the rating system
func g(phi float64) float64 {
	return 1.0 / math.Sqrt(1.0+3.0*phi*phi/(math.Pi*math.Pi))
}

// E(μ, μj, φj) - expected score function
func expectedScore(mu, muJ, phiJ float64) float64 {
	return 1.0 / (1.0 + math.Exp(-g(phiJ)*(mu-muJ)))
}

