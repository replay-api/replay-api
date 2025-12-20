package challenge_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
)

// CreateChallengeCommand represents a request to create a new challenge
type CreateChallengeCommand struct {
	MatchID          uuid.UUID
	RoundNumber      *int
	ChallengerID     uuid.UUID
	ChallengerTeamID *uuid.UUID
	GameID           string
	LobbyID          *uuid.UUID
	TournamentID     *uuid.UUID
	Type             challenge_entities.ChallengeType
	Title            string
	Description      string
	Priority         challenge_entities.ChallengePriority
}

// AddEvidenceCommand represents a request to add evidence to a challenge
type AddEvidenceCommand struct {
	ChallengeID uuid.UUID
	Type        string // screenshot, replay_clip, log, video
	URL         string
	Description string
	StartTick   *int64
	EndTick     *int64
}

// VoteOnChallengeCommand represents a vote on a challenge
type VoteOnChallengeCommand struct {
	ChallengeID uuid.UUID
	PlayerID    uuid.UUID
	TeamID      *uuid.UUID
	VoteType    string // approve, reject, abstain
	Reason      string
}

// ReviewChallengeCommand represents an admin starting a review
type ReviewChallengeCommand struct {
	ChallengeID uuid.UUID
	AdminID     uuid.UUID
}

// StartVotingCommand represents a request to start voting on a challenge
type StartVotingCommand struct {
	ChallengeID    uuid.UUID
	VotingDuration time.Duration
}

// ResolveChallengeCommand represents a request to resolve a challenge
type ResolveChallengeCommand struct {
	ChallengeID uuid.UUID
	AdminID     uuid.UUID
	Decision    string // approve, reject
	Resolution  challenge_entities.ChallengeResolution
	Notes       string
}

// CancelChallengeCommand represents a request to cancel a challenge
type CancelChallengeCommand struct {
	ChallengeID uuid.UUID
	CancellerID uuid.UUID
	Reason      string
}

// PauseMatchCommand represents a request to pause a match for a challenge
type PauseMatchCommand struct {
	ChallengeID uuid.UUID
	CurrentScore map[string]int
}

// ResumeMatchCommand represents a request to resume a match after a challenge
type ResumeMatchCommand struct {
	ChallengeID uuid.UUID
	NewScore    map[string]int
}

// CreateChallengeCommandHandler handles challenge creation
type CreateChallengeCommandHandler interface {
	Exec(ctx context.Context, cmd CreateChallengeCommand) (*challenge_entities.Challenge, error)
}

// AddEvidenceCommandHandler handles adding evidence to challenges
type AddEvidenceCommandHandler interface {
	Exec(ctx context.Context, cmd AddEvidenceCommand) (*challenge_entities.Challenge, error)
}

// VoteOnChallengeCommandHandler handles voting on challenges
type VoteOnChallengeCommandHandler interface {
	Exec(ctx context.Context, cmd VoteOnChallengeCommand) (*challenge_entities.Challenge, error)
}

// ReviewChallengeCommandHandler handles admin review start
type ReviewChallengeCommandHandler interface {
	Exec(ctx context.Context, cmd ReviewChallengeCommand) (*challenge_entities.Challenge, error)
}

// StartVotingCommandHandler handles starting the voting process
type StartVotingCommandHandler interface {
	Exec(ctx context.Context, cmd StartVotingCommand) (*challenge_entities.Challenge, error)
}

// ResolveChallengeCommandHandler handles challenge resolution
type ResolveChallengeCommandHandler interface {
	Exec(ctx context.Context, cmd ResolveChallengeCommand) (*challenge_entities.Challenge, error)
}

// CancelChallengeCommandHandler handles challenge cancellation
type CancelChallengeCommandHandler interface {
	Exec(ctx context.Context, cmd CancelChallengeCommand) (*challenge_entities.Challenge, error)
}

// PauseMatchCommandHandler handles match pause for challenges
type PauseMatchCommandHandler interface {
	Exec(ctx context.Context, cmd PauseMatchCommand) (*challenge_entities.Challenge, error)
}

// ResumeMatchCommandHandler handles match resume after challenges
type ResumeMatchCommandHandler interface {
	Exec(ctx context.Context, cmd ResumeMatchCommand) (*challenge_entities.Challenge, error)
}

