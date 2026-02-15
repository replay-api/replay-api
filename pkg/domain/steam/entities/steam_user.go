package steam_entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type SteamUser struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	VHash             string `json:"v_hash" bson:"v_hash"`
	Name              string `json:"name" bson:"name"`
	Email             string `json:"email" bson:"email"`
	Image             string `json:"image" bson:"image"`
	Steam             Steam  `json:"steam" bson:"steam"`
}

func NewSteamUser(vHash, name, email, image string, steam Steam, resourceOwner shared.ResourceOwner) *SteamUser {
	entity := shared.NewEntity(resourceOwner)
	return &SteamUser{
		BaseEntity: entity,
		VHash:      vHash,
		Name:       name,
		Email:      email,
		Image:      image,
		Steam:      steam,
	}
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
