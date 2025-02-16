package handlers

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	infocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/replay-api/replay-api/pkg/app/cs/builders"
	"github.com/replay-api/replay-api/pkg/app/cs/state"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

func ClutchStart(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e evt.Kill) {
	return func(event evt.Kill) {
		ct := make([]infocs.Player, 0)
		t := make([]infocs.Player, 0)

		gs := p.GameState()

		if gs == nil {
			msg := "Game state is nil"
			slog.Debug(msg)
			panic(msg)
		}

		roundIndex := gs.TotalRoundsPlayed()

		matchContext = matchContext.WithRound(roundIndex, gs)

		if matchContext.InClutch(roundIndex) {
			// msg := fmt.Sprintf("Already in clutch situation, round index: %d", roundIndex)
			// slog.Debug(msg)
			return
		}

		for _, player := range p.GameState().Participants().Playing() {
			if !player.IsAlive() {
				continue
			}

			if player.Team == infocs.TeamCounterTerrorists {
				ct = append(ct, *player)
			} else {
				t = append(t, *player)
			}
		}

		var playerInClutch *infocs.Player
		var opponents []infocs.Player

		if len(ct) == 1 {
			playerInClutch = &ct[0]
			opponents = t
		}

		if len(t) == 1 {
			playerInClutch = &t[0]
			opponents = ct
		}

		if playerInClutch == nil {
			msg := fmt.Sprintf("Not in clutch situation, ct players: %d, t players: %d", len(ct), len(t))
			slog.Debug(msg)
			// panic(msg)
			return
		}

		matchContext = matchContext.WithClutch(roundIndex, playerInClutch, opponents)

		b := builders.NewCSMatchStatsBuilder(p, matchContext).WithRoundsStats(matchContext.RoundContexts)

		out <- &replay_entity.GameEvent{
			ID:            uuid.New(),
			MatchID:       matchContext.MatchID,
			Type:          common.Event_ClutchStartID,
			Payload:       b.Build(),
			GameTime:      p.CurrentTime(),
			ResourceOwner: matchContext.ResourceOwner, // TODO: remover daqui ou do matchContext, esta redundante
		}
	}
}
