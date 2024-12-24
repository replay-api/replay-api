package use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	"github.com/psavelis/team-pro/replay-api/pkg/domain/steam"
	steam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"

	steam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/in"
	steam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/out"
)

type OnboardSteamUserUseCase struct {
	SteamUserWriter steam_out.SteamUserWriter
	SteamUserReader steam_out.SteamUserReader
	VHashWriter     steam_out.VHashWriter
}

func (usecase *OnboardSteamUserUseCase) Validate(ctx context.Context, steamUser *steam_entity.SteamUser) error {
	if steamUser.Steam.ID == "" {
		slog.ErrorContext(ctx, "steamID is required", "steamID", steamUser.Steam.ID)
		return steam.NewSteamIDRequiredError()
	}

	if steamUser.VHash == "" {
		slog.ErrorContext(ctx, "vHash is required", "vHash", steamUser.VHash)
		return steam.NewVHashRequiredError()
	}

	expectedVHash := usecase.VHashWriter.CreateVHash(ctx, steamUser.Steam.ID)

	if steamUser.VHash != expectedVHash {
		slog.ErrorContext(ctx, "vHash does not match", "steamID", steamUser.Steam.ID, "vHash", steamUser.VHash, "expectedVHash", expectedVHash)
		return steam.NewInvalidVHashError(steamUser.VHash)
	}

	return nil
}

func (usecase *OnboardSteamUserUseCase) Exec(ctx context.Context, steamUser *steam_entity.SteamUser) (*steam_entity.SteamUser, error) {
	vhashSearch := usecase.newSearchByVHash(ctx, steamUser.VHash)

	steamUserResult, err := usecase.SteamUserReader.Search(ctx, vhashSearch)

	if err != nil {
		slog.ErrorContext(ctx, "error getting steam user", "err",
			err)
		return nil, err
	}

	// slog.InfoContext(ctx, fmt.Sprintf("vhashSearch: %v, steamUserResult %v, len: %d", vhashSearch, steamUserResult, len(steamUserResult)))

	if len(steamUserResult) > 0 {
		if steamUser.Steam.ID != steamUserResult[0].Steam.ID {
			slog.ErrorContext(ctx, "steamID does not match", "steamID", steamUser.Steam.ID, "steamUser.Steam.ID", steamUserResult[0].Steam.ID)
			return nil, steam.NewSteamIDMismatchError(steamUser.Steam.ID)
		}

		// TODO: update with new data (!)

		return &steamUserResult[0], nil
	}

	// TODO: create User (link UserID)

	steamUser.ID = uuid.New()

	user, err := usecase.SteamUserWriter.Create(ctx, steamUser)

	if err != nil {
		slog.ErrorContext(ctx, "error creating steam user", "err", err)
		return nil, steam.NewSteamUserCreationError(fmt.Sprintf("error creating steam user: %v", steamUser.ID))
	}

	if user == nil {
		slog.ErrorContext(ctx, "error creating steam user", "err",
			err)
		return nil, steam.NewSteamUserCreationError(fmt.Sprintf("unable to create steam user: %v", steamUser.ID))
	}

	// TODO: update user profileMap steamID (futuramente conseguir unir as contas)

	return user, nil
}

func NewOnboardSteamUserUseCase(steamUserWriter steam_out.SteamUserWriter, steamUserReader steam_out.SteamUserReader, vHashWriter steam_out.VHashWriter) steam_in.OnboardSteamUserCommand {
	return &OnboardSteamUserUseCase{
		SteamUserWriter: steamUserWriter, SteamUserReader: steamUserReader, VHashWriter: vHashWriter,
	}
}

func (uc *OnboardSteamUserUseCase) newSearchByVHash(ctx context.Context, vhashString string) common.Search {
	params := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{
							Field: "VHash",
							Values: []interface{}{
								vhashString,
							},
						},
					},
				},
			},
		},
	}

	visibility := common.SearchVisibilityOptions{
		RequestSource:    common.GetResourceOwner(ctx),
		IntendedAudience: common.ClientApplicationAudienceIDKey,
	}

	result := common.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return common.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}
