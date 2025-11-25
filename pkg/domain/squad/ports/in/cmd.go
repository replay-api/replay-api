package squad_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
)

type CreateOrUpdatedSquadCommand struct {
	Name          string                                `json:"name"`
	Symbol        string                                `json:"symbol"`
	Description   string                                `json:"description"`
	GameID        common.GameIDKey                      `json:"game_id"`
	SlugURI       string                                `json:"slug_uri"`
	Members       map[string]CreateSquadMembershipInput `json:"members"`
	Base64Logo    string                                `json:"base64_logo"`
	LogoExtension string                                `json:"logo_extension"`
	Links         map[string]string                     `json:"links"`
}

type CreateSquadMembershipInput struct {
	Status squad_value_objects.SquadMembershipStatus `json:"status" bson:"status"`
	Type   squad_value_objects.SquadMembershipType   `json:"type" bson:"type"`
	Roles  []string                                  `json:"roles" bson:"roles"`
}

type CreateSquadCommandHandler interface {
	Exec(c context.Context, cmd CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error)
}

type UpdateSquadCommandHandler interface {
	Exec(c context.Context, squadID uuid.UUID, cmd CreateOrUpdatedSquadCommand) (*squad_entities.Squad, error)
}

type DeleteSquadCommandHandler interface {
	Exec(c context.Context, squadID uuid.UUID) error
}

type AddSquadMemberCommand struct {
	SquadID  uuid.UUID                                `json:"squad_id"`
	PlayerID uuid.UUID                                `json:"player_id"`
	Type     squad_value_objects.SquadMembershipType `json:"type"`
	Roles    []string                                 `json:"roles"`
}

type AddSquadMemberCommandHandler interface {
	Exec(c context.Context, cmd AddSquadMemberCommand) (*squad_entities.Squad, error)
}

type RemoveSquadMemberCommand struct {
	SquadID  uuid.UUID `json:"squad_id"`
	PlayerID uuid.UUID `json:"player_id"`
}

type RemoveSquadMemberCommandHandler interface {
	Exec(c context.Context, cmd RemoveSquadMemberCommand) (*squad_entities.Squad, error)
}

type UpdateSquadMemberRoleCommand struct {
	SquadID  uuid.UUID `json:"squad_id"`
	PlayerID uuid.UUID `json:"player_id"`
	Roles    []string  `json:"roles"`
}

type UpdateSquadMemberRoleCommandHandler interface {
	Exec(c context.Context, cmd UpdateSquadMemberRoleCommand) (*squad_entities.Squad, error)
}

type CreatePlayerProfileCommand struct {
	GameID          common.GameIDKey         `json:"game_id"`
	Nickname        string                   `json:"nickname"`
	Base64Avatar    string                   `json:"base64_avatar"`
	AvatarExtension string                   `json:"avatar_extension"`
	SlugURI         string                   `json:"slug_uri"`
	Roles           []string                 `json:"roles"`
	Description     string                   `json:"description"`
	VisibilityType  common.VisibilityTypeKey `json:"visibility_type"`
}

type CreatePlayerProfileCommandHandler interface {
	Exec(c context.Context, cmd CreatePlayerProfileCommand) (*squad_entities.PlayerProfile, error)
}

type UpdatePlayerCommand struct {
	PlayerID        uuid.UUID `json:"player_id"`
	Nickname        string    `json:"nickname"`
	Base64Avatar    string    `json:"base64_avatar"`
	AvatarExtension string    `json:"avatar_extension"`
	SlugURI         string    `json:"slug_uri"`
	Roles           []string  `json:"roles"`
	Description     string    `json:"description"`
}

type UpdatePlayerProfileCommandHandler interface {
	Exec(c context.Context, cmd UpdatePlayerCommand) (*squad_entities.PlayerProfile, error)
}

type DeletePlayerProfileCommandHandler interface {
	Exec(c context.Context, playerID uuid.UUID) error
}
