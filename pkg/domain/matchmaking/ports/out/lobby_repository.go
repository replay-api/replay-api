// Package matchmaking_out defines outbound repository interfaces
package matchmaking_out

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
) // LobbyRepository defines persistence operations for lobbies
type LobbyRepository interface {
	Save(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error
	FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingLobby, error)
	FindByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*matchmaking_entities.MatchmakingLobby, error)
	FindOpenLobbies(ctx context.Context, gameID, region, tier string, limit int) ([]*matchmaking_entities.MatchmakingLobby, error)
	FindExpiredReadyChecks(ctx context.Context) ([]*matchmaking_entities.MatchmakingLobby, error)
	Update(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error
	Delete(ctx context.Context, id uuid.UUID) error

	// SetPlayerReadyAtomic atomically sets a player's ready status.
	// Returns the updated lobby. Avoids lost-update races.
	SetPlayerReadyAtomic(ctx context.Context, lobbyID uuid.UUID, playerID uuid.UUID, isReady bool) (*matchmaking_entities.MatchmakingLobby, error)

	// TransitionStatus atomically transitions lobby status using CAS.
	// Returns true if applied, false if current status didn't match.
	TransitionStatus(ctx context.Context, lobbyID uuid.UUID, expectedStatus, newStatus matchmaking_entities.LobbyStatus, extraUpdates map[string]interface{}) (bool, error)
}

// PrizePoolRepository defines persistence operations for prize pools
type PrizePoolRepository interface {
	Save(ctx context.Context, pool *matchmaking_entities.PrizePool) error
	FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PrizePool, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) (*matchmaking_entities.PrizePool, error)
	Update(ctx context.Context, pool *matchmaking_entities.PrizePool) error
	UpdateUnsafe(ctx context.Context, pool *matchmaking_entities.PrizePool) error
	GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PrizePool, error)
	Search(ctx context.Context, s shared.Search) ([]matchmaking_entities.PrizePool, error)
	Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error)
}
