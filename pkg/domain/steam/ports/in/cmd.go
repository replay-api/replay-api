package steam_in

import (
	"context"

	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	steam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
)

type OnboardSteamUserCommand interface {
	Exec(ctx context.Context, steamUser *steam_entities.SteamUser) (*steam_entities.SteamUser, *iam_entities.RIDToken, error)
	Validate(ctx context.Context, steamUser *steam_entities.SteamUser) error
}
