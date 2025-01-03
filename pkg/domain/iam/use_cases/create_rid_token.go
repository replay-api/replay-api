package iam_use_cases

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
)

type CreateRIDTokenUseCase struct {
	RIDWriter iam_out.RIDTokenWriter
	RIDReader iam_out.RIDTokenReader
}

func NewCreateRIDTokenUseCase(rIDWriter iam_out.RIDTokenWriter, rIDReader iam_out.RIDTokenReader) iam_in.CreateRIDTokenCommand {
	return &CreateRIDTokenUseCase{
		RIDWriter: rIDWriter,
		RIDReader: rIDReader,
	}
}

func (usecase *CreateRIDTokenUseCase) Exec(ctx context.Context, reso common.ResourceOwner, source iam_entity.RIDSourceKey, aud common.IntendedAudienceKey) (*iam_entity.RIDToken, error) {
	duration, _ := time.ParseDuration("1h")
	expiresAt := time.Now().Add(duration)

	// TODO: verificar existencia, consistir usuario

	token, err := usecase.RIDWriter.Create(ctx, &iam_entity.RIDToken{
		ID:               uuid.New(),
		Key:              uuid.New(),
		Source:           source,
		ResourceOwner:    reso,
		IntendedAudience: aud,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
	})

	if err != nil {
		slog.ErrorContext(ctx, "unable to create rid token", "err", err)
		return nil, err
	}

	return token, nil
}
