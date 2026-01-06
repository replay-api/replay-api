package squad

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// SquadPermission represents the different permission levels for squad operations
type SquadPermission string

const (
	// PermissionOwner requires the user to be the squad owner (ResourceOwner.UserID or membership type Owner)
	PermissionOwner SquadPermission = "owner"
	// PermissionAdmin requires the user to be an owner or admin
	PermissionAdmin SquadPermission = "admin"
	// PermissionMember requires the user to be any active member
	PermissionMember SquadPermission = "member"
	// PermissionSelf requires the user to be performing an action on themselves (e.g., leaving squad)
	PermissionSelf SquadPermission = "self"
)

// AuthorizationResult contains details about an authorization check
type AuthorizationResult struct {
	Authorized     bool
	UserID         uuid.UUID
	MembershipType squad_value_objects.SquadMembershipType
	IsSelf         bool
	Message        string
}

// ErrForbidden represents a forbidden access error with details
type ErrForbidden struct {
	Message  string
	UserID   uuid.UUID
	SquadID  uuid.UUID
	Required SquadPermission
}

func (e *ErrForbidden) Error() string {
	return fmt.Sprintf("forbidden: %s (user: %s, squad: %s, required: %s)", e.Message, e.UserID, e.SquadID, e.Required)
}

// NewErrForbidden creates a new forbidden error
func NewErrForbidden(userID, squadID uuid.UUID, required SquadPermission, message string) *ErrForbidden {
	return &ErrForbidden{
		Message:  message,
		UserID:   userID,
		SquadID:  squadID,
		Required: required,
	}
}

// CheckSquadAuthorization checks if the current user has the required permission on a squad
func CheckSquadAuthorization(ctx context.Context, squad *squad_entities.Squad, requiredPermission SquadPermission, targetPlayerID ...uuid.UUID) AuthorizationResult {
	currentOwner := shared.GetResourceOwner(ctx)
	userID := currentOwner.UserID

	result := AuthorizationResult{
		UserID: userID,
	}

	// Check if user is the resource owner (creator) of the squad
	isResourceOwner := squad.ResourceOwner.UserID == userID

	// Find user's membership in the squad
	var userMembership *squad_value_objects.SquadMembership
	for i := range squad.Membership {
		if squad.Membership[i].UserID == userID {
			userMembership = &squad.Membership[i]
			result.MembershipType = userMembership.Type
			break
		}
	}

	// Check if this is a self-operation (target player is the current user)
	if len(targetPlayerID) > 0 {
		for _, m := range squad.Membership {
			if m.PlayerProfileID == targetPlayerID[0] && m.UserID == userID {
				result.IsSelf = true
				break
			}
		}
	}

	// Permission checks based on required level
	switch requiredPermission {
	case PermissionOwner:
		// Only owner can perform this action
		if isResourceOwner {
			result.Authorized = true
			result.Message = "authorized as resource owner"
			return result
		}
		if userMembership != nil && userMembership.Type == squad_value_objects.SquadMembershipTypeOwner {
			result.Authorized = true
			result.Message = "authorized as membership owner"
			return result
		}
		result.Message = "user is not a squad owner"

	case PermissionAdmin:
		// Owner or admin can perform this action
		if isResourceOwner {
			result.Authorized = true
			result.Message = "authorized as resource owner"
			return result
		}
		if userMembership != nil {
			switch userMembership.Type {
			case squad_value_objects.SquadMembershipTypeOwner, squad_value_objects.SquadMembershipTypeAdmin:
				result.Authorized = true
				result.Message = fmt.Sprintf("authorized as %s", userMembership.Type)
				return result
			}
		}
		result.Message = "user is not a squad owner or admin"

	case PermissionMember:
		// Any active member can perform this action
		if isResourceOwner {
			result.Authorized = true
			result.Message = "authorized as resource owner"
			return result
		}
		if userMembership != nil {
			result.Authorized = true
			result.Message = fmt.Sprintf("authorized as member with type %s", userMembership.Type)
			return result
		}
		result.Message = "user is not a squad member"

	case PermissionSelf:
		// User is performing an action on themselves
		if result.IsSelf {
			result.Authorized = true
			result.Message = "authorized for self-action"
			return result
		}
		// Also allow owners/admins to perform this action on behalf
		if isResourceOwner {
			result.Authorized = true
			result.Message = "authorized as resource owner for member action"
			return result
		}
		if userMembership != nil {
			switch userMembership.Type {
			case squad_value_objects.SquadMembershipTypeOwner, squad_value_objects.SquadMembershipTypeAdmin:
				result.Authorized = true
				result.Message = fmt.Sprintf("authorized as %s for member action", userMembership.Type)
				return result
			}
		}
		result.Message = "user can only perform this action on themselves or must be an admin"
	}

	return result
}

// MustBeSquadOwnerOrAdmin returns an error if the user is not an owner or admin
func MustBeSquadOwnerOrAdmin(ctx context.Context, squad *squad_entities.Squad) error {
	result := CheckSquadAuthorization(ctx, squad, PermissionAdmin)
	if !result.Authorized {
		return NewErrForbidden(result.UserID, squad.ID, PermissionAdmin, result.Message)
	}
	return nil
}

// MustBeSquadOwner returns an error if the user is not the squad owner
func MustBeSquadOwner(ctx context.Context, squad *squad_entities.Squad) error {
	result := CheckSquadAuthorization(ctx, squad, PermissionOwner)
	if !result.Authorized {
		return NewErrForbidden(result.UserID, squad.ID, PermissionOwner, result.Message)
	}
	return nil
}

// MustBeSquadMember returns an error if the user is not a squad member
func MustBeSquadMember(ctx context.Context, squad *squad_entities.Squad) error {
	result := CheckSquadAuthorization(ctx, squad, PermissionMember)
	if !result.Authorized {
		return NewErrForbidden(result.UserID, squad.ID, PermissionMember, result.Message)
	}
	return nil
}

// CanRemoveSquadMember checks if the user can remove a specific member
// - Owners and admins can remove any member
// - Members can remove themselves (leave squad)
func CanRemoveSquadMember(ctx context.Context, squad *squad_entities.Squad, targetPlayerID uuid.UUID) error {
	result := CheckSquadAuthorization(ctx, squad, PermissionSelf, targetPlayerID)
	if !result.Authorized {
		return NewErrForbidden(result.UserID, squad.ID, PermissionSelf, result.Message)
	}
	return nil
}

// CanUpdateMemberRole checks if the user can update a member's role
// Only owners and admins can update roles, and admins cannot demote owners
func CanUpdateMemberRole(ctx context.Context, squad *squad_entities.Squad, targetPlayerID uuid.UUID, newType squad_value_objects.SquadMembershipType) error {
	result := CheckSquadAuthorization(ctx, squad, PermissionAdmin)
	if !result.Authorized {
		return NewErrForbidden(result.UserID, squad.ID, PermissionAdmin, result.Message)
	}

	// Find target membership
	var targetMembership *squad_value_objects.SquadMembership
	for i := range squad.Membership {
		if squad.Membership[i].PlayerProfileID == targetPlayerID {
			targetMembership = &squad.Membership[i]
			break
		}
	}

	if targetMembership == nil {
		return shared.NewErrNotFound(replay_common.ResourceTypeSquad, "MemberID", targetPlayerID.String())
	}

	// Prevent admin from modifying owner's role
	if result.MembershipType == squad_value_objects.SquadMembershipTypeAdmin &&
		targetMembership.Type == squad_value_objects.SquadMembershipTypeOwner {
		return NewErrForbidden(result.UserID, squad.ID, PermissionOwner, "admins cannot modify owner's role")
	}

	// Prevent admin from promoting to owner
	if result.MembershipType == squad_value_objects.SquadMembershipTypeAdmin &&
		newType == squad_value_objects.SquadMembershipTypeOwner {
		return NewErrForbidden(result.UserID, squad.ID, PermissionOwner, "admins cannot promote members to owner")
	}

	return nil
}

