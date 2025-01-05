package google_in

import (
	"context"

	google_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/google/entities"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type OnboardGoogleUserCommand interface {
	Exec(ctx context.Context, googleUser *google_entities.GoogleUser) (*google_entities.GoogleUser, *iam_entities.RIDToken, error)
	Validate(ctx context.Context, googleUser *google_entities.GoogleUser) error
}
