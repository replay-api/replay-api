// Package matchmaking_in defines inbound command interfaces for matchmaking
package matchmaking_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
) // LobbyCommand defines operations for lobby management
type LobbyCommand interface {
	CreateLobby(ctx context.Context, cmd CreateLobbyCommand) (*matchmaking_entities.MatchmakingLobby, error)
	JoinLobby(ctx context.Context, cmd JoinLobbyCommand) error
	LeaveLobby(ctx context.Context, cmd LeaveLobbyCommand) error
	SetPlayerReady(ctx context.Context, cmd SetPlayerReadyCommand) error
	StartReadyCheck(ctx context.Context, cmd StartReadyCheckCommand) error
	StartMatch(ctx context.Context, lobbyID uuid.UUID) (uuid.UUID, error)
	CancelLobby(ctx context.Context, lobbyID uuid.UUID, reason string) error
}

// PrizePoolCommand defines operations for prize pool management
type PrizePoolCommand interface {
	CreatePrizePool(ctx context.Context, cmd CreatePrizePoolCommand) (*matchmaking_entities.PrizePool, error)
	AddPlayerContribution(ctx context.Context, cmd AddContributionCommand) error
	LockPool(ctx context.Context, prizePoolID uuid.UUID) error
	DistributePrizes(ctx context.Context, cmd DistributePrizesCommand) error
	CancelPool(ctx context.Context, prizePoolID uuid.UUID, reason string) error
}

// CreateLobbyCommand request to create a new matchmaking lobby
type CreateLobbyCommand struct {
	CreatorID        uuid.UUID
	GameID           string
	Region           string
	Tier             string
	DistributionRule matchmaking_vo.DistributionRule
	MaxPlayers       int
	AutoFill         bool
	InviteOnly       bool
}

// Validate validates the CreateLobbyCommand
func (c *CreateLobbyCommand) Validate() error {
	if c.CreatorID == uuid.Nil {
		return errors.New("creator_id is required")
	}
	if c.GameID == "" {
		return errors.New("game_id is required")
	}
	if c.Region == "" {
		return errors.New("region is required")
	}
	if c.MaxPlayers < 2 {
		return errors.New("max_players must be at least 2")
	}
	if c.MaxPlayers > 10 {
		return errors.New("max_players cannot exceed 10")
	}
	if !c.DistributionRule.IsValid() {
		return errors.New("invalid distribution_rule")
	}
	return nil
}

// JoinLobbyCommand request for a player to join a lobby
type JoinLobbyCommand struct {
	LobbyID  uuid.UUID
	PlayerID uuid.UUID
	MMR      int
}

// Validate validates the JoinLobbyCommand
func (c *JoinLobbyCommand) Validate() error {
	if c.LobbyID == uuid.Nil {
		return errors.New("lobby_id is required")
	}
	if c.PlayerID == uuid.Nil {
		return errors.New("player_id is required")
	}
	if c.MMR < 0 {
		return errors.New("mmr cannot be negative")
	}
	return nil
}

// LeaveLobbyCommand request for a player to leave a lobby
type LeaveLobbyCommand struct {
	LobbyID  uuid.UUID
	PlayerID uuid.UUID
}

// Validate validates the LeaveLobbyCommand
func (c *LeaveLobbyCommand) Validate() error {
	if c.LobbyID == uuid.Nil {
		return errors.New("lobby_id is required")
	}
	if c.PlayerID == uuid.Nil {
		return errors.New("player_id is required")
	}
	return nil
}

// SetPlayerReadyCommand request to set player ready status
type SetPlayerReadyCommand struct {
	LobbyID  uuid.UUID
	PlayerID uuid.UUID
	IsReady  bool
}

// Validate validates the SetPlayerReadyCommand
func (c *SetPlayerReadyCommand) Validate() error {
	if c.LobbyID == uuid.Nil {
		return errors.New("lobby_id is required")
	}
	if c.PlayerID == uuid.Nil {
		return errors.New("player_id is required")
	}
	return nil
}

// StartReadyCheckCommand request to begin ready check countdown
type StartReadyCheckCommand struct {
	LobbyID   uuid.UUID
	CreatorID uuid.UUID // For authorization - only creator can start
}

// Validate validates the StartReadyCheckCommand
func (c *StartReadyCheckCommand) Validate() error {
	if c.LobbyID == uuid.Nil {
		return errors.New("lobby_id is required")
	}
	if c.CreatorID == uuid.Nil {
		return errors.New("creator_id is required")
	}
	return nil
}

// CreatePrizePoolCommand request to create a new prize pool
type CreatePrizePoolCommand struct {
	MatchID              uuid.UUID
	GameID               string
	Region               string
	Currency             string
	PlatformContribution float64
	DistributionRule     matchmaking_vo.DistributionRule
}

// Validate validates the CreatePrizePoolCommand
func (c *CreatePrizePoolCommand) Validate() error {
	if c.MatchID == uuid.Nil {
		return errors.New("match_id is required")
	}
	if c.GameID == "" {
		return errors.New("game_id is required")
	}
	if c.Currency == "" {
		return errors.New("currency is required")
	}
	if c.PlatformContribution < 0 {
		return errors.New("platform_contribution cannot be negative")
	}
	if !c.DistributionRule.IsValid() {
		return errors.New("invalid distribution_rule")
	}
	return nil
}

// AddContributionCommand request to add player entry fee to prize pool
type AddContributionCommand struct {
	PrizePoolID uuid.UUID
	PlayerID    uuid.UUID
	Amount      float64
}

// Validate validates the AddContributionCommand
func (c *AddContributionCommand) Validate() error {
	if c.PrizePoolID == uuid.Nil {
		return errors.New("prize_pool_id is required")
	}
	if c.PlayerID == uuid.Nil {
		return errors.New("player_id is required")
	}
	if c.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	return nil
}

// DistributePrizesCommand request to distribute prizes after escrow period
type DistributePrizesCommand struct {
	PrizePoolID     uuid.UUID
	RankedPlayerIDs []uuid.UUID
	MVPPlayerID     *uuid.UUID
}

// Validate validates the DistributePrizesCommand
func (c *DistributePrizesCommand) Validate() error {
	if c.PrizePoolID == uuid.Nil {
		return errors.New("prize_pool_id is required")
	}
	if len(c.RankedPlayerIDs) == 0 {
		return errors.New("ranked_player_ids is required")
	}
	return nil
}
