package metadata

import (
	"context"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type ShareTokenQueryService struct {
	common.BaseQueryService[replay_entity.ShareToken]
	tokenReader replay_out.ShareTokenReader
}

func NewShareTokenQueryService(shareTokenReader replay_out.ShareTokenReader) replay_in.ShareTokenReader {
	queryableFields := map[string]bool{
		"ID":            true,
		"ResourceID":    true,
		"ResourceType":  true,
		"Status":        true,
		"ExpiresAt":     true,
		"EntityType":    true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	readableFields := map[string]bool{
		"ID":            true,
		"ResourceID":    true,
		"ResourceType":  true,
		"Status":        true,
		"ExpiresAt":     true,
		"Uri":           true,
		"EntityType":    true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	return &ShareTokenQueryService{
		BaseQueryService: common.BaseQueryService[replay_entity.ShareToken]{
			Reader:          shareTokenReader.(common.Searchable[replay_entity.ShareToken]),
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        common.UserAudienceIDKey,
		},
		tokenReader: shareTokenReader,
	}
}

func (s *ShareTokenQueryService) FindByToken(ctx context.Context, tokenID uuid.UUID) (*replay_entity.ShareToken, error) {
	return s.tokenReader.FindByToken(ctx, tokenID)
}

type ShareTokenCommandService struct {
	repository replay_out.ShareTokenWriter
}

func NewShareTokenCommandService(repository replay_out.ShareTokenWriter) replay_in.ShareTokenCommand {
	return &ShareTokenCommandService{
		repository: repository,
	}
}

func (s *ShareTokenCommandService) Create(ctx context.Context, token *replay_entity.ShareToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = now
	}

	if token.Status == "" {
		token.Status = replay_entity.ShareTokenStatusActive
	}

	// Default expiration: 30 days from now
	if token.ExpiresAt.IsZero() {
		token.ExpiresAt = now.Add(30 * 24 * time.Hour)
	}

	return s.repository.Create(ctx, token)
}

func (s *ShareTokenCommandService) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	return s.repository.Delete(ctx, tokenID)
}

func (s *ShareTokenCommandService) Update(ctx context.Context, token *replay_entity.ShareToken) error {
	token.UpdatedAt = time.Now()
	return s.repository.Update(ctx, token)
}
