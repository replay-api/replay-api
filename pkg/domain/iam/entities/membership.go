package iam_entities

import (
	"time"

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
	ID            uuid.UUID            `json:"id" bson:"_id"`
	Type          MembershipType       `json:"type" bson:"type"`
	Status        MembershipStatus     `json:"status" bson:"status"`
	ResourceOwner shared.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func (m Membership) GetID() uuid.UUID {
	return m.ID
}

func NewMembership(membershipType MembershipType, status MembershipStatus, resourceOwner shared.ResourceOwner) *Membership {
	return &Membership{
		ID:            uuid.New(),
		Type:          membershipType,
		Status:        status,
		ResourceOwner: resourceOwner,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
