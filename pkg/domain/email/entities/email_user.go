package email_entities

import (
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type EmailUser struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	VHash             string `json:"v_hash" bson:"v_hash"`
	Email             string `json:"email" bson:"email"`
	PasswordHash      string `json:"-" bson:"password_hash"`
	EmailVerified     bool   `json:"email_verified" bson:"email_verified"`
	DisplayName       string `json:"display_name" bson:"display_name"`
}

func NewEmailUser(vHash, email, passwordHash, displayName string, emailVerified bool, resourceOwner shared.ResourceOwner) *EmailUser {
	entity := shared.NewEntity(resourceOwner)
	return &EmailUser{
		BaseEntity:    entity,
		VHash:         vHash,
		Email:         email,
		PasswordHash:  passwordHash,
		EmailVerified: emailVerified,
		DisplayName:   displayName,
	}
}

func (e EmailUser) GetID() uuid.UUID {
	return e.ID
}
