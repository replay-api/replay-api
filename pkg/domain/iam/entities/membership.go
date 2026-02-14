package iam_entities

import (
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type MembershipType string

const (
	MembershipTypeOwner  MembershipType = "Owner"
	MembershipTypeAdmin  MembershipType = "Admin"
	MembershipTypeMember MembershipType = "Member"
)

type MembershipStatus string

const (
	MembershipStatusActive   MembershipStatus = "Active"
	MembershipStatusInactive MembershipStatus = "Inactive"
	MembershipStatusPending  MembershipStatus = "Pending"
)

type Membership struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	Type              MembershipType   `json:"type" bson:"type"`
	Status            MembershipStatus `json:"status" bson:"status"`
}

func (m Membership) GetID() uuid.UUID {
	return m.ID
}

func NewMembership(membershipType MembershipType, status MembershipStatus, resourceOwner shared.ResourceOwner) *Membership {
	entity := shared.NewEntity(resourceOwner)
	return &Membership{
		BaseEntity: entity,
		Type:       membershipType,
		Status:     status,
	}
}
