package google_out

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	google_entity "github.com/replay-api/replay-api/pkg/domain/google/entities"
)

type GoogleUserReader interface {
	Search(ctx context.Context, s common.Search) ([]google_entity.GoogleUser, error)
}
