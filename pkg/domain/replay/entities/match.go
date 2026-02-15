package entities

import (
	"context"
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type MatchVisibility string

const (
	MatchVisibilityPublic  MatchVisibility = "public"
	MatchVisibilitySquad   MatchVisibility = "squad"
	MatchVisibilityPrivate MatchVisibility = "private"
	MatchVisibilityCustom  MatchVisibility = "custom"
)

// MatchStatus represents the processing status of a match
type MatchStatus string

const (
	MatchStatusPending    MatchStatus = "pending"
	MatchStatusProcessing MatchStatus = "processing"
	MatchStatusCompleted  MatchStatus = "completed"
	MatchStatusFailed     MatchStatus = "failed"
)

// MatchSource indicates how the match was created/sourced
type MatchSource string

const (
	// MatchSourceReplay - Match created from replay file processing
	MatchSourceReplay MatchSource = "replay"
	// MatchSourceMatchmaking - Match created via matchmaking system
	MatchSourceMatchmaking MatchSource = "matchmaking"
	// MatchSourceExternalAPI - Match data imported from external API (e.g., Valve, FACEIT)
	MatchSourceExternalAPI MatchSource = "external_api"
	// MatchSourceManual - Match manually created by user/admin
	MatchSourceManual MatchSource = "manual"
)

// AggregteRoot
type Match struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	RegionID          replay_common.RegionIDKey `json:"region_id" bson:"region_id"`
	ReplayFileID      uuid.UUID                 `json:"replay_file_id" bson:"replay_file_id"`
	GameID            replay_common.GameIDKey   `json:"game_id" bson:"game_id"`
	MapName           string                    `json:"map_name,omitempty" bson:"map_name,omitempty"`
	Duration          float64                   `json:"duration,omitempty" bson:"duration,omitempty"` // Duration in seconds
	PlayedAt          time.Time                 `json:"played_at,omitempty" bson:"played_at,omitempty"` // When the match was played
	Mode              string                    `json:"mode,omitempty" bson:"mode,omitempty"`         // e.g., "competitive", "casual"
	Status            MatchStatus               `json:"status,omitempty" bson:"status,omitempty"`
	ServerName        string                    `json:"server_name,omitempty" bson:"server_name,omitempty"`
	Scoreboard        Scoreboard                `json:"scoreboard" bson:"scoreboard"`
	Teams             []Team                    `json:"teams" bson:"teams"`
	EventCount        int                       `json:"event_count" bson:"event_count"`
	Visibility        MatchVisibility           `json:"visibility" bson:"visibility"`
	ShareTokens       []ShareToken              `json:"share_tokens" bson:"share_tokens"`
	// Source tracking: how was this match created?
	Source        MatchSource `json:"source" bson:"source"`                                       // Source of match data (replay, matchmaking, external_api, manual)
	LinkedReplayID *uuid.UUID `json:"linked_replay_id,omitempty" bson:"linked_replay_id,omitempty"` // For matchmaking matches, links to associated replay if available
	ExternalMatchID string    `json:"external_match_id,omitempty" bson:"external_match_id,omitempty"` // External system match ID (e.g., FACEIT match ID, Valve match ID)
}

func (m Match) GetID() uuid.UUID {
	return m.ID
}

type Scoreboard struct {
	TeamScoreboards []TeamScoreboard `json:"team_scoreboards" bson:"team_scoreboards"`
	MatchMVP        *PlayerMetadata  `json:"match_mvp" bson:"match_mvp"`
}

// PlayerStatsEntry contains comprehensive esports stats for a single player
type PlayerStatsEntry struct {
	PlayerID      string  `json:"player_id" bson:"player_id"`
	Kills         int     `json:"kills" bson:"kills"`
	Deaths        int     `json:"deaths" bson:"deaths"`
	Assists       int     `json:"assists" bson:"assists"`
	KDRatio       float64 `json:"kd_ratio" bson:"kd_ratio"`
	Headshots     int     `json:"headshots" bson:"headshots"`
	HeadshotPct   float64 `json:"headshot_pct" bson:"headshot_pct"`       // Headshot percentage
	TotalDamage   int     `json:"total_damage" bson:"total_damage"`
	UtilityDamage int     `json:"utility_damage" bson:"utility_damage"`
	ADR           float64 `json:"adr" bson:"adr"`
	MVPCount      int     `json:"mvp_count" bson:"mvp_count"`
	Score         int     `json:"score" bson:"score"`
	// Advanced esports stats
	KAST            float64 `json:"kast" bson:"kast"`                           // Kill/Assist/Survived/Traded %
	ImpactRating    float64 `json:"impact_rating" bson:"impact_rating"`         // Impact rating (HLTV-style)
	OpeningKills    int     `json:"opening_kills" bson:"opening_kills"`         // First kills of a round
	OpeningDeaths   int     `json:"opening_deaths" bson:"opening_deaths"`       // First deaths of a round
	TradeKills      int     `json:"trade_kills" bson:"trade_kills"`             // Kills avenging teammates
	ClutchWins      int     `json:"clutch_wins" bson:"clutch_wins"`             // 1vX clutch wins
	ClutchAttempts  int     `json:"clutch_attempts" bson:"clutch_attempts"`     // 1vX clutch attempts
	FlashAssists    int     `json:"flash_assists" bson:"flash_assists"`         // Kills assisted by flashes
	EnemiesFlashed  int     `json:"enemies_flashed" bson:"enemies_flashed"`     // Total enemies flashed
	EntryAttempts   int     `json:"entry_attempts" bson:"entry_attempts"`       // Entry duel attempts
	EntrySuccesses  int     `json:"entry_successes" bson:"entry_successes"`     // Entry duel wins
	MultiKills      int     `json:"multi_kills" bson:"multi_kills"`             // 2+ kills in a round
	Rating2         float64 `json:"rating_2" bson:"rating_2"`                   // HLTV 2.0 rating
}

type TeamScoreboard struct {
	Team        Team               `json:"team" bson:"team"`
	Side        string             `json:"side" bson:"side"`
	TeamScore   int                `json:"team_score" bson:"team_score"`
	TeamMVP     *PlayerMetadata    `json:"team_mvp" bson:"team_mvp"`
	Players     []PlayerMetadata   `json:"players" bson:"players"`
	PlayerStats []PlayerStatsEntry `json:"player_stats" bson:"player_stats"`
	Rounds      []RoundInfo        `json:"rounds" bson:"rounds"`
	RoundStats  map[int]interface{} `json:"round_stats" bson:"round_stats"`
}

type RoundInfo struct {
	RoundNumber      int         `json:"round_number" bson:"round_number"`
	WinnerTeamID     *uuid.UUID  `json:"winner" bson:"winner"`
	RoundMVPPlayerID *uuid.UUID  `json:"round_mvp_player_id" bson:"round_mvp_player_id"`
	Events           []GameEvent `json:"events" bson:"events"`
}

func NewCS2Match(userContext context.Context, replayFileID uuid.UUID) *Match {
	resourceOwner := shared.GetResourceOwner(userContext)
	// Use NewUnrestrictedEntity for public visibility by default
	entity := shared.NewUnrestrictedEntity(resourceOwner)
	return &Match{
		BaseEntity:   entity,
		ReplayFileID: replayFileID,
		GameID:       replay_common.CS2.ID,
		Source:       MatchSourceReplay, // Matches from replay processing
	}
}

func NewCS2MatchWithOwner(resourceOwner shared.ResourceOwner, replayFileID uuid.UUID) *Match {
	// Use NewUnrestrictedEntity for public visibility by default
	entity := shared.NewUnrestrictedEntity(resourceOwner)
	return &Match{
		BaseEntity:   entity,
		ReplayFileID: replayFileID,
		GameID:       replay_common.CS2.ID,
		Source:       MatchSourceReplay, // Matches from replay processing
	}
}

// NewMatchFromMatchmaking creates a match from the matchmaking system
func NewMatchFromMatchmaking(resourceOwner shared.ResourceOwner, gameID replay_common.GameIDKey, externalMatchID string) *Match {
	entity := shared.NewUnrestrictedEntity(resourceOwner)
	return &Match{
		BaseEntity:      entity,
		GameID:          gameID,
		Source:          MatchSourceMatchmaking,
		ExternalMatchID: externalMatchID,
		Status:          MatchStatusPending,
	}
}

// NewMatchFromExternalAPI creates a match from external API data (FACEIT, Valve, etc.)
func NewMatchFromExternalAPI(resourceOwner shared.ResourceOwner, gameID replay_common.GameIDKey, externalMatchID string) *Match {
	entity := shared.NewUnrestrictedEntity(resourceOwner)
	return &Match{
		BaseEntity:      entity,
		GameID:          gameID,
		Source:          MatchSourceExternalAPI,
		ExternalMatchID: externalMatchID,
		Status:          MatchStatusCompleted,
	}
}

// LinkReplay links a replay file to a matchmaking or external API match
func (m *Match) LinkReplay(replayID uuid.UUID) {
	m.LinkedReplayID = &replayID
	m.ReplayFileID = replayID // Also set the main replay file ID for backwards compatibility
}

// HasReplay returns true if the match has an associated replay file
func (m *Match) HasReplay() bool {
	return m.ReplayFileID != uuid.Nil || (m.LinkedReplayID != nil && *m.LinkedReplayID != uuid.Nil)
}

// FinalScoreboardPayload contains the final match scoreboard data
// This is the canonical type used for scoreboard extraction during replay processing
type FinalScoreboardPayload struct {
	CTScore      int                          `json:"ct_score"`
	TScore       int                          `json:"t_score"`
	CTTeamName   string                       `json:"ct_team_name"`
	TTeamName    string                       `json:"t_team_name"`
	TotalRounds  int                          `json:"total_rounds"`
	WinnerSide   string                       `json:"winner_side"`
	Players      []PlayerScoreboardDataEntry  `json:"players"`
	Duration     float64                      `json:"duration"`
	MapName      string                       `json:"map_name"`
}

// PlayerScoreboardDataEntry contains individual player statistics for scoreboard
type PlayerScoreboardDataEntry struct {
	NetworkUserID  string  `json:"network_user_id"`
	Name           string  `json:"name"`
	Team           string  `json:"team"`
	Side           string  `json:"side"` // "CT" or "T"
	Kills          int     `json:"kills"`
	Deaths         int     `json:"deaths"`
	Assists        int     `json:"assists"`
	KDRatio        float64 `json:"kd_ratio"`
	Headshots      int     `json:"headshots"`
	TotalDamage    int     `json:"total_damage"`
	UtilityDamage  int     `json:"utility_damage"`
	ADR            float64 `json:"adr"` // Average Damage per Round
	MVPCount       int     `json:"mvp_count"`
	Score          int     `json:"score"`
	FirstKills     int     `json:"first_kills"`
	FirstDeaths    int     `json:"first_deaths"`
	TradeKills     int     `json:"trade_kills"`
	TradeDeaths    int     `json:"trade_deaths"`
	FlashAssists   int     `json:"flash_assists"`
	EnemiesFlashed int     `json:"enemies_flashed"`
}
