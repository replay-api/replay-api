package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	//	"golang.org/x/oauth2/jwt"

	"github.com/golobby/container/v3"
	"github.com/replay-api/replay-api/cmd/rest-api/routing"
	jobs "github.com/replay-api/replay-api/pkg/app/jobs"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	mongodb "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
	ioc "github.com/replay-api/replay-api/pkg/infra/ioc"
	"github.com/replay-api/replay-api/pkg/infra/kafka"
	websocket "github.com/replay-api/replay-api/pkg/infra/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	builder := ioc.NewContainerBuilder()

	// Build container with env and event publisher
	c := builder.WithEnvFile().WithEventPublisher().Build()

	// Inject MongoDB services (includes squad services)
	if err := ioc.InjectMongoDB(c); err != nil {
		slog.Error("Failed to inject MongoDB services", "error", err)
		panic(err)
	}

	// Ensure MongoDB indexes are created on startup
	// This is idempotent - indexes that already exist will be skipped
	go func() {
		var client *mongo.Client
		if err := c.Resolve(&client); err != nil {
			slog.ErrorContext(ctx, "Failed to resolve MongoDB client for index creation", "error", err)
			return
		}

		dbName := os.Getenv("MONGODB_DATABASE")
		if dbName == "" {
			dbName = "leetgaming"
		}

		slog.InfoContext(ctx, "Ensuring MongoDB indexes on startup", "database", dbName)
		if err := mongodb.CreateIndexes(ctx, client, dbName); err != nil {
			slog.ErrorContext(ctx, "Failed to create MongoDB indexes", "error", err)
		} else {
			slog.InfoContext(ctx, "MongoDB indexes ensured successfully")
		}
	}()

	// Register inbound ports (services/use cases that depend on MongoDB repositories)
	// This must be called AFTER InjectMongoDB since inbound ports depend on outbound ports (repos)
	builder.WithInboundPorts()

	// Register Squad API services (repositories, writers, command handlers)
	// This must be called AFTER InjectMongoDB and WithInboundPorts since squad services depend on MongoDB
	builder.WithSquadAPI()

	defer builder.Close(c)

	// Start WebSocket Hub
	var wsHub *websocket.WebSocketHub
	if err := c.Resolve(&wsHub); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve WebSocket hub", "error", err)
		panic(err)
	}
	go wsHub.Run(ctx)
	slog.InfoContext(ctx, "WebSocket hub started")

	// Start Prize Distribution Job
	var prizeJob *jobs.PrizeDistributionJob
	if err := c.Resolve(&prizeJob); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve PrizeDistributionJob", "error", err)
		panic(err)
	}
	go prizeJob.Run(ctx)
	slog.InfoContext(ctx, "Prize distribution job started")

	// Start Billing Consumer (if Kafka is available)
	startBillingConsumer(ctx, c)

	// Start Wallet Consumer (if Kafka is available)
	startWalletConsumer(ctx, c)

	router := routing.NewRouter(ctx, c)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.InfoContext(ctx, "Starting server on port "+port)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB max header size
	}

	if err := server.ListenAndServe(); err != nil {
		slog.ErrorContext(ctx, "Server error", "err", err)
	}

}

// startBillingConsumer starts the billing Kafka consumer if Kafka is available
func startBillingConsumer(ctx context.Context, c container.Container) {
	// Check if Kafka is enabled
	kafkaEnabled := os.Getenv("KAFKA_ENABLED")
	if kafkaEnabled != "true" {
		slog.InfoContext(ctx, "Kafka billing consumer disabled (KAFKA_ENABLED != true)")
		return
	}

	// Get Kafka client
	var kafkaClient *kafka.Client
	if err := c.Resolve(&kafkaClient); err != nil {
		slog.WarnContext(ctx, "Kafka client not available, billing consumer will not start", "error", err)
		return
	}

	if kafkaClient == nil {
		slog.InfoContext(ctx, "Kafka client is nil, billing consumer will not start")
		return
	}

	// Get billing dependencies
	var subscriptionWriter billing_out.SubscriptionWriter
	var subscriptionReader billing_out.SubscriptionReader
	var planReader billing_out.PlanReader
	var eventPublisher *kafka.EventPublisher

	if err := c.Resolve(&subscriptionWriter); err != nil {
		slog.WarnContext(ctx, "SubscriptionWriter not available for billing consumer", "error", err)
	}
	if err := c.Resolve(&subscriptionReader); err != nil {
		slog.WarnContext(ctx, "SubscriptionReader not available for billing consumer", "error", err)
	}
	if err := c.Resolve(&planReader); err != nil {
		slog.WarnContext(ctx, "PlanReader not available for billing consumer", "error", err)
	}
	if err := c.Resolve(&eventPublisher); err != nil {
		slog.WarnContext(ctx, "EventPublisher not available for billing consumer", "error", err)
	}

	// Create billing consumer
	groupID := os.Getenv("KAFKA_BILLING_CONSUMER_GROUP")
	if groupID == "" {
		groupID = "billing-consumer-group"
	}

	billingConsumer := kafka.NewBillingConsumer(kafkaClient, &kafka.BillingConsumerConfig{
		GroupID:            groupID,
		SubscriptionWriter: subscriptionWriter,
		SubscriptionReader: subscriptionReader,
		PlanReader:         planReader,
		EventPublisher:     eventPublisher,
	})

	// Start consumer in background
	go func() {
		slog.InfoContext(ctx, "Starting billing Kafka consumer", "group_id", groupID)
		if err := billingConsumer.Start(ctx); err != nil {
			slog.ErrorContext(ctx, "Billing consumer error", "error", err)
		}
	}()

	slog.InfoContext(ctx, "Billing consumer started successfully")
}

// startWalletConsumer starts the wallet Kafka consumer for financial event auditing
func startWalletConsumer(ctx context.Context, c container.Container) {
	// Check if Kafka is enabled
	kafkaEnabled := os.Getenv("KAFKA_ENABLED")
	if kafkaEnabled != "true" {
		slog.InfoContext(ctx, "Kafka wallet consumer disabled (KAFKA_ENABLED != true)")
		return
	}

	// Get Kafka client
	var kafkaClient *kafka.Client
	if err := c.Resolve(&kafkaClient); err != nil {
		slog.WarnContext(ctx, "Kafka client not available, wallet consumer will not start", "error", err)
		return
	}

	if kafkaClient == nil {
		slog.InfoContext(ctx, "Kafka client is nil, wallet consumer will not start")
		return
	}

	// Get wallet dependencies
	var walletRepo wallet_out.WalletRepository
	var ledgerRepo wallet_out.LedgerRepository
	var walletCommand wallet_in.WalletCommand
	var eventPublisher *kafka.EventPublisher

	if err := c.Resolve(&walletRepo); err != nil {
		slog.WarnContext(ctx, "WalletRepository not available for wallet consumer", "error", err)
	}
	if err := c.Resolve(&ledgerRepo); err != nil {
		slog.WarnContext(ctx, "LedgerRepository not available for wallet consumer", "error", err)
	}
	if err := c.Resolve(&walletCommand); err != nil {
		slog.WarnContext(ctx, "WalletCommand not available for wallet consumer", "error", err)
	}
	if err := c.Resolve(&eventPublisher); err != nil {
		slog.WarnContext(ctx, "EventPublisher not available for wallet consumer", "error", err)
	}

	// Create wallet consumer
	groupID := os.Getenv("KAFKA_WALLET_CONSUMER_GROUP")
	if groupID == "" {
		groupID = "wallet-consumer-group"
	}

	walletConsumer := kafka.NewWalletConsumer(kafkaClient, &kafka.WalletConsumerConfig{
		GroupID:        groupID,
		WalletRepo:     walletRepo,
		LedgerRepo:     ledgerRepo,
		WalletCommand:  walletCommand,
		EventPublisher: eventPublisher,
	})

	// Start consumer in background
	go func() {
		slog.InfoContext(ctx, "Starting wallet Kafka consumer", "group_id", groupID)
		if err := walletConsumer.Start(ctx); err != nil {
			slog.ErrorContext(ctx, "Wallet consumer error", "error", err)
		}
	}()

	slog.InfoContext(ctx, "Wallet consumer started successfully")
}
