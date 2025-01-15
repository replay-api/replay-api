package query_controllers

import (
	"github.com/golobby/container/v3"
	controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers"
	steam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
	steam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/in"
)

type SteamUserQueryController struct {
	controllers.DefaultSearchController[steam_entities.SteamUser]
}

func NewSteamUserQueryController(c container.Container) *SteamUserQueryController {
	var queryService steam_in.SteamUserReader

	err := c.Resolve(&queryService)

	if err != nil {
		panic(err)
	}

	baseController := controllers.NewDefaultSearchController(queryService)

	return &SteamUserQueryController{*baseController}
}
