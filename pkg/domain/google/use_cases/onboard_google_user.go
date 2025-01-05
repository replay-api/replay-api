package google_use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"

	"github.com/psavelis/team-pro/replay-api/pkg/domain/google"
	google_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/google/entities"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"

	google_in "github.com/psavelis/team-pro/replay-api/pkg/domain/google/ports/in"
	google_out "github.com/psavelis/team-pro/replay-api/pkg/domain/google/ports/out"
)

type OnboardGoogleUserUseCase struct {
	GoogleUserWriter  google_out.GoogleUserWriter
	GoogleUserReader  google_out.GoogleUserReader
	VHashWriter       google_out.VHashWriter
	OnboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler
}

func (usecase *OnboardGoogleUserUseCase) Validate(ctx context.Context, googleUser *google_entity.GoogleUser) error {
	if googleUser.Email == "" {
		slog.ErrorContext(ctx, "google email is required", "google.Email", googleUser.Email)
		return google.NewGoogleIDRequiredError()
	}

	if googleUser.VHash == "" {
		slog.ErrorContext(ctx, "vHash is required", "vHash", googleUser.VHash)
		return google.NewVHashRequiredError()
	}

	expectedVHash := usecase.VHashWriter.CreateVHash(ctx, googleUser.Email)

	if googleUser.VHash != expectedVHash {
		slog.ErrorContext(ctx, "vHash does not match", "google.Email", googleUser.Email, "vHash", googleUser.VHash, "expectedVHash", expectedVHash)
		return google.NewInvalidVHashError(googleUser.VHash)
	}

	return nil
}

func (usecase *OnboardGoogleUserUseCase) Exec(ctx context.Context, googleUser *google_entity.GoogleUser) (*google_entity.GoogleUser, *iam_entities.RIDToken, error) {
	vhashSearch := usecase.newSearchByVHash(ctx, googleUser.VHash)

	googleUserResult, err := usecase.GoogleUserReader.Search(ctx, vhashSearch)

	if err != nil {
		slog.ErrorContext(ctx, "error getting google user", "err",
			err)
		return nil, nil, err
	}

	// slog.InfoContext(ctx, fmt.Sprintf("vhashSearch: %v, googleUserResult %v, len: %d", vhashSearch, googleUserResult, len(googleUserResult)))

	if len(googleUserResult) > 0 {
		if googleUser.Email != googleUserResult[0].Email {
			slog.ErrorContext(ctx, "googleID does not match", "googleID", googleUser.Email, "googleUser.Email", googleUserResult[0].Email)
			return nil, nil, google.NewGoogleIDMismatchError(googleUser.Email)
		}

		// TODO: update with new data (!)

		// return &googleUserResult[0], nil, nil

		googleUser = &googleUserResult[0]

		ctx = context.WithValue(ctx, common.UserIDKey, googleUser.ResourceOwner.UserID)
		ctx = context.WithValue(ctx, common.GroupIDKey, googleUser.ResourceOwner.GroupID)
	}

	profile, ridToken, err := usecase.OnboardOpenIDUser.Exec(ctx, iam_in.OnboardOpenIDUserCommand{
		Name:   fmt.Sprintf("%s %s", googleUser.GivenName, googleUser.FamilyName),
		Source: iam_entities.RIDSource_Google,
		Key:    googleUser.Email,
	})

	if err != nil {
		slog.ErrorContext(ctx, "error creating user profile", "err", err)
		return nil, nil, google.NewGoogleUserCreationError(fmt.Sprintf("error creating user profile: %v", googleUser.Email))
	}

	if ridToken == nil {
		slog.ErrorContext(ctx, "error creating rid token", "err", err)
		return nil, nil, google.NewGoogleUserCreationError(fmt.Sprintf("error creating rid token: %v", googleUser.Email))
	}

	ctx = context.WithValue(ctx, common.UserIDKey, profile.ResourceOwner.UserID)
	ctx = context.WithValue(ctx, common.GroupIDKey, profile.ResourceOwner.GroupID)

	googleUser.ResourceOwner = common.GetResourceOwner(ctx)

	if googleUser.ID == uuid.Nil {
		googleUser.ID = profile.ResourceOwner.UserID
	}

	if len(googleUserResult) == 0 {
		slog.InfoContext(ctx, fmt.Sprintf("attempt to create google user: %v", googleUser))
		googleUser, err = usecase.GoogleUserWriter.Create(ctx, googleUser)

		if err != nil {
			slog.ErrorContext(ctx, "error creating google user: error", "err", err)
			return nil, nil, google.NewGoogleUserCreationError(fmt.Sprintf("error creating google user: %v", googleUser.ID))
		}

		if googleUser == nil {
			slog.ErrorContext(ctx, "error creating google user: user is nil", "err",
				err)
			return nil, nil, google.NewGoogleUserCreationError(fmt.Sprintf("unable to create google user: %v", googleUser))
		}
	}

	// TODO: update user profileMap googleID (futuramente conseguir unir as contas) // talvez bater email

	return googleUser, ridToken, nil
}

func NewOnboardGoogleUserUseCase(googleUserWriter google_out.GoogleUserWriter, googleUserReader google_out.GoogleUserReader, vHashWriter google_out.VHashWriter, onboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler) google_in.OnboardGoogleUserCommand {
	return &OnboardGoogleUserUseCase{
		GoogleUserWriter: googleUserWriter, GoogleUserReader: googleUserReader, VHashWriter: vHashWriter, OnboardOpenIDUser: onboardOpenIDUser,
	}
}

func (uc *OnboardGoogleUserUseCase) newSearchByVHash(ctx context.Context, vhashString string) common.Search {
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
