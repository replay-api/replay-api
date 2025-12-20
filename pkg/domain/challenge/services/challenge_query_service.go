package challenge_services

import (
	"context"

	"github.com/google/uuid"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
	challenge_in "github.com/replay-api/replay-api/pkg/domain/challenge/ports/in"
	challenge_out "github.com/replay-api/replay-api/pkg/domain/challenge/ports/out"
)

// ChallengeQueryServiceImpl implements the challenge query service
type ChallengeQueryServiceImpl struct {
	challengeRepo challenge_out.ChallengeRepository
}

// NewChallengeQueryService creates a new challenge query service
func NewChallengeQueryService(challengeRepo challenge_out.ChallengeRepository) challenge_in.ChallengeQueryService {
	return &ChallengeQueryServiceImpl{
		challengeRepo: challengeRepo,
	}
}

// GetByID retrieves a challenge by its ID
func (s *ChallengeQueryServiceImpl) GetByID(ctx context.Context, query challenge_in.GetChallengeByIDQuery) (*challenge_entities.Challenge, error) {
	return s.challengeRepo.GetByID(ctx, query.ChallengeID)
}

// GetByMatch retrieves all challenges for a specific match
func (s *ChallengeQueryServiceImpl) GetByMatch(ctx context.Context, query challenge_in.GetChallengesByMatchQuery) ([]*challenge_entities.Challenge, error) {
	return s.challengeRepo.GetByMatchID(ctx, query.MatchID, query.Search)
}

// Search searches challenges based on query criteria
func (s *ChallengeQueryServiceImpl) Search(ctx context.Context, query challenge_in.ChallengeQuery) ([]*challenge_entities.Challenge, int64, error) {
	criteria := challenge_out.ChallengeCriteria{
		MatchID:        query.MatchID,
		ChallengerID:   query.ChallengerID,
		GameID:         query.GameID,
		TournamentID:   query.TournamentID,
		LobbyID:        query.LobbyID,
		Types:          query.Types,
		Statuses:       query.Statuses,
		Priorities:     query.Priorities,
		IncludeExpired: query.IncludeExpired,
		Search:         query.Search,
	}

	return s.challengeRepo.Search(ctx, criteria)
}

// GetPendingChallenges retrieves pending challenges for review
func (s *ChallengeQueryServiceImpl) GetPendingChallenges(ctx context.Context, query challenge_in.GetPendingChallengesQuery) ([]*challenge_entities.Challenge, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	return s.challengeRepo.GetPending(ctx, query.Priority, query.GameID, limit)
}

// CountByStatus returns counts of challenges grouped by status
func (s *ChallengeQueryServiceImpl) CountByStatus(ctx context.Context, matchID *uuid.UUID) (map[challenge_entities.ChallengeStatus]int64, error) {
	return s.challengeRepo.CountByStatus(ctx, matchID)
}

// GetExpiredChallenges retrieves challenges that have expired
func (s *ChallengeQueryServiceImpl) GetExpiredChallenges(ctx context.Context, limit int) ([]*challenge_entities.Challenge, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}

	return s.challengeRepo.GetExpired(ctx, limit)
}

