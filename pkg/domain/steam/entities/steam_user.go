package steam_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type SteamUser struct {
	ID            uuid.UUID            `json:"id" bson:"_id"`
	VHash         string               `json:"v_hash" bson:"v_hash"`
	Name          string               `json:"name" bson:"name"`
	Email         string               `json:"email" bson:"email"`
	Image         string               `json:"image" bson:"image"`
	Steam         Steam                `json:"steam" bson:"steam"`
	ResourceOwner shared.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

type Steam struct {
	ID                       string    `json:"id" bson:"_id"`
	CommunityVisibilityState int       `json:"communityvisibilitystate" bson:"communityvisibilitystate"`
	ProfileState             int       `json:"profilestate" bson:"profilestate"`
	PersonaName              string    `json:"personaname" bson:"personaname"`
	ProfileURL               string    `json:"profileurl" bson:"profileurl"`
	Avatar                   string    `json:"avatar" bson:"avatar"`
	AvatarMedium             string    `json:"avatarmedium" bson:"avatarmedium"`
	AvatarFull               string    `json:"avatarfull" bson:"avatarfull"`
	AvatarHash               string    `json:"avatarhash" bson:"avatarhash"`
	PersonaState             int       `json:"personastate" bson:"personastate"`
	RealName                 string    `json:"realname" bson:"realname"`
	PrimaryClanID            string    `json:"primaryclanid" bson:"primaryclanid"`
	TimeCreated              time.Time `json:"timecreated" bson:"timecreated"`
	PersonaStateFlags        int       `json:"personastateflags" bson:"personastateflags"`
}

func (s SteamUser) GetID() uuid.UUID {
	return s.ID
}
