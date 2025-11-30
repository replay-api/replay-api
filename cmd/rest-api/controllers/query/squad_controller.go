package query_controllers

import (
	"github.com/golobby/container/v3"
	controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

type SquadQueryController struct {
	controllers.DefaultSearchController[replay_entity.Squad]
}

func NewSquadQueryController(c container.Container) *SquadQueryController {
	var queryService squad_in.SquadReader

	err := c.Resolve(&queryService)

	if err != nil {
		panic(err)
	}

	baseController := controllers.NewDefaultSearchController(queryService)

	return &SquadQueryController{*baseController}
}
