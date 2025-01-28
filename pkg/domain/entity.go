package common

import (
	"time"

	"github.com/google/uuid"
)

type BaseEntity struct {
	ID              uuid.UUID           `json:"id" bson:"_id"`
	VisibilityLevel IntendedAudienceKey `json:"visibility_level" bson:"visibility_level"`
	VisbilityType   VisibilityTypeKey   `json:"visibility_type" bson:"visibility_type"`
	ResourceOwner   ResourceOwner       `json:"resource_owner" bson:"resource_owner"`
	CreatedAt       time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" bson:"updated_at"`
}

type Entity interface {
	GetID() uuid.UUID
}

func (b BaseEntity) GetID() uuid.UUID {
	return b.ID
}

func NewEntity(resourceOwner ResourceOwner) BaseEntity {
	return BaseEntity{
		ID:              uuid.New(),
		VisibilityLevel: ClientApplicationAudienceIDKey,
		VisbilityType:   CustomVisibilityTypeKey,
		ResourceOwner:   resourceOwner,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func NewUnrestrictedEntity(resourceOwner ResourceOwner) BaseEntity {
	return BaseEntity{
		ID:              uuid.New(),
		VisibilityLevel: TenantAudienceIDKey,
		VisbilityType:   PublicVisibilityTypeKey,
		ResourceOwner:   resourceOwner,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func NewRestrictedEntity(resourceOwner ResourceOwner) BaseEntity {
	return BaseEntity{
		ID:              uuid.New(),
		VisibilityLevel: GroupAudienceIDKey,
		VisbilityType:   RestrictedVisibilityTypeKey,
		ResourceOwner:   resourceOwner,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func NewPrivateEntity(resourceOwner ResourceOwner) BaseEntity {
	return BaseEntity{
		ID:              uuid.New(),
		VisibilityLevel: UserAudienceIDKey,
		VisbilityType:   PrivateVisibilityTypeKey,
		ResourceOwner:   resourceOwner,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
