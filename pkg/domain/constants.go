package common

import (
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// Gaming-specific constants
const (
	CS2_GAME_ID   = replay_common.CS2_GAME_ID
	CSGO_GAME_ID  = replay_common.CSGO_GAME_ID
	VLRNT_GAME_ID = replay_common.VLRNT_GAME_ID
)

const (
	ClutchStatTypeKey      replay_common.StatType = "Clutch"
	EconomyStatTypeKey     replay_common.StatType = "Economy"
	StrategyStatTypeKey    replay_common.StatType = "Strategy"
	PlayerStatTypeKey      replay_common.StatType = "Player"
	PositioningStatTypeKey replay_common.StatType = "Positioning"
	UtilityStatTypeKey     replay_common.StatType = "Utility"
	BattleStatTypeKey      replay_common.StatType = "Battle"
	GameSenseStatTypeKey   replay_common.StatType = "Game Sense"
	HighlightStatTypeKey   replay_common.StatType = "Highlight"
	AreaStatTypeKey        replay_common.StatType = "Area"
)

const (
	SouthAmerica_RegionIDKey replay_common.RegionIDKey = "SA"
	NorthAmerica_RegionIDKey replay_common.RegionIDKey = "NA"
	Asia_RegionIDKey         replay_common.RegionIDKey = "AS"
	Global_RegionIDKey       replay_common.RegionIDKey = "GL"
)
