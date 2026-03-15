package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Topic constants for matchmaking events
const (
	TopicQueueEvents       = "matchmaking.queue.events"
	TopicLobbyEvents       = "matchmaking.lobby.events"
	TopicPrizePoolEvents   = "matchmaking.prizepool.events"
	TopicMatchesCreated    = "matchmaking.matches.created"
	TopicMatchesResults    = "matchmaking.matches.results"
	TopicPlayerStatus      = "matchmaking.player-status"
	TopicWebSocketBroadcast = "websocket.broadcasts"
	TopicDLQ               = "matchmaking.dlq"
	TopicReadyCheck         = "matchmaking.ready-check"
	TopicNotificationDelivery = "notifications.delivery"
)

// Replay processing topics
const (
	TopicReplaysUploaded   = "replays.uploaded"
	TopicReplaysProcessing = "replays.processing"
	TopicReplaysCompleted  = "replays.completed"
	TopicReplaysFailed     = "replays.failed"
	TopicReplaysGameEvents = "replays.game-events"
)

// Billing topics
const (
	TopicBillingSubscriptionCreated   = "billing.subscription.created"
	TopicBillingSubscriptionUpgraded  = "billing.subscription.upgraded"
	TopicBillingSubscriptionCancelled = "billing.subscription.cancelled"
	TopicBillingSubscriptionExpired   = "billing.subscription.expired"
	TopicBillingPaymentProcessed      = "billing.payment.processed"
	TopicBillingPaymentFailed         = "billing.payment.failed"
	TopicBillingPaymentRefunded       = "billing.payment.refunded"
	TopicBillingDLQ                   = "billing.dlq"
)

// Wallet topics
const (
	TopicWalletCreated           = "wallet.created"
	TopicWalletDeposit           = "wallet.deposit"
	TopicWalletWithdrawal        = "wallet.withdrawal"
	TopicWalletWithdrawalPending = "wallet.withdrawal.pending"
	TopicWalletEntryFee          = "wallet.entry_fee"
	TopicWalletPrize             = "wallet.prize"
	TopicWalletRefund            = "wallet.refund"
	TopicWalletLocked            = "wallet.locked"
	TopicWalletUnlocked          = "wallet.unlocked"
	TopicWalletDLQ               = "wallet.dlq"
)

// Messaging topics
const (
	TopicCommentCreated        = "messaging.comment.created"
	TopicCommentEdited         = "messaging.comment.edited"
	TopicCommentDeleted        = "messaging.comment.deleted"
	TopicCommentReaction       = "messaging.comment.reaction"
	TopicDirectMessageSent     = "messaging.dm.sent"
	TopicTeamMessageSent       = "messaging.team.sent"
	TopicMentionNotification   = "messaging.mention.notification"
	TopicMessagingDLQ          = "messaging.dlq"
)

// Prediction topics
const (
	TopicPredictionPlaced   = "prediction.placed"
	TopicPredictionResolved = "prediction.resolved"
	TopicPredictionDLQ      = "prediction.dlq"
)

// Event types
const (
	EventTypeQueueJoined        = "QUEUE_JOINED"
	EventTypeQueueLeft          = "QUEUE_LEFT"
	EventTypeSearching          = "SEARCHING"
	EventTypeLobbyCreated       = "LOBBY_CREATED"
	EventTypeLobbyUpdated       = "LOBBY_UPDATED"
	EventTypePlayerJoined       = "PLAYER_JOINED"
	EventTypePlayerLeft         = "PLAYER_LEFT"
	EventTypeReadyStatusChanged = "READY_STATUS_CHANGED"
	EventTypeLobbyReady         = "LOBBY_READY"
	EventTypeLobbyCancelled     = "LOBBY_CANCELLED"
	EventTypePrizePoolUpdated   = "PRIZE_POOL_UPDATED"
	EventTypeMatchCreated       = "MATCH_CREATED"
	EventTypeMatchStarted       = "MATCH_STARTED"
	EventTypeMatchCompleted     = "MATCH_COMPLETED"
	EventTypeMatchCancelled     = "MATCH_CANCELLED"

	// Ready check event types
	EventTypeReadyCheckStarted       = "READY_CHECK_STARTED"
	EventTypeReadinessConfirmed      = "READINESS_CONFIRMED"
	EventTypeReadinessDeclined       = "READINESS_DECLINED"
	EventTypeReadyCheckTimeout       = "READY_CHECK_TIMEOUT"
	EventTypeAllPlayersReady         = "ALL_PLAYERS_READY"
	EventTypeGameConnectionDelivered = "GAME_CONNECTION_DELIVERED"
)

// Replay event types
const (
	EventTypeReplayUploaded   = "REPLAY_UPLOADED"
	EventTypeReplayProcessing = "REPLAY_PROCESSING"
	EventTypeReplayCompleted  = "REPLAY_COMPLETED"
	EventTypeReplayFailed     = "REPLAY_FAILED"
	EventTypeReplayProgress   = "REPLAY_PROGRESS"
)

// Billing event types
const (
	EventTypeSubscriptionCreated   = "SUBSCRIPTION_CREATED"
	EventTypeSubscriptionUpgraded  = "SUBSCRIPTION_UPGRADED"
	EventTypeSubscriptionCancelled = "SUBSCRIPTION_CANCELLED"
	EventTypeSubscriptionExpired   = "SUBSCRIPTION_EXPIRED"
	EventTypePaymentProcessed      = "PAYMENT_PROCESSED"
	EventTypePaymentFailed         = "PAYMENT_FAILED"
	EventTypePaymentRefunded       = "PAYMENT_REFUNDED"
)

// Wallet event types
const (
	EventTypeWalletCreated           = "WALLET_CREATED"
	EventTypeWalletDeposit           = "WALLET_DEPOSIT"
	EventTypeWalletWithdrawal        = "WALLET_WITHDRAWAL"
	EventTypeWalletWithdrawalPending = "WALLET_WITHDRAWAL_PENDING"
	EventTypeWalletEntryFee          = "WALLET_ENTRY_FEE"
	EventTypeWalletPrize             = "WALLET_PRIZE"
	EventTypeWalletRefund            = "WALLET_REFUND"
	EventTypeWalletLocked            = "WALLET_LOCKED"
	EventTypeWalletUnlocked          = "WALLET_UNLOCKED"
)

// Messaging event types
const (
	EventTypeCommentCreated      = "COMMENT_CREATED"
	EventTypeCommentEdited       = "COMMENT_EDITED"
	EventTypeCommentDeleted      = "COMMENT_DELETED"
	EventTypeCommentReaction     = "COMMENT_REACTION"
	EventTypeDirectMessageSent   = "DIRECT_MESSAGE_SENT"
	EventTypeTeamMessageSent     = "TEAM_MESSAGE_SENT"
	EventTypeMentionNotification = "MENTION_NOTIFICATION"
)

// Prediction event types
const (
	EventTypePredictionPlaced   = "PREDICTION_PLACED"
	EventTypePredictionResolved = "PREDICTION_RESOLVED"
)

// Vault topics
const (
	TopicVaultCreated           = "vault.created"
	TopicVaultDeposit           = "vault.deposit"
	TopicVaultWithdrawal        = "vault.withdrawal"
	TopicVaultProposalCreated   = "vault.proposal.created"
	TopicVaultProposalApproved  = "vault.proposal.approved"
	TopicVaultProposalRejected  = "vault.proposal.rejected"
	TopicVaultProposalExecuted  = "vault.proposal.executed"
	TopicVaultProposalCancelled = "vault.proposal.cancelled"
	TopicVaultSettingsUpdated   = "vault.settings.updated"
	TopicVaultItemDeposited     = "vault.item.deposited"
	TopicVaultItemTransferred   = "vault.item.transferred"
	TopicVaultDLQ               = "vault.dlq"
)

// Vault event types
const (
	EventTypeVaultCreated           = "VAULT_CREATED"
	EventTypeVaultDeposit           = "VAULT_DEPOSIT"
	EventTypeVaultWithdrawal        = "VAULT_WITHDRAWAL"
	EventTypeVaultProposalCreated   = "VAULT_PROPOSAL_CREATED"
	EventTypeVaultProposalApproved  = "VAULT_PROPOSAL_APPROVED"
	EventTypeVaultProposalRejected  = "VAULT_PROPOSAL_REJECTED"
	EventTypeVaultProposalExecuted  = "VAULT_PROPOSAL_EXECUTED"
	EventTypeVaultProposalCancelled = "VAULT_PROPOSAL_CANCELLED"
	EventTypeVaultSettingsUpdated   = "VAULT_SETTINGS_UPDATED"
	EventTypeVaultItemDeposited     = "VAULT_ITEM_DEPOSITED"
	EventTypeVaultItemTransferred   = "VAULT_ITEM_TRANSFERRED"
)

// Analytics topics
const (
	TopicAnalyticsEntityViewed  = "analytics.entity.viewed"
	TopicAnalyticsStatsComputed = "analytics.stats.computed"
	TopicAnalyticsDLQ           = "analytics.dlq"
)

// Analytics event types
const (
	EventTypeEntityViewed  = "ENTITY_VIEWED"
	EventTypeStatsComputed = "STATS_COMPUTED"
)

// EventPublisher publishes domain events to Kafka topics
type EventPublisher struct {
	client *Client
}

// NewEventPublisher creates a new EventPublisher
func NewEventPublisher(client *Client) *EventPublisher {
	return &EventPublisher{client: client}
}

// QueueEvent represents a matchmaking queue event
type QueueEvent struct {
	EventID   uuid.UUID         `json:"event_id"`
	PlayerID  uuid.UUID         `json:"player_id"`
	GameType  string            `json:"game_type"`
	Region    string            `json:"region"`
	MMR       int               `json:"mmr"`
	QueueTime int64             `json:"queue_time"`
	EventType string            `json:"event_type"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// PublishQueueEvent publishes a queue event
func (p *EventPublisher) PublishQueueEvent(ctx context.Context, event *QueueEvent) error {
	// In development mode, client may be nil - skip publishing
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	if event.QueueTime == 0 {
		event.QueueTime = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.PlayerID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"game_type":  event.GameType,
		},
	}

	return p.client.Publish(ctx, TopicQueueEvents, msg)
}

// LobbyEvent represents a lobby lifecycle event
type LobbyEvent struct {
	EventID   uuid.UUID         `json:"event_id"`
	LobbyID   uuid.UUID         `json:"lobby_id"`
	EventType string            `json:"event_type"`
	PlayerIDs []uuid.UUID       `json:"player_ids,omitempty"`
	GameType  string            `json:"game_type"`
	Region    string            `json:"region"`
	AvgMMR    int               `json:"avg_mmr"`
	CreatedAt int64             `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// PublishLobbyEvent publishes a lobby event
func (p *EventPublisher) PublishLobbyEvent(ctx context.Context, event *LobbyEvent) error {
	event.EventID = uuid.New()
	if event.CreatedAt == 0 {
		event.CreatedAt = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.LobbyID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"lobby_id":   event.LobbyID.String(),
		},
	}

	return p.client.Publish(ctx, TopicLobbyEvents, msg)
}

// PrizePoolEvent represents a prize pool update event
type PrizePoolEvent struct {
	EventID        uuid.UUID         `json:"event_id"`
	PoolID         uuid.UUID         `json:"pool_id"`
	LobbyID        uuid.UUID         `json:"lobby_id"`
	EventType      string            `json:"event_type"`
	TotalAmount    int64             `json:"total_amount"`
	Currency       string            `json:"currency"`
	ContributorID  *uuid.UUID        `json:"contributor_id,omitempty"`
	ContributionAmt int64            `json:"contribution_amount,omitempty"`
	Timestamp      int64             `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PublishPrizePoolEvent publishes a prize pool event
func (p *EventPublisher) PublishPrizePoolEvent(ctx context.Context, event *PrizePoolEvent) error {
	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.PoolID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"lobby_id":   event.LobbyID.String(),
		},
	}

	return p.client.Publish(ctx, TopicPrizePoolEvents, msg)
}

// MatchEvent represents a match creation or result event
type MatchEvent struct {
	EventID   uuid.UUID         `json:"event_id"`
	MatchID   uuid.UUID         `json:"match_id"`
	LobbyID   uuid.UUID         `json:"lobby_id"`
	EventType string            `json:"event_type"`
	GameType  string            `json:"game_type"`
	Region    string            `json:"region"`
	PlayerIDs []uuid.UUID       `json:"player_ids"`
	Teams     []TeamInfo        `json:"teams,omitempty"`
	Result    *MatchResult      `json:"result,omitempty"`
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// TeamInfo contains team details in a match
type TeamInfo struct {
	TeamID    uuid.UUID   `json:"team_id"`
	Name      string      `json:"name"`
	PlayerIDs []uuid.UUID `json:"player_ids"`
	Side      string      `json:"side,omitempty"` // e.g., "CT", "T" for CS2
}

// MatchResult contains match outcome details
type MatchResult struct {
	WinnerTeamID  *uuid.UUID        `json:"winner_team_id,omitempty"`
	IsDraw        bool              `json:"is_draw"`
	Scores        map[string]int    `json:"scores"` // team_id -> score
	Duration      int64             `json:"duration_seconds"`
	PlayerStats   []PlayerMatchStat `json:"player_stats,omitempty"`
	CompletedAt   int64             `json:"completed_at"`
}

// PlayerMatchStat contains individual player performance
type PlayerMatchStat struct {
	PlayerID uuid.UUID `json:"player_id"`
	Kills    int       `json:"kills"`
	Deaths   int       `json:"deaths"`
	Assists  int       `json:"assists"`
	Score    int       `json:"score"`
	MMRChange int      `json:"mmr_change"`
}

// PublishMatchCreated publishes a match creation event
func (p *EventPublisher) PublishMatchCreated(ctx context.Context, event *MatchEvent) error {
	event.EventID = uuid.New()
	event.EventType = EventTypeMatchCreated
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"lobby_id":   event.LobbyID.String(),
		},
	}

	return p.client.Publish(ctx, TopicMatchesCreated, msg)
}

// PublishMatchResult publishes a match result event
func (p *EventPublisher) PublishMatchResult(ctx context.Context, event *MatchEvent) error {
	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"match_id":   event.MatchID.String(),
		},
	}

	return p.client.Publish(ctx, TopicMatchesResults, msg)
}

// WebSocketBroadcastEvent represents an event to broadcast to WebSocket clients
type WebSocketBroadcastEvent struct {
	EventID   uuid.UUID   `json:"event_id"`
	Type      string      `json:"type"`
	LobbyID   *uuid.UUID  `json:"lobby_id,omitempty"`
	TargetIDs []uuid.UUID `json:"target_ids,omitempty"` // specific user IDs, nil for broadcast
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

// ReadyCheckEvent represents a readiness confirmation lifecycle event
type ReadyCheckEvent struct {
	EventID            uuid.UUID              `json:"event_id"`
	LobbyID            uuid.UUID              `json:"lobby_id"`
	PlayerID           *uuid.UUID             `json:"player_id,omitempty"`
	EventType          string                 `json:"event_type"`
	PlayerIDs          []uuid.UUID            `json:"player_ids,omitempty"`
	Summary            interface{}            `json:"summary,omitempty"`
	GameConnectionInfo interface{}            `json:"game_connection_info,omitempty"`
	Timestamp          int64                  `json:"timestamp"`
	Metadata           map[string]string      `json:"metadata,omitempty"`
}

// PublishReadyCheckEvent publishes a ready check lifecycle event
func (p *EventPublisher) PublishReadyCheckEvent(ctx context.Context, event *ReadyCheckEvent) error {
	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	key := event.LobbyID.String()

	msg := &Message{
		Key:       key,
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"lobby_id":   event.LobbyID.String(),
		},
	}

	return p.client.Publish(ctx, TopicReadyCheck, msg)
}

// PublishWebSocketBroadcast publishes an event for WebSocket broadcast
func (p *EventPublisher) PublishWebSocketBroadcast(ctx context.Context, event *WebSocketBroadcastEvent) error {
	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	key := "broadcast"
	if event.LobbyID != nil {
		key = event.LobbyID.String()
	}

	msg := &Message{
		Key:       key,
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.Type,
		},
	}

	return p.client.Publish(ctx, TopicWebSocketBroadcast, msg)
}

// PublishPlayerStatus publishes player status update (compacted topic)
func (p *EventPublisher) PublishPlayerStatus(ctx context.Context, playerID uuid.UUID, status string, metadata map[string]string) error {
	event := map[string]interface{}{
		"player_id":  playerID,
		"status":     status,
		"updated_at": time.Now().UnixMilli(),
		"metadata":   metadata,
	}

	msg := &Message{
		Key:       playerID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"status": status,
		},
	}

	return p.client.Publish(ctx, TopicPlayerStatus, msg)
}

// PublishToDLQ publishes a failed message to the dead letter queue
func (p *EventPublisher) PublishToDLQ(ctx context.Context, originalTopic string, originalKey string, value interface{}, err error) error {
	dlqEvent := map[string]interface{}{
		"original_topic": originalTopic,
		"original_key":   originalKey,
		"value":          value,
		"error":          err.Error(),
		"timestamp":      time.Now().UnixMilli(),
	}

	msg := &Message{
		Key:       uuid.New().String(),
		Value:     dlqEvent,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"original_topic": originalTopic,
			"error_type":     "processing_failed",
		},
	}

	return p.client.Publish(ctx, TopicDLQ, msg)
}

// ═══════════════════════════════════════════════════════════════════════════════
// REPLAY PROCESSING EVENTS
// ═══════════════════════════════════════════════════════════════════════════════

// ReplayUploadedEvent is published when a replay file is uploaded
type ReplayUploadedEvent struct {
	EventID      uuid.UUID         `json:"event_id"`
	ReplayFileID uuid.UUID         `json:"replay_file_id"`
	GameID       string            `json:"game_id"`
	UserID       uuid.UUID         `json:"user_id"`
	TenantID     uuid.UUID         `json:"tenant_id"`
	FileSize     int               `json:"file_size"`
	InternalURI  string            `json:"internal_uri"`
	Timestamp    int64             `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PublishReplayUploaded publishes a replay uploaded event
func (p *EventPublisher) PublishReplayUploaded(ctx context.Context, event *ReplayUploadedEvent) error {
	// In development mode, client may be nil - skip publishing
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.ReplayFileID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": EventTypeReplayUploaded,
			"game_id":    event.GameID,
		},
	}

	return p.client.Publish(ctx, TopicReplaysUploaded, msg)
}

// ReplayProcessingEvent is published during replay processing
type ReplayProcessingEvent struct {
	EventID      uuid.UUID         `json:"event_id"`
	ReplayFileID uuid.UUID         `json:"replay_file_id"`
	MatchID      uuid.UUID         `json:"match_id,omitempty"`
	EventType    string            `json:"event_type"`
	Progress     int               `json:"progress"`       // 0-100
	Stage        string            `json:"stage"`          // "parsing", "extracting", "saving", "completed"
	EventCount   int               `json:"event_count"`
	PlayerCount  int               `json:"player_count"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Timestamp    int64             `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PublishReplayProcessing publishes a replay processing progress event
func (p *EventPublisher) PublishReplayProcessing(ctx context.Context, event *ReplayProcessingEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.ReplayFileID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"stage":      event.Stage,
		},
	}

	return p.client.Publish(ctx, TopicReplaysProcessing, msg)
}

// ReplayCompletedEvent is published when replay processing completes successfully
type ReplayCompletedEvent struct {
	EventID      uuid.UUID         `json:"event_id"`
	ReplayFileID uuid.UUID         `json:"replay_file_id"`
	MatchID      uuid.UUID         `json:"match_id"`
	GameID       string            `json:"game_id"`
	EventCount   int               `json:"event_count"`
	PlayerCount  int               `json:"player_count"`
	Duration     int64             `json:"duration_ms"`    // Processing duration
	MatchDuration int64            `json:"match_duration"` // Game duration in seconds
	Timestamp    int64             `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PublishReplayCompleted publishes a replay completed event
func (p *EventPublisher) PublishReplayCompleted(ctx context.Context, event *ReplayCompletedEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.ReplayFileID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": EventTypeReplayCompleted,
			"match_id":   event.MatchID.String(),
		},
	}

	return p.client.Publish(ctx, TopicReplaysCompleted, msg)
}

// ReplayFailedEvent is published when replay processing fails
type ReplayFailedEvent struct {
	EventID      uuid.UUID         `json:"event_id"`
	ReplayFileID uuid.UUID         `json:"replay_file_id"`
	GameID       string            `json:"game_id"`
	Stage        string            `json:"stage"`
	ErrorType    string            `json:"error_type"`
	ErrorMessage string            `json:"error_message"`
	Retryable    bool              `json:"retryable"`
	RetryCount   int               `json:"retry_count"`
	Timestamp    int64             `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PublishReplayFailed publishes a replay failed event
func (p *EventPublisher) PublishReplayFailed(ctx context.Context, event *ReplayFailedEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.ReplayFileID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": EventTypeReplayFailed,
			"error_type": event.ErrorType,
		},
	}

	return p.client.Publish(ctx, TopicReplaysFailed, msg)
}

// ═══════════════════════════════════════════════════════════════════════════════
// BILLING EVENTS
// ═══════════════════════════════════════════════════════════════════════════════

// SubscriptionEvent represents a subscription lifecycle event
type SubscriptionEvent struct {
	EventID        uuid.UUID         `json:"event_id"`
	SubscriptionID uuid.UUID         `json:"subscription_id"`
	UserID         uuid.UUID         `json:"user_id"`
	PlanID         uuid.UUID         `json:"plan_id"`
	PreviousPlanID *uuid.UUID        `json:"previous_plan_id,omitempty"`
	EventType      string            `json:"event_type"`
	Status         string            `json:"status"`
	BillingPeriod  string            `json:"billing_period"`
	StartAt        int64             `json:"start_at"`
	EndAt          *int64            `json:"end_at,omitempty"`
	IsFree         bool              `json:"is_free"`
	Reason         string            `json:"reason,omitempty"`
	Timestamp      int64             `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PublishSubscriptionCreated publishes a subscription created event
func (p *EventPublisher) PublishSubscriptionCreated(ctx context.Context, event *SubscriptionEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeSubscriptionCreated
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.SubscriptionID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
		},
	}

	return p.client.Publish(ctx, TopicBillingSubscriptionCreated, msg)
}

// PublishSubscriptionUpgraded publishes a subscription upgraded event
func (p *EventPublisher) PublishSubscriptionUpgraded(ctx context.Context, event *SubscriptionEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeSubscriptionUpgraded
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.SubscriptionID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":       event.EventType,
			"user_id":          event.UserID.String(),
			"previous_plan_id": event.PreviousPlanID.String(),
		},
	}

	return p.client.Publish(ctx, TopicBillingSubscriptionUpgraded, msg)
}

// PublishSubscriptionCancelled publishes a subscription cancelled event
func (p *EventPublisher) PublishSubscriptionCancelled(ctx context.Context, event *SubscriptionEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeSubscriptionCancelled
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.SubscriptionID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
		},
	}

	return p.client.Publish(ctx, TopicBillingSubscriptionCancelled, msg)
}

// PublishSubscriptionExpired publishes a subscription expired event
func (p *EventPublisher) PublishSubscriptionExpired(ctx context.Context, event *SubscriptionEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeSubscriptionExpired
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.SubscriptionID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
		},
	}

	return p.client.Publish(ctx, TopicBillingSubscriptionExpired, msg)
}

// PaymentEvent represents a payment lifecycle event
type PaymentEvent struct {
	EventID        uuid.UUID         `json:"event_id"`
	PaymentID      uuid.UUID         `json:"payment_id"`
	UserID         uuid.UUID         `json:"user_id"`
	SubscriptionID *uuid.UUID        `json:"subscription_id,omitempty"`
	EventType      string            `json:"event_type"`
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Provider       string            `json:"provider"` // "stripe", "paypal", etc.
	ProviderRef    string            `json:"provider_ref,omitempty"`
	Status         string            `json:"status"`
	FailureReason  string            `json:"failure_reason,omitempty"`
	Timestamp      int64             `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PublishPaymentProcessed publishes a payment processed event
func (p *EventPublisher) PublishPaymentProcessed(ctx context.Context, event *PaymentEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypePaymentProcessed
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.PaymentID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
		},
	}

	return p.client.Publish(ctx, TopicBillingPaymentProcessed, msg)
}

// PublishPaymentFailed publishes a payment failed event
func (p *EventPublisher) PublishPaymentFailed(ctx context.Context, event *PaymentEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypePaymentFailed
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.PaymentID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":     event.EventType,
			"user_id":        event.UserID.String(),
			"failure_reason": event.FailureReason,
		},
	}

	return p.client.Publish(ctx, TopicBillingPaymentFailed, msg)
}

// PublishPaymentRefunded publishes a payment refunded event
func (p *EventPublisher) PublishPaymentRefunded(ctx context.Context, event *PaymentEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypePaymentRefunded
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.PaymentID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
		},
	}

	return p.client.Publish(ctx, TopicBillingPaymentRefunded, msg)
}
// ═══════════════════════════════════════════════════════════════════════════════
// WALLET EVENTS
// ═══════════════════════════════════════════════════════════════════════════════

// WalletEvent represents a wallet transaction event
type WalletEvent struct {
	EventID        uuid.UUID         `json:"event_id"`
	WalletID       uuid.UUID         `json:"wallet_id"`
	UserID         uuid.UUID         `json:"user_id"`
	EventType      string            `json:"event_type"`
	Amount         float64           `json:"amount"`
	Currency       string            `json:"currency"`
	BalanceBefore  float64           `json:"balance_before,omitempty"`
	BalanceAfter   float64           `json:"balance_after,omitempty"`
	TransactionID  *uuid.UUID        `json:"transaction_id,omitempty"`
	LedgerEntryID  *uuid.UUID        `json:"ledger_entry_id,omitempty"`
	Description    string            `json:"description,omitempty"`
	ToAddress      string            `json:"to_address,omitempty"`
	FromAddress    string            `json:"from_address,omitempty"`
	TxHash         string            `json:"tx_hash,omitempty"`
	MatchID        *uuid.UUID        `json:"match_id,omitempty"`
	TournamentID   *uuid.UUID        `json:"tournament_id,omitempty"`
	FailureReason  string            `json:"failure_reason,omitempty"`
	LockReason     string            `json:"lock_reason,omitempty"`
	Timestamp      int64             `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PublishWalletCreated publishes a wallet created event
func (p *EventPublisher) PublishWalletCreated(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletCreated
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
		},
	}

	return p.client.Publish(ctx, TopicWalletCreated, msg)
}

// PublishWalletDeposit publishes a wallet deposit event
func (p *EventPublisher) PublishWalletDeposit(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletDeposit
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
			"amount":     fmt.Sprintf("%.2f", event.Amount),
			"currency":   event.Currency,
		},
	}

	return p.client.Publish(ctx, TopicWalletDeposit, msg)
}

// PublishWalletWithdrawal publishes a wallet withdrawal event
func (p *EventPublisher) PublishWalletWithdrawal(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletWithdrawal
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
			"amount":     fmt.Sprintf("%.2f", event.Amount),
			"currency":   event.Currency,
			"to_address": event.ToAddress,
		},
	}

	return p.client.Publish(ctx, TopicWalletWithdrawal, msg)
}

// PublishWalletWithdrawalPending publishes a pending withdrawal event
func (p *EventPublisher) PublishWalletWithdrawalPending(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletWithdrawalPending
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
			"amount":     fmt.Sprintf("%.2f", event.Amount),
			"currency":   event.Currency,
		},
	}

	return p.client.Publish(ctx, TopicWalletWithdrawalPending, msg)
}

// PublishWalletEntryFee publishes a wallet entry fee deduction event
func (p *EventPublisher) PublishWalletEntryFee(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletEntryFee
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
			"amount":     fmt.Sprintf("%.2f", event.Amount),
			"currency":   event.Currency,
		},
	}

	return p.client.Publish(ctx, TopicWalletEntryFee, msg)
}

// PublishWalletPrize publishes a wallet prize winning event
func (p *EventPublisher) PublishWalletPrize(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletPrize
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
			"amount":     fmt.Sprintf("%.2f", event.Amount),
			"currency":   event.Currency,
		},
	}

	return p.client.Publish(ctx, TopicWalletPrize, msg)
}

// PublishWalletRefund publishes a wallet refund event
func (p *EventPublisher) PublishWalletRefund(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletRefund
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
			"amount":     fmt.Sprintf("%.2f", event.Amount),
			"currency":   event.Currency,
		},
	}

	return p.client.Publish(ctx, TopicWalletRefund, msg)
}

// PublishWalletLocked publishes a wallet locked event
func (p *EventPublisher) PublishWalletLocked(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletLocked
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":  event.EventType,
			"user_id":     event.UserID.String(),
			"lock_reason": event.LockReason,
		},
	}

	return p.client.Publish(ctx, TopicWalletLocked, msg)
}

// PublishWalletUnlocked publishes a wallet unlocked event
func (p *EventPublisher) PublishWalletUnlocked(ctx context.Context, event *WalletEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeWalletUnlocked
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.WalletID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"user_id":    event.UserID.String(),
		},
	}

	return p.client.Publish(ctx, TopicWalletUnlocked, msg)
}

// GetAllWalletTopics returns all wallet-related topics
func GetAllWalletTopics() []string {
	return []string{
		TopicWalletCreated,
		TopicWalletDeposit,
		TopicWalletWithdrawal,
		TopicWalletWithdrawalPending,
		TopicWalletEntryFee,
		TopicWalletPrize,
		TopicWalletRefund,
		TopicWalletLocked,
		TopicWalletUnlocked,
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// VAULT EVENTS
// ═══════════════════════════════════════════════════════════════════════════════

// VaultEvent represents a team vault event
type VaultEvent struct {
	EventID       uuid.UUID         `json:"event_id"`
	VaultID       uuid.UUID         `json:"vault_id"`
	SquadID       uuid.UUID         `json:"squad_id"`
	ActorID       uuid.UUID         `json:"actor_id"`
	EventType     string            `json:"event_type"`
	Amount        float64           `json:"amount,omitempty"`
	Currency      string            `json:"currency,omitempty"`
	ProposalID    *uuid.UUID        `json:"proposal_id,omitempty"`
	ProposalType  string            `json:"proposal_type,omitempty"`
	ItemID        *uuid.UUID        `json:"item_id,omitempty"`
	Description   string            `json:"description,omitempty"`
	TxHash        string            `json:"tx_hash,omitempty"`
	Timestamp     int64             `json:"timestamp"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// PublishVaultCreated publishes a vault created event
func (p *EventPublisher) PublishVaultCreated(ctx context.Context, event *VaultEvent) error {
	if p.client == nil {
		return nil
	}
	event.EventID = uuid.New()
	event.EventType = EventTypeVaultCreated
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}
	msg := &Message{
		Key:       event.VaultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"squad_id":   event.SquadID.String(),
		},
	}
	return p.client.Publish(ctx, TopicVaultCreated, msg)
}

// PublishVaultDeposit publishes a vault deposit event
func (p *EventPublisher) PublishVaultDeposit(ctx context.Context, event *VaultEvent) error {
	if p.client == nil {
		return nil
	}
	event.EventID = uuid.New()
	event.EventType = EventTypeVaultDeposit
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}
	msg := &Message{
		Key:       event.VaultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"squad_id":   event.SquadID.String(),
		},
	}
	return p.client.Publish(ctx, TopicVaultDeposit, msg)
}

// PublishVaultProposalCreated publishes a vault proposal created event
func (p *EventPublisher) PublishVaultProposalCreated(ctx context.Context, event *VaultEvent) error {
	if p.client == nil {
		return nil
	}
	event.EventID = uuid.New()
	event.EventType = EventTypeVaultProposalCreated
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}
	msg := &Message{
		Key:       event.VaultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"squad_id":   event.SquadID.String(),
		},
	}
	return p.client.Publish(ctx, TopicVaultProposalCreated, msg)
}

// PublishVaultProposalApproved publishes a vault proposal approved event
func (p *EventPublisher) PublishVaultProposalApproved(ctx context.Context, event *VaultEvent) error {
	if p.client == nil {
		return nil
	}
	event.EventID = uuid.New()
	event.EventType = EventTypeVaultProposalApproved
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}
	msg := &Message{
		Key:       event.VaultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"squad_id":   event.SquadID.String(),
		},
	}
	return p.client.Publish(ctx, TopicVaultProposalApproved, msg)
}

// PublishVaultProposalExecuted publishes a vault proposal executed event
func (p *EventPublisher) PublishVaultProposalExecuted(ctx context.Context, event *VaultEvent) error {
	if p.client == nil {
		return nil
	}
	event.EventID = uuid.New()
	event.EventType = EventTypeVaultProposalExecuted
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}
	msg := &Message{
		Key:       event.VaultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"squad_id":   event.SquadID.String(),
		},
	}
	return p.client.Publish(ctx, TopicVaultProposalExecuted, msg)
}

// GetAllVaultTopics returns all vault-related topics
func GetAllVaultTopics() []string {
	return []string{
		TopicVaultCreated,
		TopicVaultDeposit,
		TopicVaultWithdrawal,
		TopicVaultProposalCreated,
		TopicVaultProposalApproved,
		TopicVaultProposalRejected,
		TopicVaultProposalExecuted,
		TopicVaultProposalCancelled,
		TopicVaultSettingsUpdated,
		TopicVaultItemDeposited,
		TopicVaultItemTransferred,
	}
}

// --- Scores Domain Events ---

// Score topics
const (
	TopicScoreSubmitted   = "scores.submitted"
	TopicScoreVerified    = "scores.verified"
	TopicScoreDisputed    = "scores.disputed"
	TopicScoreConciliated = "scores.conciliated"
	TopicScoreFinalized   = "scores.finalized"
	TopicScoreCancelled   = "scores.cancelled"
	TopicScoreDLQ         = "scores.dlq"
)

// Score event types
const (
	EventTypeScoreSubmitted   = "SCORE_SUBMITTED"
	EventTypeScoreVerified    = "SCORE_VERIFIED"
	EventTypeScoreDisputed    = "SCORE_DISPUTED"
	EventTypeScoreConciliated = "SCORE_CONCILIATED"
	EventTypeScoreFinalized   = "SCORE_FINALIZED"
	EventTypeScoreCancelled   = "SCORE_CANCELLED"
)

// ScoreEvent represents a score/match result domain event
type ScoreEvent struct {
	EventID              uuid.UUID         `json:"event_id"`
	MatchResultID        uuid.UUID         `json:"match_result_id"`
	MatchID              uuid.UUID         `json:"match_id"`
	TournamentID         *uuid.UUID        `json:"tournament_id,omitempty"`
	MatchmakingSessionID *uuid.UUID        `json:"matchmaking_session_id,omitempty"`
	GameID               string            `json:"game_id"`
	Source               string            `json:"source"`
	Status               string            `json:"status"`
	EventType            string            `json:"event_type"`
	WinnerTeamID         *uuid.UUID        `json:"winner_team_id,omitempty"`
	IsDraw               bool              `json:"is_draw"`
	TeamScores           []TeamScoreInfo   `json:"team_scores,omitempty"`
	DisputeReason        string            `json:"dispute_reason,omitempty"`
	PrizeDistributionID  *uuid.UUID        `json:"prize_distribution_id,omitempty"`
	Timestamp            int64             `json:"timestamp"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

// TeamScoreInfo represents a team's score in an event
type TeamScoreInfo struct {
	TeamID   uuid.UUID `json:"team_id"`
	TeamName string    `json:"team_name"`
	Score    int       `json:"score"`
	Position int       `json:"position"`
}

// PublishScoreSubmitted publishes a score submitted event
func (p *EventPublisher) PublishScoreSubmitted(ctx context.Context, event *ScoreEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeScoreSubmitted
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchResultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":      event.EventType,
			"match_id":        event.MatchID.String(),
			"match_result_id": event.MatchResultID.String(),
			"game_id":         event.GameID,
			"source":          event.Source,
		},
	}

	return p.client.Publish(ctx, TopicScoreSubmitted, msg)
}

// PublishScoreVerified publishes a score verified event
func (p *EventPublisher) PublishScoreVerified(ctx context.Context, event *ScoreEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeScoreVerified
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchResultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":      event.EventType,
			"match_result_id": event.MatchResultID.String(),
		},
	}

	return p.client.Publish(ctx, TopicScoreVerified, msg)
}

// PublishScoreDisputed publishes a score disputed event
func (p *EventPublisher) PublishScoreDisputed(ctx context.Context, event *ScoreEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeScoreDisputed
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchResultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":      event.EventType,
			"match_result_id": event.MatchResultID.String(),
			"dispute_reason":  event.DisputeReason,
		},
	}

	return p.client.Publish(ctx, TopicScoreDisputed, msg)
}

// PublishScoreConciliated publishes a score conciliated event
func (p *EventPublisher) PublishScoreConciliated(ctx context.Context, event *ScoreEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeScoreConciliated
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchResultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":      event.EventType,
			"match_result_id": event.MatchResultID.String(),
		},
	}

	return p.client.Publish(ctx, TopicScoreConciliated, msg)
}

// PublishScoreFinalized publishes a score finalized event
func (p *EventPublisher) PublishScoreFinalized(ctx context.Context, event *ScoreEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeScoreFinalized
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchResultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":      event.EventType,
			"match_result_id": event.MatchResultID.String(),
			"match_id":        event.MatchID.String(),
		},
	}

	return p.client.Publish(ctx, TopicScoreFinalized, msg)
}

// PublishScoreCancelled publishes a score cancelled event
func (p *EventPublisher) PublishScoreCancelled(ctx context.Context, event *ScoreEvent) error {
	if p.client == nil {
		return nil
	}

	event.EventID = uuid.New()
	event.EventType = EventTypeScoreCancelled
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchResultID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type":      event.EventType,
			"match_result_id": event.MatchResultID.String(),
		},
	}

	return p.client.Publish(ctx, TopicScoreCancelled, msg)
}

// GetAllScoreTopics returns all score-related topics
func GetAllScoreTopics() []string {
	return []string{
		TopicScoreSubmitted,
		TopicScoreVerified,
		TopicScoreDisputed,
		TopicScoreConciliated,
		TopicScoreFinalized,
		TopicScoreCancelled,
	}
}

// --- Oracle Domain Events ---

// Game import topics
const (
	TopicGameImportDiscovered = "game-import.discovered"
	TopicGameImportProcessed  = "game-import.processed"
	TopicGameImportCompleted  = "game-import.completed"
	TopicGameImportFailed     = "game-import.failed"
	TopicGameImportDLQ        = "game-import.dlq"
)

// Game import event types
const (
	EventTypeGameImportDiscovered = "GAME_IMPORT_DISCOVERED"
	EventTypeGameImportProcessed  = "GAME_IMPORT_PROCESSED"
	EventTypeGameImportCompleted  = "GAME_IMPORT_COMPLETED"
	EventTypeGameImportFailed     = "GAME_IMPORT_FAILED"
)

// GetAllGameImportTopics returns all game-import related topics
func GetAllGameImportTopics() []string {
	return []string{
		TopicGameImportDiscovered,
		TopicGameImportProcessed,
		TopicGameImportCompleted,
		TopicGameImportFailed,
	}
}

// Oracle topics
const (
	TopicOracleConsensusReached = "oracle.consensus.reached"
	TopicOraclePublished        = "oracle.published"
	TopicOracleFinalized        = "oracle.finalized"
	TopicOracleDisputed         = "oracle.disputed"
	TopicOracleExternalIngested = "oracle.external.ingested"
	TopicOracleDLQ              = "oracle.dlq"
)

// Oracle event types
const (
	EventTypeOracleConsensusReached = "ORACLE_CONSENSUS_REACHED"
	EventTypeOraclePublished        = "ORACLE_PUBLISHED"
	EventTypeOracleFinalized        = "ORACLE_FINALIZED"
	EventTypeOracleDisputed         = "ORACLE_DISPUTED"
	EventTypeOracleExternalIngested = "ORACLE_EXTERNAL_INGESTED"
)

// OracleEvent represents an oracle domain event
type OracleEvent struct {
	EventID          uuid.UUID             `json:"event_id"`
	OracleResultID   uuid.UUID             `json:"oracle_result_id"`
	MatchID          *uuid.UUID            `json:"match_id,omitempty"`
	ExternalMatchID  string                `json:"external_match_id,omitempty"`
	GameID           string                `json:"game_id"`
	Status           string                `json:"status"`
	EventType        string                `json:"event_type"`
	ConfidenceLevel  int                   `json:"confidence_level"`
	OverallAgreement float64               `json:"overall_agreement"`
	SourceType       string                `json:"source_type,omitempty"`
	SourceProvider   string                `json:"source_provider,omitempty"`
	ChainPublications []OracleChainPubInfo `json:"chain_publications,omitempty"`
	Timestamp        int64                 `json:"timestamp"`
	Metadata         map[string]string     `json:"metadata,omitempty"`
}

// OracleChainPubInfo represents chain publication info in events
type OracleChainPubInfo struct {
	ChainID int    `json:"chain_id"`
	TxHash  string `json:"tx_hash"`
	Status  string `json:"status"`
}

// GetAllOracleTopics returns all oracle-related topics
func GetAllOracleTopics() []string {
	return []string{
		TopicOracleConsensusReached,
		TopicOraclePublished,
		TopicOracleFinalized,
		TopicOracleDisputed,
		TopicOracleExternalIngested,
	}
}

// --- Tournament Event Topics ---

const (
	TopicTournamentEvents = "tournament.events"
)

// PublishTournamentEvent publishes a generic tournament lifecycle event
func (p *EventPublisher) PublishTournamentEvent(ctx context.Context, key string, event interface{}) error {
	if p.client == nil {
		return nil
	}

	msg := &Message{
		Key:       key,
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"domain": "tournament",
		},
	}

	return p.client.Publish(ctx, TopicTournamentEvents, msg)
}