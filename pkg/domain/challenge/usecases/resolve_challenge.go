package challenge_usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
	challenge_in "github.com/replay-api/replay-api/pkg/domain/challenge/ports/in"
	challenge_out "github.com/replay-api/replay-api/pkg/domain/challenge/ports/out"
)

// ResolveChallengeUseCase implements challenge resolution logic
type ResolveChallengeUseCase struct {
	challengeRepo challenge_out.ChallengeRepository
}

// NewResolveChallengeUseCase creates a new instance
func NewResolveChallengeUseCase(challengeRepo challenge_out.ChallengeRepository) *ResolveChallengeUseCase {
	return &ResolveChallengeUseCase{
		challengeRepo: challengeRepo,
	}
}

// Exec executes the resolve challenge use case
func (uc *ResolveChallengeUseCase) Exec(ctx context.Context, cmd challenge_in.ResolveChallengeCommand) (*challenge_entities.Challenge, error) {
	// 1. Validate authentication
	resourceOwner := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, shared.NewErrUnauthorized()
	}

	// 2. Validate command
	if cmd.ChallengeID == uuid.Nil {
		return nil, fmt.Errorf("challenge ID is required")
	}
	if cmd.AdminID == uuid.Nil {
		return nil, fmt.Errorf("admin ID is required")
	}
	if cmd.Decision != "approve" && cmd.Decision != "reject" {
		return nil, fmt.Errorf("decision must be 'approve' or 'reject'")
	}

	// 3. Fetch the challenge
	challenge, err := uc.challengeRepo.GetByID(ctx, cmd.ChallengeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch challenge: %w", err)
	}
	if challenge == nil {
		return nil, shared.NewErrNotFound(replay_common.ResourceTypeChallenge, "id", cmd.ChallengeID)
	}

	// 4. Verify admin can review this challenge
	if !challenge.CanBeReviewedBy(cmd.AdminID) && challenge.Status != challenge_entities.ChallengeStatusInReview {
		return nil, fmt.Errorf("challenge cannot be resolved in current state")
	}

	// 5. Apply decision
	switch cmd.Decision {
	case "approve":
		if !cmd.Resolution.IsValid() || cmd.Resolution == challenge_entities.ChallengeResolutionNone {
			return nil, fmt.Errorf("valid resolution is required for approval")
		}
		if err := challenge.Approve(cmd.AdminID, cmd.Resolution, cmd.Notes); err != nil {
			return nil, fmt.Errorf("failed to approve challenge: %w", err)
		}
	case "reject":
		if err := challenge.Reject(cmd.AdminID, cmd.Notes); err != nil {
			return nil, fmt.Errorf("failed to reject challenge: %w", err)
		}
	}

	// 6. Persist changes
	if err := uc.challengeRepo.Save(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to save challenge: %w", err)
	}

	return challenge, nil
}

