package squad_out

import (
	"context"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
)

type SquadWriter interface {
	CreateMany(createCtx context.Context, events []*squad_entities.Squad) error
	Create(createCtx context.Context, squad *squad_entities.Squad) (*squad_entities.Squad, error)
	Update(updateCtx context.Context, squad *squad_entities.Squad) (*squad_entities.Squad, error)
}

type PlayerProfileWriter interface {
	CreateMany(createCtx context.Context, players []*squad_entities.PlayerProfile) error
	Create(createCtx context.Context, player *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error)
	Update(updateCtx context.Context, player *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error)
}

type PlayerProfileHistoryWriter interface {
	CreateMany(createCtx context.Context, histories []*squad_entities.PlayerProfileHistory) error
	Create(createCtx context.Context, history *squad_entities.PlayerProfileHistory) (*squad_entities.PlayerProfileHistory, error)
}

type SquadHistoryWriter interface {
	CreateMany(createCtx context.Context, histories []*squad_entities.SquadHistory) error
	Create(createCtx context.Context, history *squad_entities.SquadHistory) (*squad_entities.SquadHistory, error)
}
