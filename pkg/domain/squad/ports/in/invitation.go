package squad_in

import (
	"context"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
)

// InvitePlayerCommand represents a request to invite a player to a squad
type InvitePlayerCommand struct {
	SquadID   uuid.UUID `json:"squad_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	Role      string    `json:"role"`
	Message   string    `json:"message,omitempty"`
}

// RequestJoinCommand represents a player's request to join a squad
type RequestJoinCommand struct {
	SquadID  uuid.UUID `json:"squad_id"`
	Message  string    `json:"message,omitempty"`
}

// RespondToInvitationCommand represents responding to an invitation
type RespondToInvitationCommand struct {
	InvitationID uuid.UUID `json:"invitation_id"`
	Accept       bool      `json:"accept"`
}

// SquadInvitationCommand defines the interface for squad invitation operations
type SquadInvitationCommand interface {
	// InvitePlayer sends an invitation from a squad to a player
	InvitePlayer(ctx context.Context, cmd InvitePlayerCommand) (*squad_entities.SquadInvitation, error)
	
	// RequestJoin creates a request from a player to join a squad
	RequestJoin(ctx context.Context, cmd RequestJoinCommand) (*squad_entities.SquadInvitation, error)
	
	// RespondToInvitation accepts or declines an invitation
	RespondToInvitation(ctx context.Context, cmd RespondToInvitationCommand) (*squad_entities.SquadInvitation, error)
	
	// CancelInvitation cancels a pending invitation
	CancelInvitation(ctx context.Context, invitationID uuid.UUID) error
	
	// GetPendingInvitations gets all pending invitations for a player
	GetPendingInvitations(ctx context.Context, playerID uuid.UUID) ([]squad_entities.SquadInvitation, error)
	
	// GetSquadInvitations gets all invitations for a squad
	GetSquadInvitations(ctx context.Context, squadID uuid.UUID) ([]squad_entities.SquadInvitation, error)
}

