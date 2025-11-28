package email_out

import (
	"context"

	common "github.com/replay-api/replay-api/pkg/domain"
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
)

type EmailUserReader interface {
	Search(ctx context.Context, s common.Search) ([]email_entities.EmailUser, error)
}
