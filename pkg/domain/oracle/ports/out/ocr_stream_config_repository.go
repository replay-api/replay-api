package oracle_out

import (
	"context"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// OCRStreamConfigRepository persists OCR stream configurations for VOD and live stream monitoring.
type OCRStreamConfigRepository interface {
	Save(ctx context.Context, config *oracle_entities.OCRStreamConfig) error
	FindByID(ctx context.Context, id uuid.UUID) (*oracle_entities.OCRStreamConfig, error)
	FindByExternalMatchID(ctx context.Context, externalMatchID string) (*oracle_entities.OCRStreamConfig, error)
	FindByVideoID(ctx context.Context, videoID string) (*oracle_entities.OCRStreamConfig, error)
	FindByStatus(ctx context.Context, status oracle_entities.OCRStreamStatus, limit int) ([]*oracle_entities.OCRStreamConfig, error)
	FindPending(ctx context.Context, limit int) ([]*oracle_entities.OCRStreamConfig, error)
	FindByGameID(ctx context.Context, gameID replay_common.GameIDKey, limit int) ([]*oracle_entities.OCRStreamConfig, error)
	Update(ctx context.Context, config *oracle_entities.OCRStreamConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
}
