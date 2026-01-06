package handlers

import (
	"github.com/google/uuid"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/replay-api/replay-api/pkg/app/cs/builders"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

func EconomyEvent(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e evt.RoundFreezetimeEnd) { // TODO: criar em RoundStart e RoundEnd para track do que foi comprado dps do freezetime
	return func(event evt.RoundFreezetimeEnd) {
		gs := p.GameState()

		roundIndex := gs.TotalRoundsPlayed()

		matchContext = matchContext.WithRound(roundIndex, gs)

		currentTick := replay_common.TickIDType(gs.IngameTick())

		// roundContext := matchContext.RoundContexts[roundIndex]

		b := builders.NewCSMatchStatsBuilder(p, matchContext).WithRoundsStats(matchContext.RoundContexts)

		out <- &replay_entity.GameEvent{
			ID:            uuid.New(),
			MatchID:       matchContext.MatchID,
			Type:          fps_events.Event_Economy,
			TickID:        currentTick,
			GameTime:      p.CurrentTime(),
			Payload:       b.Build(),
			ResourceOwner: matchContext.ResourceOwner,
		}

		// // tEconomy :=
		// ctEconomy := state.CS2TeamEconomyContext{}

		// state := cs_entity.CSEconomyStats{}
	}
}
