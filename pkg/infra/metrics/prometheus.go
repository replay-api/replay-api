package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// Business metrics
	PaymentAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_attempts_total",
			Help: "Total number of payment attempts",
		},
		[]string{"provider", "type"},
	)

	PaymentFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_failures_total",
			Help: "Total number of payment failures",
		},
		[]string{"provider", "type", "reason"},
	)

	PaymentProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_processing_duration_seconds",
			Help:    "Payment processing duration in seconds",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"provider", "type"},
	)

	StripeWebhookFailuresTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "stripe_webhook_failures_total",
			Help: "Total number of Stripe webhook processing failures",
		},
	)

	TournamentParticipants = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tournament_participants",
			Help: "Current number of tournament participants",
		},
		[]string{"tournament_id", "status"},
	)

	LobbyActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "lobby_active_connections",
			Help: "Number of active lobby WebSocket connections",
		},
	)

	MatchmakingQueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "matchmaking_queue_size",
			Help: "Current matchmaking queue size",
		},
		[]string{"game", "mode", "region"},
	)

	WalletBalanceTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wallet_balance_total",
			Help: "Total wallet balance across all users (in cents)",
		},
		[]string{"currency"},
	)

	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "collection"},
	)

	CacheHitTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hit_total",
			Help: "Total cache hits",
		},
		[]string{"cache"},
	)

	CacheMissTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_miss_total",
			Help: "Total cache misses",
		},
		[]string{"cache"},
	)

	// ============================================
	// Esports Platform Metrics
	// ============================================

	// Matchmaking Metrics
	EsportsMatchmakingQueueJoins = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_matchmaking_queue_joins_total",
			Help: "Total number of players joining matchmaking queue",
		},
		[]string{"game", "mode", "region", "tier"},
	)

	EsportsMatchmakingQueueLeaves = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_matchmaking_queue_leaves_total",
			Help: "Total number of players leaving queue",
		},
		[]string{"game", "mode", "region", "reason"},
	)

	EsportsMatchmakingWaitTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_matchmaking_wait_seconds",
			Help:    "Time players spend waiting in matchmaking queue",
			Buckets: []float64{5, 15, 30, 60, 90, 120, 180, 300, 600},
		},
		[]string{"game", "mode", "region", "tier"},
	)

	EsportsMatchmakingSkillSpread = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_matchmaking_skill_spread",
			Help:    "MMR spread in matched lobbies (max - min)",
			Buckets: []float64{50, 100, 200, 300, 500, 750, 1000},
		},
		[]string{"game", "mode", "region"},
	)

	EsportsMatchmakingMatchesCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_matchmaking_matches_created_total",
			Help: "Total matches successfully created from matchmaking",
		},
		[]string{"game", "mode", "region"},
	)

	// Lobby Metrics
	EsportsLobbyCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_lobby_created_total",
			Help: "Total lobbies created",
		},
		[]string{"game", "lobby_type", "region"},
	)

	EsportsLobbyStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "esports_lobby_status_current",
			Help: "Current lobbies by status",
		},
		[]string{"game", "status"},
	)

	EsportsLobbyFillRate = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_lobby_fill_percentage",
			Help:    "Percentage of lobby slots filled at match start",
			Buckets: []float64{0.2, 0.4, 0.6, 0.8, 0.9, 1.0},
		},
		[]string{"game", "region"},
	)

	EsportsLobbyReadyCheckTimeout = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_lobby_ready_check_timeout_total",
			Help: "Number of lobbies cancelled due to ready check timeout",
		},
		[]string{"game", "region"},
	)

	EsportsLobbyLifecycle = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_lobby_lifecycle_seconds",
			Help:    "Time from lobby creation to match start or cancellation",
			Buckets: []float64{30, 60, 120, 180, 300, 600, 900},
		},
		[]string{"game", "outcome"},
	)

	// Tournament Metrics
	EsportsTournamentCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_tournament_created_total",
			Help: "Total tournaments created",
		},
		[]string{"game", "format", "region"},
	)

	EsportsTournamentStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "esports_tournament_status_current",
			Help: "Current tournaments by status",
		},
		[]string{"game", "status"},
	)

	EsportsTournamentRegistrations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_tournament_registrations_total",
			Help: "Total tournament registrations",
		},
		[]string{"game", "format", "region"},
	)

	EsportsTournamentFillRate = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_tournament_fill_percentage",
			Help:    "Percentage of tournament capacity filled at start",
			Buckets: []float64{0.25, 0.5, 0.75, 0.9, 1.0},
		},
		[]string{"game", "format"},
	)

	EsportsTournamentPrizePool = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "esports_tournament_prize_pool_cents",
			Help: "Total prize pool value in cents",
		},
		[]string{"game", "currency"},
	)

	EsportsTournamentMatchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_tournament_match_duration_seconds",
			Help:    "Duration of tournament matches",
			Buckets: []float64{300, 600, 900, 1200, 1800, 2700, 3600},
		},
		[]string{"game", "format", "round"},
	)

	// Kafka Metrics
	EsportsKafkaMessagesProduced = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_kafka_messages_produced_total",
			Help: "Total Kafka messages produced",
		},
		[]string{"topic"},
	)

	EsportsKafkaMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_kafka_messages_consumed_total",
			Help: "Total Kafka messages consumed",
		},
		[]string{"topic", "consumer_group"},
	)

	EsportsKafkaConsumerLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "esports_kafka_consumer_lag",
			Help: "Kafka consumer lag (messages behind)",
		},
		[]string{"topic", "partition", "consumer_group"},
	)

	EsportsKafkaProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_kafka_processing_duration_seconds",
			Help:    "Time to process Kafka messages",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"topic", "consumer_group"},
	)

	EsportsKafkaDLQ = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_kafka_dlq_messages_total",
			Help: "Messages sent to dead letter queue",
		},
		[]string{"original_topic", "error_type"},
	)

	// WebSocket Metrics
	EsportsWebSocketConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "esports_websocket_connections_total",
			Help: "Current WebSocket connections",
		},
	)

	EsportsWebSocketMessagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_websocket_messages_sent_total",
			Help: "Total WebSocket messages sent",
		},
		[]string{"message_type"},
	)

	EsportsWebSocketLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_websocket_message_latency_ms",
			Help:    "WebSocket message delivery latency",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
		},
		[]string{"message_type"},
	)

	// Player Engagement Metrics
	EsportsActivePlayers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "esports_active_players_current",
			Help: "Currently active players by activity type",
		},
		[]string{"game", "activity"},
	)

	EsportsPlayerSessionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_player_session_duration_seconds",
			Help:    "Player session duration",
			Buckets: []float64{300, 600, 1800, 3600, 7200, 14400},
		},
		[]string{"game"},
	)

	// Replay Metrics
	EsportsReplayUploads = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "esports_replay_uploads_total",
			Help: "Total replay files uploaded",
		},
		[]string{"game", "source"},
	)

	EsportsReplayProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_replay_processing_seconds",
			Help:    "Time to process and analyze replay files",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"game"},
	)

	EsportsReplayFileSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "esports_replay_file_size_bytes",
			Help:    "Size of uploaded replay files",
			Buckets: []float64{1e6, 5e6, 10e6, 50e6, 100e6, 500e6},
		},
		[]string{"game"},
	)

	// ============================================
	// View Analytics & Engagement Metrics
	// ============================================

	// EntityViewsRecordedTotal counts total view events recorded
	EntityViewsRecordedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "entity_views_recorded_total",
			Help: "Total entity view events recorded",
		},
		[]string{"entity_type", "referrer_type"},
	)

	// EntityUniqueViewersTotal counts unique authenticated viewer events
	EntityUniqueViewersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "entity_unique_viewers_total",
			Help: "Total unique authenticated viewer events (approximate)",
		},
		[]string{"entity_type"},
	)

	// ProfileImpressionsTotal counts profile page impressions
	ProfileImpressionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "profile_impressions_total",
			Help: "Total profile page impressions",
		},
		[]string{"profile_type"},
	)

	// AnalyticsEventsProcessedTotal tracks Kafka consumer processing
	AnalyticsEventsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_events_processed_total",
			Help: "Total analytics events processed by Kafka consumer",
		},
		[]string{"status", "entity_type"},
	)

	// AnalyticsEventProcessingDuration tracks per-event processing latency
	AnalyticsEventProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analytics_event_processing_duration_seconds",
			Help:    "Duration of analytics event processing in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"entity_type"},
	)

	// ViewStatisticsAggregationsTotal tracks aggregation upserts
	ViewStatisticsAggregationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "view_statistics_aggregations_total",
			Help: "Total view statistics aggregation operations",
		},
		[]string{"entity_type", "operation"},
	)

	// ============================================
	// Wallet & Billing Metrics
	// ============================================

	// Wallet Balance Metrics
	LeetgamingWalletBalanceTotalUSD = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "leetgaming_wallet_balance_total_usd",
			Help: "Total wallet balance in USD (converted)",
		},
		[]string{"wallet_type"},
	)

	LeetgamingWalletBalanceTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "leetgaming_wallet_balance_total",
			Help: "Total wallet balance by currency (in cents)",
		},
		[]string{"currency"},
	)

	LeetgamingWalletsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "leetgaming_wallets_total",
			Help: "Total number of wallets",
		},
	)

	LeetgamingWalletsLockedTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "leetgaming_wallets_locked_total",
			Help: "Number of locked wallets",
		},
	)

	// Transaction Metrics
	LeetgamingWalletDepositsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_deposits_total",
			Help: "Total number of deposits",
		},
		[]string{"currency", "provider"},
	)

	LeetgamingWalletDepositsTotalUSD = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_deposits_total_usd",
			Help: "Total deposits in USD",
		},
	)

	LeetgamingWalletWithdrawalsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_withdrawals_total",
			Help: "Total number of withdrawals",
		},
		[]string{"currency", "status"},
	)

	LeetgamingWalletWithdrawalsTotalUSD = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_withdrawals_total_usd",
			Help: "Total withdrawals in USD",
		},
	)

	LeetgamingWalletPrizesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_prizes_total",
			Help: "Total prizes paid",
		},
		[]string{"currency", "source"},
	)

	LeetgamingWalletPrizesTotalUSD = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_prizes_total_usd",
			Help: "Total prizes paid in USD",
		},
	)

	LeetgamingWalletEntryFeesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_entry_fees_total",
			Help: "Total entry fees collected",
		},
		[]string{"currency", "game"},
	)

	LeetgamingWalletEntryFeesTotalUSD = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_entry_fees_total_usd",
			Help: "Total entry fees in USD",
		},
	)

	LeetgamingWalletTransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_wallet_transactions_total",
			Help: "Total wallet transactions by type",
		},
		[]string{"type", "currency"},
	)

	LeetgamingWithdrawalsPendingTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "leetgaming_withdrawals_pending_total",
			Help: "Number of pending withdrawals",
		},
	)

	// Subscription Metrics
	LeetgamingSubscriptionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "leetgaming_subscriptions_active",
			Help: "Number of active subscriptions by tier",
		},
		[]string{"tier"},
	)

	LeetgamingSubscriptionsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_subscriptions_created_total",
			Help: "Total subscriptions created",
		},
		[]string{"tier", "billing_period"},
	)

	LeetgamingSubscriptionsChurnedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_subscriptions_churned_total",
			Help: "Total subscriptions churned/cancelled",
		},
		[]string{"tier", "reason"},
	)

	LeetgamingSubscriptionRevenueUSD = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_subscription_revenue_usd",
			Help: "Subscription revenue in USD",
		},
		[]string{"tier", "billing_period"},
	)

	// Payment Metrics
	LeetgamingPaymentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_payments_total",
			Help: "Total payments processed",
		},
		[]string{"provider", "status"},
	)

	LeetgamingPaymentFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_payment_failures_total",
			Help: "Total payment failures",
		},
		[]string{"provider", "reason"},
	)

	LeetgamingRefundsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_refunds_total",
			Help: "Total refunds processed",
		},
		[]string{"reason"},
	)

	// Security & Compliance Metrics
	LeetgamingFraudAlertsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "leetgaming_fraud_alerts_total",
			Help: "Total fraud alerts triggered",
		},
		[]string{"type", "severity"},
	)

	LeetgamingLedgerEntriesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "leetgaming_ledger_entries_total",
			Help: "Total ledger entries",
		},
	)

	LeetgamingLedgerBalancedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "leetgaming_ledger_balanced_total",
			Help: "Number of balanced ledger entries (debits = credits)",
		},
	)
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		start := time.Now()
		wrapped := newResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)
		path := normalizePath(r.URL.Path)

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

func normalizePath(path string) string {
	if len(path) > 50 {
		return path[:50]
	}
	return path
}

func Handler() http.Handler {
	return promhttp.Handler()
}

func RecordDBOperation(operation, collection string, duration time.Duration) {
	DatabaseOperationDuration.WithLabelValues(operation, collection).Observe(duration.Seconds())
}

func RecordCacheHit(cache string) {
	CacheHitTotal.WithLabelValues(cache).Inc()
}

func RecordCacheMiss(cache string) {
	CacheMissTotal.WithLabelValues(cache).Inc()
}

func RecordPaymentAttempt(provider, paymentType string) {
	PaymentAttemptsTotal.WithLabelValues(provider, paymentType).Inc()
}

func RecordPaymentFailure(provider, paymentType, reason string) {
	PaymentFailuresTotal.WithLabelValues(provider, paymentType, reason).Inc()
}

func RecordPaymentDuration(provider, paymentType string, duration time.Duration) {
	PaymentProcessingDuration.WithLabelValues(provider, paymentType).Observe(duration.Seconds())
}

// ============================================
// Wallet Metric Recording Functions
// ============================================

// RecordWalletDeposit records a deposit transaction
func RecordWalletDeposit(currency, provider string, amountUSD float64) {
	LeetgamingWalletDepositsTotal.WithLabelValues(currency, provider).Inc()
	LeetgamingWalletDepositsTotalUSD.Add(amountUSD)
	LeetgamingWalletTransactionsTotal.WithLabelValues("deposit", currency).Inc()
}

// RecordWalletWithdrawal records a withdrawal transaction
func RecordWalletWithdrawal(currency, status string, amountUSD float64) {
	LeetgamingWalletWithdrawalsTotal.WithLabelValues(currency, status).Inc()
	LeetgamingWalletWithdrawalsTotalUSD.Add(amountUSD)
	LeetgamingWalletTransactionsTotal.WithLabelValues("withdrawal", currency).Inc()
}

// RecordWalletPrize records a prize payout
func RecordWalletPrize(currency, source string, amountUSD float64) {
	LeetgamingWalletPrizesTotal.WithLabelValues(currency, source).Inc()
	LeetgamingWalletPrizesTotalUSD.Add(amountUSD)
	LeetgamingWalletTransactionsTotal.WithLabelValues("prize", currency).Inc()
}

// RecordWalletEntryFee records an entry fee deduction
func RecordWalletEntryFee(currency, game string, amountUSD float64) {
	LeetgamingWalletEntryFeesTotal.WithLabelValues(currency, game).Inc()
	LeetgamingWalletEntryFeesTotalUSD.Add(amountUSD)
	LeetgamingWalletTransactionsTotal.WithLabelValues("entry_fee", currency).Inc()
}

// SetWalletBalance updates the total wallet balance gauge
func SetWalletBalance(currency string, balanceCents float64) {
	LeetgamingWalletBalanceTotal.WithLabelValues(currency).Set(balanceCents)
}

// SetWalletBalanceUSD updates the USD balance gauge
func SetWalletBalanceUSD(walletType string, balanceUSD float64) {
	LeetgamingWalletBalanceTotalUSD.WithLabelValues(walletType).Set(balanceUSD)
}

// SetTotalWallets updates the total wallets gauge
func SetTotalWallets(count float64) {
	LeetgamingWalletsTotal.Set(count)
}

// SetLockedWallets updates the locked wallets gauge
func SetLockedWallets(count float64) {
	LeetgamingWalletsLockedTotal.Set(count)
}

// SetPendingWithdrawals updates the pending withdrawals gauge
func SetPendingWithdrawals(count float64) {
	LeetgamingWithdrawalsPendingTotal.Set(count)
}

// RecordSubscriptionCreated records a new subscription
func RecordSubscriptionCreated(tier, billingPeriod string) {
	LeetgamingSubscriptionsCreatedTotal.WithLabelValues(tier, billingPeriod).Inc()
}

// RecordSubscriptionChurned records a subscription cancellation
func RecordSubscriptionChurned(tier, reason string) {
	LeetgamingSubscriptionsChurnedTotal.WithLabelValues(tier, reason).Inc()
}

// SetActiveSubscriptions updates the active subscriptions gauge by tier
func SetActiveSubscriptions(tier string, count float64) {
	LeetgamingSubscriptionsActive.WithLabelValues(tier).Set(count)
}

// RecordSubscriptionRevenue records subscription revenue
func RecordSubscriptionRevenue(tier, billingPeriod string, amountUSD float64) {
	LeetgamingSubscriptionRevenueUSD.WithLabelValues(tier, billingPeriod).Add(amountUSD)
}

// RecordPaymentProcessed records a successful payment
func RecordPaymentProcessed(provider string) {
	LeetgamingPaymentsTotal.WithLabelValues(provider, "success").Inc()
}

// RecordPaymentFailed records a payment failure
func RecordPaymentFailed(provider, reason string) {
	LeetgamingPaymentsTotal.WithLabelValues(provider, "failed").Inc()
	LeetgamingPaymentFailuresTotal.WithLabelValues(provider, reason).Inc()
}

// RecordRefund records a refund
func RecordRefund(reason string) {
	LeetgamingRefundsTotal.WithLabelValues(reason).Inc()
}

// RecordFraudAlert records a fraud detection alert
func RecordFraudAlert(alertType, severity string) {
	LeetgamingFraudAlertsTotal.WithLabelValues(alertType, severity).Inc()
}

// RecordLedgerEntry records a ledger entry
func RecordLedgerEntry(balanced bool) {
	LeetgamingLedgerEntriesTotal.Inc()
	if balanced {
		LeetgamingLedgerBalancedTotal.Inc()
	}
}

// ============================================
// View Analytics Metric Recording Functions
// ============================================

// RecordEntityView records an entity view event
func RecordEntityView(entityType, referrerType string) {
	EntityViewsRecordedTotal.WithLabelValues(entityType, referrerType).Inc()
}

// RecordUniqueViewer records a unique authenticated viewer
func RecordUniqueViewer(entityType string) {
	EntityUniqueViewersTotal.WithLabelValues(entityType).Inc()
}

// RecordProfileImpression records a profile page impression
func RecordProfileImpression(profileType string) {
	ProfileImpressionsTotal.WithLabelValues(profileType).Inc()
}

// RecordAnalyticsEventProcessed records a processed analytics event
func RecordAnalyticsEventProcessed(status, entityType string) {
	AnalyticsEventsProcessedTotal.WithLabelValues(status, entityType).Inc()
}

// RecordAnalyticsProcessingDuration records analytics event processing time
func RecordAnalyticsProcessingDuration(entityType string, duration time.Duration) {
	AnalyticsEventProcessingDuration.WithLabelValues(entityType).Observe(duration.Seconds())
}

// RecordViewStatisticsAggregation records a view statistics aggregation operation
func RecordViewStatisticsAggregation(entityType, operation string) {
	ViewStatisticsAggregationsTotal.WithLabelValues(entityType, operation).Inc()
}
