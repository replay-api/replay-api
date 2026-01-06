package squad_entities

import (
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// PlayerStatistics represents aggregated statistics for a player
type PlayerStatistics struct {
	PlayerID      uuid.UUID               `json:"player_id" bson:"player_id"`
	GameID        replay_common.GameIDKey `json:"game_id" bson:"game_id"`
	TotalMatches  int                     `json:"total_matches" bson:"total_matches"`
	TotalWins     int                     `json:"total_wins" bson:"total_wins"`
	TotalLosses   int                     `json:"total_losses" bson:"total_losses"`
	TotalDraws    int                     `json:"total_draws" bson:"total_draws"`
	WinRate       float64                 `json:"win_rate" bson:"win_rate"`
	TotalKills    int                     `json:"total_kills" bson:"total_kills"`
	TotalDeaths   int                     `json:"total_deaths" bson:"total_deaths"`
	TotalAssists  int                  `json:"total_assists" bson:"total_assists"`
	KDRatio       float64              `json:"kd_ratio" bson:"kd_ratio"`
	KDARatio      float64              `json:"kda_ratio" bson:"kda_ratio"`
	TotalHeadshots int                 `json:"total_headshots" bson:"total_headshots"`
	HeadshotRate  float64              `json:"headshot_rate" bson:"headshot_rate"`
	TotalDamage   int                  `json:"total_damage" bson:"total_damage"`
	AvgDamagePerRound float64          `json:"avg_damage_per_round" bson:"avg_damage_per_round"`
	TotalRoundsPlayed int              `json:"total_rounds_played" bson:"total_rounds_played"`
	AvgRoundsPerMatch float64          `json:"avg_rounds_per_match" bson:"avg_rounds_per_match"`
	MVPCount      int                  `json:"mvp_count" bson:"mvp_count"`
	RecentMatches []RecentMatchSummary `json:"recent_matches" bson:"recent_matches"`
	LastUpdated   time.Time            `json:"last_updated" bson:"last_updated"`
}

// RecentMatchSummary represents a summary of a recent match
type RecentMatchSummary struct {
	MatchID   uuid.UUID `json:"match_id" bson:"match_id"`
	GameID    string    `json:"game_id" bson:"game_id"`
	Result    string    `json:"result" bson:"result"` // "win", "loss", "draw"
	Kills     int       `json:"kills" bson:"kills"`
	Deaths    int       `json:"deaths" bson:"deaths"`
	Assists   int       `json:"assists" bson:"assists"`
	Score     string    `json:"score" bson:"score"` // e.g., "16-14"
	PlayedAt  time.Time `json:"played_at" bson:"played_at"`
}

// NewPlayerStatistics creates a new PlayerStatistics instance
func NewPlayerStatistics(playerID uuid.UUID, gameID replay_common.GameIDKey) *PlayerStatistics {
	return &PlayerStatistics{
		PlayerID:      playerID,
		GameID:        gameID,
		RecentMatches: make([]RecentMatchSummary, 0),
		LastUpdated:   time.Now(),
	}
}

// CalculateRatios calculates derived statistics
func (ps *PlayerStatistics) CalculateRatios() {
	// Win Rate
	if ps.TotalMatches > 0 {
		ps.WinRate = float64(ps.TotalWins) / float64(ps.TotalMatches) * 100
	}

	// K/D Ratio
	if ps.TotalDeaths > 0 {
		ps.KDRatio = float64(ps.TotalKills) / float64(ps.TotalDeaths)
	} else if ps.TotalKills > 0 {
		ps.KDRatio = float64(ps.TotalKills)
	}

	// KDA Ratio
	if ps.TotalDeaths > 0 {
		ps.KDARatio = float64(ps.TotalKills+ps.TotalAssists) / float64(ps.TotalDeaths)
	} else if ps.TotalKills+ps.TotalAssists > 0 {
		ps.KDARatio = float64(ps.TotalKills + ps.TotalAssists)
	}

	// Headshot Rate
	if ps.TotalKills > 0 {
		ps.HeadshotRate = float64(ps.TotalHeadshots) / float64(ps.TotalKills) * 100
	}

	// Average Damage per Round
	if ps.TotalRoundsPlayed > 0 {
		ps.AvgDamagePerRound = float64(ps.TotalDamage) / float64(ps.TotalRoundsPlayed)
	}

	// Average Rounds per Match
	if ps.TotalMatches > 0 {
		ps.AvgRoundsPerMatch = float64(ps.TotalRoundsPlayed) / float64(ps.TotalMatches)
	}
}

// AddMatchResult adds a match result to the statistics
func (ps *PlayerStatistics) AddMatchResult(result string, kills, deaths, assists, headshots, damage, rounds int) {
	ps.TotalMatches++
	
	switch result {
	case "win":
		ps.TotalWins++
	case "loss":
		ps.TotalLosses++
	case "draw":
		ps.TotalDraws++
	}
	
	ps.TotalKills += kills
	ps.TotalDeaths += deaths
	ps.TotalAssists += assists
	ps.TotalHeadshots += headshots
	ps.TotalDamage += damage
	ps.TotalRoundsPlayed += rounds
	
	ps.CalculateRatios()
	ps.LastUpdated = time.Now()
}

