package email_out

import (
	"context"

	shared "github.com/resource-ownership/go-common/pkg/common"
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
)

type EmailUserReader interface {
	Search(ctx context.Context, s shared.Search) ([]email_entities.EmailUser, error)
}
