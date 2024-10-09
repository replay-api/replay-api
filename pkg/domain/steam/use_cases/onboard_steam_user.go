package use_cases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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

func (usecase *OnboardSteamUserUseCase) Validate(ctx context.Context, steamID string, vHash string) error {
	if steamID == "" {
		slog.ErrorContext(ctx, "steamID is required", "steamID", steamID)
		return steam.NewSteamIDRequiredError()
	}

	if vHash == "" {
		slog.ErrorContext(ctx, "vHash is required", "vHash", vHash)
		return steam.NewVHashRequiredError()
	}

	expectedVHash := usecase.VHashWriter.CreateVHash(ctx, steamID)

	if vHash != expectedVHash {
		slog.ErrorContext(ctx, "vHash does not match", "steamID", steamID, "vHash", vHash, "expectedVHash", expectedVHash)
		return steam.NewInvalidVHashError(vHash)
	}

	return nil
}

func (usecase *OnboardSteamUserUseCase) Exec(ctx context.Context, steamID string, vHash string) (*steam_entity.SteamUser, error) {
	vhashSearch := usecase.newSearchByVHash(ctx, vHash)

	steamUserResult, err := usecase.SteamUserReader.Search(ctx, vhashSearch)

	if err != nil {
		slog.ErrorContext(ctx, "error getting steam user", "err",
			err)
		return nil, err
	}

	if len(steamUserResult) > 0 {
		if steamID != steamUserResult[0].Steam.ID {
			slog.ErrorContext(ctx, "steamID does not match", "steamID", steamID, "steamUser.Steam.ID", steamUserResult[0].Steam.ID)
			return nil, steam.NewSteamIDMismatchError(steamID)
		}

		return &steamUserResult[0], nil
	}

	userID := uuid.New()

	// TODO: deprecated! apenas usar o display_name do common.User
	user, err := usecase.SteamUserWriter.Create(ctx, steam_entity.SteamUser{
		ID:    userID,
		VHash: vHash,
		Name:  "",                                                                                                                                                                                                                                                                             // TODO: deprecated! apenas usar o display_name do common.User                                                                                                                                                                                                                                                                          // TODO: deprecated! apenas usar o display_name do common.User                                                                                                                                                                                                                                                                             // TODO: deprecated! apenas usar o display_name do common.User
		Email: "",                                                                                                                                                                                                                                                                             // TODO: deprecated! apenas usar o display_name do common.User                                                                                                                                                                                                                                                                         //
		Image: "",                                                                                                                                                                                                                                                                             // rm
		Steam: steam_entity.Steam{ID: steamID, CommunityVisibilityState: 0, ProfileState: 0, PersonaName: "", ProfileURL: "", Avatar: "", AvatarMedium: "", AvatarFull: "", AvatarHash: "", PersonaState: 0, RealName: "", PrimaryClanID: "", TimeCreated: time.Time{}, PersonaStateFlags: 0}, // TODO: TimeCreated talvez possa ser solicitado o consentimento do usuario para exibição
		ResourceOwner: common.ResourceOwner{
			TenantID: common.TeamPROTenantID,
			UserID:   userID,
			GroupID:  uuid.Nil,
			ClientID: common.TeamPROAppClientID,
		},
	})

	if err != nil {
		slog.ErrorContext(ctx, "error creating steam user", "err", err)
		return nil, steam.NewSteamUserCreationError(fmt.Sprintf("error creating steam user: %v", userID))
	}

	if user == nil {
		slog.ErrorContext(ctx, "error creating steam user", "err",
			err)
		return nil, steam.NewSteamUserCreationError(fmt.Sprintf("unable to create steam user: %v", userID))
	}

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
