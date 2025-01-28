package squad_value_objects

import (
	"time"
)

type SquadMembershipType string

const (
	SquadMembershipTypeOwner  SquadMembershipType = "Owner"
	SquadMembershipTypeAdmin  SquadMembershipType = "Admin"
	SquadMembershipTypeMember SquadMembershipType = "Member"
)

type SquadMembershipStatus string

const (
	SquadMembershipStatusActive   SquadMembershipStatus = "Active"
	SquadMembershipStatusInactive SquadMembershipStatus = "Inactive"
	SquadMembershipStatusInvited  SquadMembershipStatus = "Invited"
)

type SquadMembership struct {
	Type    SquadMembershipType                 `json:"type" bson:"type"`
	Status  map[time.Time]SquadMembershipStatus `json:"status" bson:"status"`
	History map[time.Time]SquadMembershipType   `json:"history" bson:"history"`
}
