package steam_out

import (
	"context"

	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"
)

type SteamUserWriter interface {
	Create(ctx context.Context, user *steam_entity.SteamUser) (*steam_entity.SteamUser, error)
}

type VHashWriter interface {
	CreateVHash(ctx context.Context, steamID string) string
}
