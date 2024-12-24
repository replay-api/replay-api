package iam_use_cases

import (
	"context"
	"log/slog"

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

	userProfileResult, err := uc.ProfileReader.Search(ctx, profileSourceKeySearch)

	if err != nil {
		slog.ErrorContext(ctx, "error getting user profile", "err",
			err)
		return nil, err
	}

	if len(userProfileResult) > 0 {
		return &userProfileResult[0], nil
	}

	// create user, group, profile

	// retornar user profile
}
