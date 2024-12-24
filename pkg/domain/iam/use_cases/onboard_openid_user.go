package iam_use_cases

import (
	"context"
	"log/slog"

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

	user := iam_entities.NewUser(cmd.Name, rxn)

	user, err = uc.UserWriter.Create(ctx, user)

	if err != nil {
		slog.ErrorContext(ctx, "error creating user", "err",
			err)
		return nil, err
	}

	group := iam_entities.NewGroup(user.ID, "private:default", iam_entities.GroupTypeSystem, rxn)

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
