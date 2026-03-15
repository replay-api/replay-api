package oracle_in

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// --- Query Handlers ---

// OracleQueryHandler handles all read operations for oracle results
type OracleQueryHandler interface {
	GetOracleResult(ctx context.Context, query GetOracleResultQuery) (*OracleResultDTO, error)
	GetOracleResultByMatchID(ctx context.Context, query GetOracleResultByMatchIDQuery) (*OracleResultDTO, error)
	ListOracleResults(ctx context.Context, query ListOracleResultsQuery) (*OracleResultListDTO, error)
	GetSubmissionsForResult(ctx context.Context, query GetSubmissionsQuery) ([]ScoreSubmissionDTO, error)
	GetPublicationStatus(ctx context.Context, query GetPublicationStatusQuery) ([]ChainPublicationDTO, error)
}

// --- Query DTOs ---

type GetOracleResultQuery struct {
	OracleResultID uuid.UUID `json:"oracle_result_id"`
}

func (q GetOracleResultQuery) Validate() error {
	if q.OracleResultID == uuid.Nil {
		return fmt.Errorf("oracle_result_id is required")
	}
	return nil
}

type GetOracleResultByMatchIDQuery struct {
	MatchID uuid.UUID `json:"match_id"`
}

func (q GetOracleResultByMatchIDQuery) Validate() error {
	if q.MatchID == uuid.Nil {
		return fmt.Errorf("match_id is required")
	}
	return nil
}

type ListOracleResultsQuery struct {
	GameID   *replay_common.GameIDKey   `json:"game_id,omitempty"`
	Status   *oracle_vo.OracleStatus    `json:"status,omitempty"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
}

func (q ListOracleResultsQuery) Validate() error {
	if q.Page < 0 {
		return fmt.Errorf("page must be >= 0")
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		return fmt.Errorf("page_size must be between 1 and 100")
	}
	return nil
}

type GetSubmissionsQuery struct {
	OracleResultID uuid.UUID `json:"oracle_result_id"`
}

func (q GetSubmissionsQuery) Validate() error {
	if q.OracleResultID == uuid.Nil {
		return fmt.Errorf("oracle_result_id is required")
	}
	return nil
}

type GetPublicationStatusQuery struct {
	OracleResultID uuid.UUID `json:"oracle_result_id"`
}

func (q GetPublicationStatusQuery) Validate() error {
	if q.OracleResultID == uuid.Nil {
		return fmt.Errorf("oracle_result_id is required")
	}
	return nil
}

// --- List Response DTO ---

type OracleResultListDTO struct {
	Results    []OracleResultDTO `json:"results"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}
