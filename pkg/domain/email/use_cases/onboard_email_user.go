package email_use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/replay-api/replay-api/pkg/domain/email"
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	email_in "github.com/replay-api/replay-api/pkg/domain/email/ports/in"
	email_out "github.com/replay-api/replay-api/pkg/domain/email/ports/out"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
)

type OnboardEmailUserUseCase struct {
	EmailUserWriter   email_out.EmailUserWriter
	EmailUserReader   email_out.EmailUserReader
	VHashWriter       email_out.VHashWriter
	PasswordHasher    email_out.PasswordHasher
	OnboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler
}

func (usecase *OnboardEmailUserUseCase) Validate(ctx context.Context, emailUser *email_entities.EmailUser, password string) error {
	if emailUser.Email == "" {
		slog.ErrorContext(ctx, "email is required", "email", emailUser.Email)
		return email.NewEmailRequiredError()
	}

	if password == "" {
		slog.ErrorContext(ctx, "password is required")
		return email.NewPasswordRequiredError()
	}

	if len(password) < 8 {
		slog.ErrorContext(ctx, "password is too weak")
		return email.NewPasswordTooWeakError()
	}

	if emailUser.VHash == "" {
		slog.ErrorContext(ctx, "vHash is required", "vHash", emailUser.VHash)
		return email.NewVHashRequiredError()
	}

	expectedVHash := usecase.VHashWriter.CreateVHash(ctx, emailUser.Email)

	if emailUser.VHash != expectedVHash {
		slog.ErrorContext(ctx, "vHash does not match", "email", emailUser.Email, "vHash", emailUser.VHash, "expectedVHash", expectedVHash)
		return email.NewInvalidVHashError(emailUser.VHash)
	}

	return nil
}

func (usecase *OnboardEmailUserUseCase) Exec(ctx context.Context, emailUser *email_entities.EmailUser, password string) (*email_entities.EmailUser, *iam_entities.RIDToken, error) {
	// Check if email already exists
	emailSearch := usecase.newSearchByEmail(ctx, emailUser.Email)
	existingUsers, err := usecase.EmailUserReader.Search(ctx, emailSearch)

	if err != nil {
		slog.ErrorContext(ctx, "error checking for existing email user", "err", err)
		return nil, nil, err
	}

	if len(existingUsers) > 0 {
		slog.ErrorContext(ctx, "email already exists", "email", emailUser.Email)
		return nil, nil, email.NewEmailAlreadyExistsError(emailUser.Email)
	}

	// Hash password
	hashedPassword, err := usecase.PasswordHasher.HashPassword(ctx, password)
	if err != nil {
		slog.ErrorContext(ctx, "error hashing password", "err", err)
		return nil, nil, email.NewEmailUserCreationError("failed to hash password")
	}
	emailUser.PasswordHash = hashedPassword

	// Set display name if not provided
	if emailUser.DisplayName == "" {
		// Extract name from email (before @)
		for i, c := range emailUser.Email {
			if c == '@' {
				emailUser.DisplayName = emailUser.Email[:i]
				break
			}
		}
	}

	// Create profile and user through IAM
	profile, ridToken, err := usecase.OnboardOpenIDUser.Exec(ctx, iam_in.OnboardOpenIDUserCommand{
		Name:           emailUser.DisplayName,
		Source:         iam_entities.RIDSource_Email,
		Key:            emailUser.Email,
		ProfileDetails: emailUser,
	})

	if err != nil {
		slog.ErrorContext(ctx, "error creating user profile", "err", err)
		return nil, nil, email.NewEmailUserCreationError(fmt.Sprintf("error creating user profile: %v", emailUser.Email))
	}

	if ridToken == nil {
		slog.ErrorContext(ctx, "error creating rid token", "err", err)
		return nil, nil, email.NewEmailUserCreationError(fmt.Sprintf("error creating rid token: %v", emailUser.Email))
	}

	ctx = context.WithValue(ctx, shared.UserIDKey, profile.ResourceOwner.UserID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, profile.ResourceOwner.GroupID)

	emailUser.ResourceOwner = shared.GetResourceOwner(ctx)

	if emailUser.ID == uuid.Nil {
		emailUser.ID = profile.ResourceOwner.UserID
	}

	// Create email user record
	slog.InfoContext(ctx, fmt.Sprintf("attempt to create email user: %v", emailUser.Email))
	emailUser, err = usecase.EmailUserWriter.Create(ctx, emailUser)

	if err != nil {
		slog.ErrorContext(ctx, "error creating email user: error", "err", err)
		return nil, nil, email.NewEmailUserCreationError(fmt.Sprintf("error creating email user: %v", emailUser.ID))
	}

	if emailUser == nil {
		slog.ErrorContext(ctx, "error creating email user: user is nil", "err", err)
		return nil, nil, email.NewEmailUserCreationError(fmt.Sprintf("unable to create email user: %v", emailUser))
	}

	return emailUser, ridToken, nil
}

func NewOnboardEmailUserUseCase(
	emailUserWriter email_out.EmailUserWriter,
	emailUserReader email_out.EmailUserReader,
	vHashWriter email_out.VHashWriter,
	passwordHasher email_out.PasswordHasher,
	onboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler,
) email_in.OnboardEmailUserCommand {
	return &OnboardEmailUserUseCase{
		EmailUserWriter:   emailUserWriter,
		EmailUserReader:   emailUserReader,
		VHashWriter:       vHashWriter,
		PasswordHasher:    passwordHasher,
		OnboardOpenIDUser: onboardOpenIDUser,
	}
}

func (uc *OnboardEmailUserUseCase) newSearchByEmail(ctx context.Context, emailString string) shared.Search {
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
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
