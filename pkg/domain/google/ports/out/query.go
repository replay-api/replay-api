package google_out

import (
	"context"

	shared "github.com/resource-ownership/go-common/pkg/common"
	google_entity "github.com/replay-api/replay-api/pkg/domain/google/entities"
)

type GoogleUserReader interface {
	Search(ctx context.Context, s shared.Search) ([]google_entity.GoogleUser, error)
}
