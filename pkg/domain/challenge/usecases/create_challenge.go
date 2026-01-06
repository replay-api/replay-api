package challenge_usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
	challenge_in "github.com/replay-api/replay-api/pkg/domain/challenge/ports/in"
	challenge_out "github.com/replay-api/replay-api/pkg/domain/challenge/ports/out"
)

// CreateChallengeUseCase implements challenge creation logic
type CreateChallengeUseCase struct {
	challengeRepo challenge_out.ChallengeRepository
}

// NewCreateChallengeUseCase creates a new instance
func NewCreateChallengeUseCase(challengeRepo challenge_out.ChallengeRepository) *CreateChallengeUseCase {
	return &CreateChallengeUseCase{
		challengeRepo: challengeRepo,
	}
}

// Exec executes the create challenge use case
func (uc *CreateChallengeUseCase) Exec(ctx context.Context, cmd challenge_in.CreateChallengeCommand) (*challenge_entities.Challenge, error) {
	// 1. Validate authentication
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, shared.NewErrUnauthorized()
	}

	// 2. Validate the challenger owns the request
	if cmd.ChallengerID != resourceOwner.UserID {
		return nil, shared.NewErrForbidden()
	}

	// 3. Validate command
	if cmd.MatchID == uuid.Nil {
		return nil, fmt.Errorf("match ID is required")
	}
	if cmd.GameID == "" {
		return nil, fmt.Errorf("game ID is required")
	}
	if cmd.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if cmd.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	// 4. Check for duplicate pending challenges from same player on same match
	existingChallenges, err := uc.challengeRepo.GetByMatchID(ctx, cmd.MatchID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing challenges: %w", err)
	}

	for _, existing := range existingChallenges {
		if existing.ChallengerID == cmd.ChallengerID &&
			(existing.Status == challenge_entities.ChallengeStatusPending ||
				existing.Status == challenge_entities.ChallengeStatusVotePending ||
				existing.Status == challenge_entities.ChallengeStatusInReview) {
			return nil, fmt.Errorf("you already have a pending challenge for this match")
		}
	}

	// 5. Create the challenge
	challenge, err := challenge_entities.NewChallenge(
		resourceOwner,
		cmd.MatchID,
		cmd.ChallengerID,
		cmd.GameID,
		cmd.Type,
		cmd.Title,
		cmd.Description,
		cmd.Priority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create challenge: %w", err)
	}

	// 6. Set optional fields
	if cmd.RoundNumber != nil {
		challenge.RoundNumber = cmd.RoundNumber
	}
	if cmd.ChallengerTeamID != nil {
		challenge.ChallengerTeamID = cmd.ChallengerTeamID
	}
	if cmd.LobbyID != nil {
		challenge.LobbyID = cmd.LobbyID
	}
	if cmd.TournamentID != nil {
		challenge.TournamentID = cmd.TournamentID
	}

	// 7. Validate the challenge
	if err := challenge.Validate(); err != nil {
		return nil, fmt.Errorf("challenge validation failed: %w", err)
	}

	// 8. Persist the challenge
	if err := uc.challengeRepo.Save(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to save challenge: %w", err)
	}

	return challenge, nil
}

