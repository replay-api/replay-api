package common

import "github.com/google/uuid"

var (
	// Default TenantID for TeamPRO (random and valid UUID)
	TeamPROTenantID = uuid.MustParse("a3a80810-f91c-4391-9eff-6d47a13bebde")

	// Default ClientID for TeamPRO (random and valid UUID)
	TeamPROAppClientID = uuid.MustParse("ff96c01f-a741-4429-a0cd-2868d408c42f")

	ServerClientID = uuid.MustParse("ff96c01f-a741-4429-a0cd-2868d408c42f")
)

const (
	// semantic aliases
	ALLOW = true
	DENY  = false
)

type StatType = string

const (
	ClutchStatTypeKey      StatType = "Clutch"
	EconomyStatTypeKey     StatType = "Economy"
	StrategyStatTypeKey    StatType = "Strategy"
	PlayerStatTypeKey      StatType = "Player"
	PositioningStatTypeKey StatType = "Positioning"
	UtilityStatTypeKey     StatType = "Utility"
	BattleStatTypeKey      StatType = "Battle"
	GameSenseStatTypeKey   StatType = "Game Sense"
	HighlightStatTypeKey   StatType = "Highlight"
	AreaStatTypeKey        StatType = "Area" // TODO: redundancia com MapRegion?
)

type RegionIDKey = string

const (
	SouthAmerica_RegionIDKey RegionIDKey = "SA"
	NorthAmerica_RegionIDKey RegionIDKey = "NA"
	Asia_RegionIDKey         RegionIDKey = "AS"
	Global_RegionIDKey       RegionIDKey = "GL"
	// TODO: espelhar do cs2 ou usar mais granular como por server msm?
)
