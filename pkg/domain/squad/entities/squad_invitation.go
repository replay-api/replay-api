package squad_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// InvitationStatus represents the state of an invitation
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusCanceled InvitationStatus = "canceled"
)

// InvitationType indicates who initiated the invitation
type InvitationType string

const (
	// InvitationTypeSquadToPlayer - Squad invites player to join
	InvitationTypeSquadToPlayer InvitationType = "squad_to_player"
	// InvitationTypePlayerToSquad - Player requests to join squad
	InvitationTypePlayerToSquad InvitationType = "player_to_squad"
)

// SquadInvitation represents an invitation to join a squad
type SquadInvitation struct {
	ID              uuid.UUID            `json:"id" bson:"_id"`
	SquadID         uuid.UUID            `json:"squad_id" bson:"squad_id"`
	SquadName       string               `json:"squad_name" bson:"squad_name"`
	PlayerProfileID uuid.UUID            `json:"player_profile_id" bson:"player_profile_id"`
	PlayerName      string               `json:"player_name" bson:"player_name"`
	InviterID       uuid.UUID            `json:"inviter_id" bson:"inviter_id"`
	InviterName     string               `json:"inviter_name" bson:"inviter_name"`
	InvitationType  InvitationType       `json:"invitation_type" bson:"invitation_type"`
	Status          InvitationStatus     `json:"status" bson:"status"`
	Role            string               `json:"role" bson:"role"` // Proposed role in squad
	Message         string               `json:"message,omitempty" bson:"message,omitempty"`
	ExpiresAt       time.Time            `json:"expires_at" bson:"expires_at"`
	RespondedAt     *time.Time           `json:"responded_at,omitempty" bson:"responded_at,omitempty"`
	ResourceOwner   shared.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt       time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" bson:"updated_at"`
}

// NewSquadInvitation creates a new squad invitation
func NewSquadInvitation(
	squadID uuid.UUID,
	squadName string,
	playerProfileID uuid.UUID,
	playerName string,
	inviterID uuid.UUID,
	inviterName string,
	invitationType InvitationType,
	role string,
	message string,
	expirationDays int,
	rxn shared.ResourceOwner,
) *SquadInvitation {
	now := time.Now()
	return &SquadInvitation{
		ID:              uuid.New(),
		SquadID:         squadID,
		SquadName:       squadName,
		PlayerProfileID: playerProfileID,
		PlayerName:      playerName,
		InviterID:       inviterID,
		InviterName:     inviterName,
		InvitationType:  invitationType,
		Status:          InvitationStatusPending,
		Role:            role,
		Message:         message,
		ExpiresAt:       now.AddDate(0, 0, expirationDays),
		ResourceOwner:   rxn,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// GetID returns the invitation ID
func (i SquadInvitation) GetID() uuid.UUID {
	return i.ID
}

// IsExpired checks if the invitation has expired
func (i SquadInvitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsPending checks if the invitation is still pending
func (i SquadInvitation) IsPending() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}

// Accept marks the invitation as accepted
func (i *SquadInvitation) Accept() {
	now := time.Now()
	i.Status = InvitationStatusAccepted
	i.RespondedAt = &now
	i.UpdatedAt = now
}

// Decline marks the invitation as declined
func (i *SquadInvitation) Decline() {
	now := time.Now()
	i.Status = InvitationStatusDeclined
	i.RespondedAt = &now
	i.UpdatedAt = now
}

// Cancel marks the invitation as canceled
func (i *SquadInvitation) Cancel() {
	i.Status = InvitationStatusCanceled
	i.UpdatedAt = time.Now()
}

// MarkExpired marks the invitation as expired
func (i *SquadInvitation) MarkExpired() {
	i.Status = InvitationStatusExpired
	i.UpdatedAt = time.Now()
}

