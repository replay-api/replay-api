package kafka

import (
	"context"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

// ReplayEventPublisherAdapter implements the replay_out.ReplayEventPublisher interface
// using the Kafka EventPublisher for message broker communication.
type ReplayEventPublisherAdapter struct {
	publisher *EventPublisher
}

// NewReplayEventPublisherAdapter creates a new adapter for replay event publishing
func NewReplayEventPublisherAdapter(publisher *EventPublisher) replay_out.ReplayEventPublisher {
	return &ReplayEventPublisherAdapter{publisher: publisher}
}

// PublishReplayUploaded publishes an event when a replay file is uploaded
func (a *ReplayEventPublisherAdapter) PublishReplayUploaded(ctx context.Context, replayFile *replay_entity.ReplayFile) error {
	if a.publisher == nil {
		return nil
	}

	event := &ReplayUploadedEvent{
		ReplayFileID: replayFile.ID,
		GameID:       string(replayFile.GameID),
		UserID:       replayFile.ResourceOwner.UserID,
		TenantID:     replayFile.ResourceOwner.TenantID,
		FileSize:     replayFile.Size,
		InternalURI:  replayFile.InternalURI,
		Metadata: map[string]string{
			"network_id": string(replayFile.NetworkID),
			"status":     string(replayFile.Status),
		},
	}

	return a.publisher.PublishReplayUploaded(ctx, event)
}

// PublishReplayProcessing publishes progress events during replay processing
func (a *ReplayEventPublisherAdapter) PublishReplayProcessing(ctx context.Context, replayFileID uuid.UUID, stage string, progress int, eventCount int, playerCount int) error {
	if a.publisher == nil {
		return nil
	}

	eventType := EventTypeReplayProcessing
	if progress >= 100 {
		eventType = EventTypeReplayProgress
	}

	event := &ReplayProcessingEvent{
		ReplayFileID: replayFileID,
		EventType:    eventType,
		Progress:     progress,
		Stage:        stage,
		EventCount:   eventCount,
		PlayerCount:  playerCount,
	}

	return a.publisher.PublishReplayProcessing(ctx, event)
}

// PublishReplayCompleted publishes an event when replay processing completes successfully
func (a *ReplayEventPublisherAdapter) PublishReplayCompleted(ctx context.Context, replayFile *replay_entity.ReplayFile, matchID uuid.UUID, eventCount int, playerCount int, processingDurationMs int64) error {
	if a.publisher == nil {
		return nil
	}

	event := &ReplayCompletedEvent{
		ReplayFileID:  replayFile.ID,
		MatchID:       matchID,
		GameID:        string(replayFile.GameID),
		EventCount:    eventCount,
		PlayerCount:   playerCount,
		Duration:      processingDurationMs,
		MatchDuration: 0, // Will be filled by processing
		Metadata: map[string]string{
			"internal_uri": replayFile.InternalURI,
		},
	}

	return a.publisher.PublishReplayCompleted(ctx, event)
}

// PublishReplayFailed publishes an event when replay processing fails
func (a *ReplayEventPublisherAdapter) PublishReplayFailed(ctx context.Context, replayFile *replay_entity.ReplayFile, stage string, errorType string, errorMessage string, retryable bool, retryCount int) error {
	if a.publisher == nil {
		return nil
	}

	event := &ReplayFailedEvent{
		ReplayFileID: replayFile.ID,
		GameID:       string(replayFile.GameID),
		Stage:        stage,
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		Retryable:    retryable,
		RetryCount:   retryCount,
	}

	return a.publisher.PublishReplayFailed(ctx, event)
}
