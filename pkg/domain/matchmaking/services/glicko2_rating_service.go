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
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
)

// Glicko2RatingService implements the Glicko-2 rating algorithm
// Used by competitive e-sports platforms for fair skill-based matchmaking
type Glicko2RatingService struct {
	ratingRepository matchmaking_out.PlayerRatingRepository
}

// NewGlicko2RatingService creates a new Glicko-2 rating service
func NewGlicko2RatingService(ratingRepository matchmaking_out.PlayerRatingRepository) matchmaking_in.RatingService {
	return &Glicko2RatingService{
		ratingRepository: ratingRepository,
	}
}

// GetPlayerRating retrieves or creates a player's rating
func (s *Glicko2RatingService) GetPlayerRating(ctx context.Context, playerID uuid.UUID, gameID common.GameIDKey) (*matchmaking_entities.PlayerRating, error) {
	rating, err := s.ratingRepository.FindByPlayerAndGame(ctx, playerID, gameID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find player rating", "player_id", playerID, "game_id", gameID, "error", err)
		return nil, fmt.Errorf("failed to find player rating: %w", err)
	}

	// Create new rating if none exists
	if rating == nil {
		resourceOwner := common.GetResourceOwner(ctx)
		rating = matchmaking_entities.NewPlayerRating(playerID, gameID, resourceOwner)
		
		if err := s.ratingRepository.Save(ctx, rating); err != nil {
			slog.ErrorContext(ctx, "Failed to create player rating", "player_id", playerID, "error", err)
			return nil, fmt.Errorf("failed to create player rating: %w", err)
		}
		
		slog.InfoContext(ctx, "Created new player rating", "player_id", playerID, "game_id", gameID, "rating", rating.Rating)
	}

	// Apply inactivity decay if needed
	if rating.LastMatchAt != nil {
		daysSinceLastMatch := int(time.Since(*rating.LastMatchAt).Hours() / 24)
		if daysSinceLastMatch > 7 {
			rating.ApplyInactivityDecay(daysSinceLastMatch)
		}
	}

	return rating, nil
}

// UpdateRatingsAfterMatch updates ratings for all players after a match
// This implements the full Glicko-2 algorithm
func (s *Glicko2RatingService) UpdateRatingsAfterMatch(ctx context.Context, cmd matchmaking_in.UpdateRatingsCommand) error {
	slog.InfoContext(ctx, "Updating ratings after match",
		"match_id", cmd.MatchID,
		"winners_count", len(cmd.WinnerPlayerIDs),
		"losers_count", len(cmd.LoserPlayerIDs),
	)

	// Get all player ratings
	allPlayerIDs := append(cmd.WinnerPlayerIDs, cmd.LoserPlayerIDs...)
	ratings := make(map[uuid.UUID]*matchmaking_entities.PlayerRating)

	for _, playerID := range allPlayerIDs {
		rating, err := s.GetPlayerRating(ctx, playerID, cmd.GameID)
		if err != nil {
			return fmt.Errorf("failed to get rating for player %s: %w", playerID, err)
		}
		ratings[playerID] = rating
	}

	// Calculate average ratings for each team
	winnerAvgRating, winnerAvgRD := s.calculateTeamAverages(ratings, cmd.WinnerPlayerIDs)
	loserAvgRating, loserAvgRD := s.calculateTeamAverages(ratings, cmd.LoserPlayerIDs)

	now := time.Now()

	// Update winner ratings
	for _, playerID := range cmd.WinnerPlayerIDs {
		rating := ratings[playerID]
		oldRating := rating.Rating
		
		s.updateSingleRating(rating, loserAvgRating, loserAvgRD, 1.0) // score = 1.0 for win
		
		rating.Wins++
		rating.MatchesPlayed++
		rating.WinStreak++
		rating.LastMatchAt = &now
		rating.UpdatedAt = now
		
		if rating.Rating > rating.PeakRating {
			rating.PeakRating = rating.Rating
		}
		
		// Record history (keep last 50)
		change := matchmaking_entities.RatingChange{
			MatchID:        cmd.MatchID,
			OldRating:      oldRating,
			NewRating:      rating.Rating,
			Change:         rating.Rating - oldRating,
			Result:         "win",
			OpponentRating: loserAvgRating,
			Timestamp:      now,
		}
		rating.RatingHistory = append(rating.RatingHistory, change)
		if len(rating.RatingHistory) > 50 {
			rating.RatingHistory = rating.RatingHistory[len(rating.RatingHistory)-50:]
		}
		
		if err := s.ratingRepository.Update(ctx, rating); err != nil {
			slog.ErrorContext(ctx, "Failed to update winner rating", "player_id", playerID, "error", err)
			return fmt.Errorf("failed to update winner rating: %w", err)
		}
		
		slog.InfoContext(ctx, "Updated winner rating",
			"player_id", playerID,
			"old_rating", oldRating,
			"new_rating", rating.Rating,
			"change", rating.Rating-oldRating,
		)
	}

	// Update loser ratings
	for _, playerID := range cmd.LoserPlayerIDs {
		rating := ratings[playerID]
		oldRating := rating.Rating
		
		s.updateSingleRating(rating, winnerAvgRating, winnerAvgRD, 0.0) // score = 0.0 for loss
		
		rating.Losses++
		rating.MatchesPlayed++
		rating.WinStreak = 0
		rating.LastMatchAt = &now
		rating.UpdatedAt = now
		
		change := matchmaking_entities.RatingChange{
			MatchID:        cmd.MatchID,
			OldRating:      oldRating,
			NewRating:      rating.Rating,
			Change:         rating.Rating - oldRating,
			Result:         "loss",
			OpponentRating: winnerAvgRating,
			Timestamp:      now,
		}
		rating.RatingHistory = append(rating.RatingHistory, change)
		if len(rating.RatingHistory) > 50 {
			rating.RatingHistory = rating.RatingHistory[len(rating.RatingHistory)-50:]
		}
		
		if err := s.ratingRepository.Update(ctx, rating); err != nil {
			slog.ErrorContext(ctx, "Failed to update loser rating", "player_id", playerID, "error", err)
			return fmt.Errorf("failed to update loser rating: %w", err)
		}
		
		slog.InfoContext(ctx, "Updated loser rating",
			"player_id", playerID,
			"old_rating", oldRating,
			"new_rating", rating.Rating,
			"change", rating.Rating-oldRating,
		)
	}

	return nil
}

// calculateTeamAverages calculates average rating and RD for a team
func (s *Glicko2RatingService) calculateTeamAverages(ratings map[uuid.UUID]*matchmaking_entities.PlayerRating, playerIDs []uuid.UUID) (avgRating, avgRD float64) {
	if len(playerIDs) == 0 {
		return matchmaking_entities.DefaultRating, matchmaking_entities.DefaultRatingDeviation
	}

	var totalRating, totalRD float64
	for _, id := range playerIDs {
		if r, ok := ratings[id]; ok {
			totalRating += r.Rating
			totalRD += r.RatingDeviation
		}
	}

	return totalRating / float64(len(playerIDs)), totalRD / float64(len(playerIDs))
}

// updateSingleRating applies the Glicko-2 update for a single player
func (s *Glicko2RatingService) updateSingleRating(player *matchmaking_entities.PlayerRating, oppRating, oppRD, score float64) {
	// Step 1: Convert to Glicko-2 scale
	mu := (player.Rating - matchmaking_entities.DefaultRating) / matchmaking_entities.ScaleFactor
	phi := player.RatingDeviation / matchmaking_entities.ScaleFactor
	sigma := player.Volatility

	muJ := (oppRating - matchmaking_entities.DefaultRating) / matchmaking_entities.ScaleFactor
	phiJ := oppRD / matchmaking_entities.ScaleFactor

	// Step 2: Calculate g(φ) and E
	gPhiJ := 1.0 / math.Sqrt(1.0+3.0*phiJ*phiJ/(math.Pi*math.Pi))
	E := 1.0 / (1.0 + math.Exp(-gPhiJ*(mu-muJ)))

	// Step 3: Calculate v (variance)
	v := 1.0 / (gPhiJ * gPhiJ * E * (1.0 - E))

	// Step 4: Calculate delta
	delta := v * gPhiJ * (score - E)

	// Step 5: Update volatility (simplified Illinois algorithm)
	a := math.Log(sigma * sigma)
	tau := matchmaking_entities.Tau
	
	// Convergence tolerance
	epsilon := 0.000001
	
	// Illinois algorithm for finding new volatility
	A := a
	B := 0.0
	if delta*delta > phi*phi+v {
		B = math.Log(delta*delta - phi*phi - v)
	} else {
		k := 1.0
		for s.f(a-k*tau, delta, phi, v, a) < 0 {
			k++
		}
		B = a - k*tau
	}

	fA := s.f(A, delta, phi, v, a)
	fB := s.f(B, delta, phi, v, a)

	for math.Abs(B-A) > epsilon {
		C := A + (A-B)*fA/(fB-fA)
		fC := s.f(C, delta, phi, v, a)

		if fC*fB < 0 {
			A = B
			fA = fB
		} else {
			fA = fA / 2
		}

		B = C
		fB = fC
	}

	newSigma := math.Exp(A / 2)

	// Step 6: Update phi* (pre-rating period)
	phiStar := math.Sqrt(phi*phi + newSigma*newSigma)

	// Step 7: Update phi and mu
	newPhi := 1.0 / math.Sqrt(1.0/(phiStar*phiStar)+1.0/v)
	newMu := mu + newPhi*newPhi*gPhiJ*(score-E)

	// Step 8: Convert back to original scale
	player.Rating = newMu*matchmaking_entities.ScaleFactor + matchmaking_entities.DefaultRating
	player.RatingDeviation = newPhi * matchmaking_entities.ScaleFactor
	player.Volatility = newSigma

	// Clamp rating to reasonable bounds
	player.Rating = math.Max(100, math.Min(4000, player.Rating))
	player.RatingDeviation = math.Max(30, math.Min(matchmaking_entities.DefaultRatingDeviation, player.RatingDeviation))
}

// f function for volatility calculation
func (s *Glicko2RatingService) f(x, delta, phi, v, a float64) float64 {
	ex := math.Exp(x)
	num := ex * (delta*delta - phi*phi - v - ex)
	denom := 2.0 * (phi*phi + v + ex) * (phi*phi + v + ex)
	tau := matchmaking_entities.Tau
	return num/denom - (x-a)/(tau*tau)
}

// GetLeaderboard returns top players by rating
func (s *Glicko2RatingService) GetLeaderboard(ctx context.Context, gameID common.GameIDKey, limit int) ([]*matchmaking_entities.PlayerRating, error) {
	return s.ratingRepository.GetTopPlayers(ctx, gameID, limit)
}

// GetRankDistribution returns the distribution of players across ranks
func (s *Glicko2RatingService) GetRankDistribution(ctx context.Context, gameID common.GameIDKey) (map[matchmaking_entities.Rank]int, error) {
	return s.ratingRepository.GetRankDistribution(ctx, gameID)
}

// Ensure Glicko2RatingService implements RatingService
var _ matchmaking_in.RatingService = (*Glicko2RatingService)(nil)

