package entities

import (
	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type CSRoundStats struct {
	TickID common.TickIDType

	MatchID          uuid.UUID
	RoundNumber      int
	WinnerTeamID     TeamIDType
	PlayerStats      []CSPlayerStats
	TeamEconomyStats map[string]*CSTeamEconomyStats
	ClutchStats      *CSClutchStats
}

// In Discovery
type CSRoundStatSeries struct {
	StartTick common.TickIDType
	EndTick   common.TickIDType
	TickIDs   []common.TickIDType
	// PositioningStatsByTick map[common.TickIDType]*CSPositioningStats            // TODO: review, default: pegar o ultimo
	PlayersStatsByTick map[common.TickIDType]map[common.PlayerIDType]CSPlayerStats // TODO: review, default: pegar o último
	TeamStatsByTick    map[common.TickIDType]map[TeamIDType]CSTeamStats            // TODO: review, default: pegar o último

	AreaStatsByTick map[common.TickIDType]*CSMapRegionStats // default: pegar o último = resultado final

	StrategyStatsByTick map[common.TickIDType]map[CSStrategyIDType]CSStrategyStats // geral=pegar o ultimo
	UtilityStatsByTick  map[common.TickIDType]map[CSUtilityIDType]CSUtilityStats   // geral=pegar o ultimo?

	// Tick Ranges
	TickRangesOfPlayerUtilityUsage map[common.PlayerIDType]map[CSUtilityIDType]CSTickRange
	TickRangesOfTeamUtilityUsage   map[TeamIDType]map[CSUtilityIDType]CSTickRange

	TickRangesOfPlayerStrategy map[common.PlayerIDType]map[CSStrategyIDType][]CSTickRange
	TickRangesOfTeamStrategy   map[TeamIDType]map[CSStrategyIDType][]CSTickRange
}
