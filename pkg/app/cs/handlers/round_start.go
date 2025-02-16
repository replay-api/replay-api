package handlers

import (
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

func RoundStart(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e evt.RoundStart) {
	return func(event evt.RoundStart) {
		gs := p.GameState()

		roundIndex := gs.TotalRoundsPlayed()

		matchContext = matchContext.WithRound(roundIndex, gs)

		roundContext := matchContext.RoundContexts[roundIndex]

		if roundContext.RoundNumber == 1 || roundContext.RoundNumber == 16 {
			roundContext.SetRoundType(state.CSRoundTypePistol)
		}

		// TODO:
		// matchContext.RoundContexts[roundIndex] = roundContext
	}
}
