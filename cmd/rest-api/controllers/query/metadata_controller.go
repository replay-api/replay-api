package query_controllers

import (
	"github.com/golobby/container/v3"

	controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
)

type ReplayMetadataQueryController struct {
	controllers.DefaultSearchController[replay_entity.ReplayFile]
}

func NewReplayMetadataQueryController(container container.Container) *ReplayMetadataQueryController {
	var replayFileReader replay_in.ReplayFileReader

	err := container.Resolve(&replayFileReader)
	if err != nil {
		panic(err)
	}

	baseController := controllers.NewDefaultSearchController(replayFileReader)

	return &ReplayMetadataQueryController{
		*baseController,
	}
}
