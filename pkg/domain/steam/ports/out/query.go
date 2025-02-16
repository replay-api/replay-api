package steam_out

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	e "github.com/replay-api/replay-api/pkg/domain/steam/entities"
)

type SteamUserReader interface {
	Search(ctx context.Context, s common.Search) ([]e.SteamUser, error)
}
