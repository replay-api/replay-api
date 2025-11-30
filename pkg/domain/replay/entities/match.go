package entities

import (
	"context"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type MatchVisibility string

const (
	MatchVisibilityPublic  MatchVisibility = "public"
	MatchVisibilitySquad   MatchVisibility = "squad"
	MatchVisibilityPrivate MatchVisibility = "private"
	MatchVisibilityCustom  MatchVisibility = "custom"
)

// AggregteRoot
type Match struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	RegionID      common.RegionIDKey   `json:"region_id" bson:"region_id"`
	ReplayFileID  uuid.UUID            `json:"replay_file_id" bson:"replay_file_id"`
	GameID        common.GameIDKey     `json:"game_id" bson:"game_id"`
	Scoreboard    Scoreboard           `json:"scoreboard" bson:"scoreboard"`
	Teams         []Team               `json:"teams" bson:"teams"`
	Events        []*GameEvent         `json:"game_events" bson:"game_events"`
	Visibility    MatchVisibility      `json:"visibility" bson:"visibility"`
	ShareTokens   []ShareToken         `json:"share_tokens" bson:"share_tokens"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func (m Match) GetID() uuid.UUID {
	return m.ID
}

type Scoreboard struct {
	TeamScoreboards []TeamScoreboard `json:"team_scoreboards" bson:"team_scoreboards"`
	MatchMVP        *PlayerMetadata  `json:"match_mvp" bson:"match_mvp"`
}

type TeamScoreboard struct {
	Team        Team                      `json:"team" bson:"team"`
	Side        string                    `json:"side" bson:"side"`
	TeamScore   int                       `json:"team_score" bson:"team_score"`
	TeamMVP     *PlayerMetadata           `json:"team_mvp" bson:"team_mvp"`
	Players     []PlayerMetadata          `json:"players" bson:"playerss"`
	PlayerStats map[uuid.UUID]interface{} `json:"player_stats" bson:"player_stats"`
	Rounds      []RoundInfo               `json:"rounds" bson:"rounds"`
	RoundStats  map[int]interface{}       `json:"round_stats" bson:"round_stats"`
}

type RoundInfo struct {
	RoundNumber      int         `json:"round_number" bson:"round_number"`
	WinnerTeamID     *uuid.UUID  `json:"winner" bson:"winner"`
	RoundMVPPlayerID *uuid.UUID  `json:"round_mvp_player_id" bson:"round_mvp_player_id"`
	Events           []GameEvent `json:"events" bson:"events"`
}

func NewCS2Match(userContext context.Context, replayFileID uuid.UUID) *Match {
	return &Match{
		ID:            uuid.New(),
		ReplayFileID:  replayFileID,
		GameID:        common.CS2.ID,
		ResourceOwner: common.GetResourceOwner(userContext),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
