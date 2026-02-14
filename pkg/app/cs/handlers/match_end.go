package handlers

import (
	"fmt"
	"log/slog"
	"strconv"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	evt "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	event_factory "github.com/replay-api/replay-api/pkg/app/cs/factories"
	state "github.com/replay-api/replay-api/pkg/app/cs/state"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

// Re-export types from entities package for backwards compatibility
type FinalScoreboardPayload = replay_entity.FinalScoreboardPayload
type PlayerScoreboardData = replay_entity.PlayerScoreboardDataEntry

// AnnouncementWinPanelMatch captures the final scoreboard when match ends
func AnnouncementWinPanelMatch(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e evt.AnnouncementWinPanelMatch) {
	return func(event evt.AnnouncementWinPanelMatch) {
		gs := p.GameState()

		if gs == nil {
			slog.Warn("Game state is nil at match end")
			return
		}

		slog.Info("AnnouncementWinPanelMatch - capturing final scoreboard")

		teamCT := gs.TeamCounterTerrorists()
		teamT := gs.TeamTerrorists()

		ctScore := teamCT.Score()
		tScore := teamT.Score()
		totalRounds := ctScore + tScore

		winnerSide := "tie"
		if ctScore > tScore {
			winnerSide = "CT"
		} else if tScore > ctScore {
			winnerSide = "T"
		}

		// Finalize accumulated stats (calculate multi-kills, etc.)
		if matchContext.StatsAccumulator != nil {
			matchContext.StatsAccumulator.FinalizeMultiKills()
		}

		// Build player scoreboard data
		players := make([]PlayerScoreboardData, 0)
		
		for _, player := range gs.Participants().Playing() {
			if player == nil {
				continue
			}

			steamID := strconv.FormatUint(player.SteamID64, 10)

			kd := float64(0)
			if player.Deaths() > 0 {
				kd = float64(player.Kills()) / float64(player.Deaths())
			} else if player.Kills() > 0 {
				kd = float64(player.Kills())
			}

			adr := float64(0)
			if totalRounds > 0 {
				adr = float64(player.TotalDamage()) / float64(totalRounds)
			}

			side := "CT"
			teamName := teamCT.ClanName()
			if player.Team == 2 { // Terrorist
				side = "T"
				teamName = teamT.ClanName()
			}

			// Get accumulated stats from our tracker
			headshots := 0
			openingKills := 0
			openingDeaths := 0
			tradeKills := 0
			tradeDeaths := 0
			flashAssists := 0
			enemiesFlashed := 0

			if matchContext.StatsAccumulator != nil {
				accumulatedStats := matchContext.StatsAccumulator.GetPlayerStats(steamID)
				if accumulatedStats != nil {
					headshots = accumulatedStats.Headshots
					openingKills = accumulatedStats.OpeningKills
					openingDeaths = accumulatedStats.OpeningDeaths
					tradeKills = accumulatedStats.TradeKills
					tradeDeaths = accumulatedStats.RoundsTraded // TradeDeaths = rounds where player was traded
					flashAssists = accumulatedStats.FlashAssists
					enemiesFlashed = accumulatedStats.EnemiesFlashed
				}
			}

			playerData := PlayerScoreboardData{
				NetworkUserID:   steamID,
				Name:            player.Name,
				Team:            teamName,
				Side:            side,
				Kills:           player.Kills(),
				Deaths:          player.Deaths(),
				Assists:         player.Assists(),
				KDRatio:         kd,
				Headshots:       headshots,
				TotalDamage:     player.TotalDamage(),
				UtilityDamage:   player.UtilityDamage(),
				ADR:             adr,
				MVPCount:        player.MVPs(),
				Score:           player.Score(),
				FirstKills:      openingKills,
				FirstDeaths:     openingDeaths,
				TradeKills:      tradeKills,
				TradeDeaths:     tradeDeaths,
				FlashAssists:    flashAssists,
				EnemiesFlashed:  enemiesFlashed,
			}

			players = append(players, playerData)
		}

		payload := FinalScoreboardPayload{
			CTScore:     ctScore,
			TScore:      tScore,
			CTTeamName:  teamCT.ClanName(),
			TTeamName:   teamT.ClanName(),
			TotalRounds: totalRounds,
			WinnerSide:  winnerSide,
			Players:     players,
			Duration:    p.Header().PlaybackTime.Seconds(),
			MapName:     p.Header().MapName,
		}

		currentTick := replay_common.TickIDType(gs.IngameTick())

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_MatchEndID,
			matchContext,
			totalRounds,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error(fmt.Sprintf("AnnouncementWinPanelMatch: unable to create game event: %s", err.Error()))
			return
		}

		slog.Info("AnnouncementWinPanelMatch - final scoreboard captured",
			"ct_score", ctScore,
			"t_score", tScore,
			"player_count", len(players))

		out <- gameEvent
	}
}

// GameHalfEnded captures score at halftime
func GameHalfEnded(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e evt.GameHalfEnded) {
	return func(event evt.GameHalfEnded) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		slog.Info("GameHalfEnded - halftime reached")

		teamCT := gs.TeamCounterTerrorists()
		teamT := gs.TeamTerrorists()

		payload := map[string]interface{}{
			"ct_score":    teamCT.Score(),
			"t_score":     teamT.Score(),
			"ct_team":     teamCT.ClanName(),
			"t_team":      teamT.ClanName(),
			"total_rounds": gs.TotalRoundsPlayed(),
		}

		currentTick := replay_common.TickIDType(gs.IngameTick())
		roundIndex := gs.TotalRoundsPlayed()

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_HalftimeID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("GameHalfEnded: unable to create game event", "err", err)
			return
		}

		out <- gameEvent
	}
}

// ScoreUpdated captures every score change
func ScoreUpdated(p dem.Parser, matchContext *state.CS2MatchContext, out chan *replay_entity.GameEvent) func(e evt.ScoreUpdated) {
	return func(event evt.ScoreUpdated) {
		gs := p.GameState()

		if gs == nil {
			return
		}

		teamCT := gs.TeamCounterTerrorists()
		teamT := gs.TeamTerrorists()

		// Determine which team's score changed
		teamSide := "unknown"
		if event.TeamState != nil && event.TeamState.Team() == 3 {
			teamSide = "CT"
		} else if event.TeamState != nil && event.TeamState.Team() == 2 {
			teamSide = "T"
		}

		payload := map[string]interface{}{
			"ct_score":     teamCT.Score(),
			"t_score":      teamT.Score(),
			"old_score":    event.OldScore,
			"new_score":    event.NewScore,
			"team_side":    teamSide,
			"total_rounds": gs.TotalRoundsPlayed(),
		}

		currentTick := replay_common.TickIDType(gs.IngameTick())
		roundIndex := gs.TotalRoundsPlayed()

		gameEvent, err := event_factory.NewGameEvent(
			fps_events.Event_ScoreUpdatedID,
			matchContext,
			roundIndex,
			currentTick,
			p.CurrentTime(),
			payload,
		)

		if err != nil {
			slog.Error("ScoreUpdated: unable to create game event", "err", err)
			return
		}

		out <- gameEvent
	}
}
