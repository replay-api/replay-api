package squad_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
)

const defaultExpirationDays = 7

type SquadInvitationUseCase struct {
	invitationWriter squad_out.SquadInvitationWriter
	invitationReader squad_out.SquadInvitationReader
	squadReader      squad_in.SquadReader
	squadWriter      squad_out.SquadWriter
	playerReader     squad_in.PlayerProfileReader
	historyWriter    squad_out.SquadHistoryWriter
}

func NewSquadInvitationUseCase(
	invitationWriter squad_out.SquadInvitationWriter,
	invitationReader squad_out.SquadInvitationReader,
	squadReader squad_in.SquadReader,
	squadWriter squad_out.SquadWriter,
	playerReader squad_in.PlayerProfileReader,
	historyWriter squad_out.SquadHistoryWriter,
) squad_in.SquadInvitationCommand {
	return &SquadInvitationUseCase{
		invitationWriter: invitationWriter,
		invitationReader: invitationReader,
		squadReader:      squadReader,
		squadWriter:      squadWriter,
		playerReader:     playerReader,
		historyWriter:    historyWriter,
	}
}

// InvitePlayer sends an invitation from a squad to a player
func (uc *SquadInvitationUseCase) InvitePlayer(ctx context.Context, cmd squad_in.InvitePlayerCommand) (*squad_entities.SquadInvitation, error) {
	slog.InfoContext(ctx, "Inviting player to squad", "squad_id", cmd.SquadID, "player_id", cmd.PlayerID)

	// Verify authentication
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, common.NewErrUnauthorized()
	}

	// Get squad and verify user is owner/admin
	squad, err := uc.getSquad(ctx, cmd.SquadID)
	if err != nil {
		return nil, err
	}

	if !uc.canManageMembers(squad, resourceOwner.UserID) {
		return nil, fmt.Errorf("only squad owners or admins can invite players")
	}

	// Get player profile
	player, err := uc.getPlayer(ctx, cmd.PlayerID)
	if err != nil {
		return nil, err
	}

	// Check if player is already a member
	for _, member := range squad.Membership {
		if member.PlayerProfileID == cmd.PlayerID {
			return nil, fmt.Errorf("player is already a member of this squad")
		}
	}

	// Check for existing pending invitation
	existing, err := uc.invitationReader.GetPendingBySquadAndPlayer(ctx, cmd.SquadID, cmd.PlayerID)
	if err != nil {
		slog.WarnContext(ctx, "Error checking existing invitation", "error", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("a pending invitation already exists")
	}

	// Get inviter name
	inviter, _ := uc.getPlayer(ctx, resourceOwner.UserID)
	inviterName := "Unknown"
	if inviter != nil {
		inviterName = inviter.Nickname
	}

	// Create invitation
	invitation := squad_entities.NewSquadInvitation(
		cmd.SquadID,
		squad.Name,
		cmd.PlayerID,
		player.Nickname,
		resourceOwner.UserID,
		inviterName,
		squad_entities.InvitationTypeSquadToPlayer,
		cmd.Role,
		cmd.Message,
		defaultExpirationDays,
		resourceOwner,
	)

	createdInvitation, err := uc.invitationWriter.Create(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create invitation", "error", err)
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Log history
	history := squad_entities.NewSquadHistory(cmd.SquadID, resourceOwner.UserID, squad_entities.SquadMembershipRequest, resourceOwner)
	_, _ = uc.historyWriter.Create(ctx, history)

	slog.InfoContext(ctx, "Player invited to squad", "invitation_id", createdInvitation.ID)
	return createdInvitation, nil
}

// RequestJoin creates a request from a player to join a squad
func (uc *SquadInvitationUseCase) RequestJoin(ctx context.Context, cmd squad_in.RequestJoinCommand) (*squad_entities.SquadInvitation, error) {
	slog.InfoContext(ctx, "Player requesting to join squad", "squad_id", cmd.SquadID)

	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, common.NewErrUnauthorized()
	}

	// Get squad
	squad, err := uc.getSquad(ctx, cmd.SquadID)
	if err != nil {
		return nil, err
	}

	// Get player profile
	player, err := uc.getPlayer(ctx, resourceOwner.UserID)
	if err != nil {
		return nil, err
	}

	// Check if already a member
	for _, member := range squad.Membership {
		if member.PlayerProfileID == resourceOwner.UserID {
			return nil, fmt.Errorf("you are already a member of this squad")
		}
	}

	// Check for existing pending request
	existing, err := uc.invitationReader.GetPendingBySquadAndPlayer(ctx, cmd.SquadID, resourceOwner.UserID)
	if err != nil {
		slog.WarnContext(ctx, "Error checking existing invitation", "error", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("a pending request already exists")
	}

	// Create join request
	invitation := squad_entities.NewSquadInvitation(
		cmd.SquadID,
		squad.Name,
		resourceOwner.UserID,
		player.Nickname,
		resourceOwner.UserID,
		player.Nickname,
		squad_entities.InvitationTypePlayerToSquad,
		"member", // Default role for join requests
		cmd.Message,
		defaultExpirationDays,
		resourceOwner,
	)

	createdInvitation, err := uc.invitationWriter.Create(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create join request", "error", err)
		return nil, fmt.Errorf("failed to create join request: %w", err)
	}

	// Log history
	history := squad_entities.NewSquadHistory(cmd.SquadID, resourceOwner.UserID, squad_entities.SquadMemberJoinRequest, resourceOwner)
	_, _ = uc.historyWriter.Create(ctx, history)

	slog.InfoContext(ctx, "Join request created", "invitation_id", createdInvitation.ID)
	return createdInvitation, nil
}

// RespondToInvitation accepts or declines an invitation
func (uc *SquadInvitationUseCase) RespondToInvitation(ctx context.Context, cmd squad_in.RespondToInvitationCommand) (*squad_entities.SquadInvitation, error) {
	slog.InfoContext(ctx, "Responding to invitation", "invitation_id", cmd.InvitationID, "accept", cmd.Accept)

	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return nil, common.NewErrUnauthorized()
	}

	// Get invitation
	invitation, err := uc.invitationReader.GetByID(ctx, cmd.InvitationID)
	if err != nil {
		return nil, common.NewErrNotFound(common.ResourceTypeSquad, "invitation", cmd.InvitationID.String())
	}

	if !invitation.IsPending() {
		return nil, fmt.Errorf("invitation is no longer pending")
	}

	// Verify authorization based on invitation type
	if invitation.InvitationType == squad_entities.InvitationTypeSquadToPlayer {
		// Player must respond to squad invitation
		if invitation.PlayerProfileID != resourceOwner.UserID {
			return nil, common.NewErrUnauthorized()
		}
	} else {
		// Squad admin must respond to join request
		squad, err := uc.getSquad(ctx, invitation.SquadID)
		if err != nil {
			return nil, err
		}
		if !uc.canManageMembers(squad, resourceOwner.UserID) {
			return nil, fmt.Errorf("only squad owners or admins can respond to join requests")
		}
	}

	if cmd.Accept {
		invitation.Accept()

		// Add player to squad
		squad, err := uc.getSquad(ctx, invitation.SquadID)
		if err != nil {
			return nil, err
		}

		// Determine membership type from role
		memberType := squad_value_objects.SquadMembershipTypeMember
		if invitation.Role == "admin" {
			memberType = squad_value_objects.SquadMembershipTypeAdmin
		}

		// Get the user ID for this player - use inviter ID as fallback
		userID := invitation.InviterID // This should be the player's user ID

		newMember := squad_value_objects.NewSquadMembership(
			userID,
			invitation.PlayerProfileID,
			memberType,
			[]string{invitation.Role},
			squad_value_objects.SquadMembershipStatusActive,
			memberType,
		)
		squad.Membership = append(squad.Membership, *newMember)

		_, err = uc.squadWriter.Update(ctx, squad)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to add member to squad", "error", err)
			return nil, fmt.Errorf("failed to add member: %w", err)
		}

		// Log history
		history := squad_entities.NewSquadHistory(invitation.SquadID, invitation.PlayerProfileID, squad_entities.SquadMemberJoined, resourceOwner)
		_, _ = uc.historyWriter.Create(ctx, history)
	} else {
		invitation.Decline()

		// Log history
		action := squad_entities.SquadMemberRequestDeclined
		if invitation.InvitationType == squad_entities.InvitationTypeSquadToPlayer {
			action = squad_entities.SquadMemberRequestDeclined
		}
		history := squad_entities.NewSquadHistory(invitation.SquadID, invitation.PlayerProfileID, action, resourceOwner)
		_, _ = uc.historyWriter.Create(ctx, history)
	}

	// Update invitation
	updatedInvitation, err := uc.invitationWriter.Update(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update invitation", "error", err)
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	return updatedInvitation, nil
}

// CancelInvitation cancels a pending invitation
func (uc *SquadInvitationUseCase) CancelInvitation(ctx context.Context, invitationID uuid.UUID) error {
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return common.NewErrUnauthorized()
	}

	invitation, err := uc.invitationReader.GetByID(ctx, invitationID)
	if err != nil {
		return common.NewErrNotFound(common.ResourceTypeSquad, "invitation", invitationID.String())
	}

	if !invitation.IsPending() {
		return fmt.Errorf("invitation is no longer pending")
	}

	// Verify authorization
	if invitation.InviterID != resourceOwner.UserID {
		// Check if user is squad admin
		squad, err := uc.getSquad(ctx, invitation.SquadID)
		if err != nil {
			return err
		}
		if !uc.canManageMembers(squad, resourceOwner.UserID) {
			return common.NewErrUnauthorized()
		}
	}

	invitation.Cancel()
	_, err = uc.invitationWriter.Update(ctx, invitation)
	if err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	slog.InfoContext(ctx, "Invitation canceled", "invitation_id", invitationID)
	return nil
}

// GetPendingInvitations gets all pending invitations for a player
func (uc *SquadInvitationUseCase) GetPendingInvitations(ctx context.Context, playerID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
	return uc.invitationReader.GetByPlayerID(ctx, playerID)
}

// GetSquadInvitations gets all invitations for a squad
func (uc *SquadInvitationUseCase) GetSquadInvitations(ctx context.Context, squadID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
	return uc.invitationReader.GetBySquadID(ctx, squadID)
}

// Helper functions

func (uc *SquadInvitationUseCase) getSquad(ctx context.Context, squadID uuid.UUID) (*squad_entities.Squad, error) {
	searchParams := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{Field: "ID", Values: []interface{}{squadID.String()}, Operator: common.EqualsOperator},
					},
				},
			},
		},
	}
	resultOpts := common.SearchResultOptions{Limit: 1}

	compiledSearch, err := uc.squadReader.Compile(ctx, searchParams, resultOpts)
	if err != nil {
		return nil, err
	}

	squads, err := uc.squadReader.Search(ctx, *compiledSearch)
	if err != nil {
		return nil, err
	}

	if len(squads) == 0 {
		return nil, common.NewErrNotFound(common.ResourceTypeSquad, "ID", squadID.String())
	}

	return &squads[0], nil
}

func (uc *SquadInvitationUseCase) getPlayer(ctx context.Context, playerID uuid.UUID) (*squad_entities.PlayerProfile, error) {
	search := squad_entities.NewSearchByID(ctx, playerID)
	players, err := uc.playerReader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	if len(players) == 0 {
		return nil, common.NewErrNotFound(common.ResourceTypePlayerProfile, "ID", playerID.String())
	}

	return &players[0], nil
}

func (uc *SquadInvitationUseCase) canManageMembers(squad *squad_entities.Squad, userID uuid.UUID) bool {
	// Check if user is the owner (ResourceOwner.UserID)
	if squad.ResourceOwner.UserID == userID {
		return true
	}

	// Check if user has admin role
	for _, member := range squad.Membership {
		if member.PlayerProfileID == userID || member.UserID == userID {
			return member.Type == squad_value_objects.SquadMembershipTypeOwner ||
				member.Type == squad_value_objects.SquadMembershipTypeAdmin
		}
	}

	return false
}

