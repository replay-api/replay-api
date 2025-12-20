package challenge_out

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
)

// ChallengeRepository defines persistence operations for challenges
type ChallengeRepository interface {
	// Save persists a challenge (create or update)
	Save(ctx context.Context, challenge *challenge_entities.Challenge) error

	// GetByID retrieves a challenge by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*challenge_entities.Challenge, error)

	// GetByMatchID retrieves all challenges for a match
	GetByMatchID(ctx context.Context, matchID uuid.UUID, search *common.Search) ([]*challenge_entities.Challenge, error)

	// GetByChallengerID retrieves challenges submitted by a player
	GetByChallengerID(ctx context.Context, challengerID uuid.UUID, search *common.Search) ([]*challenge_entities.Challenge, error)

	// Search searches challenges based on criteria
	Search(ctx context.Context, criteria ChallengeCriteria) ([]*challenge_entities.Challenge, int64, error)

	// GetPending retrieves pending challenges requiring review
	GetPending(ctx context.Context, priority *challenge_entities.ChallengePriority, gameID *string, limit int) ([]*challenge_entities.Challenge, error)

	// GetExpired retrieves expired challenges
	GetExpired(ctx context.Context, limit int) ([]*challenge_entities.Challenge, error)

	// CountByStatus counts challenges grouped by status
	CountByStatus(ctx context.Context, matchID *uuid.UUID) (map[challenge_entities.ChallengeStatus]int64, error)

	// Delete deletes a challenge by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByMatchID deletes all challenges for a match
	DeleteByMatchID(ctx context.Context, matchID uuid.UUID) error
}

// ChallengeCriteria represents search criteria for the repository
type ChallengeCriteria struct {
	MatchID        *uuid.UUID
	ChallengerID   *uuid.UUID
	GameID         *string
	TournamentID   *uuid.UUID
	LobbyID        *uuid.UUID
	Types          []challenge_entities.ChallengeType
	Statuses       []challenge_entities.ChallengeStatus
	Priorities     []challenge_entities.ChallengePriority
	IncludeExpired bool
	ResourceOwner  *common.ResourceOwner
	Search         *common.Search
}

