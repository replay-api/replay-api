// Package entities provides CS2-specific stat implementations
// extending the generic stats domain
package entities

import (
	"time"

	"github.com/google/uuid"
)

// ==============================================================================
// CS2 Player Stats - Full extraction from demo files
// ==============================================================================

// CS2PlayerMatchStats contains comprehensive CS2 player statistics for a match
type CS2PlayerMatchStats struct {
	FPSPlayerStats
	
	// CS2-specific identifiers
	SteamID64     string `json:"steam_id_64" bson:"steam_id_64"`
	PlayerName    string `json:"player_name" bson:"player_name"`
	TeamSide      string `json:"team_side" bson:"team_side"` // "CT" or "T"
	
	// ====================
	// Economy Stats
	// ====================
	TotalMoneyEarned    int     `json:"total_money_earned" bson:"total_money_earned"`
	TotalMoneySpent     int     `json:"total_money_spent" bson:"total_money_spent"`
	MoneyWasted         int     `json:"money_wasted" bson:"money_wasted"` // Lost on death
	AvgLoadoutValue     float64 `json:"avg_loadout_value" bson:"avg_loadout_value"`
	PistolRoundBudget   int     `json:"pistol_round_budget" bson:"pistol_round_budget"`
	
	// Economy round performance
	EcoRoundKills       int `json:"eco_round_kills" bson:"eco_round_kills"`
	ForceBuyKills       int `json:"force_buy_kills" bson:"force_buy_kills"`
	FullBuyKills        int `json:"full_buy_kills" bson:"full_buy_kills"`
	
	// ====================
	// Weapon Stats
	// ====================
	WeaponStats         map[string]*WeaponPerformance `json:"weapon_stats" bson:"weapon_stats"`
	PrimaryWeaponKills  int    `json:"primary_weapon_kills" bson:"primary_weapon_kills"`
	SecondaryWeaponKills int   `json:"secondary_weapon_kills" bson:"secondary_weapon_kills"`
	KnifeKills          int    `json:"knife_kills" bson:"knife_kills"`
	
	// ====================
	// Damage Breakdown
	// ====================
	DamageByWeapon      map[string]int `json:"damage_by_weapon" bson:"damage_by_weapon"`
	DamageByHitbox      map[string]int `json:"damage_by_hitbox" bson:"damage_by_hitbox"`
	SelfDamage          int            `json:"self_damage" bson:"self_damage"`
	TeamDamage          int            `json:"team_damage" bson:"team_damage"`
	
	// ====================
	// Grenade/Utility Stats
	// ====================
	GrenadeStats        GrenadeStatsSummary `json:"grenade_stats" bson:"grenade_stats"`
	SmokesThrown        int     `json:"smokes_thrown" bson:"smokes_thrown"`
	FlashesThrown       int     `json:"flashes_thrown" bson:"flashes_thrown"`
	HEsThrown           int     `json:"hes_thrown" bson:"hes_thrown"`
	MolotovsThrown      int     `json:"molotovs_thrown" bson:"molotovs_thrown"`
	DecoysThrown        int     `json:"decoys_thrown" bson:"decoys_thrown"`
	
	// Flash effectiveness
	FlashDurationTotal  float64 `json:"flash_duration_total" bson:"flash_duration_total"`
	FlashDurationAvg    float64 `json:"flash_duration_avg" bson:"flash_duration_avg"`
	TeamFlashCount      int     `json:"team_flash_count" bson:"team_flash_count"` // Bad flashes
	
	// Molotov/Incendiary effectiveness
	MolotovDamage       int     `json:"molotov_damage" bson:"molotov_damage"`
	MolotovKills        int     `json:"molotov_kills" bson:"molotov_kills"`
	
	// HE Grenade effectiveness
	HEDamage            int     `json:"he_damage" bson:"he_damage"`
	HEKills             int     `json:"he_kills" bson:"he_kills"`
	
	// ====================
	// Positioning Stats
	// ====================
	AvgPositionHeatmap  []HeatmapPoint `json:"avg_position_heatmap,omitempty" bson:"avg_position_heatmap,omitempty"`
	KillPositions       []KillPosition `json:"kill_positions,omitempty" bson:"kill_positions,omitempty"`
	DeathPositions      []Position3D   `json:"death_positions,omitempty" bson:"death_positions,omitempty"`
	
	// Site presence (for CT)
	ASitePresence       float64 `json:"a_site_presence" bson:"a_site_presence"` // % time at A
	BSitePresence       float64 `json:"b_site_presence" bson:"b_site_presence"` // % time at B
	MidPresence         float64 `json:"mid_presence" bson:"mid_presence"`
	
	// ====================
	// Bomb/Objective Stats (CT/T specific)
	// ====================
	BombPlants          int     `json:"bomb_plants" bson:"bomb_plants"`
	BombDefuses         int     `json:"bomb_defuses" bson:"bomb_defuses"`
	PlantAfterKill      int     `json:"plant_after_kill" bson:"plant_after_kill"`
	DefuseUnderPressure int     `json:"defuse_under_pressure" bson:"defuse_under_pressure"` // Defuse with enemies alive
	
	// ====================
	// Round-by-Round Tracking
	// ====================
	RoundByRoundStats   []RoundPlayerStats `json:"round_by_round_stats,omitempty" bson:"round_by_round_stats,omitempty"`
	
	// ====================
	// Advanced Metrics
	// ====================
	Rating2             float64 `json:"rating_2" bson:"rating_2"`           // HLTV 2.0 rating
	KAST                float64 `json:"kast" bson:"kast"`
	ImpactRating        float64 `json:"impact_rating" bson:"impact_rating"`
	
	// Consistency metrics
	KillsStdDev         float64 `json:"kills_std_dev" bson:"kills_std_dev"`         // How consistent
	RatingStdDev        float64 `json:"rating_std_dev" bson:"rating_std_dev"`
	
	// Timing stats
	AvgTimeToFirstKill  float64 `json:"avg_time_to_first_kill" bson:"avg_time_to_first_kill"`
	AvgReactionTime     float64 `json:"avg_reaction_time" bson:"avg_reaction_time"`
}

// WeaponPerformance tracks per-weapon statistics
type WeaponPerformance struct {
	WeaponID       string  `json:"weapon_id" bson:"weapon_id"`
	WeaponName     string  `json:"weapon_name" bson:"weapon_name"`
	WeaponType     string  `json:"weapon_type" bson:"weapon_type"` // "rifle", "smg", "pistol", etc.
	
	Kills          int     `json:"kills" bson:"kills"`
	Headshots      int     `json:"headshots" bson:"headshots"`
	HeadshotPct    float64 `json:"headshot_pct" bson:"headshot_pct"`
	Damage         int     `json:"damage" bson:"damage"`
	ShotsTotal     int     `json:"shots_total" bson:"shots_total"`
	ShotsHit       int     `json:"shots_hit" bson:"shots_hit"`
	Accuracy       float64 `json:"accuracy" bson:"accuracy"`
	
	// Time with weapon
	TimeEquipped   float64 `json:"time_equipped" bson:"time_equipped"`
	TimesFired     int     `json:"times_fired" bson:"times_fired"`
}

// GrenadeStatsSummary aggregates grenade usage
type GrenadeStatsSummary struct {
	TotalThrown     int     `json:"total_thrown" bson:"total_thrown"`
	TotalDamage     int     `json:"total_damage" bson:"total_damage"`
	TotalKills      int     `json:"total_kills" bson:"total_kills"`
	AvgDamagePerNade float64 `json:"avg_damage_per_nade" bson:"avg_damage_per_nade"`
	
	// Per-type breakdown
	Flashbangs      GrenadeTypeStats `json:"flashbangs" bson:"flashbangs"`
	Smokes          GrenadeTypeStats `json:"smokes" bson:"smokes"`
	HEGrenades      GrenadeTypeStats `json:"he_grenades" bson:"he_grenades"`
	Molotovs        GrenadeTypeStats `json:"molotovs" bson:"molotovs"`
	Decoys          GrenadeTypeStats `json:"decoys" bson:"decoys"`
}

// GrenadeTypeStats for each grenade type
type GrenadeTypeStats struct {
	Count           int     `json:"count" bson:"count"`
	Damage          int     `json:"damage" bson:"damage"`
	Kills           int     `json:"kills" bson:"kills"`
	EnemiesAffected int     `json:"enemies_affected" bson:"enemies_affected"`
	AvgEffectTime   float64 `json:"avg_effect_time" bson:"avg_effect_time"` // Flash duration, smoke time, etc.
}

// Position3D represents a 3D position on the map
type Position3D struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
	Z float64 `json:"z" bson:"z"`
}

// HeatmapPoint for positional heatmaps
type HeatmapPoint struct {
	Position  Position3D `json:"position" bson:"position"`
	Intensity float64    `json:"intensity" bson:"intensity"` // Time or frequency
}

// KillPosition records where a kill happened
type KillPosition struct {
	KillerPos Position3D `json:"killer_pos" bson:"killer_pos"`
	VictimPos Position3D `json:"victim_pos" bson:"victim_pos"`
	Distance  float64    `json:"distance" bson:"distance"`
	Weapon    string     `json:"weapon" bson:"weapon"`
	Headshot  bool       `json:"headshot" bson:"headshot"`
	Tick      int        `json:"tick" bson:"tick"`
	Round     int        `json:"round" bson:"round"`
}

// RoundPlayerStats tracks per-round player performance
type RoundPlayerStats struct {
	RoundNumber     int     `json:"round_number" bson:"round_number"`
	Kills           int     `json:"kills" bson:"kills"`
	Deaths          int     `json:"deaths" bson:"deaths"`
	Assists         int     `json:"assists" bson:"assists"`
	Damage          int     `json:"damage" bson:"damage"`
	Survived        bool    `json:"survived" bson:"survived"`
	WasTraded       bool    `json:"was_traded" bson:"was_traded"`
	MoneySpent      int     `json:"money_spent" bson:"money_spent"`
	LoadoutValue    int     `json:"loadout_value" bson:"loadout_value"`
	UtilityUsed     int     `json:"utility_used" bson:"utility_used"`
	OpeningKill     bool    `json:"opening_kill" bson:"opening_kill"`
	OpeningDeath    bool    `json:"opening_death" bson:"opening_death"`
	ClutchSituation bool    `json:"clutch_situation" bson:"clutch_situation"`
	ClutchWin       bool    `json:"clutch_win" bson:"clutch_win"`
	MultiKills      int     `json:"multi_kills" bson:"multi_kills"` // 2, 3, 4, or 5
}

// ==============================================================================
// CS2 Team Match Stats
// ==============================================================================

// CS2TeamMatchStats contains comprehensive CS2 team statistics
type CS2TeamMatchStats struct {
	GenericTeamStats
	
	TeamName        string `json:"team_name" bson:"team_name"`
	TeamSide        string `json:"team_side" bson:"team_side"` // Starting side
	
	// Round wins by type
	EliminationWins     int `json:"elimination_wins" bson:"elimination_wins"`
	BombExplosionWins   int `json:"bomb_explosion_wins" bson:"bomb_explosion_wins"`
	DefuseWins          int `json:"defuse_wins" bson:"defuse_wins"`
	TimeoutWins         int `json:"timeout_wins" bson:"timeout_wins"`
	
	// Economy
	PistolRoundsWon     int     `json:"pistol_rounds_won" bson:"pistol_rounds_won"`
	PistolRoundsPlayed  int     `json:"pistol_rounds_played" bson:"pistol_rounds_played"`
	EcoRoundsWon        int     `json:"eco_rounds_won" bson:"eco_rounds_won"`
	EcoRoundsPlayed     int     `json:"eco_rounds_played" bson:"eco_rounds_played"`
	AntiEcoWins         int     `json:"anti_eco_wins" bson:"anti_eco_wins"`
	ForceBuyWins        int     `json:"force_buy_wins" bson:"force_buy_wins"`
	AvgEquipmentValue   float64 `json:"avg_equipment_value" bson:"avg_equipment_value"`
	
	// Utility usage
	TotalSmokesUsed     int `json:"total_smokes_used" bson:"total_smokes_used"`
	TotalFlashesUsed    int `json:"total_flashes_used" bson:"total_flashes_used"`
	TotalNadesUsed      int `json:"total_nades_used" bson:"total_nades_used"`
	TotalMolotovsUsed   int `json:"total_molotovs_used" bson:"total_molotovs_used"`
	
	// Site control (CT side)
	ASiteHolds          int `json:"a_site_holds" bson:"a_site_holds"`
	ASiteRetakes        int `json:"a_site_retakes" bson:"a_site_retakes"`
	BSiteHolds          int `json:"b_site_holds" bson:"b_site_holds"`
	BSiteRetakes        int `json:"b_site_retakes" bson:"b_site_retakes"`
	
	// Executes (T side)
	ASiteExecutes       int `json:"a_site_executes" bson:"a_site_executes"`
	ASiteExecuteWins    int `json:"a_site_execute_wins" bson:"a_site_execute_wins"`
	BSiteExecutes       int `json:"b_site_executes" bson:"b_site_executes"`
	BSiteExecuteWins    int `json:"b_site_execute_wins" bson:"b_site_execute_wins"`
	
	// Team coordination
	TradeKillsTeam      int     `json:"trade_kills_team" bson:"trade_kills_team"`
	TradeEfficiency     float64 `json:"trade_efficiency" bson:"trade_efficiency"`
	FlashAssistsTeam    int     `json:"flash_assists_team" bson:"flash_assists_team"`
	TeamFlashesGiven    int     `json:"team_flashes_given" bson:"team_flashes_given"` // Bad
	
	// First engagement stats
	FirstKillsTotal     int     `json:"first_kills_total" bson:"first_kills_total"`
	FirstDeathsTotal    int     `json:"first_deaths_total" bson:"first_deaths_total"`
	FirstEngagementWinRate float64 `json:"first_engagement_win_rate" bson:"first_engagement_win_rate"`
}

// ==============================================================================
// CS2 Match Analysis
// ==============================================================================

// CS2MatchAnalysis provides comprehensive match analysis
type CS2MatchAnalysis struct {
	MatchID         uuid.UUID `json:"match_id" bson:"match_id"`
	ReplayFileID    uuid.UUID `json:"replay_file_id" bson:"replay_file_id"`
	
	// Match metadata
	MapName         string    `json:"map_name" bson:"map_name"`
	Duration        float64   `json:"duration" bson:"duration"`
	TotalRounds     int       `json:"total_rounds" bson:"total_rounds"`
	OvertimeRounds  int       `json:"overtime_rounds" bson:"overtime_rounds"`
	
	// Team stats
	Team1Stats      CS2TeamMatchStats      `json:"team1_stats" bson:"team1_stats"`
	Team2Stats      CS2TeamMatchStats      `json:"team2_stats" bson:"team2_stats"`
	
	// Player stats
	PlayerStats     []CS2PlayerMatchStats  `json:"player_stats" bson:"player_stats"`
	
	// Round analysis
	RoundAnalysis   []CS2RoundAnalysis     `json:"round_analysis" bson:"round_analysis"`
	
	// Kill timeline
	KillTimeline    []KillEvent            `json:"kill_timeline" bson:"kill_timeline"`
	
	// Economy flow
	EconomyTimeline []EconomySnapshot      `json:"economy_timeline" bson:"economy_timeline"`
	
	// Match momentum
	MomentumPoints  []MomentumPoint        `json:"momentum_points" bson:"momentum_points"`
	
	// Highlights
	Highlights      []HighlightEvent       `json:"highlights" bson:"highlights"`
	
	// Analysis timestamps
	AnalyzedAt      time.Time `json:"analyzed_at" bson:"analyzed_at"`
}

// CS2RoundAnalysis provides detailed round-by-round analysis
type CS2RoundAnalysis struct {
	RoundNumber     int       `json:"round_number" bson:"round_number"`
	WinnerSide      string    `json:"winner_side" bson:"winner_side"`
	WinReason       string    `json:"win_reason" bson:"win_reason"`
	Duration        float64   `json:"duration" bson:"duration"`
	
	// Economy at round start
	CTEconomy       int       `json:"ct_economy" bson:"ct_economy"`
	TEconomy        int       `json:"t_economy" bson:"t_economy"`
	CTBuyType       string    `json:"ct_buy_type" bson:"ct_buy_type"` // "full", "force", "eco", "pistol"
	TBuyType        string    `json:"t_buy_type" bson:"t_buy_type"`
	
	// Engagement stats
	FirstKillSide   string    `json:"first_kill_side" bson:"first_kill_side"`
	FirstKillTime   float64   `json:"first_kill_time" bson:"first_kill_time"`
	TotalKills      int       `json:"total_kills" bson:"total_kills"`
	
	// Bomb events
	BombPlanted     bool      `json:"bomb_planted" bson:"bomb_planted"`
	PlantSite       string    `json:"plant_site" bson:"plant_site"`
	PlantTime       float64   `json:"plant_time" bson:"plant_time"`
	BombDefused     bool      `json:"bomb_defused" bson:"bomb_defused"`
	
	// Player performance
	RoundMVP        string    `json:"round_mvp" bson:"round_mvp"` // Player ID
	ClutchPlayer    string    `json:"clutch_player,omitempty" bson:"clutch_player,omitempty"`
	ClutchVs        int       `json:"clutch_vs,omitempty" bson:"clutch_vs,omitempty"` // 1v2, 1v3, etc.
	ClutchWon       bool      `json:"clutch_won" bson:"clutch_won"`
}

// KillEvent represents a kill in the timeline
type KillEvent struct {
	Tick            int        `json:"tick" bson:"tick"`
	Round           int        `json:"round" bson:"round"`
	Time            float64    `json:"time" bson:"time"` // Time in round
	KillerID        string     `json:"killer_id" bson:"killer_id"`
	KillerName      string     `json:"killer_name" bson:"killer_name"`
	KillerSide      string     `json:"killer_side" bson:"killer_side"`
	VictimID        string     `json:"victim_id" bson:"victim_id"`
	VictimName      string     `json:"victim_name" bson:"victim_name"`
	VictimSide      string     `json:"victim_side" bson:"victim_side"`
	Weapon          string     `json:"weapon" bson:"weapon"`
	Headshot        bool       `json:"headshot" bson:"headshot"`
	Wallbang        bool       `json:"wallbang" bson:"wallbang"`
	NoScope         bool       `json:"no_scope" bson:"no_scope"`
	ThroughSmoke    bool       `json:"through_smoke" bson:"through_smoke"`
	Flashed         bool       `json:"flashed" bson:"flashed"` // Victim was flashed
	IsOpeningKill   bool       `json:"is_opening_kill" bson:"is_opening_kill"`
	IsTrade         bool       `json:"is_trade" bson:"is_trade"`
	AssisterID      string     `json:"assister_id,omitempty" bson:"assister_id,omitempty"`
	FlashAssisterID string     `json:"flash_assister_id,omitempty" bson:"flash_assister_id,omitempty"`
	KillerPos       Position3D `json:"killer_pos" bson:"killer_pos"`
	VictimPos       Position3D `json:"victim_pos" bson:"victim_pos"`
	Distance        float64    `json:"distance" bson:"distance"`
}

// EconomySnapshot captures team economy at a point in time
type EconomySnapshot struct {
	Round           int   `json:"round" bson:"round"`
	CTTotalMoney    int   `json:"ct_total_money" bson:"ct_total_money"`
	TTotalMoney     int   `json:"t_total_money" bson:"t_total_money"`
	CTSpent         int   `json:"ct_spent" bson:"ct_spent"`
	TSpent          int   `json:"t_spent" bson:"t_spent"`
	CTEquipValue    int   `json:"ct_equip_value" bson:"ct_equip_value"`
	TEquipValue     int   `json:"t_equip_value" bson:"t_equip_value"`
	CTLossBonus     int   `json:"ct_loss_bonus" bson:"ct_loss_bonus"`
	TLossBonus      int   `json:"t_loss_bonus" bson:"t_loss_bonus"`
}

// MomentumPoint tracks match momentum
type MomentumPoint struct {
	Round           int     `json:"round" bson:"round"`
	Score           string  `json:"score" bson:"score"` // "8-5" format
	Momentum        float64 `json:"momentum" bson:"momentum"` // -1 to 1, negative = T, positive = CT
	WinStreak       int     `json:"win_streak" bson:"win_streak"` // Current win streak
	WinStreakSide   string  `json:"win_streak_side" bson:"win_streak_side"`
}

// HighlightEvent for notable moments
type HighlightEvent struct {
	Type            string  `json:"type" bson:"type"` // "ace", "clutch", "multi_kill", "eco_win", etc.
	Round           int     `json:"round" bson:"round"`
	Tick            int     `json:"tick" bson:"tick"`
	PlayerID        string  `json:"player_id" bson:"player_id"`
	PlayerName      string  `json:"player_name" bson:"player_name"`
	Description     string  `json:"description" bson:"description"`
	Significance    float64 `json:"significance" bson:"significance"` // 0-1 how impressive
}
