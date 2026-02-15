package handlers

import (
	"log/slog"
	"strconv"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	"github.com/replay-api/replay-api/pkg/app/cs/state"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	"github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

// BombPlantedPayload contains data for bomb plant events
type BombPlantedPayload struct {
	PlayerName     string      `json:"player_name"`
	PlayerSteamID  string      `json:"player_steam_id,omitempty"`
	PlayerPosition *Position3D `json:"player_position,omitempty"`
	BombSite       string      `json:"bomb_site"` // "A" or "B"
}

// BombDefusedPayload contains data for bomb defuse events
type BombDefusedPayload struct {
	PlayerName     string      `json:"player_name"`
	PlayerSteamID  string      `json:"player_steam_id,omitempty"`
	PlayerPosition *Position3D `json:"player_position,omitempty"`
	BombSite       string      `json:"bomb_site"` // "A" or "B"
}

// BombPlanted handles bomb plant events
func BombPlanted(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.BombPlanted) {
	return func(event evt.BombPlanted) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get player info
		playerName := ""
		playerSteamID := ""
		var playerPosition *Position3D

		if event.Player != nil {
			playerName = event.Player.Name
			playerSteamID = strconv.FormatUint(event.Player.SteamID64, 10)
			
			pos := event.Player.LastAlivePosition
			playerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Get bombsite
		bombSite := "Unknown"
		switch event.Site {
		case evt.BombsiteA:
			bombSite = "A"
		case evt.BombsiteB:
			bombSite = "B"
		}

		slog.Debug("BombPlanted",
			"player", playerName,
			"site", bombSite,
			"round", roundIndex+1)

		payload := BombPlantedPayload{
			PlayerName:     playerName,
			PlayerSteamID:  playerSteamID,
			PlayerPosition: playerPosition,
			BombSite:       bombSite,
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_BombPlantedID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create bomb planted event", "err", err)
			return
		}

		out <- gameEvent
	}
}

// BombDefused handles bomb defuse events
func BombDefused(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.BombDefused) {
	return func(event evt.BombDefused) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get player info
		playerName := ""
		playerSteamID := ""
		var playerPosition *Position3D

		if event.Player != nil {
			playerName = event.Player.Name
			playerSteamID = strconv.FormatUint(event.Player.SteamID64, 10)
			
			pos := event.Player.LastAlivePosition
			playerPosition = &Position3D{
				X: pos.X,
				Y: pos.Y,
				Z: pos.Z,
			}
		}

		// Get bombsite
		bombSite := "Unknown"
		switch event.Site {
		case evt.BombsiteA:
			bombSite = "A"
		case evt.BombsiteB:
			bombSite = "B"
		}

		slog.Debug("BombDefused",
			"player", playerName,
			"site", bombSite,
			"round", roundIndex+1)

		payload := BombDefusedPayload{
			PlayerName:     playerName,
			PlayerSteamID:  playerSteamID,
			PlayerPosition: playerPosition,
			BombSite:       bombSite,
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_BombDefusedID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create bomb defused event", "err", err)
			return
		}

		out <- gameEvent
	}
}
