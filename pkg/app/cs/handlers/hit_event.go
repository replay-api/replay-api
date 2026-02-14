package handlers

import (
	"log/slog"
	"strconv"

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

func HitEvent(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.PlayerHurt) {
	return func(event evt.PlayerHurt) {
		// slog.Info(fmt.Sprintf("%s event", fps_events.Event_HitID), "event", event)

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

		// Extract attacker info for stats tracking
		attackerSteamID := ""
		attackerName := ""
		attackerTeam := ""
		if event.Attacker != nil {
			attackerSteamID = strconv.FormatUint(event.Attacker.SteamID64, 10)
			attackerName = event.Attacker.Name
			if event.Attacker.Team == 3 {
				attackerTeam = "CT"
			} else if event.Attacker.Team == 2 {
				attackerTeam = "T"
			}
		}
		
		// Extract victim info
		victimSteamID := ""
		victimTeam := ""
		if event.Player != nil {
			victimSteamID = strconv.FormatUint(event.Player.SteamID64, 10)
			if event.Player.Team == 3 {
				victimTeam = "CT"
			} else if event.Player.Team == 2 {
				victimTeam = "T"
			}
		}
		
		// Track damage in accumulator
		if matchContext.StatsAccumulator != nil && attackerSteamID != "" {
			damage := event.HealthDamage + event.ArmorDamage
			weaponName := ""
			if event.Weapon != nil {
				weaponName = event.Weapon.String()
			}
			
			// Determine hitbox from hit group
			hitbox := ""
			switch event.HitGroup {
			case 1:
				hitbox = "head"
			case 2, 3, 6:
				hitbox = "body" // chest, stomach, neck
			case 4, 5:
				hitbox = "arms"
			case 7, 8:
				hitbox = "legs"
			}
			
			isSelfDamage := attackerSteamID == victimSteamID
			matchContext.StatsAccumulator.RecordDamage(
				attackerSteamID, attackerName, attackerTeam,
				victimSteamID, victimTeam,
				damage, weaponName, hitbox, isSelfDamage,
			)
			
			// Record hit for accuracy
			matchContext.StatsAccumulator.RecordWeaponHit(attackerSteamID, attackerName, attackerTeam)
		}

		payload := cs_entity.CSHitStats{
			// SourcePlayerID: sourcePlayerID,
			// TODO: ticket + spec (angles data, values etc)
			Damage:   event.HealthDamage + event.ArmorDamage, // REVIEW
			Location: cs_entity.CSHitBoxType(event.HitGroup),
		}

		battleContext.Hits[currentTick] = payload

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_HitID,
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
