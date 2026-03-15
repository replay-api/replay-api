package handlers

import (
	"log/slog"
	"strconv"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/msg"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	"github.com/replay-api/replay-api/pkg/app/cs/state"
	"github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
)

// CS2GrenadeEventPayload contains data for CS2 grenade events extracted from GenericGameEvent
type CS2GrenadeEventPayload struct {
	ThrowerName     string      `json:"thrower_name,omitempty"`
	ThrowerSteamID  string      `json:"thrower_steam_id,omitempty"`
	ThrowerTeam     string      `json:"thrower_team,omitempty"`
	ThrowerPosition *Position3D `json:"thrower_position,omitempty"`
	GrenadePosition *Position3D `json:"grenade_position"`
	GrenadeType     string      `json:"grenade_type"` // "he", "flash", "smoke", "molotov", "incendiary", "decoy"
	EntityID        int         `json:"entity_id,omitempty"`
}

// extractFloat32FromData extracts a float32 value from GenericGameEvent data
func extractFloat32FromData(data map[string]*msg.CSVCMsg_GameEventKeyT, key string) float32 {
	if val, ok := data[key]; ok && val != nil {
		return val.GetValFloat()
	}
	return 0
}

// extractIntFromData extracts an int value from GenericGameEvent data
func extractIntFromData(data map[string]*msg.CSVCMsg_GameEventKeyT, key string) int {
	if val, ok := data[key]; ok && val != nil {
		// Try short first (most common for userid/entityid)
		if s := val.GetValShort(); s != 0 {
			return int(s)
		}
		// Then try long
		return int(val.GetValLong())
	}
	return 0
}

func GenericGameEvent(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.GenericGameEvent) {
	return func(event evt.GenericGameEvent) {
		gs := p.GameState()
		if gs == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		matchContext = matchContext.WithRound(roundIndex, gs)
		currentTick := replay_common.TickIDType(gs.IngameTick())

		// Handle CS2 grenade events - extract data directly from proto message
		switch event.Name {
		case "hegrenade_detonate":
			handleCS2GrenadeEventDirect(p, matchContext, out, event.Name, event.Data, roundIndex, currentTick, "he", fps_events.Event_HeGrenadeExplodeID)
		case "flashbang_detonate":
			handleCS2GrenadeEventDirect(p, matchContext, out, event.Name, event.Data, roundIndex, currentTick, "flash", fps_events.Event_FlashExplodeID)
		case "smokegrenade_detonate":
			handleCS2GrenadeEventDirect(p, matchContext, out, event.Name, event.Data, roundIndex, currentTick, "smoke", fps_events.Event_SmokeStartID)
		case "inferno_startburn":
			handleCS2GrenadeEventDirect(p, matchContext, out, event.Name, event.Data, roundIndex, currentTick, "molotov", fps_events.Event_InfernoStartID)
		case "decoy_started":
			handleCS2GrenadeEventDirect(p, matchContext, out, event.Name, event.Data, roundIndex, currentTick, "decoy", fps_events.Event_DecoyStartID)
		default:
			// Log other events for debugging at debug level
			slog.Debug("GenericGameEvent skipped", "name", event.Name)
		}
	}
}

// handleCS2GrenadeEventDirect processes a CS2 grenade event using direct proto message access
func handleCS2GrenadeEventDirect(
	p dem.Parser,
	matchContext *state.CS2MatchContext,
	out chan *entities.GameEvent,
	eventName string,
	data map[string]*msg.CSVCMsg_GameEventKeyT,
	roundIndex int,
	currentTick replay_common.TickIDType,
	grenadeType string,
	eventID fps_events.EventIDKey,
) {
	// Extract position directly from proto data
	x := extractFloat32FromData(data, "x")
	y := extractFloat32FromData(data, "y")
	z := extractFloat32FromData(data, "z")
	entityID := extractIntFromData(data, "entityid")
	userID := extractIntFromData(data, "userid")

	grenadePosition := &Position3D{
		X: float64(x),
		Y: float64(y),
		Z: float64(z),
	}

	// Try to get thrower info from game state using userid
	gs := p.GameState()
	throwerName := ""
	throwerSteamID := ""
	throwerTeam := "Unknown"
	var throwerPosition *Position3D

	if userID > 0 && gs != nil {
		// Try to find the player by user ID
		for _, player := range gs.Participants().All() {
			if player != nil && player.UserID == userID {
				throwerName = player.Name
				throwerSteamID = strconv.FormatUint(player.SteamID64, 10)
				switch player.Team {
				case 2:
					throwerTeam = "T"
				case 3:
					throwerTeam = "CT"
				}
				pos := player.LastAlivePosition
				throwerPosition = &Position3D{
					X: pos.X,
					Y: pos.Y,
					Z: pos.Z,
				}
				break
			}
		}
	}

	slog.Info("CS2 Grenade Event CAPTURED",
		"event", eventName,
		"type", grenadeType,
		"x", x,
		"y", y,
		"z", z,
		"thrower", throwerName,
		"round", roundIndex+1,
		"entityID", entityID)

	payload := CS2GrenadeEventPayload{
		ThrowerName:     throwerName,
		ThrowerSteamID:  throwerSteamID,
		ThrowerTeam:     throwerTeam,
		ThrowerPosition: throwerPosition,
		GrenadePosition: grenadePosition,
		GrenadeType:     grenadeType,
		EntityID:        entityID,
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
		slog.Error("unable to create CS2 grenade event", "err", err, "event", eventName)
		return
	}

	out <- gameEvent
}
