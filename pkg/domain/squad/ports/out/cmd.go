package squad_out

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

type SquadWriter interface {
	CreateMany(createCtx context.Context, events []*squad_entities.Squad) error
	Create(createCtx context.Context, events *squad_entities.Squad) (*squad_entities.Squad, error)
	Update(ctx context.Context, squad *squad_entities.Squad) (*squad_entities.Squad, error)
	Delete(ctx context.Context, squadID uuid.UUID) error
}

type PlayerProfileWriter interface {
	CreateMany(createCtx context.Context, events []*squad_entities.PlayerProfile) error
	Create(createCtx context.Context, events *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error)
	Update(ctx context.Context, profile *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error)
	Delete(ctx context.Context, profileID uuid.UUID) error
}

type PlayerProfileHistoryWriter interface {
	CreateMany(createCtx context.Context, events []*squad_entities.PlayerProfileHistory) error
	Create(createCtx context.Context, events *squad_entities.PlayerProfileHistory) (*squad_entities.PlayerProfileHistory, error)
}
