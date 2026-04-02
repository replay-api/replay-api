package scores_services

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	replay_entities "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	scores_entities "github.com/replay-api/replay-api/pkg/domain/scores/entities"
	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// ReplayScoreBridge converts replay match data into scores domain types
// and submits them through the ScoreVerificationService.
type ReplayScoreBridge struct {
	verificationService *ScoreVerificationService
}

// NewReplayScoreBridge creates a new bridge between replay and scores domains
func NewReplayScoreBridge(verificationService *ScoreVerificationService) *ReplayScoreBridge {
	return &ReplayScoreBridge{
		verificationService: verificationService,
	}
}

// OnMatchProcessed is the callback to be wired into ProcessReplayFileUseCase.OnMatchProcessed.
// It extracts team and player results from the match scoreboard and submits them to the scores domain.
func (b *ReplayScoreBridge) OnMatchProcessed(ctx context.Context, match *replay_entities.Match, replayFileID uuid.UUID) error {
	if match == nil || len(match.Scoreboard.TeamScoreboards) < 2 {
		slog.InfoContext(ctx, "replay match has insufficient scoreboard data, skipping score submission",
			slog.String("match_id", match.ID.String()),
		)
		return nil
	}

	teamResults := buildTeamResultsFromMatch(match)
	playerResults := buildPlayerResultsFromMatch(match)

	duration := time.Duration(match.Duration * float64(time.Second))

	// Count rounds
	roundsPlayed := 0
	for _, ts := range match.Scoreboard.TeamScoreboards {
		roundsPlayed += ts.TeamScore
	}

	return b.verificationService.ProcessReplayCompleted(
		ctx,
		match.ID,
		replayFileID,
		match.GameID,
		match.MapName,
		match.Mode,
		teamResults,
		playerResults,
		match.PlayedAt,
		duration,
		roundsPlayed,
		nil, // matchmakingSessionID — not available from replay context
		nil, // tournamentID — not available from replay context
	)
}

// buildTeamResultsFromMatch converts replay TeamScoreboards to scores domain TeamResults
func buildTeamResultsFromMatch(match *replay_entities.Match) []scores_entities.TeamResult {
	results := make([]scores_entities.TeamResult, 0, len(match.Scoreboard.TeamScoreboards))

	for i, ts := range match.Scoreboard.TeamScoreboards {
		playerIDs := make([]uuid.UUID, 0, len(ts.Players))
		for _, p := range ts.Players {
			playerIDs = append(playerIDs, p.GetID())
		}

		// Determine position based on score
		position := i + 1

		results = append(results, scores_entities.TeamResult{
			TeamID:   ts.Team.GetID(),
			TeamName: ts.Team.Name,
			Score:    ts.TeamScore,
			Position: position,
			Players:  playerIDs,
		})
	}

	// Adjust positions: team with higher score gets position 1
	if len(results) >= 2 && results[1].Score > results[0].Score {
		results[0].Position = 2
		results[1].Position = 1
	}

	return results
}

// buildPlayerResultsFromMatch converts replay PlayerStatsEntry to scores domain PlayerResults,
// carrying all advanced esports statistics through the flexible Stats map.
func buildPlayerResultsFromMatch(match *replay_entities.Match) []scores_entities.PlayerResult {
	results := make([]scores_entities.PlayerResult, 0)

	for _, ts := range match.Scoreboard.TeamScoreboards {
		teamID := ts.Team.GetID()

		for _, stats := range ts.PlayerStats {
			playerID := uuid.Nil
			steamID := stats.PlayerID
			displayName := ""

			// Try to find the player's UUID and display name from player metadata
			for _, p := range ts.Players {
				if p.NetworkUserID == stats.PlayerID || p.GetID().String() == stats.PlayerID {
					playerID = p.GetID()
					steamID = p.NetworkUserID
					displayName = p.Name
					break
				}
			}

			// Determine MVP status by finding the player with the highest MVP count
			isMVP := false
			if match.Scoreboard.MatchMVP != nil && match.Scoreboard.MatchMVP.GetID() == playerID {
				isMVP = true
			}

			// Build the flexible stats map with ALL advanced esports data
			advancedStats := map[string]interface{}{
				// Player identity (for reconciliation fallback)
				"steam_id":     steamID,
				"display_name": displayName,
				"side":         ts.Side,
				// Core combat stats
				"kd_ratio":       stats.KDRatio,
				"headshots":      stats.Headshots,
				"headshot_pct":   stats.HeadshotPct,
				"total_damage":   stats.TotalDamage,
				"utility_damage": stats.UtilityDamage,
				"adr":            stats.ADR,
				"mvp_count":      stats.MVPCount,
				// Advanced esports metrics
				"kast":            stats.KAST,
				"impact_rating":   stats.ImpactRating,
				"rating_2":        stats.Rating2,
				"opening_kills":   stats.OpeningKills,
				"opening_deaths":  stats.OpeningDeaths,
				"trade_kills":     stats.TradeKills,
				"clutch_wins":     stats.ClutchWins,
				"clutch_attempts": stats.ClutchAttempts,
				"flash_assists":   stats.FlashAssists,
				"enemies_flashed": stats.EnemiesFlashed,
				"entry_attempts":  stats.EntryAttempts,
				"entry_successes": stats.EntrySuccesses,
				"multi_kills":     stats.MultiKills,
				// Source metadata
				"source": "replay_file",
			}

			results = append(results, scores_entities.PlayerResult{
				PlayerID: playerID,
				TeamID:   teamID,
				Score:    stats.Score,
				Kills:    stats.Kills,
				Deaths:   stats.Deaths,
				Assists:  stats.Assists,
				Rating:   stats.Rating2, // Use HLTV 2.0 rating as primary rating
				IsMVP:    isMVP,
				Stats:    advancedStats,
			})
		}
	}

	return results
}

// MatchResultFromReplayCmd is a helper to build a SubmitReplayResultCommand from a Match entity.
// Useful for re-processing / backfilling scores from existing matches.
func MatchResultFromReplayCmd(match *replay_entities.Match, replayFileID uuid.UUID) scores_in.SubmitReplayResultCommand {
	teamResults := buildTeamResultsFromMatch(match)
	playerResults := buildPlayerResultsFromMatch(match)
	duration := time.Duration(match.Duration * float64(time.Second))

	roundsPlayed := 0
	for _, ts := range match.Scoreboard.TeamScoreboards {
		roundsPlayed += ts.TeamScore
	}

	return scores_in.SubmitReplayResultCommand{
		MatchID:       match.ID,
		ReplayID:      replayFileID,
		GameID:        replay_common.GameIDKey(match.GameID),
		MapName:       match.MapName,
		Mode:          match.Mode,
		TeamResults:   teamResults,
		PlayerResults: playerResults,
		PlayedAt:      match.PlayedAt,
		Duration:      duration,
		RoundsPlayed:  roundsPlayed,
	}
}
