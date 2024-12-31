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
	UserReader     iam_out.UserReader
	UserWriter     iam_out.UserWriter
	ProfileReader  iam_out.ProfileReader
	ProfileWriter  iam_out.ProfileWriter
	GroupWriter    iam_out.GroupWriter
	CreateRIDToken iam_in.CreateRIDTokenCommand
}

func NewOnboardOpenIDUserUseCase(userReader iam_out.UserReader, userWriter iam_out.UserWriter, profileReader iam_out.ProfileReader, profileWriter iam_out.ProfileWriter, groupWriter iam_out.GroupWriter, createRIDToken iam_in.CreateRIDTokenCommand) *OnboardOpenIDUserUseCase {
	return &OnboardOpenIDUserUseCase{
		UserReader:     userReader,
		UserWriter:     userWriter,
		ProfileReader:  profileReader,
		ProfileWriter:  profileWriter,
		GroupWriter:    groupWriter,
		CreateRIDToken: createRIDToken,
	}
}

func (uc *OnboardOpenIDUserUseCase) Exec(ctx context.Context, cmd iam_in.OnboardOpenIDUserCommand) (*iam_entities.Profile, *iam_entities.RIDToken, error) {
	profileSourceKeySearch := uc.newSearchByProfileSourceKey(ctx, cmd.Source, cmd.Key)

	profiles, err := uc.ProfileReader.Search(ctx, profileSourceKeySearch)

	if err != nil {
		slog.ErrorContext(ctx, "error getting user profile", "err",
			err)
		return nil, nil, err
	}

	slog.InfoContext(ctx, fmt.Sprintf("profileSourceKeySearch: %v, profiles %v, len: %d", profileSourceKeySearch, profiles, len(profiles)))

	if len(profiles) > 0 {
		ridToken, err := uc.CreateRIDToken.Exec(ctx, profiles[0].GetResourceOwner(ctx), cmd.Source, common.ClientApplicationAudienceIDKey)

		if err != nil {
			slog.ErrorContext(ctx, "error creating rid token", "err",
				err)
			return nil, nil, err
		}

		return &profiles[0], ridToken, nil
	}

	rxn := common.GetResourceOwner(ctx)

	if rxn.UserID == uuid.Nil {
		return nil, nil, fmt.Errorf("invalid resource owner: no user id")
	}

	if rxn.GroupID == uuid.Nil {
		return nil, nil, fmt.Errorf("invalid resource owner: no group id")
	}

	user := iam_entities.NewUser(rxn.UserID, cmd.Name, rxn)

	user, err = uc.UserWriter.Create(ctx, user)

	rxn.UserID = user.ID

	if err != nil {
		slog.ErrorContext(ctx, "error creating user", "err",
			err)
		return nil, nil, err
	}

	group := iam_entities.NewGroup(rxn.GroupID, "private:default", iam_entities.GroupTypeSystem, rxn)

	rxn.GroupID = group.ID

	group, err = uc.GroupWriter.Create(ctx, group)

	if err != nil {
		slog.ErrorContext(ctx, "error creating group", "err",
			err)
		return nil, nil, err
	}

	profile := iam_entities.NewProfile(user.ID, group.ID, cmd.Source, cmd.Key, nil, rxn)

	profile, err = uc.ProfileWriter.Create(ctx, profile)

	if err != nil {
		slog.ErrorContext(ctx, "error creating user profile", "err",
			err)
		return nil, nil, err
	}

	ridToken, err := uc.CreateRIDToken.Exec(ctx, profiles[0].GetResourceOwner(ctx), cmd.Source, common.ClientApplicationAudienceIDKey)

	if err != nil {
		slog.ErrorContext(ctx, "error creating rid token", "err",
			err)
		return nil, nil, err
	}

	if ridToken == nil {
		return nil, nil, fmt.Errorf("failed to create rid token: token is nil")
	}

	slog.InfoContext(ctx, "onboarded user", "user", user, "group", group, "profile", profile, "ridToken", ridToken)

	return profile, ridToken, nil
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
