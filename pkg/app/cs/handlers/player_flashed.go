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

// FlashPayload contains data for a player being flashed
type FlashPayload struct {
	AttackerName     string      `json:"attacker_name"`
	AttackerSteamID  string      `json:"attacker_steam_id,omitempty"`
	AttackerPosition *Position3D `json:"attacker_position,omitempty"`
	VictimName       string      `json:"victim_name"`
	VictimSteamID    string      `json:"victim_steam_id,omitempty"`
	VictimPosition   *Position3D `json:"victim_position,omitempty"`
	FlashDuration    float64     `json:"flash_duration"` // Duration in seconds
	IsTeamFlash      bool        `json:"is_team_flash"`
	Distance         float64     `json:"distance,omitempty"`
}

// PlayerFlashed handles flash events to track enemies flashed
// This is called when a player is blinded by a flashbang
func PlayerFlashed(p dem.Parser, matchContext *state.CS2MatchContext, out chan *entities.GameEvent) func(e evt.PlayerFlashed) {
	return func(event evt.PlayerFlashed) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		// Skip if no attacker (flasher) or no player flashed
		if event.Attacker == nil || event.Player == nil {
			return
		}

		roundIndex := gs.TotalRoundsPlayed()
		currentTick := replay_common.TickIDType(gs.IngameTick())
		matchContext = matchContext.WithRound(roundIndex, gs)

		// Get team information
		attackerTeam := event.Attacker.Team
		playerTeam := event.Player.Team
		isTeamFlash := attackerTeam == playerTeam

		// Skip if duration is too short (less than 0.3 seconds) 
		flashDuration := event.FlashDuration().Seconds()
		if flashDuration < 0.3 {
			return
		}

		// Get attacker info
		attackerSteamID := strconv.FormatUint(event.Attacker.SteamID64, 10)
		attackerName := event.Attacker.Name
		attackerSide := "CT"
		if attackerTeam == 2 {
			attackerSide = "T"
		}

		// Get victim info
		victimSteamID := strconv.FormatUint(event.Player.SteamID64, 10)
		victimName := event.Player.Name

		// Extract position data
		var attackerPosition *Position3D
		var victimPosition *Position3D
		var distance float64

		attackerPos := event.Attacker.LastAlivePosition
		attackerPosition = &Position3D{
			X: attackerPos.X,
			Y: attackerPos.Y,
			Z: attackerPos.Z,
		}

		victimPos := event.Player.LastAlivePosition
		victimPosition = &Position3D{
			X: victimPos.X,
			Y: victimPos.Y,
			Z: victimPos.Z,
		}

		// Calculate distance
		dx := attackerPosition.X - victimPosition.X
		dy := attackerPosition.Y - victimPosition.Y
		dz := attackerPosition.Z - victimPosition.Z
		distance = math.Sqrt(dx*dx + dy*dy + dz*dz)

		// Record enemy flashed in stats accumulator (only enemy flashes)
		if !isTeamFlash && matchContext.StatsAccumulator != nil {
			matchContext.StatsAccumulator.RecordEnemyFlashed(attackerSteamID, attackerName, attackerSide)
		}

		slog.Debug("PlayerFlashed",
			"attacker", attackerName,
			"victim", victimName,
			"duration", flashDuration,
			"isTeamFlash", isTeamFlash)

		// Create flash event payload
		payload := FlashPayload{
			AttackerName:     attackerName,
			AttackerSteamID:  attackerSteamID,
			AttackerPosition: attackerPosition,
			VictimName:       victimName,
			VictimSteamID:    victimSteamID,
			VictimPosition:   victimPosition,
			FlashDuration:    flashDuration,
			IsTeamFlash:      isTeamFlash,
			Distance:         distance,
		}

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_PlayerFlashedID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("unable to create flash event", "err", err)
			return
		}

		out <- gameEvent
	}
}
