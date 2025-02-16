package handlers

import (
	"log/slog"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	"github.com/replay-api/replay-api/pkg/app/cs/state"
	common "github.com/replay-api/replay-api/pkg/domain"
	"github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

func GenericGameEvent(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.GenericGameEvent) {
	return func(event evt.GenericGameEvent) {
		// slog.Info("GenericGameEvent: %v", event.Name, event.Data)

		gs := p.GameState()

		roundIndex := gs.TotalRoundsPlayed()

		matchContext = matchContext.WithRound(roundIndex, gs)
		currentTick := common.TickIDType(gs.IngameTick())

		gameEvent, err := event_factory.NewGameEvent(
			common.Event_GenericGameEventID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			event,
		)

		if err != nil {
			slog.Error("unable to create generic game event")
			return
		}

		// out <- *gameEvent
		slog.Warn("skipping generic event", "gameEvent.Type", gameEvent.Type)
	}
}
