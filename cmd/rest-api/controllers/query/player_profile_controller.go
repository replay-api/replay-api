package query_controllers

import (
	"github.com/golobby/container/v3"
	controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

type PlayerProfileQueryController struct {
	controllers.DefaultSearchController[squad_entities.PlayerProfile]
}

func NewPlayerProfileQueryController(c container.Container) *PlayerProfileQueryController {
	var queryService squad_in.PlayerProfileReader

	err := c.Resolve(&queryService)

	if err != nil {
		panic(err)
	}

	baseController := controllers.NewDefaultSearchController(queryService)

	return &PlayerProfileQueryController{*baseController}
}
