// Package matchmaking_in defines inbound command interfaces for matchmaking
package matchmaking_in

import (
	"context"

	"github.com/google/uuid"
matchmaking_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/entities"
matchmaking_vo "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/value-objects"
)// LobbyCommand defines operations for lobby management
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

// JoinLobbyCommand request for a player to join a lobby
type JoinLobbyCommand struct {
	LobbyID  uuid.UUID
	PlayerID uuid.UUID
	MMR      int
}

// LeaveLobbyCommand request for a player to leave a lobby
type LeaveLobbyCommand struct {
	LobbyID  uuid.UUID
	PlayerID uuid.UUID
}

// SetPlayerReadyCommand request to set player ready status
type SetPlayerReadyCommand struct {
	LobbyID  uuid.UUID
	PlayerID uuid.UUID
	IsReady  bool
}

// StartReadyCheckCommand request to begin ready check countdown
type StartReadyCheckCommand struct {
	LobbyID uuid.UUID
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

// AddContributionCommand request to add player entry fee to prize pool
type AddContributionCommand struct {
	PrizePoolID uuid.UUID
	PlayerID    uuid.UUID
	Amount      float64
}

// DistributePrizesCommand request to distribute prizes after escrow period
type DistributePrizesCommand struct {
	PrizePoolID     uuid.UUID
	RankedPlayerIDs []uuid.UUID
	MVPPlayerID     *uuid.UUID
}
