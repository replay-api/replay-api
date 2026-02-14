// Package stats provides game-agnostic statistical entities and interfaces
// that can be extended by specific game implementations (CS2, Dota2, Valorant, etc.)
package entities

import (
	"time"

	"github.com/google/uuid"
)

// GameType identifies the game for which stats are being tracked
type GameType string

const (
	GameTypeCS2      GameType = "cs2"
	GameTypeCSGO     GameType = "csgo"
	GameTypeDota2    GameType = "dota2"
	GameTypeValorant GameType = "valorant"
	GameTypeLoL      GameType = "lol"
	GameTypeOverwatch GameType = "overwatch"
)

// StatCategory represents the category of a statistic
type StatCategory string

const (
	// Universal categories across all games
	StatCategoryCombat      StatCategory = "combat"      // Kills, deaths, damage
	StatCategoryEconomy     StatCategory = "economy"     // Money, items, resources
	StatCategoryObjective   StatCategory = "objective"   // Game objectives (bomb, capture, etc.)
	StatCategoryUtility     StatCategory = "utility"     // Abilities, items, grenades
	StatCategoryTeamplay    StatCategory = "teamplay"    // Assists, trades, support
	StatCategoryPositioning StatCategory = "positioning" // Map control, rotations
	StatCategoryTiming      StatCategory = "timing"      // Reaction times, round timing
	StatCategoryPerformance StatCategory = "performance" // Ratings, impact scores
)

// ==============================================================================
// Generic Player Stats (Game-Agnostic Base)
// ==============================================================================

// GenericPlayerStats contains universal player statistics applicable to all games
type GenericPlayerStats struct {
	PlayerID    uuid.UUID `json:"player_id" bson:"player_id"`
	GameType    GameType  `json:"game_type" bson:"game_type"`
	MatchID     uuid.UUID `json:"match_id" bson:"match_id"`
	
	// Combat Stats - Universal across FPS/MOBA
	Kills         int     `json:"kills" bson:"kills"`
	Deaths        int     `json:"deaths" bson:"deaths"`
	Assists       int     `json:"assists" bson:"assists"`
	KDRatio       float64 `json:"kd_ratio" bson:"kd_ratio"`
	KDARatio      float64 `json:"kda_ratio" bson:"kda_ratio"`
	TotalDamage   int     `json:"total_damage" bson:"total_damage"`
	DamagePerUnit float64 `json:"damage_per_unit" bson:"damage_per_unit"` // ADR for FPS, DPM for MOBA
	
	// Performance Metrics - Universal
	Rating         float64 `json:"rating" bson:"rating"`           // Game-specific rating (HLTV, Dotabuff, etc.)
	ImpactScore    float64 `json:"impact_score" bson:"impact_score"`
	Participation  float64 `json:"participation" bson:"participation"` // Kill participation %
	
	// Economy Stats - Universal (CS money, Dota gold, etc.)
	ResourcesEarned  int `json:"resources_earned" bson:"resources_earned"`
	ResourcesSpent   int `json:"resources_spent" bson:"resources_spent"`
	ResourcesWasted  int `json:"resources_wasted" bson:"resources_wasted"`
	
	// Time-based Stats
	PlayTime         float64   `json:"play_time" bson:"play_time"` // Total time played in match
	TimeAlive        float64   `json:"time_alive" bson:"time_alive"`
	TimeDead         float64   `json:"time_dead" bson:"time_dead"`
	
	// Meta information
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// ==============================================================================
// Generic Team Stats (Game-Agnostic Base)
// ==============================================================================

// GenericTeamStats contains universal team statistics
type GenericTeamStats struct {
	TeamID      uuid.UUID `json:"team_id" bson:"team_id"`
	GameType    GameType  `json:"game_type" bson:"game_type"`
	MatchID     uuid.UUID `json:"match_id" bson:"match_id"`
	
	// Core team metrics
	Score           int     `json:"score" bson:"score"`
	RoundsWon       int     `json:"rounds_won" bson:"rounds_won"`
	RoundsLost      int     `json:"rounds_lost" bson:"rounds_lost"`
	TotalKills      int     `json:"total_kills" bson:"total_kills"`
	TotalDeaths     int     `json:"total_deaths" bson:"total_deaths"`
	TotalDamage     int     `json:"total_damage" bson:"total_damage"`
	
	// Economy
	TotalResourcesEarned int `json:"total_resources_earned" bson:"total_resources_earned"`
	TotalResourcesSpent  int `json:"total_resources_spent" bson:"total_resources_spent"`
	
	// Team coordination metrics
	TradeKills         int     `json:"trade_kills" bson:"trade_kills"`
	Crossfires         int     `json:"crossfires" bson:"crossfires"`
	TeamFlashes        int     `json:"team_flashes" bson:"team_flashes"` // Flashes on teammates (bad)
	SupportActions     int     `json:"support_actions" bson:"support_actions"`
	
	// Round breakdown
	PistolRoundsWon    int `json:"pistol_rounds_won" bson:"pistol_rounds_won"`
	EcoRoundsWon       int `json:"eco_rounds_won" bson:"eco_rounds_won"`
	ForceRoundsWon     int `json:"force_rounds_won" bson:"force_rounds_won"`
	FullBuyRoundsWon   int `json:"full_buy_rounds_won" bson:"full_buy_rounds_won"`
}

// ==============================================================================
// Generic Match Stats (Game-Agnostic Base)
// ==============================================================================

// GenericMatchStats contains universal match-level statistics
type GenericMatchStats struct {
	MatchID     uuid.UUID `json:"match_id" bson:"match_id"`
	GameType    GameType  `json:"game_type" bson:"game_type"`
	
	// Match outcome
	WinnerTeamID *uuid.UUID `json:"winner_team_id,omitempty" bson:"winner_team_id,omitempty"`
	IsDraw       bool       `json:"is_draw" bson:"is_draw"`
	
	// Match metadata
	Duration     float64   `json:"duration" bson:"duration"` // Duration in seconds
	TotalRounds  int       `json:"total_rounds" bson:"total_rounds"`
	MapName      string    `json:"map_name" bson:"map_name"`
	GameMode     string    `json:"game_mode" bson:"game_mode"`
	
	// Aggregate stats
	TotalKills         int `json:"total_kills" bson:"total_kills"`
	TotalDamage        int `json:"total_damage" bson:"total_damage"`
	TotalResourceFlow  int `json:"total_resource_flow" bson:"total_resource_flow"`
	
	// Match quality metrics
	AverageRating      float64 `json:"average_rating" bson:"average_rating"`
	SkillDisparity     float64 `json:"skill_disparity" bson:"skill_disparity"` // How balanced the match was
	CompetitivenessScore float64 `json:"competitiveness_score" bson:"competitiveness_score"`
	
	// Timestamps
	PlayedAt  time.Time `json:"played_at" bson:"played_at"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// ==============================================================================
// Generic Round Stats (for round-based games like CS2, Valorant)
// ==============================================================================

// GenericRoundStats contains per-round statistics
type GenericRoundStats struct {
	MatchID     uuid.UUID `json:"match_id" bson:"match_id"`
	RoundNumber int       `json:"round_number" bson:"round_number"`
	GameType    GameType  `json:"game_type" bson:"game_type"`
	
	// Round outcome
	WinnerTeamID *uuid.UUID `json:"winner_team_id,omitempty" bson:"winner_team_id,omitempty"`
	WinReason    string     `json:"win_reason" bson:"win_reason"` // elimination, objective, time
	
	// Round metrics
	Duration     float64 `json:"duration" bson:"duration"`
	TotalKills   int     `json:"total_kills" bson:"total_kills"`
	TotalDamage  int     `json:"total_damage" bson:"total_damage"`
	
	// First blood info
	FirstBloodPlayerID  *uuid.UUID `json:"first_blood_player_id,omitempty" bson:"first_blood_player_id,omitempty"`
	FirstBloodTime      float64    `json:"first_blood_time" bson:"first_blood_time"`
	
	// Economy snapshot (start of round)
	Team1Economy int `json:"team1_economy" bson:"team1_economy"`
	Team2Economy int `json:"team2_economy" bson:"team2_economy"`
}

// ==============================================================================
// Extended Stats Interfaces (Game-Specific Extensions)
// ==============================================================================

// FPSPlayerStats extends GenericPlayerStats with FPS-specific metrics
type FPSPlayerStats struct {
	GenericPlayerStats
	
	// FPS-specific combat
	Headshots       int     `json:"headshots" bson:"headshots"`
	HeadshotPct     float64 `json:"headshot_pct" bson:"headshot_pct"`
	BodyShots       int     `json:"body_shots" bson:"body_shots"`
	LegShots        int     `json:"leg_shots" bson:"leg_shots"`
	
	// Accuracy
	ShotsTotal      int     `json:"shots_total" bson:"shots_total"`
	ShotsHit        int     `json:"shots_hit" bson:"shots_hit"`
	Accuracy        float64 `json:"accuracy" bson:"accuracy"`
	
	// Opening duels
	OpeningKills    int     `json:"opening_kills" bson:"opening_kills"`
	OpeningDeaths   int     `json:"opening_deaths" bson:"opening_deaths"`
	OpeningDuelWinRate float64 `json:"opening_duel_win_rate" bson:"opening_duel_win_rate"`
	
	// Trades
	TradeKills      int     `json:"trade_kills" bson:"trade_kills"`
	WasTraded       int     `json:"was_traded" bson:"was_traded"`
	TradeEfficiency float64 `json:"trade_efficiency" bson:"trade_efficiency"`
	
	// Clutches
	ClutchAttempts  int     `json:"clutch_attempts" bson:"clutch_attempts"`
	ClutchWins      int     `json:"clutch_wins" bson:"clutch_wins"`
	ClutchWinRate   float64 `json:"clutch_win_rate" bson:"clutch_win_rate"`
	
	// Multi-kills
	DoubleKills     int `json:"double_kills" bson:"double_kills"`
	TripleKills     int `json:"triple_kills" bson:"triple_kills"`
	QuadKills       int `json:"quad_kills" bson:"quad_kills"`
	Aces            int `json:"aces" bson:"aces"`
	
	// Utility (grenades/abilities)
	UtilityDamage   int `json:"utility_damage" bson:"utility_damage"`
	EnemiesFlashed  int `json:"enemies_flashed" bson:"enemies_flashed"`
	FlashAssists    int `json:"flash_assists" bson:"flash_assists"`
	
	// Survival
	KAST            float64 `json:"kast" bson:"kast"`
	RoundsSurvived  int     `json:"rounds_survived" bson:"rounds_survived"`
	SurvivalRate    float64 `json:"survival_rate" bson:"survival_rate"`
}

// MOBAPlayerStats extends GenericPlayerStats with MOBA-specific metrics
type MOBAPlayerStats struct {
	GenericPlayerStats
	
	// MOBA-specific
	LastHits        int     `json:"last_hits" bson:"last_hits"`
	Denies          int     `json:"denies" bson:"denies"`
	GPM             float64 `json:"gpm" bson:"gpm"` // Gold per minute
	XPM             float64 `json:"xpm" bson:"xpm"` // Experience per minute
	
	// Hero/Champion specific
	HeroID          string `json:"hero_id" bson:"hero_id"`
	Level           int    `json:"level" bson:"level"`
	
	// Objectives
	TowerDamage     int `json:"tower_damage" bson:"tower_damage"`
	TowersDestroyed int `json:"towers_destroyed" bson:"towers_destroyed"`
	BossKills       int `json:"boss_kills" bson:"boss_kills"` // Roshan/Baron kills
	
	// Vision
	WardsPlaced     int `json:"wards_placed" bson:"wards_placed"`
	WardsDestroyed  int `json:"wards_destroyed" bson:"wards_destroyed"`
	
	// Healing/Support
	HealingDone     int `json:"healing_done" bson:"healing_done"`
	DamageMitigated int `json:"damage_mitigated" bson:"damage_mitigated"`
	
	// Items
	ItemsCompleted  int `json:"items_completed" bson:"items_completed"`
	NeutralItems    int `json:"neutral_items" bson:"neutral_items"`
}

// ==============================================================================
// Stat Aggregation Types
// ==============================================================================

// StatAggregation represents how a stat should be aggregated over time
type StatAggregation string

const (
	AggregationSum     StatAggregation = "sum"
	AggregationAvg     StatAggregation = "avg"
	AggregationMax     StatAggregation = "max"
	AggregationMin     StatAggregation = "min"
	AggregationLast    StatAggregation = "last"
	AggregationRate    StatAggregation = "rate"    // Per-time-unit
	AggregationPercent StatAggregation = "percent" // Percentage calculation
)

// StatDefinition describes a statistic and how it should be processed
type StatDefinition struct {
	Key         string          `json:"key" bson:"key"`
	Name        string          `json:"name" bson:"name"`
	Description string          `json:"description" bson:"description"`
	Category    StatCategory    `json:"category" bson:"category"`
	Aggregation StatAggregation `json:"aggregation" bson:"aggregation"`
	GameTypes   []GameType      `json:"game_types" bson:"game_types"` // Which games support this stat
	Unit        string          `json:"unit" bson:"unit"`             // "count", "percent", "seconds", etc.
	HigherBetter bool           `json:"higher_better" bson:"higher_better"`
}

// ==============================================================================
// Performance Benchmarks
// ==============================================================================

// PerformanceBenchmark provides context for stat evaluation
type PerformanceBenchmark struct {
	StatKey     string   `json:"stat_key" bson:"stat_key"`
	GameType    GameType `json:"game_type" bson:"game_type"`
	Rank        string   `json:"rank" bson:"rank"` // Optional: rank-specific benchmarks
	
	// Percentile thresholds
	P10         float64 `json:"p10" bson:"p10"`   // Bottom 10%
	P25         float64 `json:"p25" bson:"p25"`   // Bottom 25%
	P50         float64 `json:"p50" bson:"p50"`   // Median
	P75         float64 `json:"p75" bson:"p75"`   // Top 25%
	P90         float64 `json:"p90" bson:"p90"`   // Top 10%
	P99         float64 `json:"p99" bson:"p99"`   // Top 1%
	
	// Pro benchmarks (if available)
	ProAverage  float64 `json:"pro_average" bson:"pro_average"`
	ProTop      float64 `json:"pro_top" bson:"pro_top"`
}

// ==============================================================================
// Time Series Stats (for graphs and trends)
// ==============================================================================

// TimeSeriesPoint represents a single data point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Value     float64   `json:"value" bson:"value"`
	MatchID   uuid.UUID `json:"match_id,omitempty" bson:"match_id,omitempty"`
}

// PlayerStatsTrend tracks a player's stats over time
type PlayerStatsTrend struct {
	PlayerID  uuid.UUID         `json:"player_id" bson:"player_id"`
	GameType  GameType          `json:"game_type" bson:"game_type"`
	StatKey   string            `json:"stat_key" bson:"stat_key"`
	Period    string            `json:"period" bson:"period"` // "daily", "weekly", "monthly"
	Points    []TimeSeriesPoint `json:"points" bson:"points"`
	
	// Computed trend info
	TrendDirection string  `json:"trend_direction" bson:"trend_direction"` // "up", "down", "stable"
	TrendSlope     float64 `json:"trend_slope" bson:"trend_slope"`
	MovingAvg      float64 `json:"moving_avg" bson:"moving_avg"`
}
