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

// AddEvidenceUseCase implements adding evidence to challenges
type AddEvidenceUseCase struct {
	challengeRepo challenge_out.ChallengeRepository
}

// NewAddEvidenceUseCase creates a new instance
func NewAddEvidenceUseCase(challengeRepo challenge_out.ChallengeRepository) *AddEvidenceUseCase {
	return &AddEvidenceUseCase{
		challengeRepo: challengeRepo,
	}
}

// Exec executes the add evidence use case
func (uc *AddEvidenceUseCase) Exec(ctx context.Context, cmd challenge_in.AddEvidenceCommand) (*challenge_entities.Challenge, error) {
	// 1. Validate authentication
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, common.NewErrUnauthorized()
	}

	// 2. Validate command
	if cmd.ChallengeID == uuid.Nil {
		return nil, fmt.Errorf("challenge ID is required")
	}
	if cmd.Type == "" {
		return nil, fmt.Errorf("evidence type is required")
	}
	if cmd.URL == "" {
		return nil, fmt.Errorf("evidence URL is required")
	}

	// Validate evidence type
	validTypes := []string{"screenshot", "replay_clip", "log", "video"}
	isValidType := false
	for _, vt := range validTypes {
		if cmd.Type == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return nil, fmt.Errorf("evidence type must be one of: screenshot, replay_clip, log, video")
	}

	// 3. Fetch the challenge
	challenge, err := uc.challengeRepo.GetByID(ctx, cmd.ChallengeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch challenge: %w", err)
	}
	if challenge == nil {
		return nil, common.NewErrNotFound(common.ResourceTypeChallenge, "id", cmd.ChallengeID)
	}

	// 4. Verify the user is the challenger (only challenger can add evidence)
	if challenge.ChallengerID != resourceOwner.UserID {
		return nil, common.NewErrForbidden()
	}

	// 5. Build tick range if provided
	var tickRange *challenge_entities.TickRange
	if cmd.StartTick != nil && cmd.EndTick != nil {
		tickRange = &challenge_entities.TickRange{
			StartTick: *cmd.StartTick,
			EndTick:   *cmd.EndTick,
		}
	}

	// 6. Add the evidence
	if err := challenge.AddEvidence(cmd.Type, cmd.URL, cmd.Description, tickRange); err != nil {
		return nil, fmt.Errorf("failed to add evidence: %w", err)
	}

	// 7. Persist changes
	if err := uc.challengeRepo.Save(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to save challenge: %w", err)
	}

	return challenge, nil
}

