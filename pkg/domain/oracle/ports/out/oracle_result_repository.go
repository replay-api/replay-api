package oracle_out

import (
	"context"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
)

// OracleResultRepository defines persistence operations for oracle results
type OracleResultRepository interface {
	Save(ctx context.Context, result *oracle_entities.OracleResult) error
	FindByID(ctx context.Context, id uuid.UUID) (*oracle_entities.OracleResult, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) (*oracle_entities.OracleResult, error)
	FindByExternalMatchID(ctx context.Context, externalMatchID string) (*oracle_entities.OracleResult, error)
	FindByStatus(ctx context.Context, status oracle_vo.OracleStatus, limit int, offset int) ([]*oracle_entities.OracleResult, int64, error)
	FindPendingPublication(ctx context.Context) ([]*oracle_entities.OracleResult, error)
	FindPublishedBefore(ctx context.Context, before time.Time) ([]*oracle_entities.OracleResult, error)
	Update(ctx context.Context, result *oracle_entities.OracleResult) error
	Count(ctx context.Context, filter OracleResultFilter) (int64, error)
	Search(ctx context.Context, filter OracleResultFilter, limit int, offset int) ([]*oracle_entities.OracleResult, int64, error)
}

// OracleResultFilter defines the filter criteria for searching oracle results
type OracleResultFilter struct {
	GameID          *string    `json:"game_id,omitempty"`
	Status          *string    `json:"status,omitempty"`
	MatchID         *uuid.UUID `json:"match_id,omitempty"`
	ExternalMatchID *string    `json:"external_match_id,omitempty"`
	MinConfidence   *int       `json:"min_confidence,omitempty"`
}
