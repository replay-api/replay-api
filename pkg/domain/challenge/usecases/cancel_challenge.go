package challenge_usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
	challenge_in "github.com/replay-api/replay-api/pkg/domain/challenge/ports/in"
	challenge_out "github.com/replay-api/replay-api/pkg/domain/challenge/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// CancelChallengeUseCase implements challenge cancellation logic
type CancelChallengeUseCase struct {
	challengeRepo challenge_out.ChallengeRepository
}

// NewCancelChallengeUseCase creates a new instance
func NewCancelChallengeUseCase(challengeRepo challenge_out.ChallengeRepository) *CancelChallengeUseCase {
	return &CancelChallengeUseCase{
		challengeRepo: challengeRepo,
	}
}

// Exec executes the cancel challenge use case
func (uc *CancelChallengeUseCase) Exec(ctx context.Context, cmd challenge_in.CancelChallengeCommand) (*challenge_entities.Challenge, error) {
	// 1. Validate authentication
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, shared.NewErrUnauthorized()
	}

	// 2. Validate command
	if cmd.ChallengeID == uuid.Nil {
		return nil, fmt.Errorf("challenge ID is required")
	}

	// 3. Fetch the challenge
	challenge, err := uc.challengeRepo.GetByID(ctx, cmd.ChallengeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch challenge: %w", err)
	}
	if challenge == nil {
		return nil, shared.NewErrNotFound(replay_common.ResourceTypeChallenge, "id", cmd.ChallengeID)
	}

	// 4. Verify the user is the challenger or an admin
	if challenge.ChallengerID != resourceOwner.UserID && cmd.CancellerID != resourceOwner.UserID {
		return nil, shared.NewErrForbidden()
	}

	// Only the challenger can cancel their own challenge (or admin can cancel any)
	// For now, we allow the challenger to cancel
	if challenge.ChallengerID != resourceOwner.UserID {
		// Check if the canceller is an admin (TODO: implement admin role check)
		// For now, we'll just check if the user is the original challenger
		return nil, shared.NewErrForbidden()
	}

	// 5. Cancel the challenge
	reason := cmd.Reason
	if reason == "" {
		reason = "Cancelled by challenger"
	}
	if err := challenge.Cancel(reason); err != nil {
		return nil, fmt.Errorf("failed to cancel challenge: %w", err)
	}

	// 6. Persist changes
	if err := uc.challengeRepo.Save(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to save challenge: %w", err)
	}

	return challenge, nil
}

