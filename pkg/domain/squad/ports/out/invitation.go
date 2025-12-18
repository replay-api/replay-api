package squad_out

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
)

// SquadInvitationWriter defines write operations for squad invitations
type SquadInvitationWriter interface {
	Create(ctx context.Context, invitation *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error)
	Update(ctx context.Context, invitation *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// SquadInvitationReader defines read operations for squad invitations
type SquadInvitationReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*squad_entities.SquadInvitation, error)
	GetBySquadID(ctx context.Context, squadID uuid.UUID) ([]squad_entities.SquadInvitation, error)
	GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]squad_entities.SquadInvitation, error)
	GetPendingBySquadAndPlayer(ctx context.Context, squadID, playerID uuid.UUID) (*squad_entities.SquadInvitation, error)
}

