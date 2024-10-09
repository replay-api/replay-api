package state

import (
	"time"

	cs2 "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	cs_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/cs/entities"
)

type CS2StrategyContext struct {
	Team         *cs2.Team
	CurrentSide  cs_entity.CSTeamSideIDType
	Map          string
	RoundNumber  int
	RoundEconomy *CSEconomyContext

	// Bombsite Play
	BombsiteAggression        map[string]int             // Maps bombsite to aggression level
	BombsitePlants            map[string]int             // Number of plants per bombsite
	SuccessfulBombsiteEntries map[string]int             // Successful entries per bombsite (kills on site entry)
	BombsiteEntryTimings      map[string][]time.Duration // Timings of bombsite entries (for timing analysis)
	BombsiteControlTimings    map[string][]time.Duration // How long a team held a site after taking control
	BombsiteDefenses          map[string]int             // Number of successful defenses per bombsite (for CT side)

	// Map Control
	MapControl map[cs_entity.CSAreaIDType]float64 // Zone control percentage
	// MapControlTimings map[string][]TimeRange // When a team had control of a zone

	// Utility Usage
	UtilityUsage        map[string]int           // Count per utility type
	UtilitySuccess      map[string]int           // Successful utility usage (e.g., flash that got a kill)
	UtilityCoordination map[string][]*cs2.Player // Players involved in coordinated utility plays

	// Player-Specific
	PlayerPositions  map[int]string // Last known position (granular zones if possible)
	PlayerRoles      map[int]string // Inferred roles (e.g., "entry fragger", "support", "lurker")
	PlayerAggression map[int]int    // Tendency to engage in duels

	// Round Outcomes
	RoundWinConditions map[string]int // Count of different win conditions (e.g., "Elimination", "Bomb Explosion", "Time")
	RoundLossReasons   map[string]int // Count of loss reasons (e.g., "Elimination", "Bomb Defused", "Time")

	// General Strategies (if detectable)
	UsedStrategies       []string       // List of identified strategies (e.g., "Fast A", "Default B")
	StrategySuccess      map[string]int // Win rate for each strategy
	MostFrequentStrategy string

	// Additional Insights
	AverageTimeToSite time.Duration // Average time taken to reach a bombsite (for T side)
	RetakeSuccessRate float64       // Percentage of successful retakes (for CT side)

	// ... other fields you deem relevant ...
}

// ... (constructor and methods remain similar, but need to update to handle new fields) ...

func NewCS2StrategyContext(team *cs2.Team, currentSide cs_entity.CSTeamSideIDType, mapName string) *CS2StrategyContext {
	return &CS2StrategyContext{
		Team:               team,
		CurrentSide:        currentSide,
		Map:                mapName,
		RoundNumber:        0,   // Initialize to 0, will be updated as rounds progress
		RoundEconomy:       nil, // Initialized later when round starts
		BombsiteAggression: make(map[string]int),
		MapControl:         make(map[cs_entity.CSAreaIDType]float64),
		UtilityUsage:       make(map[string]int),
		PlayerPositions:    make(map[int]string),
	}
}

// UpdateBombsiteAggression increments aggression for a bombsite (e.g., after a successful attack)
func (ctx *CS2StrategyContext) UpdateBombsiteAggression(bombsite string, increment int) {
	ctx.BombsiteAggression[bombsite] += increment
}

// UpdateMapControl updates the control percentage for a map zone
func (ctx *CS2StrategyContext) UpdateMapControl(zone cs_entity.CSAreaIDType, control float64) {
	ctx.MapControl[zone] = control
}

// UpdateUtilityUsage increments the usage count for a specific utility type
func (ctx *CS2StrategyContext) UpdateUtilityUsage(utilityType string) {
	ctx.UtilityUsage[utilityType]++
}

// UpdatePlayerPositions updates the position of a player on the map
func (ctx *CS2StrategyContext) UpdatePlayerPositions(player *cs2.Player, position string) {
	ctx.PlayerPositions[player.UserID] = position
}

// ... (Add other methods to analyze and update the context based on game events)
