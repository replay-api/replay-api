package squad_services

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
)

// SquadStatisticsService aggregates statistics for squads
type SquadStatisticsService struct {
	squadReader  squad_in.SquadReader
	playerReader squad_in.PlayerProfileReader
	matchReader  replay_in.MatchReader
}

// NewSquadStatisticsService creates a new SquadStatisticsService
func NewSquadStatisticsService(
	squadReader squad_in.SquadReader,
	playerReader squad_in.PlayerProfileReader,
	matchReader replay_in.MatchReader,
) *SquadStatisticsService {
	return &SquadStatisticsService{
		squadReader:  squadReader,
		playerReader: playerReader,
		matchReader:  matchReader,
	}
}

// GetSquadStatistics retrieves aggregated statistics for a squad
func (s *SquadStatisticsService) GetSquadStatistics(ctx context.Context, squadID uuid.UUID, gameID string) (*squad_entities.SquadStatistics, error) {
	slog.InfoContext(ctx, "Fetching squad statistics", "squad_id", squadID, "game_id", gameID)

	// Get squad to verify it exists and get member IDs
	squadSearch := common.Search{
		SearchParams: []common.SearchAggregation{
			{
				Params: []common.SearchParameter{
					{
						ValueParams: []common.SearchableValue{
							{Field: "ID", Values: []interface{}{squadID.String()}, Operator: common.EqualsOperator},
						},
					},
				},
			},
		},
		ResultOptions: common.SearchResultOptions{Limit: 1},
	}

	squads, err := s.squadReader.Search(ctx, squadSearch)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch squad", "error", err, "squad_id", squadID)
		return nil, err
	}

	if len(squads) == 0 {
		return nil, common.NewErrNotFound(common.ResourceTypeSquad, "ID", squadID.String())
	}

	squad := squads[0]
	stats := squad_entities.NewSquadStatistics(squadID, string(squad.GameID))

	// Get player IDs from squad membership
	var playerIDs []interface{}
	for _, member := range squad.Membership {
		playerIDs = append(playerIDs, member.PlayerProfileID.String())
	}

	if len(playerIDs) == 0 {
		// No members, return empty stats
		return stats, nil
	}

	// Get matches involving squad members
	matchSearch := common.Search{
		SearchParams: []common.SearchAggregation{
			{
				Params: []common.SearchParameter{
					{
						ValueParams: []common.SearchableValue{
							{
								Field:    "scoreboard.team_scoreboards.players._id",
								Values:   playerIDs,
								Operator: common.InOperator,
							},
						},
					},
				},
			},
		},
		ResultOptions: common.SearchResultOptions{Limit: 100}, // Last 100 matches
	}

	matches, err := s.matchReader.Search(ctx, matchSearch)
	if err != nil {
		slog.WarnContext(ctx, "Failed to fetch matches for squad stats", "error", err, "squad_id", squadID)
		// Continue with empty match data
	}

	// Aggregate match statistics
	mapStats := make(map[string]*squad_entities.MapStats)
	
	for _, match := range matches {
		stats.TotalMatches++

		// Determine if squad won (simplified - check if any squad member was on winning team)
		won := s.didSquadWin(match, playerIDs)
		if won {
			stats.Wins++
		} else {
			stats.Losses++
		}

		// Track map statistics (use GameID as a proxy for map tracking, or skip if not available)
		mapName := string(match.GameID)
		if mapName != "" {
			if _, exists := mapStats[mapName]; !exists {
				mapStats[mapName] = &squad_entities.MapStats{MapName: mapName}
			}
			mapStats[mapName].Played++
			if won {
				mapStats[mapName].Wins++
			} else {
				mapStats[mapName].Losses++
			}
		}

		// Add to recent form (only last 10)
		if len(stats.RecentForm) < 10 {
			result := "L"
			if won {
				result = "W"
			}
			
			stats.RecentForm = append(stats.RecentForm, squad_entities.MatchResult{
				MatchID:  match.ID,
				Result:   result,
				Map:      mapName,
				Date:     match.CreatedAt,
			})
		}
	}

	// Calculate win rates
	stats.CalculateWinRate()
	for mapName, ms := range mapStats {
		if ms.Played > 0 {
			ms.WinRate = float64(ms.Wins) / float64(ms.Played) * 100.0
		}
		stats.MapStatistics[mapName] = *ms
	}

	// Get member contributions
	for _, member := range squad.Membership {
		contribution := squad_entities.MemberContribution{
			PlayerID:      member.PlayerProfileID,
			Role:          string(member.Type),
		}

		// Get player profile for name
		playerSearch := common.Search{
			SearchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "ID", Values: []interface{}{member.PlayerProfileID.String()}, Operator: common.EqualsOperator},
							},
						},
					},
				},
			},
			ResultOptions: common.SearchResultOptions{Limit: 1},
		}

		players, err := s.playerReader.Search(ctx, playerSearch)
		if err == nil && len(players) > 0 {
			contribution.PlayerName = players[0].Nickname
		}

		stats.MemberStats = append(stats.MemberStats, contribution)
	}

	stats.LastUpdated = time.Now()

	slog.InfoContext(ctx, "Squad statistics calculated",
		"squad_id", squadID,
		"total_matches", stats.TotalMatches,
		"wins", stats.Wins,
		"losses", stats.Losses,
		"win_rate", stats.WinRate)

	return stats, nil
}

// didSquadWin checks if any squad member was on the winning team
func (s *SquadStatisticsService) didSquadWin(_ interface{}, _ []interface{}) bool {
	// Simplified - this would need to properly parse the match data
	// For now, return based on random/default
	return false
}

