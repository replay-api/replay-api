package iam_use_cases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

type VerifyRIDUseCase struct {
	RIDWriter iam_out.RIDTokenWriter
	RIDReader iam_out.RIDTokenReader
}

func NewVerifyRIDUseCase(rIDWriter iam_out.RIDTokenWriter, rIDReader iam_out.RIDTokenReader) iam_in.VerifyRIDKeyCommand {
	return &VerifyRIDUseCase{
		RIDWriter: rIDWriter,
		RIDReader: rIDReader,
	}
}

func (usecase *VerifyRIDUseCase) Exec(ctx context.Context, key uuid.UUID) (shared.ResourceOwner, shared.IntendedAudienceKey, error) {
	s := usecase.newSearchByValidKey(ctx, key)

	tokenResult, err := usecase.RIDReader.Search(ctx, s)

	if err != nil {
		slog.ErrorContext(ctx, "error getting rid token by key", "err",
			err)

		return shared.ResourceOwner{}, shared.UserAudienceIDKey, err
	}

	if len(tokenResult) == 0 || tokenResult[0].ID == uuid.Nil {
		err = fmt.Errorf("invalid rid key")
		slog.ErrorContext(ctx, err.Error(), "key", key)

		return shared.ResourceOwner{}, shared.UserAudienceIDKey, err
	}

	if len(tokenResult) > 1 {
		slog.ErrorContext(ctx, "multiple rid tokens for the same key", "result", tokenResult, "key", key)

		// TODO: recovery: notificar conta, temp lock?, desabilitar refresh

		return shared.ResourceOwner{}, shared.UserAudienceIDKey, fmt.Errorf("invalid RID token")
	}

	return tokenResult[0].ResourceOwner, tokenResult[0].IntendedAudience, nil
}

func (uc *VerifyRIDUseCase) newSearchByValidKey(ctx context.Context, key uuid.UUID) shared.Search {
	notBefore := time.Now()
	params := []shared.SearchAggregation{
		{
			Params: []shared.SearchParameter{
				{
					ValueParams: []shared.SearchableValue{
						{
							Field: "ID",
							Values: []interface{}{
								key,
							},
						},
					},
					DateParams: []shared.SearchableDateRange{
						{
							Field: "ExpiresAt",
							Min:   &notBefore,
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
