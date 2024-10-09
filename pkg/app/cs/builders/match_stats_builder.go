package builders

import (
	"log/slog"
	"strconv"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	cs2 "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	"github.com/psavelis/team-pro/replay-api/pkg/app/cs/state"
	cs_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/cs/entities"
)

type CS2MatchStatsBuilder struct {
	MatchContext *state.CS2MatchContext
	MatchStats   cs_entity.CSMatchStats
	Parser       dem.Parser
}

func NewCSMatchStatsBuilder(p dem.Parser, matchContext *state.CS2MatchContext) *CS2MatchStatsBuilder {
	return &CS2MatchStatsBuilder{
		Parser:       p,
		MatchContext: matchContext,
		MatchStats:   cs_entity.NewCSMatchStats(matchContext.MatchID, matchContext.ResourceOwner, len(matchContext.RoundContexts)),
	}
}

func (builder *CS2MatchStatsBuilder) WithRoundsStats(RoundContexts map[int]*state.CS2RoundContext) *CS2MatchStatsBuilder {
	for _, v := range RoundContexts {
		builder = builder.withRoundStats(v)
	}

	return builder
}

func (builder *CS2MatchStatsBuilder) StatsFromPlayerWithRound(roundNumber int, player cs2.Player) cs_entity.CSPlayerStats {
	// return e.CSPlayerStats{}

	// TODO: IMPLEMENTAR CONTEXTO DE TRADE!
	// trades := min(player.Kills(), player.Deaths(), player.Assists())

	return cs_entity.CSPlayerStats{
		NetworkPlayerID:   strconv.Itoa(player.UserID),
		TimesFragged:      player.Kills(),
		TimesEliminated:   player.Deaths(),
		Assists:           player.Assists(),
		TotalDamage:       player.TotalDamage(),
		TotalRoundsPlayed: roundNumber, // TODO: verificar se houver disconnect etc
		// DMR:               cs_entity.CalculateDMR(player.TotalDamage(), roundIndex+1),
		// Headshots:         player.Headshots(),
		LastAlivePosition: player.LastAlivePosition,
		// Inventory:         player.Inventory,
		// TradeFrags: trades,
		// DeathsToKillsRatio: cs_entity.CalculateKDR(player.Deaths(), player.Kills()),
		// KAST:       e.CalculateKAST(player.Kills(), player.Assists(), roundContext.RoundNumber-player.Deaths(), trades, roundContext.RoundNumber),
		// TODO: fork do demoinfo e adicionar esses campos:
		// 	Headshots          int        `json:"headshots" bson:"headshots"`                         // Headshots is the number of kills a player gets by shooting an enemy in the head
		// 	EntryFrags         int        `json:"entry_frags" bson:"entry_frags"`                     // EntryFrags is the number of kills a player gets when they are the first to kill an enemy
		// 	TradeFrags         int        `json:"trade_frags" bson:"trade_frags"`                     // TradeFrags is the number of kills a player gets when they are the second to kill an enemy
		// 	FirstKills         int        `json:"first_kills" bson:"first_kills"`                     // FirstKills is the number of times a player gets the first kill in a round
		// 	Clutches           int        `json:"clutches" bson:"clutches"`                           // Clutches is the number of times a player wins a round when they are the last player alive
		// 	Flashes            int        `json:"flashes" bson:"flashes"`                             // Flashes is the number of times a player flashes an enemy
		// 	DeathsToKillsRatio float64    `json:"deaths_to_kills_ratio" bson:"deaths_to_kills_ratio"` // DeathsToKillsRatio is the ratio of deaths to kills
		// }

	}
}

func (builder *CS2MatchStatsBuilder) withRoundStats(roundContext *state.CS2RoundContext) *CS2MatchStatsBuilder {
	if roundContext.RoundNumber == 0 {
		msg := "round number cannot be 0"
		slog.Error(msg)
		panic(msg)
	}

	if roundContext.RoundNumber > len(builder.MatchContext.RoundContexts) {
		msg := "round number is greater than the total number of rounds"
		slog.Error(msg)
		panic(msg)
	}

	if roundContext.RoundNumber < 0 {
		msg := "round number cannot be negative"
		slog.Error(msg)
		panic(msg)
	}

	roundIndex := roundContext.RoundNumber - 1

	clutchStats := builder.GetClutchStats(roundContext)
	playerStats := builder.GetPlayerStatsWithRound(roundContext)
	teamEconomyStats := builder.GetTeamEconomyStats(roundContext)

	builder.MatchStats.RoundsStats[roundIndex] = cs_entity.CSRoundStats{
		// WinnerNetworkTeamID: roundContext.WinnerNetworkTeamID,
		RoundNumber:      roundContext.RoundNumber,
		PlayerStats:      playerStats,
		ClutchStats:      clutchStats,
		TeamEconomyStats: teamEconomyStats,
	}

	return builder
}

// * parses and retrieves stats for economy statistic model
func (builder *CS2MatchStatsBuilder) GetTeamEconomyStats(r *state.CS2RoundContext) map[cs_entity.TeamHashIDType]*cs_entity.CSTeamEconomyStats {
	teamEconomyStats := make(map[cs_entity.TeamHashIDType]*cs_entity.CSTeamEconomyStats)

	if r.TeamCT != "" && r.TeamContext[r.TeamCT] != nil {
		teamEconomyStats[r.TeamCT] = &cs_entity.CSTeamEconomyStats{
			State: r.TeamContext[r.TeamCT].EconomyContext.State,
		}
	}

	if r.TeamT != "" && r.TeamContext[r.TeamT] != nil {
		teamEconomyStats[r.TeamT] = &cs_entity.CSTeamEconomyStats{
			State: r.TeamContext[r.TeamT].EconomyContext.State,
		}
	}

	return teamEconomyStats
}

func (builder *CS2MatchStatsBuilder) GetPlayerStatsWithRound(r *state.CS2RoundContext) []cs_entity.CSPlayerStats {
	participants := builder.Parser.GameState().Participants().All()

	playerStats := make([]cs_entity.CSPlayerStats, len(participants))

	for i, player := range participants {
		playerStats[i] = builder.StatsFromPlayerWithRound(r.RoundNumber, *player)
	}

	return playerStats
}

func (builder *CS2MatchStatsBuilder) GetClutchStats(r *state.CS2RoundContext) *cs_entity.CSClutchStats {
	var clutchStats *cs_entity.CSClutchStats

	currentClutch := r.GetClutch()
	if currentClutch != nil && currentClutch.GetPlayer() != nil {
		clutchStats = &cs_entity.CSClutchStats{
			RoundNumber:     r.RoundNumber,
			NetworkPlayerID: currentClutch.GetNetworkPlayerID(),
			Status:          r.GetClutch().Status,
			OpponentsStats:  make([]cs_entity.CSPlayerStats, len(currentClutch.GetOpponents())),
		}

		for i, opponent := range r.GetClutch().GetOpponents() {
			clutchStats.OpponentsStats[i] = builder.StatsFromPlayerWithRound(r.RoundNumber, opponent)
		}
	} else {
		clutchStats = &cs_entity.CSClutchStats{
			RoundNumber: r.RoundNumber,
			Status:      cs_entity.NotInClutchSituation,
		}
	}

	return clutchStats
}

func (builder *CS2MatchStatsBuilder) Build() cs_entity.CSMatchStats {
	return builder.MatchStats
}

func (builder *CS2MatchStatsBuilder) BuildWithHeader() *cs_entity.CSMatchStats {
	return &cs_entity.CSMatchStats{
		MatchID:       builder.MatchStats.MatchID,
		GameState:     builder.MatchStats.GameState,
		Rules:         builder.MatchStats.Rules,
		RoundsStats:   builder.MatchStats.RoundsStats,
		ResourceOwner: builder.MatchStats.ResourceOwner,
		Header:        &builder.MatchContext.Header,
	}
}
