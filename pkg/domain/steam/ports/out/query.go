package steam_out

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	e "github.com/replay-api/replay-api/pkg/domain/steam/entities"
)

type SteamUserReader interface {
	common.Searchable[e.SteamUser]
}
