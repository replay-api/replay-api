package steam_in

import (
	"context"

	e "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
)

type OnboardSteamUserCommand interface {
	Exec(ctx context.Context, steamUser *e.SteamUser) (*e.SteamUser, error)
	Validate(ctx context.Context, steamUser *e.SteamUser) error
}
