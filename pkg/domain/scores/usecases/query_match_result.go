package scores_usecases

import (
	"context"
	"fmt"

	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
)

type matchResultQueryHandler struct {
	repository scores_out.MatchResultRepository
}

// NewMatchResultQueryHandler creates a new query handler for match results
func NewMatchResultQueryHandler(
	repository scores_out.MatchResultRepository,
) scores_in.MatchResultQueryHandler {
	return &matchResultQueryHandler{
		repository: repository,
	}
}

func (h *matchResultQueryHandler) GetMatchResult(ctx context.Context, query scores_in.GetMatchResultQuery) (*scores_in.MatchResultDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	result, err := h.repository.FindByID(ctx, query.MatchResultID)
	if err != nil {
		return nil, fmt.Errorf("match result not found: %w", err)
	}

	dto := scores_in.MatchResultToDTO(result)
	return &dto, nil
}

func (h *matchResultQueryHandler) GetMatchResultByMatchID(ctx context.Context, query scores_in.GetMatchResultByMatchIDQuery) (*scores_in.MatchResultDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	result, err := h.repository.FindByMatchID(ctx, query.MatchID)
	if err != nil {
		return nil, fmt.Errorf("match result not found: %w", err)
	}

	dto := scores_in.MatchResultToDTO(result)
	return &dto, nil
}

func (h *matchResultQueryHandler) ListMatchResults(ctx context.Context, query scores_in.ListMatchResultsQuery) (*scores_in.MatchResultListDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	filter := scores_out.MatchResultFilter{}
	if query.GameID != nil {
		gameID := string(*query.GameID)
		filter.GameID = &gameID
	}
	if query.TournamentID != nil {
		filter.TournamentID = query.TournamentID
	}
	if query.MatchmakingSessionID != nil {
		filter.MatchmakingSessionID = query.MatchmakingSessionID
	}
	if query.Status != nil {
		status := string(*query.Status)
		filter.Status = &status
	}
	if query.PlayerID != nil {
		filter.PlayerID = query.PlayerID
	}
	if query.TeamID != nil {
		filter.TeamID = query.TeamID
	}

	offset := query.Page * query.PageSize
	results, total, err := h.repository.Search(ctx, filter, query.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search match results: %w", err)
	}

	dtos := make([]scores_in.MatchResultDTO, len(results))
	for i, r := range results {
		dtos[i] = scores_in.MatchResultToDTO(r)
	}

	return &scores_in.MatchResultListDTO{
		Results:    dtos,
		TotalCount: total,
		Page:       query.Page,
		PageSize:   query.PageSize,
	}, nil
}

func (h *matchResultQueryHandler) GetMatchResultsByTournament(ctx context.Context, query scores_in.GetTournamentResultsQuery) (*scores_in.MatchResultListDTO, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	results, err := h.repository.FindByTournamentID(ctx, query.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find tournament results: %w", err)
	}

	dtos := make([]scores_in.MatchResultDTO, len(results))
	for i, r := range results {
		dtos[i] = scores_in.MatchResultToDTO(r)
	}

	return &scores_in.MatchResultListDTO{
		Results:    dtos,
		TotalCount: int64(len(results)),
		Page:       0,
		PageSize:   len(results),
	}, nil
}
