package challenge_usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
	challenge_in "github.com/replay-api/replay-api/pkg/domain/challenge/ports/in"
	challenge_out "github.com/replay-api/replay-api/pkg/domain/challenge/ports/out"
)

// VoteOnChallengeUseCase implements voting on challenges
type VoteOnChallengeUseCase struct {
	challengeRepo challenge_out.ChallengeRepository
}

// NewVoteOnChallengeUseCase creates a new instance
func NewVoteOnChallengeUseCase(challengeRepo challenge_out.ChallengeRepository) *VoteOnChallengeUseCase {
	return &VoteOnChallengeUseCase{
		challengeRepo: challengeRepo,
	}
}

// Exec executes the vote on challenge use case
func (uc *VoteOnChallengeUseCase) Exec(ctx context.Context, cmd challenge_in.VoteOnChallengeCommand) (*challenge_entities.Challenge, error) {
	// 1. Validate authentication
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, common.NewErrUnauthorized()
	}

	// 2. Validate the voter is the authenticated user
	if cmd.PlayerID != resourceOwner.UserID {
		return nil, common.NewErrForbidden()
	}

	// 3. Validate command
	if cmd.ChallengeID == uuid.Nil {
		return nil, fmt.Errorf("challenge ID is required")
	}
	if cmd.VoteType != "approve" && cmd.VoteType != "reject" && cmd.VoteType != "abstain" {
		return nil, fmt.Errorf("vote type must be 'approve', 'reject', or 'abstain'")
	}

	// 4. Fetch the challenge
	challenge, err := uc.challengeRepo.GetByID(ctx, cmd.ChallengeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch challenge: %w", err)
	}
	if challenge == nil {
		return nil, common.NewErrNotFound(common.ResourceTypeChallenge, "id", cmd.ChallengeID)
	}

	// 5. Voter cannot be the challenger
	if cmd.PlayerID == challenge.ChallengerID {
		return nil, fmt.Errorf("challenger cannot vote on their own challenge")
	}

	// 6. Check voting deadline
	if challenge.IsVotingExpired() {
		return nil, fmt.Errorf("voting period has expired")
	}

	// 7. Add the vote
	if err := challenge.AddVote(cmd.PlayerID, cmd.TeamID, cmd.VoteType, cmd.Reason); err != nil {
		return nil, fmt.Errorf("failed to add vote: %w", err)
	}

	// 8. Persist changes
	if err := uc.challengeRepo.Save(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to save challenge: %w", err)
	}

	return challenge, nil
}

