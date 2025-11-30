package handlers

import (
	"fmt"
	"log/slog"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/replay-api/replay-api/pkg/app/cs/builders"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

func RoundEnd(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e infocs.RoundEnd) {
	return func(event infocs.RoundEnd) {
		// slog.Info("RoundEnd event: %v", "event", event)

		gs := p.GameState()

		if gs == nil {
			msg := "Game state is nil"
			slog.Debug(msg)
			panic(msg)
		}

		roundIndex := gs.TotalRoundsPlayed()

		matchContext = matchContext.WithRound(roundIndex, gs)

		b := builders.NewCSMatchStatsBuilder(p, matchContext).WithRoundsStats(matchContext.RoundContexts)

		payload := b.Build()

		gameEvent, err := event_factory.NewGameEvent(
			common.Event_RoundEndID,
			matchContext,
			roundIndex,
			common.TickIDType(gs.IngameTick()),
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error(fmt.Sprintf("RoundEnd: unable to create game event due to %s", err.Error()), "err", err)
			return
		}

		out <- gameEvent
	}
}
