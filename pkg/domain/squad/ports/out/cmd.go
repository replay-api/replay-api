package squad_out

import (
	"context"

	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

type SquadWriter interface {
	CreateMany(createCtx context.Context, events []*squad_entities.Squad) error
	Create(createCtx context.Context, events *squad_entities.Squad) (*squad_entities.Squad, error)
}

type PlayerProfileWriter interface {
	CreateMany(createCtx context.Context, events []*squad_entities.PlayerProfile) error
	Create(createCtx context.Context, events *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error)
}

type PlayerProfileHistoryWriter interface {
	CreateMany(createCtx context.Context, events []*squad_entities.PlayerProfileHistory) error
	Create(createCtx context.Context, events *squad_entities.PlayerProfileHistory) (*squad_entities.PlayerProfileHistory, error)
}
