package squad_entities

import (
	"time"

	"github.com/google/uuid"
)

// SquadStatistics aggregates performance data for a squad
type SquadStatistics struct {
	SquadID         uuid.UUID              `json:"squad_id" bson:"squad_id"`
	GameID          string                 `json:"game_id" bson:"game_id"`
	TotalMatches    int                    `json:"total_matches" bson:"total_matches"`
	Wins            int                    `json:"wins" bson:"wins"`
	Losses          int                    `json:"losses" bson:"losses"`
	Draws           int                    `json:"draws" bson:"draws"`
	WinRate         float64                `json:"win_rate" bson:"win_rate"`
	CurrentStreak   int                    `json:"current_streak" bson:"current_streak"` // Positive for win streak, negative for loss streak
	LongestWinStreak int                   `json:"longest_win_streak" bson:"longest_win_streak"`
	TournamentWins  int                    `json:"tournament_wins" bson:"tournament_wins"`
	TournamentTop3  int                    `json:"tournament_top_3" bson:"tournament_top_3"`
	TotalPrizeWon   float64                `json:"total_prize_won" bson:"total_prize_won"`
	PrizeCurrency   string                 `json:"prize_currency" bson:"prize_currency"`
	AverageRating   float64                `json:"average_rating" bson:"average_rating"`
	HighestRating   float64                `json:"highest_rating" bson:"highest_rating"`
	MemberStats     []MemberContribution   `json:"member_stats" bson:"member_stats"`
	RecentForm      []MatchResult          `json:"recent_form" bson:"recent_form"` // Last 5-10 matches
	MapStatistics   map[string]MapStats    `json:"map_statistics" bson:"map_statistics"`
	LastUpdated     time.Time              `json:"last_updated" bson:"last_updated"`
}

// MemberContribution shows individual member contributions to squad performance
type MemberContribution struct {
	PlayerID      uuid.UUID `json:"player_id" bson:"player_id"`
	PlayerName    string    `json:"player_name" bson:"player_name"`
	MatchesPlayed int       `json:"matches_played" bson:"matches_played"`
	WinRate       float64   `json:"win_rate" bson:"win_rate"`
	Rating        float64   `json:"rating" bson:"rating"`
	Role          string    `json:"role" bson:"role"`
}

// MatchResult represents a simplified match result for form display
type MatchResult struct {
	MatchID   uuid.UUID `json:"match_id" bson:"match_id"`
	Opponent  string    `json:"opponent" bson:"opponent"`
	Result    string    `json:"result" bson:"result"` // "W", "L", "D"
	Score     string    `json:"score" bson:"score"`   // e.g., "16-12"
	Map       string    `json:"map" bson:"map"`
	Date      time.Time `json:"date" bson:"date"`
}

// MapStats tracks performance on specific maps
type MapStats struct {
	MapName    string  `json:"map_name" bson:"map_name"`
	Played     int     `json:"played" bson:"played"`
	Wins       int     `json:"wins" bson:"wins"`
	Losses     int     `json:"losses" bson:"losses"`
	WinRate    float64 `json:"win_rate" bson:"win_rate"`
}

// NewSquadStatistics creates a new empty SquadStatistics
func NewSquadStatistics(squadID uuid.UUID, gameID string) *SquadStatistics {
	return &SquadStatistics{
		SquadID:       squadID,
		GameID:        gameID,
		TotalMatches:  0,
		Wins:          0,
		Losses:        0,
		Draws:         0,
		WinRate:       0.0,
		CurrentStreak: 0,
		MemberStats:   []MemberContribution{},
		RecentForm:    []MatchResult{},
		MapStatistics: make(map[string]MapStats),
		LastUpdated:   time.Now(),
	}
}

// CalculateWinRate computes the win rate
func (s *SquadStatistics) CalculateWinRate() {
	if s.TotalMatches > 0 {
		s.WinRate = float64(s.Wins) / float64(s.TotalMatches) * 100.0
	}
}

// AddMatchResult adds a match to recent form
func (s *SquadStatistics) AddMatchResult(result MatchResult) {
	s.RecentForm = append([]MatchResult{result}, s.RecentForm...)
	// Keep only last 10 matches
	if len(s.RecentForm) > 10 {
		s.RecentForm = s.RecentForm[:10]
	}
}

// UpdateStreak updates the win/loss streak
func (s *SquadStatistics) UpdateStreak(won bool) {
	if won {
		if s.CurrentStreak >= 0 {
			s.CurrentStreak++
			if s.CurrentStreak > s.LongestWinStreak {
				s.LongestWinStreak = s.CurrentStreak
			}
		} else {
			s.CurrentStreak = 1
		}
	} else {
		if s.CurrentStreak <= 0 {
			s.CurrentStreak--
		} else {
			s.CurrentStreak = -1
		}
	}
}

