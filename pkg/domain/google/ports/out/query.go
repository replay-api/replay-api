package google_out

import (
	"context"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	google_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/google/entities"
)

type GoogleUserReader interface {
	Search(ctx context.Context, s common.Search) ([]google_entity.GoogleUser, error)
}
