package use_cases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
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

func (usecase *VerifyRIDUseCase) Exec(ctx context.Context, key uuid.UUID) (common.ResourceOwner, error) {
	s := usecase.newSearchByValidKey(ctx, key)

	tokenResult, err := usecase.RIDReader.Search(ctx, s)

	if err != nil {
		slog.ErrorContext(ctx, "error getting rid token by key", "err",
			err)

		return common.ResourceOwner{}, err
	}

	if len(tokenResult) == 0 || tokenResult[0].ID == uuid.Nil {
		err = fmt.Errorf("invalid rid key")
		slog.ErrorContext(ctx, err.Error(), "key", key)

		return common.ResourceOwner{}, err
	}

	if len(tokenResult) > 1 {
		slog.ErrorContext(ctx, "multiple rid tokens for the same key", "result", tokenResult, "key", key)

		// TODO: recovery: notificar conta, temp lock?, desabilitar refresh

		return common.ResourceOwner{}, fmt.Errorf("invalid RID token")
	}

	return tokenResult[0].ResourceOwner, nil
}

func (uc *VerifyRIDUseCase) newSearchByValidKey(ctx context.Context, key uuid.UUID) common.Search {
	notBefore := time.Now()
	params := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{
							Field: "Key",
							Values: []interface{}{
								key,
							},
						},
					},
					DateParams: []common.SearchableDateRange{
						{
							Field: "ExpiresAt",
							Min:   &notBefore,
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
