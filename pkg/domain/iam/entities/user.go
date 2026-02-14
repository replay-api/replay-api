package iam_entities

import (
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type UserIDKey uuid.UUID

type User struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	Name              string `json:"name" bson:"name"`
}

func NewUser(userID uuid.UUID, name string, resourceOwner shared.ResourceOwner) *User {
	entity := shared.NewEntity(resourceOwner)
	entity.ID = userID
	return &User{
		BaseEntity: entity,
		Name:       name,
	}
}

// func (u *User) GetID() uuid.UUID {
// 	return u.ID
// }

func (u User) GetID() uuid.UUID {
	return u.ID
}
