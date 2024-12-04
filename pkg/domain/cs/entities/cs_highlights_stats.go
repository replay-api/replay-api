package entities

type CSHighlightStats struct {
	HighlightNumber  int
	Tick             float64
	HighlightedEvent interface{}
	LongStartTick    *float64
	LongEndTick      *float64
	ShortStartTick   *float64
	ShortEndTick     *float64
	EventStats       []CSEventStats // events per tick map[tick][]event

	// ClutchStats     CSClutchStats
	// EconomyStats    CSEconomyStats
	// PositioningStats []CSPositioningStats
	// WeaponStats []CSWeaponStats
	// DamageStats []CSDamageStats
	// ClutchStats []CSClutchStats
	// UtilityStats []CSUtilityStats
	// EconomyStats []CSEconomyStats
	// RoundStats []CSRoundStats
	// PlayerStats []CSPlayerStats
	// TeamStats []CSTeamStats
	// MatchStats []CSMatchStats
	// GameStats []CSGameStats

}
