package squad_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	squad_value_objects "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/value-objects"
)

type Member struct {
	ID            uuid.UUID                              `json:"id" bson:"_id"`
	UserID        uuid.UUID                              `json:"user_id" bson:"user_id"`
	RIDSource     iam_entities.RIDSourceKey              `json:"rid_source" bson:"rid_source"`
	Name          string                                 `json:"name" bson:"name"`
	ShortName     string                                 `json:"short_name" bson:"short_name"`
	Profiles      map[string]squad_value_objects.Profile `json:"profiles" bson:"profiles"`
	ResourceOwner common.ResourceOwner                   `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time                              `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time                              `json:"updated_at" bson:"updated_at"`
}
