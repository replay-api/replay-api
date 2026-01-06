package handlers

import (
	"log/slog"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	"github.com/replay-api/replay-api/pkg/domain/replay/entities"

	// replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"

	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	cs_entity "github.com/replay-api/replay-api/pkg/domain/cs/entities"
)

func WeaponFire(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.WeaponFire) {
	return func(event evt.WeaponFire) {
		// slog.Info(fmt.Sprintf("%s event", fps_events.Event_WeaponFireID), "event", event)

		gs := p.GameState()

		if gs == nil {
			msg := "Game state is nil"
			slog.Debug(msg)

			panic(msg)
		}

		roundIndex := gs.TotalRoundsPlayed()

		matchContext := matchContext.WithRound(roundIndex, gs)

		battleContext := matchContext.RoundContexts[roundIndex].BattleContext

		currentTick := replay_common.TickIDType(gs.IngameTick())

		// sourcePlayerID := fmt.Sprintf("%d", event.Shooter.SteamID64) // TODO: ticket + spec (angles data, values etc)

		payload := cs_entity.CSHitStats{
			// SourcePlayerID: sourcePlayerID,
			// TODO: ticket + spec (angles data, values etc)
			// Damage: event.Shooter.FlashTick

		}

		battleContext.Hits[currentTick] = payload

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_WeaponFireID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create weapon_fire event", "err", err)
			return
		}

		out <- gameEvent
	}
}
