package entities

type Profile struct {
	// ID uuid.UUID `json:"id" bson:"_id"` // TODO: rever se precisa realmente manter o default como id (já q não utiliza o default objectId) ProfileID ou ID
	// // ProfileType??

	// // Profile (ver se precisa mesmo normalizar, talvez não precise, pois o player pode representar a entidade de perfil)
	// CurrentDisplayName string     `json:"display_name" bson:"display_name"`
	// NameHistory        []string   `json:"name_history" bson:"name_history"`
	// AvatarURI          string     `json:"avatar_uri" bson:"avatar_uri"`
	// Description        string     `json:"description" bson:"description"`
	// VerifiedAt         *time.Time `json:"verified_at" bson:"verified_at"`

	// // Team (TODO: especificar)
	// CurrentDisplayClanTag string   `json:"display_clan_tag" bson:"display_clan_tag"`
	// ClanHistory           []string `json:"clan_history" bson:"clan_history"` // TODO: Usar o ID do time

	// // i18n
	// City             string            `json:"city" bson:"city"`
	// Country          string            `json:"country" bson:"country"`
	// Language         string            `json:"language" bson:"language"`
	// Timezone         string            `json:"timezone" bson:"timezone"`
	// ExternalProfiles map[string]string `json:"external_profiles" bson:"external_profiles"`

	// ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	// CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	// UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

// func (e Profile) GetID() uuid.UUID {
// 	return e.ID
// }
