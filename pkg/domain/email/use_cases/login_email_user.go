package email_use_cases

import (
	"context"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	"github.com/replay-api/replay-api/pkg/domain/email"
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	email_in "github.com/replay-api/replay-api/pkg/domain/email/ports/in"
	email_out "github.com/replay-api/replay-api/pkg/domain/email/ports/out"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
)

type LoginEmailUserUseCase struct {
	EmailUserReader    email_out.EmailUserReader
	VHashWriter        email_out.VHashWriter
	PasswordHasher     email_out.PasswordHasher
	CreateRIDToken     iam_in.CreateRIDTokenCommand
}

func (usecase *LoginEmailUserUseCase) Validate(ctx context.Context, emailAddr string, password string, vHash string) error {
	if emailAddr == "" {
		slog.ErrorContext(ctx, "email is required")
		return email.NewEmailRequiredError()
	}

	if password == "" {
		slog.ErrorContext(ctx, "password is required")
		return email.NewPasswordRequiredError()
	}

	if vHash == "" {
		slog.ErrorContext(ctx, "vHash is required")
		return email.NewVHashRequiredError()
	}

	expectedVHash := usecase.VHashWriter.CreateVHash(ctx, emailAddr)

	if vHash != expectedVHash {
		slog.ErrorContext(ctx, "vHash does not match", "email", emailAddr, "vHash", vHash, "expectedVHash", expectedVHash)
		return email.NewInvalidVHashError(vHash)
	}

	return nil
}

func (usecase *LoginEmailUserUseCase) Exec(ctx context.Context, emailAddr string, password string, vHash string) (*email_entities.EmailUser, *iam_entities.RIDToken, error) {
	// Find email user by email
	emailSearch := usecase.newSearchByEmail(ctx, emailAddr)
	emailUsers, err := usecase.EmailUserReader.Search(ctx, emailSearch)

	if err != nil {
		slog.ErrorContext(ctx, "error searching for email user", "err", err)
		return nil, nil, err
	}

	if len(emailUsers) == 0 {
		slog.ErrorContext(ctx, "email user not found", "email", emailAddr)
		return nil, nil, email.NewEmailUserNotFoundError(emailAddr)
	}

	emailUser := &emailUsers[0]

	// Verify password
	err = usecase.PasswordHasher.ComparePassword(ctx, emailUser.PasswordHash, password)
	if err != nil {
		slog.ErrorContext(ctx, "invalid password", "email", emailAddr)
		return nil, nil, email.NewInvalidPasswordError()
	}

	// Create new RID token
	ridToken, err := usecase.CreateRIDToken.Exec(
		ctx,
		emailUser.ResourceOwner,
		iam_entities.RIDSource_Email,
		common.UserAudienceIDKey,
	)

	if err != nil {
		slog.ErrorContext(ctx, "error creating rid token", "err", err)
		return nil, nil, err
	}

	return emailUser, ridToken, nil
}

func NewLoginEmailUserUseCase(
	emailUserReader email_out.EmailUserReader,
	vHashWriter email_out.VHashWriter,
	passwordHasher email_out.PasswordHasher,
	createRIDToken iam_in.CreateRIDTokenCommand,
) email_in.LoginEmailUserCommand {
	return &LoginEmailUserUseCase{
		EmailUserReader:    emailUserReader,
		VHashWriter:        vHashWriter,
		PasswordHasher:     passwordHasher,
		CreateRIDToken:     createRIDToken,
	}
}

func (uc *LoginEmailUserUseCase) newSearchByEmail(ctx context.Context, emailString string) common.Search {
	params := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{
							Field: "Email",
							Values: []interface{}{
								emailString,
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
