package google_out

import (
	"context"

	google_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/google/entities"
)

type GoogleUserWriter interface {
	Create(ctx context.Context, user *google_entity.GoogleUser) (*google_entity.GoogleUser, error)
}

type VHashWriter interface {
	CreateVHash(ctx context.Context, steamID string) string
}
