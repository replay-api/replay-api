package query_controllers

import (
	"github.com/golobby/container/v3"
	controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
)

type MatchQueryController struct {
	controllers.DefaultSearchController[replay_entity.Match]
}

func NewMatchQueryController(c container.Container) *MatchQueryController {
	var queryService replay_in.MatchReader

	err := c.Resolve(&queryService)

	if err != nil {
		panic(err)
	}

	baseController := controllers.NewDefaultSearchController(queryService)

	return &MatchQueryController{*baseController}
}
