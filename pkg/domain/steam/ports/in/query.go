package steam_in

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	steam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
)

type SteamUserReader interface {
	common.Searchable[steam_entities.SteamUser]
}
