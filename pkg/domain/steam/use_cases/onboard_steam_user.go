package use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	"github.com/replay-api/replay-api/pkg/domain/steam"
	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"

	steam_in "github.com/replay-api/replay-api/pkg/domain/steam/ports/in"
	steam_out "github.com/replay-api/replay-api/pkg/domain/steam/ports/out"
)

type OnboardSteamUserUseCase struct {
	SteamUserWriter   steam_out.SteamUserWriter
	SteamUserReader   steam_out.SteamUserReader
	VHashWriter       steam_out.VHashWriter
	OnboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler
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

func (usecase *OnboardSteamUserUseCase) Exec(ctx context.Context, steamUser *steam_entity.SteamUser) (*steam_entity.SteamUser, *iam_entities.RIDToken, error) {
	vhashSearch := usecase.newSearchByVHash(ctx, steamUser.VHash)

	steamUserResult, err := usecase.SteamUserReader.Search(ctx, vhashSearch)

	if err != nil {
		slog.ErrorContext(ctx, "error getting steam user", "err",
			err)
		return nil, nil, err
	}

	// slog.InfoContext(ctx, fmt.Sprintf("vhashSearch: %v, steamUserResult %v, len: %d", vhashSearch, steamUserResult, len(steamUserResult)))

	if len(steamUserResult) > 0 {
		if steamUser.Steam.ID != steamUserResult[0].Steam.ID {
			slog.ErrorContext(ctx, "steamID does not match", "steamID", steamUser.Steam.ID, "steamUser.Steam.ID", steamUserResult[0].Steam.ID)
			return nil, nil, steam.NewSteamIDMismatchError(steamUser.Steam.ID)
		}

		// TODO: update with new data (!)

		// return &steamUserResult[0], nil, nil

		steamUser = &steamUserResult[0]

		ctx = context.WithValue(ctx, shared.UserIDKey, steamUser.ResourceOwner.UserID)
		ctx = context.WithValue(ctx, shared.GroupIDKey, steamUser.ResourceOwner.GroupID)
	}

	profile, ridToken, err := usecase.OnboardOpenIDUser.Exec(ctx, iam_in.OnboardOpenIDUserCommand{
		Name:           steamUser.Steam.RealName,
		Source:         iam_entities.RIDSource_Steam,
		Key:            steamUser.Steam.ID,
		ProfileDetails: steamUser.Steam,
	})

	if err != nil {
		slog.ErrorContext(ctx, "error creating user profile", "err", err)
		return nil, nil, steam.NewSteamUserCreationError(fmt.Sprintf("error creating user profile: %v", steamUser.Steam.ID))
	}

	if ridToken == nil {
		slog.ErrorContext(ctx, "error creating rid token", "err", err)
		return nil, nil, steam.NewSteamUserCreationError(fmt.Sprintf("error creating rid token: %v", steamUser.Steam.ID))
	}

	ctx = context.WithValue(ctx, shared.UserIDKey, profile.ResourceOwner.UserID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, profile.ResourceOwner.GroupID)

	steamUser.ResourceOwner = shared.GetResourceOwner(ctx)

	if steamUser.ID == uuid.Nil {
		steamUser.ID = profile.ResourceOwner.UserID
	}

	if len(steamUserResult) == 0 {
		slog.InfoContext(ctx, fmt.Sprintf("attempt to create steam user: %v", steamUser))
		steamUser, err = usecase.SteamUserWriter.Create(ctx, steamUser)

		if err != nil {
			slog.ErrorContext(ctx, "error creating steam user: error", "err", err)
			return nil, nil, steam.NewSteamUserCreationError(fmt.Sprintf("error creating steam user: %v", steamUser.ID))
		}

		if steamUser == nil {
			slog.ErrorContext(ctx, "error creating steam user: user is nil", "err",
				err)
			return nil, nil, steam.NewSteamUserCreationError(fmt.Sprintf("unable to create steam user: %v", steamUser.ID))
		}
	}

	// TODO: update user profileMap steamID (futuramente conseguir unir as contas)

	return steamUser, ridToken, nil
}

func NewOnboardSteamUserUseCase(steamUserWriter steam_out.SteamUserWriter, steamUserReader steam_out.SteamUserReader, vHashWriter steam_out.VHashWriter, onboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler) steam_in.OnboardSteamUserCommand {
	return &OnboardSteamUserUseCase{
		SteamUserWriter: steamUserWriter, SteamUserReader: steamUserReader, VHashWriter: vHashWriter, OnboardOpenIDUser: onboardOpenIDUser,
	}
}

func (uc *OnboardSteamUserUseCase) newSearchByVHash(ctx context.Context, vhashString string) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
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

	visibility := shared.SearchVisibilityOptions{
		RequestSource:    shared.GetResourceOwner(ctx),
		IntendedAudience: shared.ClientApplicationAudienceIDKey,
	}

	result := shared.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	return shared.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}
