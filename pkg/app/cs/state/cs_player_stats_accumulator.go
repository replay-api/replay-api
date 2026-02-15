package state

import (
	"sync"
	"time"
)

// PlayerStatsAccumulator tracks advanced player statistics during demo parsing
// These stats cannot be obtained directly from demoinfocs Player object
// and must be accumulated from individual events throughout the match
type PlayerStatsAccumulator struct {
	mu sync.RWMutex

	// Per-player accumulated stats (key: steamID64 as string)
	playerStats map[string]*AccumulatedPlayerStats

	// Round tracking for opening kills/deaths
	currentRound      int
	roundFirstKill    map[int]string // round -> killer steamID (first kill of round)
	roundFirstDeath   map[int]string // round -> victim steamID (first death of round)
	roundKillsPerPlayer map[int]map[string]int // round -> steamID -> kill count (for multi-kills)

	// Trade kill tracking - recent deaths with timestamp
	recentDeaths []RecentDeath
}

// AccumulatedPlayerStats holds stats that need to be accumulated from events
type AccumulatedPlayerStats struct {
	SteamID64      string
	Name           string
	Team           string // "CT" or "T"
	
	// Kill-based stats
	Headshots      int
	OpeningKills   int  // First kills of rounds
	OpeningDeaths  int  // First deaths of rounds
	TradeKills     int  // Kills within trade window of teammate death
	TradeDeaths    int  // Deaths that were traded by teammates
	
	// Multi-kill tracking
	MultiKillRounds  int // Rounds with 2+ kills
	DoubleKills      int // 2 kills in a round
	TripleKills      int // 3 kills in a round
	QuadKills        int // 4 kills in a round
	Aces             int // 5 kills in a round
	
	// Utility stats
	FlashAssists     int // Kills where victim was flashed by this player
	EnemiesFlashed   int // Total enemies blinded by this player
	TeamFlashes      int // Flashes on teammates (bad)
	SmokesThrown     int
	FlashesThrown    int
	HEsThrown        int
	MolotovsThrown   int
	UtilityDamage    int
	
	// Clutch stats
	ClutchAttempts int
	ClutchWins     int
	Clutch1v1Wins  int
	Clutch1v2Wins  int
	Clutch1v3Wins  int
	Clutch1v4Wins  int
	Clutch1v5Wins  int
	
	// Survival tracking for KAST
	RoundsPlayed     int
	RoundsSurvived   int
	RoundsWithKill   int
	RoundsWithAssist int
	RoundsTraded     int // Died but was traded
	
	// Damage tracking
	TotalDamage      int
	DamageByWeapon   map[string]int
	DamageByHitbox   map[string]int // "head", "body", "legs"
	SelfDamage       int
	TeamDamage       int
	
	// Weapon kills
	WeaponKills      map[string]int
	WeaponHeadshots  map[string]int
	KnifeKills       int
	
	// Accuracy (shots tracking)
	ShotsTotal       int
	ShotsHit         int
	
	// Economy tracking
	MoneySpentTotal  int
	MoneyEarnedTotal int
	
	// Bomb/objective
	BombPlants       int
	BombDefuses      int
	
	// Special kills
	WallbangKills    int
	NoScopeKills     int
	ThroughSmokeKills int
	AirborneKills    int
	BlindKills       int // Killed while flashed
}

// RecentDeath tracks a death for trade kill detection
type RecentDeath struct {
	VictimSteamID  string
	KillerSteamID  string
	VictimTeam     string
	Timestamp      time.Duration
	Round          int
	WasTraded      bool
}

const (
	// TradeWindow is the maximum time after a death for a kill to count as a trade
	TradeWindow = 5 * time.Second
)

// NewPlayerStatsAccumulator creates a new stats accumulator
func NewPlayerStatsAccumulator() *PlayerStatsAccumulator {
	return &PlayerStatsAccumulator{
		playerStats:         make(map[string]*AccumulatedPlayerStats),
		roundFirstKill:      make(map[int]string),
		roundFirstDeath:     make(map[int]string),
		roundKillsPerPlayer: make(map[int]map[string]int),
		recentDeaths:        make([]RecentDeath, 0),
		currentRound:        0,
	}
}

// GetOrCreatePlayer gets or creates stats for a player
func (a *PlayerStatsAccumulator) GetOrCreatePlayer(steamID64, name, team string) *AccumulatedPlayerStats {
	a.mu.Lock()
	defer a.mu.Unlock()

	if stats, exists := a.playerStats[steamID64]; exists {
		// Update name/team if changed
		if name != "" {
			stats.Name = name
		}
		if team != "" {
			stats.Team = team
		}
		return stats
	}

	stats := &AccumulatedPlayerStats{
		SteamID64: steamID64,
		Name:      name,
		Team:      team,
	}
	a.playerStats[steamID64] = stats
	return stats
}

// SetCurrentRound updates the current round number
func (a *PlayerStatsAccumulator) SetCurrentRound(round int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if round > a.currentRound {
		a.currentRound = round
		// Initialize kill tracking for new round
		if _, exists := a.roundKillsPerPlayer[round]; !exists {
			a.roundKillsPerPlayer[round] = make(map[string]int)
		}
	}
}

// RecordKill records a kill event and updates relevant stats
func (a *PlayerStatsAccumulator) RecordKill(
	killerSteamID, killerName, killerTeam string,
	victimSteamID, victimName, victimTeam string,
	isHeadshot bool,
	currentTime time.Duration,
	round int,
) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Ensure round is set
	if round > a.currentRound {
		a.currentRound = round
	}

	// Get or create killer stats
	killer := a.getOrCreatePlayerLocked(killerSteamID, killerName, killerTeam)
	
	// Track headshot
	if isHeadshot {
		killer.Headshots++
	}

	// Track opening kill (first kill of the round)
	if _, hasFirstKill := a.roundFirstKill[round]; !hasFirstKill {
		a.roundFirstKill[round] = killerSteamID
		killer.OpeningKills++
	}

	// Track opening death (first death of the round)
	if _, hasFirstDeath := a.roundFirstDeath[round]; !hasFirstDeath {
		a.roundFirstDeath[round] = victimSteamID
		victim := a.getOrCreatePlayerLocked(victimSteamID, victimName, victimTeam)
		victim.OpeningDeaths++
	}

	// Track multi-kills per round
	if _, exists := a.roundKillsPerPlayer[round]; !exists {
		a.roundKillsPerPlayer[round] = make(map[string]int)
	}
	a.roundKillsPerPlayer[round][killerSteamID]++

	// Check for trade kill - did killer avenge a recently dead teammate?
	for i := range a.recentDeaths {
		death := &a.recentDeaths[i]
		// Same round, teammate was killed by this victim, within trade window
		if death.Round == round &&
			death.VictimTeam == killerTeam &&
			death.KillerSteamID == victimSteamID &&
			!death.WasTraded &&
			(currentTime-death.Timestamp) <= TradeWindow {
			
			killer.TradeKills++
			death.WasTraded = true
			
			// The dead teammate was traded
			tradedPlayer := a.getOrCreatePlayerLocked(death.VictimSteamID, "", death.VictimTeam)
			tradedPlayer.RoundsTraded++
			break
		}
	}

	// Record this death for potential future trade
	a.recentDeaths = append(a.recentDeaths, RecentDeath{
		VictimSteamID: victimSteamID,
		KillerSteamID: killerSteamID,
		VictimTeam:    victimTeam,
		Timestamp:     currentTime,
		Round:         round,
		WasTraded:     false,
	})

	// Mark killer as having a kill this round
	killer.RoundsWithKill++
}

// RecordAssist records an assist
func (a *PlayerStatsAccumulator) RecordAssist(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.RoundsWithAssist++
}

// RecordFlashAssist records when a player's flash led to a kill
func (a *PlayerStatsAccumulator) RecordFlashAssist(flasherSteamID, flasherName, flasherTeam string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(flasherSteamID, flasherName, flasherTeam)
	player.FlashAssists++
}

// RecordEnemyFlashed records when a player flashes an enemy
func (a *PlayerStatsAccumulator) RecordEnemyFlashed(flasherSteamID, flasherName, flasherTeam string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(flasherSteamID, flasherName, flasherTeam)
	player.EnemiesFlashed++
}

// RecordClutchAttempt records a clutch attempt
func (a *PlayerStatsAccumulator) RecordClutchAttempt(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.ClutchAttempts++
}

// RecordClutchWin records a clutch win
func (a *PlayerStatsAccumulator) RecordClutchWin(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.ClutchWins++
}

// RecordRoundSurvival records that a player survived a round
func (a *PlayerStatsAccumulator) RecordRoundSurvival(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.RoundsSurvived++
}

// RecordRoundPlayed records that a player participated in a round
func (a *PlayerStatsAccumulator) RecordRoundPlayed(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.RoundsPlayed++
}

// GetPlayerStats returns the accumulated stats for a player
func (a *PlayerStatsAccumulator) GetPlayerStats(steamID64 string) *AccumulatedPlayerStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if stats, exists := a.playerStats[steamID64]; exists {
		// Return a copy to prevent external modification
		copy := *stats
		return &copy
	}
	return nil
}

// GetAllPlayerStats returns all accumulated player stats
func (a *PlayerStatsAccumulator) GetAllPlayerStats() map[string]*AccumulatedPlayerStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*AccumulatedPlayerStats)
	for k, v := range a.playerStats {
		copy := *v
		result[k] = &copy
	}
	return result
}

// CalculateKAST calculates KAST percentage for a player
// KAST = (Rounds with Kill + Rounds with Assist + Rounds Survived + Rounds Traded) / Total Rounds
func (a *PlayerStatsAccumulator) CalculateKAST(steamID64 string, totalRounds int) float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats, exists := a.playerStats[steamID64]
	if !exists || totalRounds == 0 {
		return 0
	}

	// Count unique rounds where player contributed
	// Note: A player can have multiple contributions in one round
	// We need to track unique rounds, which is more complex
	// For now, use a simplified calculation
	kastRounds := stats.RoundsWithKill + stats.RoundsWithAssist + stats.RoundsSurvived + stats.RoundsTraded
	
	// Cap at total rounds (in case of double counting)
	if kastRounds > totalRounds {
		kastRounds = totalRounds
	}

	return float64(kastRounds) / float64(totalRounds) * 100.0
}

// FinalizeMultiKills calculates multi-kill rounds at end of match
func (a *PlayerStatsAccumulator) FinalizeMultiKills() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, roundKills := range a.roundKillsPerPlayer {
		for steamID, kills := range roundKills {
			if stats, exists := a.playerStats[steamID]; exists {
				switch kills {
				case 2:
					stats.DoubleKills++
					stats.MultiKillRounds++
				case 3:
					stats.TripleKills++
					stats.MultiKillRounds++
				case 4:
					stats.QuadKills++
					stats.MultiKillRounds++
				case 5:
					stats.Aces++
					stats.MultiKillRounds++
				}
			}
		}
	}
}

// CleanupOldDeaths removes deaths older than trade window to save memory
func (a *PlayerStatsAccumulator) CleanupOldDeaths(currentTime time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := currentTime - (TradeWindow * 2)
	newDeaths := make([]RecentDeath, 0, len(a.recentDeaths)/2)
	
	for _, death := range a.recentDeaths {
		if death.Timestamp > cutoff {
			newDeaths = append(newDeaths, death)
		}
	}
	
	a.recentDeaths = newDeaths
}

// getOrCreatePlayerLocked gets or creates player stats (must be called with lock held)
func (a *PlayerStatsAccumulator) getOrCreatePlayerLocked(steamID64, name, team string) *AccumulatedPlayerStats {
	if stats, exists := a.playerStats[steamID64]; exists {
		if name != "" {
			stats.Name = name
		}
		if team != "" {
			stats.Team = team
		}
		return stats
	}

	stats := &AccumulatedPlayerStats{
		SteamID64:       steamID64,
		Name:            name,
		Team:            team,
		DamageByWeapon:  make(map[string]int),
		DamageByHitbox:  make(map[string]int),
		WeaponKills:     make(map[string]int),
		WeaponHeadshots: make(map[string]int),
	}
	a.playerStats[steamID64] = stats
	return stats
}

// RecordDamage records damage dealt by a player
func (a *PlayerStatsAccumulator) RecordDamage(attackerSteamID, attackerName, attackerTeam string,
	victimSteamID, victimTeam string, damage int, weapon string, hitbox string, isSelfDamage bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(attackerSteamID, attackerName, attackerTeam)
	
	if isSelfDamage {
		player.SelfDamage += damage
		return
	}
	
	// Check if team damage
	if attackerTeam == victimTeam && attackerSteamID != victimSteamID {
		player.TeamDamage += damage
		return
	}
	
	player.TotalDamage += damage
	
	if weapon != "" {
		if player.DamageByWeapon == nil {
			player.DamageByWeapon = make(map[string]int)
		}
		player.DamageByWeapon[weapon] += damage
	}
	
	if hitbox != "" {
		if player.DamageByHitbox == nil {
			player.DamageByHitbox = make(map[string]int)
		}
		player.DamageByHitbox[hitbox] += damage
	}
}

// RecordWeaponFire records a shot fired
func (a *PlayerStatsAccumulator) RecordWeaponFire(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.ShotsTotal++
}

// RecordWeaponHit records a hit
func (a *PlayerStatsAccumulator) RecordWeaponHit(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.ShotsHit++
}

// RecordGrenadeThrown records grenade usage
func (a *PlayerStatsAccumulator) RecordGrenadeThrown(steamID64, name, team string, grenadeType string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	
	switch grenadeType {
	case "flashbang":
		player.FlashesThrown++
	case "smoke", "smokegrenade":
		player.SmokesThrown++
	case "hegrenade", "he":
		player.HEsThrown++
	case "molotov", "incendiary", "molotov_projectile", "incgrenade":
		player.MolotovsThrown++
	}
}

// RecordUtilityDamage records damage from utility
func (a *PlayerStatsAccumulator) RecordUtilityDamage(steamID64, name, team string, damage int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.UtilityDamage += damage
}

// RecordTeamFlash records when a player flashes a teammate
func (a *PlayerStatsAccumulator) RecordTeamFlash(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.TeamFlashes++
}

// RecordBombPlant records a bomb plant
func (a *PlayerStatsAccumulator) RecordBombPlant(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.BombPlants++
}

// RecordBombDefuse records a bomb defuse
func (a *PlayerStatsAccumulator) RecordBombDefuse(steamID64, name, team string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.BombDefuses++
}

// RecordSpecialKill records special kill types
func (a *PlayerStatsAccumulator) RecordSpecialKill(steamID64 string, isWallbang, isNoScope, isThroughSmoke, isAirborne, isBlind bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	stats, exists := a.playerStats[steamID64]
	if !exists {
		return
	}
	
	if isWallbang {
		stats.WallbangKills++
	}
	if isNoScope {
		stats.NoScopeKills++
	}
	if isThroughSmoke {
		stats.ThroughSmokeKills++
	}
	if isAirborne {
		stats.AirborneKills++
	}
	if isBlind {
		stats.BlindKills++
	}
}

// RecordClutchWinDetails records detailed clutch information
func (a *PlayerStatsAccumulator) RecordClutchWinDetails(steamID64 string, enemiesAlive int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	stats, exists := a.playerStats[steamID64]
	if !exists {
		return
	}
	
	switch enemiesAlive {
	case 1:
		stats.Clutch1v1Wins++
	case 2:
		stats.Clutch1v2Wins++
	case 3:
		stats.Clutch1v3Wins++
	case 4:
		stats.Clutch1v4Wins++
	case 5:
		stats.Clutch1v5Wins++
	}
}

// RecordWeaponKill records a kill with a specific weapon
func (a *PlayerStatsAccumulator) RecordWeaponKill(steamID64 string, weapon string, isHeadshot bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	stats, exists := a.playerStats[steamID64]
	if !exists {
		return
	}
	
	if stats.WeaponKills == nil {
		stats.WeaponKills = make(map[string]int)
	}
	stats.WeaponKills[weapon]++
	
	if isHeadshot {
		if stats.WeaponHeadshots == nil {
			stats.WeaponHeadshots = make(map[string]int)
		}
		stats.WeaponHeadshots[weapon]++
	}
	
	if weapon == "knife" || weapon == "knife_t" || weapon == "knife_ct" {
		stats.KnifeKills++
	}
}

// RecordMoneyEvent records economy events
func (a *PlayerStatsAccumulator) RecordMoneyEvent(steamID64, name, team string, earned int, spent int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	player := a.getOrCreatePlayerLocked(steamID64, name, team)
	player.MoneyEarnedTotal += earned
	player.MoneySpentTotal += spent
}
