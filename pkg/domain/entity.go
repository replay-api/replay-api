package common

import (
	"time"

	"github.com/google/uuid"
)

type BaseEntity struct {
	ID            uuid.UUID     `json:"id" bson:"_id"`
	ResourceOwner ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" bson:"updated_at"`
}

type Entity interface {
	GetID() uuid.UUID
}

func (b BaseEntity) GetID() uuid.UUID {
	return b.ID
}

func NewEntity(resourceOwner ResourceOwner) BaseEntity {
	return BaseEntity{
		ID:            uuid.New(),
		ResourceOwner: resourceOwner,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
