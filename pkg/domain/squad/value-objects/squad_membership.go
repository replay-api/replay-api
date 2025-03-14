package squad_value_objects

import (
	"time"

	"github.com/google/uuid"
)

type SquadMembershipType string

const (
	SquadMembershipTypeOwner    SquadMembershipType = "Owner"
	SquadMembershipTypeAdmin    SquadMembershipType = "Admin"
	SquadMembershipTypeMember   SquadMembershipType = "Member"
	SquadMembershipTypeGuest    SquadMembershipType = "Guest"
	SquadMembershipTypeInactive SquadMembershipType = "Inactive"
)

type SquadMembershipStatus string

const (
	SquadMembershipStatusActive   SquadMembershipStatus = "Active"
	SquadMembershipStatusInactive SquadMembershipStatus = "Inactive"
	SquadMembershipStatusInvited  SquadMembershipStatus = "Invited"
)

type SquadMembership struct {
	UserID          uuid.UUID                           `json:"user_id" bson:"user_id"`
	PlayerProfileID uuid.UUID                           `json:"player_profile_id" bson:"player_profile_id"`
	Type            SquadMembershipType                 `json:"type" bson:"type"`
	Roles           []string                            `json:"roles" bson:"roles"`
	Status          map[time.Time]SquadMembershipStatus `json:"status" bson:"status"`
	History         map[time.Time]SquadMembershipType   `json:"history" bson:"history"`
}

func NewSquadMembership(userID uuid.UUID, playerProfileID uuid.UUID, membershipType SquadMembershipType, roles []string, status SquadMembershipStatus, history SquadMembershipType) *SquadMembership {
	return &SquadMembership{
		UserID:          userID,
		PlayerProfileID: playerProfileID,
		Type:            membershipType,
		Roles:           roles,
		Status:          map[time.Time]SquadMembershipStatus{time.Now(): status},
		History:         map[time.Time]SquadMembershipType{time.Now(): history},
	}
}
