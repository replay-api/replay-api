package steam

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type GoogleUser struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	VHash         string               `json:"v_hash" bson:"v_hash"`
	Sub           string               `json:"sub" bson:"sub"`
	Hd            string               `json:"hd" bson:"hd"`
	GivenName     string               `json:"given_name" bson:"given_name"`
	FamilyName    string               `json:"family_name" bson:"family_name"`
	Email         string               `json:"email" bson:"email"`
	Locale        string               `json:"locale" bson:"locale"`
	EmailVerified bool                 `json:"email_verified" bson:"email_verified"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

func (s GoogleUser) GetID() uuid.UUID {
	return s.ID
}
