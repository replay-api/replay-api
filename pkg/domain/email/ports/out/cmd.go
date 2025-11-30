package email_out

import (
	"context"

	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
)

type EmailUserWriter interface {
	Create(ctx context.Context, user *email_entities.EmailUser) (*email_entities.EmailUser, error)
}

type VHashWriter interface {
	CreateVHash(ctx context.Context, email string) string
}

type PasswordHasher interface {
	HashPassword(ctx context.Context, password string) (string, error)
	ComparePassword(ctx context.Context, hashedPassword string, password string) error
}
