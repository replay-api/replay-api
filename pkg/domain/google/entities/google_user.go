package google

import (
	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type GoogleUser struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	VHash             string `json:"v_hash" bson:"v_hash"`
	Sub               string `json:"sub" bson:"sub"`
	Hd                string `json:"hd" bson:"hd"`
	GivenName         string `json:"given_name" bson:"given_name"`
	FamilyName        string `json:"family_name" bson:"family_name"`
	Email             string `json:"email" bson:"email"`
	Locale            string `json:"locale" bson:"locale"`
	EmailVerified     bool   `json:"email_verified" bson:"email_verified"`
}

func NewGoogleUser(vHash, sub, hd, givenName, familyName, email, locale string, emailVerified bool, resourceOwner shared.ResourceOwner) *GoogleUser {
	entity := shared.NewEntity(resourceOwner)
	return &GoogleUser{
		BaseEntity:    entity,
		VHash:         vHash,
		Sub:           sub,
		Hd:            hd,
		GivenName:     givenName,
		FamilyName:    familyName,
		Email:         email,
		Locale:        locale,
		EmailVerified: emailVerified,
	}
}

func (s GoogleUser) GetID() uuid.UUID {
	return s.ID
}
