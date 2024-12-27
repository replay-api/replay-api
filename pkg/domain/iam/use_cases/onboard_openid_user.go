package iam_use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
)

type OnboardOpenIDUserUseCase struct {
	UserReader    iam_out.UserReader
	UserWriter    iam_out.UserWriter
	ProfileReader iam_out.ProfileReader
	ProfileWriter iam_out.ProfileWriter
	GroupWriter   iam_out.GroupWriter
}

func NewOnboardOpenIDUserUseCase(userReader iam_out.UserReader, userWriter iam_out.UserWriter, profileReader iam_out.ProfileReader, profileWriter iam_out.ProfileWriter, groupWriter iam_out.GroupWriter) *OnboardOpenIDUserUseCase {
	return &OnboardOpenIDUserUseCase{
		UserReader:    userReader,
		UserWriter:    userWriter,
		ProfileReader: profileReader,
		ProfileWriter: profileWriter,
		GroupWriter:   groupWriter,
	}
}

func (uc *OnboardOpenIDUserUseCase) Exec(ctx context.Context, cmd iam_in.OnboardOpenIDUserCommand) (*iam_entities.Profile, error) {
	profileSourceKeySearch := uc.newSearchByProfileSourceKey(ctx, cmd.Source, cmd.Key)

	profiles, err := uc.ProfileReader.Search(ctx, profileSourceKeySearch)

	if err != nil {
		slog.ErrorContext(ctx, "error getting user profile", "err",
			err)
		return nil, err
	}

	if len(profiles) > 0 {
		return &profiles[0], nil
	}

	rxn := common.GetResourceOwner(ctx)

	if rxn.UserID == uuid.Nil {
		return nil, fmt.Errorf("invalid resource owner: no user id")
	}

	if rxn.GroupID == uuid.Nil {
		return nil, fmt.Errorf("invalid resource owner: no group id")
	}

	user := iam_entities.NewUser(rxn.UserID, cmd.Name, rxn)

	user, err = uc.UserWriter.Create(ctx, user)

	rxn.UserID = user.ID

	if err != nil {
		slog.ErrorContext(ctx, "error creating user", "err",
			err)
		return nil, err
	}

	group := iam_entities.NewGroup(rxn.GroupID, "private:default", iam_entities.GroupTypeSystem, rxn)

	rxn.GroupID = group.ID

	group, err = uc.GroupWriter.Create(ctx, group)

	if err != nil {
		slog.ErrorContext(ctx, "error creating group", "err",
			err)
		return nil, err
	}

	profile := iam_entities.NewProfile(user.ID, group.ID, cmd.Source, cmd.Key, nil, rxn)

	profile, err = uc.ProfileWriter.Create(ctx, profile)

	if err != nil {
		slog.ErrorContext(ctx, "error creating user profile", "err",
			err)
		return nil, err
	}

	return profile, nil
}

func (uc *OnboardOpenIDUserUseCase) newSearchByProfileSourceKey(ctx context.Context, source iam_entities.RIDSourceKey, key string) common.Search {
	params := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					Operator: common.EqualsOperator,
					ValueParams: []common.SearchableValue{
						{
							Field: "RIDSource",
							Values: []interface{}{
								source,
							},
						},
						{
							Field: "SourceKey",
							Values: []interface{}{
								key,
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
