package email_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type EmailUser struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	VHash         string               `json:"v_hash" bson:"v_hash"`
	Email         string               `json:"email" bson:"email"`
	PasswordHash  string               `json:"-" bson:"password_hash"`
	EmailVerified bool                 `json:"email_verified" bson:"email_verified"`
	DisplayName   string               `json:"display_name" bson:"display_name"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func (e EmailUser) GetID() uuid.UUID {
	return e.ID
}
