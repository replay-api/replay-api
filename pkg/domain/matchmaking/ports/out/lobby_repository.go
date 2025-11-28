// Package matchmaking_out defines outbound repository interfaces
package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
) // LobbyRepository defines persistence operations for lobbies
type LobbyRepository interface {
	Save(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error
	FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingLobby, error)
	FindByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*matchmaking_entities.MatchmakingLobby, error)
	FindOpenLobbies(ctx context.Context, gameID, region, tier string, limit int) ([]*matchmaking_entities.MatchmakingLobby, error)
	Update(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PrizePoolRepository defines persistence operations for prize pools
type PrizePoolRepository interface {
	Save(ctx context.Context, pool *matchmaking_entities.PrizePool) error
	FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PrizePool, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) (*matchmaking_entities.PrizePool, error)
	FindPendingDistributions(ctx context.Context, limit int) ([]*matchmaking_entities.PrizePool, error)
	Update(ctx context.Context, pool *matchmaking_entities.PrizePool) error
	Delete(ctx context.Context, id uuid.UUID) error
}
