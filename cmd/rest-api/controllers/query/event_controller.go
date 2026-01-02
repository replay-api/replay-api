package query_controllers

import (
	"log/slog"

	"github.com/golobby/container/v3"
	controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
)

type EventQueryController struct {
	controllers.DefaultSearchController[replay_entity.GameEvent]
}

func NewEventQueryController(c container.Container) *EventQueryController {
	var queryService replay_in.EventReader

	err := c.Resolve(&queryService)

	if err != nil {
		slog.Warn("EventReader not available - event queries will be disabled", "error", err)
	}

	baseController := controllers.NewDefaultSearchController(queryService)

	return &EventQueryController{*baseController}
}
