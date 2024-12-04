package handlers

import (
	"github.com/google/uuid"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	csinfo "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/psavelis/team-pro/replay-api/pkg/app/cs/builders"
	"github.com/psavelis/team-pro/replay-api/pkg/app/cs/state"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	csDomain "github.com/psavelis/team-pro/replay-api/pkg/domain/cs/entities"
	e "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

func ClutchEnd(p dem.Parser, matchContext *state.CS2MatchContext, out chan e.GameEvent) func(e evt.RoundEnd) error {
	return func(event evt.RoundEnd) error {
		gs := p.GameState()
		roundIndex := gs.TotalRoundsPlayed() - 1 // last round

		matchContext = matchContext.WithRound(roundIndex, gs)

		if !matchContext.InClutch(roundIndex) {
			return nil
		}

		playerInClutch := *matchContext.GetClutchPlayer(roundIndex)

		isWinner := event.Winner == playerInClutch.Team

		var opponentsState []*csinfo.Player
		var result csDomain.ClutchSituationStatusKey

		if isWinner {
			result = csDomain.ClutchWonKey
			opponentsState = event.LoserState.Members()
		} else {
			result = csDomain.ClutchLostKey
			opponentsState = event.WinnerState.Members()
		}

		remainingOpponents := make([]csinfo.Player, len(opponentsState))
		for k, player := range opponentsState {
			remainingOpponents[k] = *player
		}

		matchContext = matchContext.UpdateClutchState(roundIndex, result, remainingOpponents)

		b := builders.NewCSMatchStatsBuilder(p, matchContext).WithRoundsStats(matchContext.RoundContexts)

		out <- e.GameEvent{
			ID:            uuid.New(),
			MatchID:       matchContext.MatchID,
			Type:          common.Event_ClutchEndID,
			Payload:       b.Build(),
			GameTime:      p.CurrentTime(),
			ResourceOwner: matchContext.ResourceOwner, // TODO: remover daqui ou do matchContext, esta redundante
		}

		return nil
	}
}
