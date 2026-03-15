package kafka

import (
	"context"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
)

// OracleEventPublisherAdapter adapts the Kafka EventPublisher to the oracle domain port
type OracleEventPublisherAdapter struct {
	publisher *EventPublisher
}

// Compile-time interface satisfaction check
var _ oracle_out.OracleEventPublisher = (*OracleEventPublisherAdapter)(nil)

// NewOracleEventPublisherAdapter creates a new adapter
func NewOracleEventPublisherAdapter(publisher *EventPublisher) *OracleEventPublisherAdapter {
	return &OracleEventPublisherAdapter{publisher: publisher}
}

func (a *OracleEventPublisherAdapter) toOracleEvent(result *oracle_entities.OracleResult) *OracleEvent {
	event := &OracleEvent{
		OracleResultID: result.ID,
		MatchID:        result.MatchID,
		GameID:         string(result.GameID),
		Status:         string(result.Status),
	}

	if result.ExternalMatchID != nil {
		event.ExternalMatchID = *result.ExternalMatchID
	}

	if result.ConsensusResult != nil {
		event.ConfidenceLevel = result.ConsensusResult.ConfidenceLevel
		event.OverallAgreement = result.ConsensusResult.AgreementRatio
	}

	if len(result.Publications) > 0 {
		pubs := make([]OracleChainPubInfo, len(result.Publications))
		for i, pub := range result.Publications {
			pubs[i] = OracleChainPubInfo{
				ChainID: int(pub.ChainID),
				TxHash:  pub.TxHash,
				Status:  pub.Status,
			}
		}
		event.ChainPublications = pubs
	}

	return event
}

func (a *OracleEventPublisherAdapter) PublishConsensusReached(ctx context.Context, result *oracle_entities.OracleResult) error {
	if a.publisher.client == nil {
		return nil
	}

	event := a.toOracleEvent(result)
	event.EventID = uuid.New()
	event.EventType = EventTypeOracleConsensusReached
	event.Timestamp = time.Now().UnixMilli()

	msg := &Message{
		Key:       result.ID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":       event.EventType,
			"oracle_result_id": result.ID.String(),
			"game_id":          string(result.GameID),
		},
	}

	return a.publisher.client.Publish(ctx, TopicOracleConsensusReached, msg)
}

func (a *OracleEventPublisherAdapter) PublishScorePublished(ctx context.Context, result *oracle_entities.OracleResult) error {
	if a.publisher.client == nil {
		return nil
	}

	event := a.toOracleEvent(result)
	event.EventID = uuid.New()
	event.EventType = EventTypeOraclePublished
	event.Timestamp = time.Now().UnixMilli()

	msg := &Message{
		Key:       result.ID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":       event.EventType,
			"oracle_result_id": result.ID.String(),
			"game_id":          string(result.GameID),
		},
	}

	return a.publisher.client.Publish(ctx, TopicOraclePublished, msg)
}

func (a *OracleEventPublisherAdapter) PublishScoreFinalized(ctx context.Context, result *oracle_entities.OracleResult) error {
	if a.publisher.client == nil {
		return nil
	}

	event := a.toOracleEvent(result)
	event.EventID = uuid.New()
	event.EventType = EventTypeOracleFinalized
	event.Timestamp = time.Now().UnixMilli()

	msg := &Message{
		Key:       result.ID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":       event.EventType,
			"oracle_result_id": result.ID.String(),
		},
	}

	return a.publisher.client.Publish(ctx, TopicOracleFinalized, msg)
}

func (a *OracleEventPublisherAdapter) PublishScoreDisputed(ctx context.Context, result *oracle_entities.OracleResult) error {
	if a.publisher.client == nil {
		return nil
	}

	event := a.toOracleEvent(result)
	event.EventID = uuid.New()
	event.EventType = EventTypeOracleDisputed
	event.Timestamp = time.Now().UnixMilli()

	msg := &Message{
		Key:       result.ID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":       event.EventType,
			"oracle_result_id": result.ID.String(),
		},
	}

	return a.publisher.client.Publish(ctx, TopicOracleDisputed, msg)
}

func (a *OracleEventPublisherAdapter) PublishExternalIngested(ctx context.Context, result *oracle_entities.OracleResult, sub oracle_entities.ScoreSubmission) error {
	if a.publisher.client == nil {
		return nil
	}

	event := a.toOracleEvent(result)
	event.EventID = uuid.New()
	event.EventType = EventTypeOracleExternalIngested
	event.Timestamp = time.Now().UnixMilli()
	event.SourceType = string(sub.SourceType)
	event.SourceProvider = sub.ProviderMatchID

	msg := &Message{
		Key:       result.ID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":       event.EventType,
			"oracle_result_id": result.ID.String(),
			"source_type":      string(sub.SourceType),
			"game_id":          string(result.GameID),
		},
	}

	return a.publisher.client.Publish(ctx, TopicOracleExternalIngested, msg)
}
