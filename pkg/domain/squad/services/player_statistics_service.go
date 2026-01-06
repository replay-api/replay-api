package squad_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// PlayerStatisticsService aggregates player statistics from match history
type PlayerStatisticsService struct {
	matchReader replay_out.MatchMetadataReader
}

// NewPlayerStatisticsService creates a new PlayerStatisticsService
func NewPlayerStatisticsService(matchReader replay_out.MatchMetadataReader) squad_in.PlayerStatisticsReader {
	return &PlayerStatisticsService{
		matchReader: matchReader,
	}
}

// GetPlayerStatistics retrieves aggregated statistics for a player
func (s *PlayerStatisticsService) GetPlayerStatistics(ctx context.Context, playerID uuid.UUID, gameID *replay_common.GameIDKey) (*squad_entities.PlayerStatistics, error) {
	slog.InfoContext(ctx, "Getting player statistics", "player_id", playerID, "game_id", gameID)

	// Default to CS2 if no game specified
	defaultGameID := replay_common.CS2.ID
	if gameID != nil {
		defaultGameID = *gameID
	}

	stats := squad_entities.NewPlayerStatistics(playerID, defaultGameID)

	// Get player's match history using MatchMetadataReader
	matches, err := s.getPlayerMatches(ctx, playerID, 100)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get player match history", "error", err, "player_id", playerID)
		// Return empty stats if we can't get match history
		return stats, nil
	}

	// Aggregate statistics from matches
	for _, match := range matches {
		// Determine match result for this player
		playerTeamIdx, result := s.determineMatchResult(&match, playerID)
		
		// Get player stats from the match
		kills, deaths, assists, headshots, damage, rounds := s.extractPlayerStats(&match, playerID, playerTeamIdx)
		
		// Add to aggregated stats
		stats.AddMatchResult(result, kills, deaths, assists, headshots, damage, rounds)

		// Track if player was MVP
		if match.Scoreboard.MatchMVP != nil && match.Scoreboard.MatchMVP.UserID != nil && *match.Scoreboard.MatchMVP.UserID == playerID {
			stats.MVPCount++
		}

		// Add to recent matches (keep last 10)
		if len(stats.RecentMatches) < 10 {
			stats.RecentMatches = append(stats.RecentMatches, squad_entities.RecentMatchSummary{
				MatchID:  match.ID,
				GameID:   string(match.GameID),
				Result:   result,
				Kills:    kills,
				Deaths:   deaths,
				Assists:  assists,
				Score:    s.formatScore(&match),
				PlayedAt: match.CreatedAt,
			})
		}
	}

	slog.InfoContext(ctx, "Player statistics calculated",
		"player_id", playerID,
		"total_matches", stats.TotalMatches,
		"win_rate", stats.WinRate,
		"kd_ratio", stats.KDRatio,
	)

	return stats, nil
}

// getPlayerMatches retrieves matches for a player using the match reader
func (s *PlayerStatisticsService) getPlayerMatches(ctx context.Context, playerID uuid.UUID, limit int) ([]replay_entity.Match, error) {
	// Search for matches containing this player in the scoreboard
	searchParams := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field:    "scoreboard.team_scoreboards.players._id",
							Values:   []interface{}{playerID.String()},
							Operator: shared.EqualsOperator,
						},
					},
				},
			},
			AggregationClause: shared.AndAggregationClause,
		},
	}

	resultOptions := shared.SearchResultOptions{
		Limit: uint(limit), // #nosec G115 - limit is validated to be non-negative
		Skip:  0,
	}

	compiledSearch, err := s.matchReader.Compile(ctx, searchParams, resultOptions)
	if err != nil {
		return nil, fmt.Errorf("compile search: %w", err)
	}

	results, err := s.matchReader.Search(ctx, *compiledSearch)
	if err != nil {
		return nil, fmt.Errorf("search matches: %w", err)
	}

	return results, nil
}

// determineMatchResult determines if the player won, lost, or drew the match
// Returns the player's team index and the result
func (s *PlayerStatisticsService) determineMatchResult(match *replay_entity.Match, playerID uuid.UUID) (int, string) {
	if len(match.Scoreboard.TeamScoreboards) < 2 {
		return -1, "unknown"
	}

	// Find which team the player is on
	playerTeamIdx := -1
	for i, ts := range match.Scoreboard.TeamScoreboards {
		for _, player := range ts.Players {
			if player.UserID != nil && *player.UserID == playerID {
				playerTeamIdx = i
				break
			}
		}
		if playerTeamIdx >= 0 {
			break
		}
	}

	if playerTeamIdx < 0 {
		return -1, "unknown"
	}

	// Compare team scores
	playerTeamScore := match.Scoreboard.TeamScoreboards[playerTeamIdx].TeamScore
	opponentTeamIdx := 1 - playerTeamIdx
	opponentScore := match.Scoreboard.TeamScoreboards[opponentTeamIdx].TeamScore

	if playerTeamScore > opponentScore {
		return playerTeamIdx, "win"
	} else if playerTeamScore < opponentScore {
		return playerTeamIdx, "loss"
	}
	return playerTeamIdx, "draw"
}

// extractPlayerStats extracts individual player stats from a match
func (s *PlayerStatisticsService) extractPlayerStats(match *replay_entity.Match, playerID uuid.UUID, teamIdx int) (kills, deaths, assists, headshots, damage, rounds int) {
	if teamIdx < 0 || teamIdx >= len(match.Scoreboard.TeamScoreboards) {
		return 0, 0, 0, 0, 0, 0
	}

	ts := match.Scoreboard.TeamScoreboards[teamIdx]
	
	// Try to get stats from PlayerStats map
	if ts.PlayerStats != nil {
		if playerStats, ok := ts.PlayerStats[playerID]; ok {
			// Type assert and extract stats
			if statsMap, ok := playerStats.(map[string]interface{}); ok {
				if k, ok := statsMap["frags"].(float64); ok {
					kills = int(k)
				}
				if d, ok := statsMap["times_eliminated"].(float64); ok {
					deaths = int(d)
				}
				if a, ok := statsMap["assists"].(float64); ok {
					assists = int(a)
				}
				if h, ok := statsMap["headshots"].(float64); ok {
					headshots = int(h)
				}
				if dmg, ok := statsMap["total_damage"].(float64); ok {
					damage = int(dmg)
				}
				if r, ok := statsMap["total_rounds_played"].(float64); ok {
					rounds = int(r)
				}
			}
		}
	}

	return kills, deaths, assists, headshots, damage, rounds
}

// formatScore formats the match score as a string
func (s *PlayerStatisticsService) formatScore(match *replay_entity.Match) string {
	if len(match.Scoreboard.TeamScoreboards) < 2 {
		return "0-0"
	}
	
	score1 := match.Scoreboard.TeamScoreboards[0].TeamScore
	score2 := match.Scoreboard.TeamScoreboards[1].TeamScore
	return fmt.Sprintf("%d-%d", score1, score2)
}

// Ensure PlayerStatisticsService implements PlayerStatisticsReader
var _ squad_in.PlayerStatisticsReader = (*PlayerStatisticsService)(nil)

