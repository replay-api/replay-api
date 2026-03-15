package oracle_usecases

import (
	"context"
	"fmt"

	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
)

// oracleQueryHandler implements oracle_in.OracleQueryHandler
type oracleQueryHandler struct {
	repository oracle_out.OracleResultRepository
}

// NewOracleQueryHandler creates a new query handler
func NewOracleQueryHandler(repository oracle_out.OracleResultRepository) oracle_in.OracleQueryHandler {
	return &oracleQueryHandler{repository: repository}
}

func (h *oracleQueryHandler) GetOracleResult(ctx context.Context, query oracle_in.GetOracleResultQuery) (*oracle_in.OracleResultDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	result, err := h.repository.FindByID(ctx, query.OracleResultID)
	if err != nil {
		return nil, fmt.Errorf("oracle result not found: %w", err)
	}

	return oracle_in.MapOracleResultToDTO(result), nil
}

func (h *oracleQueryHandler) GetOracleResultByMatchID(ctx context.Context, query oracle_in.GetOracleResultByMatchIDQuery) (*oracle_in.OracleResultDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	result, err := h.repository.FindByMatchID(ctx, query.MatchID)
	if err != nil {
		return nil, fmt.Errorf("oracle result not found: %w", err)
	}

	return oracle_in.MapOracleResultToDTO(result), nil
}

func (h *oracleQueryHandler) ListOracleResults(ctx context.Context, query oracle_in.ListOracleResultsQuery) (*oracle_in.OracleResultListDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	filter := oracle_out.OracleResultFilter{}
	if query.GameID != nil {
		s := string(*query.GameID)
		filter.GameID = &s
	}
	if query.Status != nil {
		s := string(*query.Status)
		filter.Status = &s
	}

	offset := query.Page * query.PageSize
	results, totalCount, err := h.repository.Search(ctx, filter, query.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search oracle results: %w", err)
	}

	dtos := make([]oracle_in.OracleResultDTO, len(results))
	for i, r := range results {
		dtos[i] = *oracle_in.MapOracleResultToDTO(r)
	}

	return &oracle_in.OracleResultListDTO{
		Results:    dtos,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
	}, nil
}

func (h *oracleQueryHandler) GetSubmissionsForResult(ctx context.Context, query oracle_in.GetSubmissionsQuery) ([]oracle_in.ScoreSubmissionDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	result, err := h.repository.FindByID(ctx, query.OracleResultID)
	if err != nil {
		return nil, fmt.Errorf("oracle result not found: %w", err)
	}

	dtos := make([]oracle_in.ScoreSubmissionDTO, len(result.Submissions))
	for i := range result.Submissions {
		dtos[i] = *oracle_in.MapSubmissionToDTO(&result.Submissions[i])
	}

	return dtos, nil
}

func (h *oracleQueryHandler) GetPublicationStatus(ctx context.Context, query oracle_in.GetPublicationStatusQuery) ([]oracle_in.ChainPublicationDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	result, err := h.repository.FindByID(ctx, query.OracleResultID)
	if err != nil {
		return nil, fmt.Errorf("oracle result not found: %w", err)
	}

	dtos := make([]oracle_in.ChainPublicationDTO, len(result.Publications))
	for i, pub := range result.Publications {
		dtos[i] = oracle_in.ChainPublicationDTO{
			ChainID:         pub.ChainID,
			CAIP2:           pub.CAIP2,
			ContractAddress: pub.ContractAddress,
			TxHash:          pub.TxHash,
			BlockNumber:     pub.BlockNumber,
			GasUsed:         pub.GasUsed,
			Status:          pub.Status,
			PublishedAt:     pub.PublishedAt,
			ConfirmedAt:     pub.ConfirmedAt,
		}
	}

	return dtos, nil
}

// --- Unused import guard ---
var _ oracle_vo.OracleStatus // force import
