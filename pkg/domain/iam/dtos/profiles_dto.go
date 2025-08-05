package iam_dtos

import (
	"github.com/google/uuid"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
)

type ProfilesDTO struct {
	ActiveProfiles map[iam_entities.ProfileType]uuid.UUID `json:"active_profiles"` // key is the squad or player or tournament
	Profiles       []iam_entities.Profile                 `json:"profiles"`
}
