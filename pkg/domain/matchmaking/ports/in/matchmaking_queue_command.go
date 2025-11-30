package matchmaking_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
)

// JoinMatchmakingQueueCommandHandler handles joining matchmaking queue
type JoinMatchmakingQueueCommandHandler interface {
	Exec(ctx context.Context, cmd JoinMatchmakingQueueCommand) (*matchmaking_entities.MatchmakingSession, error)
}

// LeaveMatchmakingQueueCommandHandler handles leaving matchmaking queue
type LeaveMatchmakingQueueCommandHandler interface {
	Exec(ctx context.Context, cmd LeaveMatchmakingQueueCommand) error
}

// JoinMatchmakingQueueCommand request to join matchmaking queue
type JoinMatchmakingQueueCommand struct {
	PlayerID    uuid.UUID
	SquadID     *uuid.UUID
	GameID      string
	GameMode    string
	Region      string
	Tier        matchmaking_entities.MatchmakingTier
	PlayerMMR   int
	PlayerRole  *string // for role-based matching (tank, dps, support)
	TeamFormat  TeamFormat
	MaxPing     int
	PriorityBoost bool
}

// Validate validates the JoinMatchmakingQueueCommand
func (c *JoinMatchmakingQueueCommand) Validate() error {
	if c.PlayerID == uuid.Nil {
		return errors.New("player_id is required")
	}
	if c.GameID == "" {
		return errors.New("game_id is required")
	}
	if c.GameMode == "" {
		return errors.New("game_mode is required")
	}
	if c.Region == "" {
		return errors.New("region is required")
	}
	if !c.TeamFormat.IsValid() {
		return errors.New("invalid team_format")
	}
	if c.PlayerMMR < 0 {
		return errors.New("player_mmr cannot be negative")
	}
	if c.MaxPing <= 0 {
		c.MaxPing = 150 // default max ping
	}
	if c.PlayerRole != nil && !PlayerRole(*c.PlayerRole).IsValid() {
		return errors.New("invalid player_role")
	}
	return nil
}

// LeaveMatchmakingQueueCommand request to leave matchmaking queue
type LeaveMatchmakingQueueCommand struct {
	SessionID uuid.UUID
	PlayerID  uuid.UUID
}

// Validate validates the LeaveMatchmakingQueueCommand
func (c *LeaveMatchmakingQueueCommand) Validate() error {
	if c.SessionID == uuid.Nil {
		return errors.New("session_id is required")
	}
	if c.PlayerID == uuid.Nil {
		return errors.New("player_id is required")
	}
	return nil
}

// TeamFormat defines the team size for matchmaking
type TeamFormat string

const (
	TeamFormat1v1 TeamFormat = "1v1"
	TeamFormat2v2 TeamFormat = "2v2"
	TeamFormat3v3 TeamFormat = "3v3"
	TeamFormat4v4 TeamFormat = "4v4"
	TeamFormat5v5 TeamFormat = "5v5"
)

// GetTeamSize returns the number of players per team
func (tf TeamFormat) GetTeamSize() int {
	switch tf {
	case TeamFormat1v1:
		return 1
	case TeamFormat2v2:
		return 2
	case TeamFormat3v3:
		return 3
	case TeamFormat4v4:
		return 4
	case TeamFormat5v5:
		return 5
	default:
		return 1
	}
}

// GetTotalPlayers returns total players needed for match
func (tf TeamFormat) GetTotalPlayers() int {
	return tf.GetTeamSize() * 2
}

// IsValid checks if team format is valid
func (tf TeamFormat) IsValid() bool {
	switch tf {
	case TeamFormat1v1, TeamFormat2v2, TeamFormat3v3, TeamFormat4v4, TeamFormat5v5:
		return true
	default:
		return false
	}
}

// PlayerRole defines player roles for team-based games
type PlayerRole string

const (
	PlayerRoleTank    PlayerRole = "tank"
	PlayerRoleDPS     PlayerRole = "dps"
	PlayerRoleSupport PlayerRole = "support"
	PlayerRoleFlex    PlayerRole = "flex"
)

// IsValid checks if player role is valid
func (pr PlayerRole) IsValid() bool {
	switch pr {
	case PlayerRoleTank, PlayerRoleDPS, PlayerRoleSupport, PlayerRoleFlex:
		return true
	default:
		return false
	}
}
