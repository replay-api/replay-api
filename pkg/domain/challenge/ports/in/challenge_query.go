package challenge_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
)

// ChallengeQuery represents search criteria for challenges
type ChallengeQuery struct {
	MatchID       *uuid.UUID
	ChallengerID  *uuid.UUID
	GameID        *string
	TournamentID  *uuid.UUID
	LobbyID       *uuid.UUID
	Types         []challenge_entities.ChallengeType
	Statuses      []challenge_entities.ChallengeStatus
	Priorities    []challenge_entities.ChallengePriority
	IncludeExpired bool
	Search        *common.Search
}

// GetChallengeByIDQuery represents a request to get a challenge by ID
type GetChallengeByIDQuery struct {
	ChallengeID uuid.UUID
}

// GetChallengesByMatchQuery represents a request to get challenges for a match
type GetChallengesByMatchQuery struct {
	MatchID uuid.UUID
	Search  *common.Search
}

// GetPendingChallengesQuery represents a request for pending challenges requiring review
type GetPendingChallengesQuery struct {
	Priority *challenge_entities.ChallengePriority
	GameID   *string
	Limit    int
}

// ChallengeQueryService provides read operations for challenges
type ChallengeQueryService interface {
	// GetByID retrieves a challenge by its ID
	GetByID(ctx context.Context, query GetChallengeByIDQuery) (*challenge_entities.Challenge, error)

	// GetByMatch retrieves all challenges for a specific match
	GetByMatch(ctx context.Context, query GetChallengesByMatchQuery) ([]*challenge_entities.Challenge, error)

	// Search searches challenges based on query criteria
	Search(ctx context.Context, query ChallengeQuery) ([]*challenge_entities.Challenge, int64, error)

	// GetPendingChallenges retrieves pending challenges for review
	GetPendingChallenges(ctx context.Context, query GetPendingChallengesQuery) ([]*challenge_entities.Challenge, error)

	// CountByStatus returns counts of challenges grouped by status
	CountByStatus(ctx context.Context, matchID *uuid.UUID) (map[challenge_entities.ChallengeStatus]int64, error)

	// GetExpiredChallenges retrieves challenges that have expired
	GetExpiredChallenges(ctx context.Context, limit int) ([]*challenge_entities.Challenge, error)
}

