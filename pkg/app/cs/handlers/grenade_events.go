package handlers

import (
	"log/slog"
	"strconv"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	"github.com/replay-api/replay-api/pkg/app/cs/state"
	"github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
)

// GrenadeEventPayload contains data for grenade explosion events
type GrenadeEventPayload struct {
	ThrowerName     string      `json:"thrower_name"`
	ThrowerSteamID  string      `json:"thrower_steam_id,omitempty"`
	ThrowerTeam     string      `json:"thrower_team,omitempty"`
	ThrowerPosition *Position3D `json:"thrower_position,omitempty"`
	GrenadePosition *Position3D `json:"grenade_position"`
	GrenadeType     string      `json:"grenade_type"` // "he", "flash", "smoke", "molotov", "incendiary", "decoy"
	Damage          int         `json:"damage,omitempty"`         // For HE grenades
	PlayersHit      int         `json:"players_hit,omitempty"`    // Number of players affected
	EnemiesHit      int         `json:"enemies_hit,omitempty"`    // Number of enemies affected
	TeamHit         int         `json:"team_hit,omitempty"`       // Number of teammates affected
}

// SmokeEventPayload contains data for smoke grenade events
type SmokeEventPayload struct {
	ThrowerName     string      `json:"thrower_name"`
	ThrowerSteamID  string      `json:"thrower_steam_id,omitempty"`
	ThrowerTeam     string      `json:"thrower_team,omitempty"`
	ThrowerPosition *Position3D `json:"thrower_position,omitempty"`
	SmokePosition   *Position3D `json:"smoke_position"`
	Duration        float64     `json:"duration,omitempty"` // Duration in seconds
	IsExtinguished  bool        `json:"is_extinguished,omitempty"`
}

// HeGrenadeExplode handles HE grenade explosion events
func HeGrenadeExplode(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.HeExplode) {
	return func(event evt.HeExplode) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get thrower info
		throwerName := ""
		throwerSteamID := ""
		throwerTeam := "Unknown"
		var throwerPosition *Position3D

		if event.Thrower != nil {
			throwerName = event.Thrower.Name
			throwerSteamID = strconv.FormatUint(event.Thrower.SteamID64, 10)
			
			switch event.Thrower.Team {
			case 2:
				throwerTeam = "T"
			case 3:
				throwerTeam = "CT"
			}
			
			pos := event.Thrower.LastAlivePosition
			throwerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Get grenade position
		grenadePosition := &Position3D{
			X: event.Position.X,
			Y: event.Position.Y,
			Z: event.Position.Z,
		}

		slog.Info("HeGrenadeExplode TRIGGERED",
			"thrower", throwerName,
			"position", grenadePosition,
			"round", roundIndex+1)

		payload := GrenadeEventPayload{
			ThrowerName:     throwerName,
			ThrowerSteamID:  throwerSteamID,
			ThrowerTeam:     throwerTeam,
			ThrowerPosition: throwerPosition,
			GrenadePosition: grenadePosition,
			GrenadeType:     "he",
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_HeGrenadeExplodeID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create HE grenade event", "err", err)
			return
		}

		out <- gameEvent
	}
}

// FlashExplode handles flashbang explosion events (different from PlayerFlashed)
func FlashExplode(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.FlashExplode) {
	return func(event evt.FlashExplode) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get thrower info
		throwerName := ""
		throwerSteamID := ""
		throwerTeam := "Unknown"
		var throwerPosition *Position3D

		if event.Thrower != nil {
			throwerName = event.Thrower.Name
			throwerSteamID = strconv.FormatUint(event.Thrower.SteamID64, 10)
			
			switch event.Thrower.Team {
			case 2:
				throwerTeam = "T"
			case 3:
				throwerTeam = "CT"
			}
			
			pos := event.Thrower.LastAlivePosition
			throwerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Get grenade position
		grenadePosition := &Position3D{
			X: event.Position.X,
			Y: event.Position.Y,
			Z: event.Position.Z,
		}

		slog.Info("FlashExplode TRIGGERED",
			"thrower", throwerName,
			"position", grenadePosition,
			"round", roundIndex+1)

		payload := GrenadeEventPayload{
			ThrowerName:     throwerName,
			ThrowerSteamID:  throwerSteamID,
			ThrowerTeam:     throwerTeam,
			ThrowerPosition: throwerPosition,
			GrenadePosition: grenadePosition,
			GrenadeType:     "flash",
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_FlashExplodeID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create flash grenade event", "err", err)
			return
		}

		out <- gameEvent
	}
}

// SmokeStart handles smoke grenade activation
func SmokeStart(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.SmokeStart) {
	return func(event evt.SmokeStart) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get thrower info
		throwerName := ""
		throwerSteamID := ""
		throwerTeam := "Unknown"
		var throwerPosition *Position3D

		if event.Thrower != nil {
			throwerName = event.Thrower.Name
			throwerSteamID = strconv.FormatUint(event.Thrower.SteamID64, 10)
			
			switch event.Thrower.Team {
			case 2:
				throwerTeam = "T"
			case 3:
				throwerTeam = "CT"
			}
			
			pos := event.Thrower.LastAlivePosition
			throwerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Get smoke position
		smokePosition := &Position3D{
			X: event.Position.X,
			Y: event.Position.Y,
			Z: event.Position.Z,
		}

		slog.Debug("SmokeStart",
			"thrower", throwerName,
			"position", smokePosition,
			"round", roundIndex+1)

		payload := SmokeEventPayload{
			ThrowerName:     throwerName,
			ThrowerSteamID:  throwerSteamID,
			ThrowerTeam:     throwerTeam,
			ThrowerPosition: throwerPosition,
			SmokePosition:   smokePosition,
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_SmokeStartID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create smoke start event", "err", err)
			return
		}

		out <- gameEvent
	}
}

// InfernoStart handles molotov/incendiary fire start
func InfernoStart(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.InfernoStart) {
	return func(event evt.InfernoStart) {
		gs := p.GameState()

		if gs == nil || event.Inferno == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get thrower info from Inferno
		throwerName := ""
		throwerSteamID := ""
		throwerTeam := "Unknown"
		var throwerPosition *Position3D

		if thrower := event.Inferno.Thrower(); thrower != nil {
			throwerName = thrower.Name
			throwerSteamID = strconv.FormatUint(thrower.SteamID64, 10)
			
			switch thrower.Team {
			case 2:
				throwerTeam = "T"
			case 3:
				throwerTeam = "CT"
			}
			
			pos := thrower.LastAlivePosition
			throwerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Get inferno position from the fires collection
		// Use the first fire's position as the center
		var infernoPosition *Position3D
		fires := event.Inferno.Fires()
		activeFires := fires.Active().List()
		if len(activeFires) > 0 {
			firstFire := activeFires[0]
			infernoPosition = &Position3D{
				X: firstFire.X,
				Y: firstFire.Y,
				Z: firstFire.Z,
			}
		}

		// Fallback to thrower position if no fire positions available
		if infernoPosition == nil && throwerPosition != nil {
			infernoPosition = throwerPosition
		}
		if infernoPosition == nil {
			infernoPosition = &Position3D{X: 0, Y: 0, Z: 0}
		}

		// Determine grenade type based on team (T=molotov, CT=incendiary)
		grenadeType := "molotov"
		if throwerTeam == "CT" {
			grenadeType = "incendiary"
		}

		slog.Debug("InfernoStart",
			"thrower", throwerName,
			"type", grenadeType,
			"position", infernoPosition,
			"round", roundIndex+1)

		payload := GrenadeEventPayload{
			ThrowerName:     throwerName,
			ThrowerSteamID:  throwerSteamID,
			ThrowerTeam:     throwerTeam,
			ThrowerPosition: throwerPosition,
			GrenadePosition: infernoPosition,
			GrenadeType:     grenadeType,
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_InfernoStartID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create inferno start event", "err", err)
			return
		}

		out <- gameEvent
	}
}

// DecoyStart handles decoy grenade activation
func DecoyStart(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.DecoyStart) {
	return func(event evt.DecoyStart) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get thrower info
		throwerName := ""
		throwerSteamID := ""
		throwerTeam := "Unknown"
		var throwerPosition *Position3D

		if event.Thrower != nil {
			throwerName = event.Thrower.Name
			throwerSteamID = strconv.FormatUint(event.Thrower.SteamID64, 10)
			
			switch event.Thrower.Team {
			case 2:
				throwerTeam = "T"
			case 3:
				throwerTeam = "CT"
			}
			
			pos := event.Thrower.LastAlivePosition
			throwerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Get decoy position
		decoyPosition := &Position3D{
			X: event.Position.X,
			Y: event.Position.Y,
			Z: event.Position.Z,
		}

		slog.Debug("DecoyStart",
			"thrower", throwerName,
			"position", decoyPosition,
			"round", roundIndex+1)

		payload := GrenadeEventPayload{
			ThrowerName:     throwerName,
			ThrowerSteamID:  throwerSteamID,
			ThrowerTeam:     throwerTeam,
			ThrowerPosition: throwerPosition,
			GrenadePosition: decoyPosition,
			GrenadeType:     "decoy",
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_DecoyStartID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create decoy start event", "err", err)
			return
		}

		out <- gameEvent
	}
}

// GrenadeProjectileDestroy handles all grenade detonation events in CS2
// This is the most reliable way to capture grenade explosions in Source 2 demos
func GrenadeProjectileDestroy(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.GrenadeProjectileDestroy) {
	return func(event evt.GrenadeProjectileDestroy) {
		slog.Info("GrenadeProjectileDestroy CALLED - Entry point")
		gs := p.GameState()

		if gs == nil || event.Projectile == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get grenade type from WeaponInstance
		grenadeType := "unknown"
		if event.Projectile.WeaponInstance != nil {
			switch event.Projectile.WeaponInstance.Type {
			case common.EqHE:
				grenadeType = "he"
			case common.EqFlash:
				grenadeType = "flash"
			case common.EqSmoke:
				grenadeType = "smoke"
			case common.EqMolotov:
				grenadeType = "molotov"
			case common.EqIncendiary:
				grenadeType = "incendiary"
			case common.EqDecoy:
				grenadeType = "decoy"
			default:
				return // Not a grenade we care about
			}
		} else {
			return // No weapon instance
		}

		// Get thrower info
		throwerName := ""
		throwerSteamID := ""
		throwerTeam := "Unknown"
		var throwerPosition *Position3D

		if event.Projectile.Thrower != nil {
			throwerName = event.Projectile.Thrower.Name
			throwerSteamID = strconv.FormatUint(event.Projectile.Thrower.SteamID64, 10)
			
			switch event.Projectile.Thrower.Team {
			case 2:
				throwerTeam = "T"
			case 3:
				throwerTeam = "CT"
			}
			
			pos := event.Projectile.Thrower.LastAlivePosition
			throwerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Get grenade final position
		grenadePosition := event.Projectile.Position()
		position3D := &Position3D{
			X: grenadePosition.X,
			Y: grenadePosition.Y,
			Z: grenadePosition.Z,
		}

		slog.Info("GrenadeProjectileDestroy",
			"type", grenadeType,
			"thrower", throwerName,
			"position", position3D,
			"round", roundIndex+1)

		payload := GrenadeEventPayload{
			ThrowerName:     throwerName,
			ThrowerSteamID:  throwerSteamID,
			ThrowerTeam:     throwerTeam,
			ThrowerPosition: throwerPosition,
			GrenadePosition: position3D,
			GrenadeType:     grenadeType,
		}

		// Map grenade type to appropriate event ID
		var eventID fps_events.EventIDKey
		switch grenadeType {
		case "he":
			eventID = fps_events.Event_HeGrenadeExplodeID
		case "flash":
			eventID = fps_events.Event_FlashExplodeID
		case "smoke":
			eventID = fps_events.Event_SmokeStartID
		case "molotov", "incendiary":
			eventID = fps_events.Event_InfernoStartID
		case "decoy":
			eventID = fps_events.Event_DecoyStartID
		default:
			return
		}

		gameEvent, err := event_factory.NewGameEvent(
			eventID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create grenade event", "type", grenadeType, "err", err)
			return
		}

		out <- gameEvent
	}
}

// GrenadeProjectileThrow handles grenade throw events (for testing/debugging)
func GrenadeProjectileThrow(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.GrenadeProjectileThrow) {
	return func(event evt.GrenadeProjectileThrow) {
		slog.Info("GrenadeProjectileThrow CALLED", 
			"projectile", event.Projectile != nil,
			"hasWeapon", event.Projectile != nil && event.Projectile.WeaponInstance != nil)
		// Not emitting events, just logging for debugging
	}
}
