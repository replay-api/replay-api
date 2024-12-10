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
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

func ClutchProgress(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e evt.Kill) {
	return func(event evt.Kill) {
		roundIndex := p.GameState().TotalRoundsPlayed()

		if !matchContext.InClutch(roundIndex) {
			return
		}

		if event.Killer == nil {
			return
		}

		playerInClutch := *matchContext.GetClutchPlayer(roundIndex)

		fragger := *event.Killer

		isClutchPlayer := fragger.SteamID64 == playerInClutch.SteamID64
		isNotSelf := fragger.SteamID64 != (*event.Victim).SteamID64
		isNotFriendlyFire := event.Killer.GetTeam() != event.Victim.GetTeam()

		isProgress := isClutchPlayer && isNotSelf && isNotFriendlyFire
		if !isProgress {
			return
		}

		var opponents []csinfo.Player

		for _, player := range p.GameState().Participants().Playing() {
			if !player.IsAlive() {
				continue
			}

			if player.Team != playerInClutch.Team {
				opponents = append(opponents, *player)
			}
		}

		matchContext = matchContext.UpdateClutchState(roundIndex, csDomain.ClutchProgressKey, opponents)

		b := builders.NewCSMatchStatsBuilder(p, matchContext).WithRoundsStats(matchContext.RoundContexts)

		// slog.Info("ClutchProgress event: %v", "event", event)

		out <- &replay_entity.GameEvent{
			ID:            uuid.New(),
			MatchID:       matchContext.MatchID,
			Type:          common.Event_ClutchProgressID,
			Payload:       b.Build(),
			GameTime:      p.CurrentTime(),
			ResourceOwner: matchContext.ResourceOwner, // TODO: remover daqui ou do matchContext, esta redundante
		}
	}
}
