package handlers

import (
	"log/slog"
	"math"
	"strconv"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	"github.com/replay-api/replay-api/pkg/app/cs/state"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	"github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

// Position3D represents a 3D coordinate in the game world
type Position3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type KillPayload struct {
	KillerName      string      `json:"killer_name"`
	KillerSteamID   string      `json:"killer_steam_id,omitempty"`
	KillerPosition  *Position3D `json:"killer_position,omitempty"`
	VictimName      string      `json:"victim_name"`
	VictimSteamID   string      `json:"victim_steam_id,omitempty"`
	VictimPosition  *Position3D `json:"victim_position,omitempty"`
	Weapon          string      `json:"weapon"`
	Headshot        bool        `json:"headshot"`
	IsOpeningKill   bool        `json:"is_opening_kill,omitempty"`
	IsTradeKill     bool        `json:"is_trade_kill,omitempty"`
	IsWallbang      bool        `json:"is_wallbang,omitempty"`
	IsNoScope       bool        `json:"is_no_scope,omitempty"`
	IsThroughSmoke  bool        `json:"is_through_smoke,omitempty"`
	IsAttackerBlind bool        `json:"is_attacker_blind,omitempty"`
	AssisterName    string      `json:"assister_name,omitempty"`
	AssisterSteamID string      `json:"assister_steam_id,omitempty"`
	FlashAssister   string      `json:"flash_assister,omitempty"`
	Distance        float64     `json:"distance,omitempty"`
}

func Kill(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.Kill) {
	return func(event evt.Kill) {
		gs := p.GameState()

		if gs == nil {
			msg := "Game state is nil"
			slog.Debug(msg)
			panic(msg)
		}

		roundIndex := gs.TotalRoundsPlayed()

		matchContext = matchContext.WithRound(roundIndex, gs)
		currentTick := replay_common.TickIDType(gs.IngameTick())

		// Extract killer info
		killerName := ""
		killerSteamID := ""
		killerTeam := ""
		if event.Killer != nil {
			killerName = event.Killer.Name
			killerSteamID = strconv.FormatUint(event.Killer.SteamID64, 10)
			if event.Killer.Team == 3 {
				killerTeam = "CT"
			} else if event.Killer.Team == 2 {
				killerTeam = "T"
			}
		}

		// Extract victim info
		victimName := ""
		victimSteamID := ""
		victimTeam := ""
		if event.Victim != nil {
			victimName = event.Victim.Name
			victimSteamID = strconv.FormatUint(event.Victim.SteamID64, 10)
			if event.Victim.Team == 3 {
				victimTeam = "CT"
			} else if event.Victim.Team == 2 {
				victimTeam = "T"
			}
		}

		// Extract assister info
		assisterName := ""
		assisterSteamID := ""
		assisterTeam := ""
		if event.Assister != nil {
			assisterName = event.Assister.Name
			assisterSteamID = strconv.FormatUint(event.Assister.SteamID64, 10)
			if event.Assister.Team == 3 {
				assisterTeam = "CT"
			} else if event.Assister.Team == 2 {
				assisterTeam = "T"
			}
		}

		// Track stats in accumulator
		if matchContext.StatsAccumulator != nil && killerSteamID != "" && victimSteamID != "" {
			// Set current round for tracking
			matchContext.StatsAccumulator.SetCurrentRound(roundIndex + 1) // rounds are 1-indexed for display
			
			// Record the kill (handles headshots, opening kills, trade kills)
			matchContext.StatsAccumulator.RecordKill(
				killerSteamID, killerName, killerTeam,
				victimSteamID, victimName, victimTeam,
				event.IsHeadshot,
				p.CurrentTime(),
				roundIndex+1,
			)
			
			// Record weapon-specific kill stats
			weaponName := event.Weapon.String()
			matchContext.StatsAccumulator.RecordWeaponKill(killerSteamID, weaponName, event.IsHeadshot)
			
			// Record special kill types
			isWallbang := event.PenetratedObjects > 0
			isNoScope := event.NoScope
			isThroughSmoke := event.ThroughSmoke
			isAirborne := event.AttackerBlind // Reusing for attacker state
			isBlind := event.AttackerBlind
			matchContext.StatsAccumulator.RecordSpecialKill(killerSteamID, isWallbang, isNoScope, isThroughSmoke, isAirborne, isBlind)

			// Record assist if present
			if assisterSteamID != "" {
				matchContext.StatsAccumulator.RecordAssist(assisterSteamID, assisterName, assisterTeam)
			}

			// Track flash assist if victim was flashed
			if event.AssistedFlash && event.Assister != nil {
				matchContext.StatsAccumulator.RecordFlashAssist(assisterSteamID, assisterName, assisterTeam)
			}

			// Periodic cleanup of old deaths to save memory
			matchContext.StatsAccumulator.CleanupOldDeaths(p.CurrentTime())
		}

		// Check if this was an opening kill (for payload)
		isOpeningKill := false
		if matchContext.StatsAccumulator != nil {
			stats := matchContext.StatsAccumulator.GetPlayerStats(killerSteamID)
			// Simple check: if first kill in accumulator for this round
			isOpeningKill = stats != nil && stats.OpeningKills > 0
		}

		// Extract position data
		var killerPosition *Position3D
		var victimPosition *Position3D
		var distance float64

		if event.Killer != nil {
			pos := event.Killer.LastAlivePosition
			killerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		if event.Victim != nil {
			pos := event.Victim.LastAlivePosition
			victimPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Calculate distance between killer and victim
		if killerPosition != nil && victimPosition != nil {
			dx := killerPosition.X - victimPosition.X
			dy := killerPosition.Y - victimPosition.Y
			dz := killerPosition.Z - victimPosition.Z
			distance = math.Sqrt(dx*dx + dy*dy + dz*dz)
		}

		// Extract special kill flags
		isWallbang := event.PenetratedObjects > 0
		isNoScope := event.NoScope
		isThroughSmoke := event.ThroughSmoke
		isAttackerBlind := event.AttackerBlind

		payload := KillPayload{
			KillerName:      killerName,
			KillerSteamID:   killerSteamID,
			KillerPosition:  killerPosition,
			VictimName:      victimName,
			VictimSteamID:   victimSteamID,
			VictimPosition:  victimPosition,
			Weapon:          event.Weapon.String(),
			Headshot:        event.IsHeadshot,
			IsOpeningKill:   isOpeningKill,
			IsWallbang:      isWallbang,
			IsNoScope:       isNoScope,
			IsThroughSmoke:  isThroughSmoke,
			IsAttackerBlind: isAttackerBlind,
			AssisterName:    assisterName,
			AssisterSteamID: assisterSteamID,
			Distance:        distance,
		}

		if event.AssistedFlash && event.Assister != nil {
			payload.FlashAssister = assisterName
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_KillID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create kill event", "err", err)
			return
		}

		out <- gameEvent
	}
}