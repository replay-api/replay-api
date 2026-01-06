package entities

import (
	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type CSRoundStats struct {
	TickID replay_common.TickIDType

	MatchID          uuid.UUID
	RoundNumber      int
	WinnerTeamID     TeamIDType
	PlayerStats      []*CSPlayerStats
	TeamEconomyStats map[string]*CSTeamEconomyStats
	ClutchStats      *CSClutchStats
}

// In Discovery
type CSRoundStatSeries struct {
	StartTick replay_common.TickIDType
	EndTick   replay_common.TickIDType
	TickIDs   []replay_common.TickIDType
	// PositioningStatsByTick map[replay_common.TickIDType]*CSPositioningStats            // TODO: review, default: pegar o ultimo
	PlayersStatsByTick map[replay_common.TickIDType]map[shared.PlayerIDType]CSPlayerStats // TODO: review, default: pegar o último
	TeamStatsByTick    map[replay_common.TickIDType]map[TeamIDType]CSTeamStats            // TODO: review, default: pegar o último

	AreaStatsByTick map[replay_common.TickIDType]*CSMapRegionStats // default: pegar o último = resultado final

	StrategyStatsByTick map[replay_common.TickIDType]map[CSStrategyIDType]CSStrategyStats // geral=pegar o ultimo
	UtilityStatsByTick  map[replay_common.TickIDType]map[CSUtilityIDType]CSUtilityStats   // geral=pegar o ultimo?

	// Tick Ranges
	TickRangesOfPlayerUtilityUsage map[shared.PlayerIDType]map[CSUtilityIDType]CSTickRange
	TickRangesOfTeamUtilityUsage   map[TeamIDType]map[CSUtilityIDType]CSTickRange

	TickRangesOfPlayerStrategy map[shared.PlayerIDType]map[CSStrategyIDType][]CSTickRange
	TickRangesOfTeamStrategy   map[TeamIDType]map[CSStrategyIDType][]CSTickRange
}
