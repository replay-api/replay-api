package steam_out

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	e "github.com/replay-api/replay-api/pkg/domain/steam/entities"
)

type SteamUserReader interface {
	shared.Searchable[e.SteamUser]
}
