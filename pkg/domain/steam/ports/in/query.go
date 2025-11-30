package steam_in

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	steam_entities "github.com/replay-api/replay-api/pkg/domain/steam/entities"
)

type SteamUserReader interface {
	common.Searchable[steam_entities.SteamUser]
}
