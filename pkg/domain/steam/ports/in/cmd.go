package steam_in

import (
	"context"

	e "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
)

type OnboardSteamUserCommand interface {
	Exec(ctx context.Context, steamID string, vHash string) (*e.SteamUser, error)
	Validate(ctx context.Context, steamID string, vHash string) error
}
