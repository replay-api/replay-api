package steam_in

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	steam_entities "github.com/replay-api/replay-api/pkg/domain/steam/entities"
)

type SteamUserReader interface {
	shared.Searchable[steam_entities.SteamUser]
}
