package email_in

import (
	"context"

	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
)

type OnboardEmailUserCommand interface {
	Exec(ctx context.Context, emailUser *email_entities.EmailUser, password string) (*email_entities.EmailUser, *iam_entities.RIDToken, error)
	Validate(ctx context.Context, emailUser *email_entities.EmailUser, password string) error
}

type LoginEmailUserCommand interface {
	Exec(ctx context.Context, email string, password string, vHash string) (*email_entities.EmailUser, *iam_entities.RIDToken, error)
	Validate(ctx context.Context, email string, password string, vHash string) error
}
