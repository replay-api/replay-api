package iam_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type UserIDKey uuid.UUID

type User struct {
	ID            uuid.UUID            `json:"-" bson:"_id"`
	Name          string               `json:"name" bson:"name"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func NewUser(userID uuid.UUID, name string, resourceOwner common.ResourceOwner) *User {
	return &User{
		ID:            userID,
		Name:          name,
		ResourceOwner: resourceOwner,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// func (u *User) GetID() uuid.UUID {
// 	return u.ID
// }

func (u User) GetID() uuid.UUID {
	return u.ID
}
