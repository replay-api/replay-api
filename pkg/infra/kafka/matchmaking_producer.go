package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/replay-api/replay-api/pkg/infra/events/schemas"
)

// ErrResourceOwnershipInvalid is returned when tenant_id, client_id, or resource_owner_id are missing/invalid.
var ErrResourceOwnershipInvalid = errors.New("resource ownership validation failed: tenant_id, client_id, and resource_owner_id are required")

// ErrProducerDisabled is returned when the producer is disabled via config.
var ErrProducerDisabled = errors.New("matchmaking producer is disabled")

// MatchmakingProducer publishes PlayerQueued and MatchCompleted events to Kafka
// using the canonical schemas (#16) with partition key strategy and resource ownership validation.
type MatchmakingProducer struct {
	client *Client
	cfg    *MatchmakingProducerConfig
}

// NewMatchmakingProducer creates a new MatchmakingProducer.
func NewMatchmakingProducer(client *Client, cfg *MatchmakingProducerConfig) *MatchmakingProducer {
	if cfg == nil {
		cfg = DefaultMatchmakingProducerConfig()
	}
	return &MatchmakingProducer{client: client, cfg: cfg}
}

// PlayerQueuedInput holds the input for publishing a PlayerQueued event.
type PlayerQueuedInput struct {
	PlayerID            string
	GameID              string
	Region              string
	TenantID            string
	ClientID            string
	ResourceOwnerID     string
	MinMMR              *int32
	MaxMMR              *int32
	PriorityBoost       *int32
	ResourcePermissions []string
	CorrelationID       string
}

// ValidateResourceOwnership validates that tenant_id, client_id, and resource_owner_id are present.
func (p *PlayerQueuedInput) ValidateResourceOwnership() error {
	if strings.TrimSpace(p.TenantID) == "" {
		return fmt.Errorf("%w: tenant_id is empty", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(p.ClientID) == "" {
		return fmt.Errorf("%w: client_id is empty", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(p.ResourceOwnerID) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty", ErrResourceOwnershipInvalid)
	}
	return nil
}

// MatchCompletedInput holds the input for publishing a MatchCompleted event.
type MatchCompletedInput struct {
	MatchID            string
	PlayerIDs          []string
	WinnerTeamID       string
	IsDraw             bool
	CompletedAtEpochMs int64
	TenantID           string
	ClientID           string
	ResourceOwnerID    string
	CorrelationID      string
}

// ValidateResourceOwnership validates that tenant_id, client_id, and resource_owner_id are present.
func (p *MatchCompletedInput) ValidateResourceOwnership() error {
	if strings.TrimSpace(p.TenantID) == "" {
		return fmt.Errorf("%w: tenant_id is empty", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(p.ClientID) == "" {
		return fmt.Errorf("%w: client_id is empty", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(p.ResourceOwnerID) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty", ErrResourceOwnershipInvalid)
	}
	return nil
}

// PublishPlayerQueued publishes a PlayerQueued event to the commands topic.
// Partition key: game_id + "-" + region (per Epic §10) for ordering by game/region.
// Returns error if validation fails, producer is disabled, or retries are exhausted.
func (mp *MatchmakingProducer) PublishPlayerQueued(ctx context.Context, input *PlayerQueuedInput) error {
	if !mp.cfg.Enabled {
		slog.DebugContext(ctx, "Matchmaking producer disabled, skipping PlayerQueued")
		return ErrProducerDisabled
	}

	if err := input.ValidateResourceOwnership(); err != nil {
		slog.ErrorContext(ctx, "PlayerQueued resource ownership validation failed",
			"player_id", input.PlayerID,
			"game_id", input.GameID,
			"error", err)
		return err
	}

	eventID := uuid.New().String()
	now := timestamppb.Now()

	payload := &schemas.PlayerQueuedPayload{
		PlayerId:            input.PlayerID,
		GameId:              input.GameID,
		Region:              input.Region,
		TenantId:            input.TenantID,
		ClientId:            input.ClientID,
		ResourcePermissions: input.ResourcePermissions,
	}
	if input.MinMMR != nil && input.MaxMMR != nil {
		payload.SkillRange = &schemas.SkillRange{
			MinMmr: *input.MinMMR,
			MaxMmr: *input.MaxMMR,
		}
	}
	if input.PriorityBoost != nil {
		payload.PriorityBoost = input.PriorityBoost
	}

	envelope := &schemas.EventEnvelope{
		Id:                eventID,
		Type:              schemas.EventTypePlayerQueued,
		Source:            mp.cfg.Source,
		Specversion:       schemas.CloudEventsSpecVersion,
		Time:              now,
		Subject:           input.PlayerID,
		ResourceOwnerId:   input.ResourceOwnerID,
		CorrelationId:     input.CorrelationID,
		DataschemaVersion: schemas.SchemaVersionV1,
	}

	event := &schemas.MatchmakingEvent{
		Envelope: envelope,
		Data:     &schemas.MatchmakingEvent_PlayerQueued{PlayerQueued: payload},
	}

	// Partition key: game_id + region for ordering and consumer parallelism (Epic §10)
	partitionKey := input.GameID + "-" + input.Region

	return mp.publishWithRetry(ctx, mp.cfg.TopicCommands, partitionKey, event)
}

// PublishMatchCompleted publishes a MatchCompleted event to the matches topic.
// Partition key: match_id (per Epic §10) for ordering by match.
func (mp *MatchmakingProducer) PublishMatchCompleted(ctx context.Context, input *MatchCompletedInput) error {
	if !mp.cfg.Enabled {
		slog.DebugContext(ctx, "Matchmaking producer disabled, skipping MatchCompleted")
		return ErrProducerDisabled
	}

	if err := input.ValidateResourceOwnership(); err != nil {
		slog.ErrorContext(ctx, "MatchCompleted resource ownership validation failed",
			"match_id", input.MatchID,
			"error", err)
		return err
	}

	eventID := uuid.New().String()
	now := timestamppb.Now()

	if input.CompletedAtEpochMs == 0 {
		input.CompletedAtEpochMs = time.Now().UnixMilli()
	}

	payload := &schemas.MatchCompletedPayload{
		MatchId:            input.MatchID,
		PlayerIds:          input.PlayerIDs,
		WinnerTeamId:       input.WinnerTeamID,
		IsDraw:             input.IsDraw,
		CompletedAtEpochMs: input.CompletedAtEpochMs,
		TenantId:           input.TenantID,
		ClientId:           input.ClientID,
	}

	envelope := &schemas.EventEnvelope{
		Id:                eventID,
		Type:              schemas.EventTypeMatchCompleted,
		Source:            mp.cfg.Source,
		Specversion:       schemas.CloudEventsSpecVersion,
		Time:              now,
		Subject:           input.MatchID,
		ResourceOwnerId:   input.ResourceOwnerID,
		CorrelationId:     input.CorrelationID,
		DataschemaVersion: schemas.SchemaVersionV1,
	}

	event := &schemas.MatchmakingEvent{
		Envelope: envelope,
		Data:     &schemas.MatchmakingEvent_MatchCompleted{MatchCompleted: payload},
	}

	// Partition key: match_id for ordering and consumer parallelism (Epic §10)
	partitionKey := input.MatchID

	return mp.publishWithRetry(ctx, mp.cfg.TopicMatches, partitionKey, event)
}

// publishWithRetry serialises the event to JSON (protojson) and publishes with retry/backoff.
func (mp *MatchmakingProducer) publishWithRetry(ctx context.Context, topic, key string, event *schemas.MatchmakingEvent) error {
	value, err := protojson.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal matchmaking event: %w", err)
	}

	headers := map[string]string{
		"ce_type": event.Envelope.Type,
		"ce_source": event.Envelope.Source,
	}

	var lastErr error
	backoff := mp.cfg.InitialBackoff

	for attempt := 0; attempt <= mp.cfg.MaxRetries; attempt++ {
		produceCtx, cancel := context.WithTimeout(ctx, mp.cfg.ProduceTimeout)
		err := mp.client.PublishRaw(produceCtx, topic, key, value, headers)
		cancel()

		if err == nil {
			slog.DebugContext(ctx, "Published matchmaking event",
				"topic", topic,
				"key", key,
				"type", event.Envelope.Type)
			return nil
		}

		lastErr = err
		slog.WarnContext(ctx, "Matchmaking produce failed, retrying",
			"topic", topic,
			"key", key,
			"attempt", attempt+1,
			"max_retries", mp.cfg.MaxRetries,
			"error", err)

		if attempt < mp.cfg.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff = min(backoff*2, mp.cfg.MaxBackoff)
			}
		}
	}

	slog.ErrorContext(ctx, "Matchmaking produce failed after retries",
		"topic", topic,
		"key", key,
		"error", lastErr)
	return fmt.Errorf("produce failed after %d retries: %w", mp.cfg.MaxRetries, lastErr)
}
