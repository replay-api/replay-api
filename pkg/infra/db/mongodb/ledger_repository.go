package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ledgerEntriesCollection = "ledger_entries"
)

// LedgerRepository implements MongoDB persistence for ledger entries
type LedgerRepository struct {
	db *mongo.Database
}

// NewLedgerRepository creates a new MongoDB ledger repository
func NewLedgerRepository(db *mongo.Database) wallet_out.LedgerRepository {
	repo := &LedgerRepository{db: db}
	repo.ensureIndexes()
	return repo
}

// ensureIndexes creates required indexes for performance and uniqueness
func (r *LedgerRepository) ensureIndexes() {
	ctx := context.Background()
	collection := r.db.Collection(ledgerEntriesCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "idempotency_key", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "account_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "transaction_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "account_id", Value: 1},
				{Key: "currency", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "created_at", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "metadata.payment_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "metadata.approval_status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "created_by", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("failed to create ledger indexes", "error", err)
	} else {
		slog.Info("ledger repository indexes created successfully")
	}
}

// CreateTransaction atomically creates all entries in a transaction
func (r *LedgerRepository) CreateTransaction(ctx context.Context, entries []*wallet_entities.LedgerEntry) error {
	if len(entries) == 0 {
		return fmt.Errorf("no entries to create")
	}

	collection := r.db.Collection(ledgerEntriesCollection)

	// Validate all entries before insertion
	for _, entry := range entries {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("entry validation failed: %w", err)
		}
	}

	// Convert to BSON documents
	docs := make([]interface{}, len(entries))
	for i, entry := range entries {
		docs[i] = entry
	}

	// Use MongoDB session for transaction atomicity
	session, err := r.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// Execute in transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, err := collection.InsertMany(sessCtx, docs)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	slog.InfoContext(ctx, "ledger transaction created",
		"transaction_id", entries[0].TransactionID,
		"entry_count", len(entries))

	return nil
}

// CreateEntry creates a single ledger entry
func (r *LedgerRepository) CreateEntry(ctx context.Context, entry *wallet_entities.LedgerEntry) error {
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("entry validation failed: %w", err)
	}

	collection := r.db.Collection(ledgerEntriesCollection)
	_, err := collection.InsertOne(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to create entry: %w", err)
	}

	return nil
}

// FindByID retrieves a ledger entry by ID
func (r *LedgerRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.LedgerEntry, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{"_id": id}
	var entry wallet_entities.LedgerEntry

	err := collection.FindOne(ctx, filter).Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("ledger entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find entry: %w", err)
	}

	return &entry, nil
}

// FindByTransactionID retrieves all entries for a transaction
func (r *LedgerRepository) FindByTransactionID(ctx context.Context, txID uuid.UUID) ([]*wallet_entities.LedgerEntry, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{"transaction_id": txID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*wallet_entities.LedgerEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// FindByAccountID retrieves all entries for an account (wallet)
func (r *LedgerRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID, limit int, offset int) ([]*wallet_entities.LedgerEntry, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{"account_id": accountID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*wallet_entities.LedgerEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// FindByAccountAndCurrency retrieves entries for specific currency
func (r *LedgerRepository) FindByAccountAndCurrency(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) ([]*wallet_entities.LedgerEntry, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{
		"account_id": accountID,
		"currency":   currency,
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*wallet_entities.LedgerEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// FindByIdempotencyKey checks if an entry with this key exists
func (r *LedgerRepository) FindByIdempotencyKey(ctx context.Context, key string) (*wallet_entities.LedgerEntry, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{"idempotency_key": key}
	var entry wallet_entities.LedgerEntry

	err := collection.FindOne(ctx, filter).Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find entry: %w", err)
	}

	return &entry, nil
}

// ExistsByIdempotencyKey checks if an idempotency key has been used
func (r *LedgerRepository) ExistsByIdempotencyKey(ctx context.Context, key string) bool {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{"idempotency_key": key}
	count, err := collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false
	}

	return count > 0
}

// FindByDateRange retrieves entries within a date range
func (r *LedgerRepository) FindByDateRange(ctx context.Context, accountID uuid.UUID, from time.Time, to time.Time) ([]*wallet_entities.LedgerEntry, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{
		"account_id": accountID,
		"created_at": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*wallet_entities.LedgerEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// CalculateBalance calculates the current balance for an account
// For asset accounts: Balance = SUM(debits) - SUM(credits)
func (r *LedgerRepository) CalculateBalance(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) (wallet_vo.Amount, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	// Aggregate: sum debits and credits separately
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"account_id": accountID,
			"currency":   currency,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": "$entry_type",
			"total": bson.M{
				"$sum": "$amount.cents",
			},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return wallet_vo.NewAmount(0), fmt.Errorf("failed to aggregate balance: %w", err)
	}
	defer cursor.Close(ctx)

	var debitTotal int64 = 0
	var creditTotal int64 = 0

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Total int64  `bson:"total"`
		}
		if err := cursor.Decode(&result); err != nil {
			return wallet_vo.NewAmount(0), err
		}

		if result.ID == string(wallet_entities.EntryTypeDebit) {
			debitTotal = result.Total
		} else {
			creditTotal = result.Total
		}
	}

	// For asset accounts: Balance = Debits - Credits
	balanceCents := debitTotal - creditTotal
	return wallet_vo.NewAmountFromCents(balanceCents), nil
}

// GetAccountHistory retrieves transaction history with pagination and filters
func (r *LedgerRepository) GetAccountHistory(ctx context.Context, accountID uuid.UUID, filters wallet_out.HistoryFilters) ([]*wallet_entities.LedgerEntry, int64, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	// Build filter
	filter := bson.M{"account_id": accountID}

	if filters.Currency != nil {
		filter["currency"] = *filters.Currency
	}
	if filters.AssetType != nil {
		filter["asset_type"] = *filters.AssetType
	}
	if filters.EntryType != nil {
		filter["entry_type"] = *filters.EntryType
	}
	if filters.OperationType != nil {
		filter["metadata.operation_type"] = *filters.OperationType
	}
	if filters.FromDate != nil || filters.ToDate != nil {
		dateFilter := bson.M{}
		if filters.FromDate != nil {
			dateFilter["$gte"] = *filters.FromDate
		}
		if filters.ToDate != nil {
			dateFilter["$lte"] = *filters.ToDate
		}
		filter["created_at"] = dateFilter
	}
	if filters.MinAmount != nil || filters.MaxAmount != nil {
		amountFilter := bson.M{}
		if filters.MinAmount != nil {
			amountFilter["$gte"] = filters.MinAmount.ToCents()
		}
		if filters.MaxAmount != nil {
			amountFilter["$lte"] = filters.MaxAmount.ToCents()
		}
		filter["amount.cents"] = amountFilter
	}

	// Get total count
	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Build options
	sortOrder := -1 // desc by default
	if filters.SortOrder == "asc" {
		sortOrder = 1
	}

	sortField := "created_at"
	if filters.SortBy != "" {
		sortField = filters.SortBy
	}

	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetLimit(int64(filters.Limit)).
		SetSkip(int64(filters.Offset))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*wallet_entities.LedgerEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, 0, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, totalCount, nil
}

// FindPendingApprovals retrieves entries pending manual review
func (r *LedgerRepository) FindPendingApprovals(ctx context.Context, limit int) ([]*wallet_entities.LedgerEntry, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{
		"metadata.approval_status": wallet_entities.ApprovalStatusPendingReview,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*wallet_entities.LedgerEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// UpdateApprovalStatus updates the approval status
func (r *LedgerRepository) UpdateApprovalStatus(ctx context.Context, entryID uuid.UUID, status wallet_entities.ApprovalStatus, approverID uuid.UUID) error {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{"_id": entryID}
	update := bson.M{
		"$set": bson.M{
			"metadata.approval_status": status,
			"metadata.approver_id":     approverID,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update approval status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("entry not found")
	}

	return nil
}

// MarkAsReversed marks an entry as reversed
func (r *LedgerRepository) MarkAsReversed(ctx context.Context, entryID uuid.UUID, reversalEntryID uuid.UUID) error {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{"_id": entryID}
	update := bson.M{
		"$set": bson.M{
			"is_reversed": true,
			"reversed_by": reversalEntryID,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark as reversed: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("entry not found")
	}

	return nil
}

// GetDailyTransactionCount gets count of transactions for an account in last 24h
func (r *LedgerRepository) GetDailyTransactionCount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{
		"account_id": accountID,
		"created_at": bson.M{
			"$gte": time.Now().Add(-24 * time.Hour),
		},
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return count, nil
}

// GetDailyTransactionVolume gets total volume for an account in last 24h
func (r *LedgerRepository) GetDailyTransactionVolume(ctx context.Context, accountID uuid.UUID, currency wallet_vo.Currency) (wallet_vo.Amount, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"account_id": accountID,
			"currency":   currency,
			"created_at": bson.M{
				"$gte": time.Now().Add(-24 * time.Hour),
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"total": bson.M{
				"$sum": "$amount.cents",
			},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return wallet_vo.NewAmount(0), fmt.Errorf("failed to aggregate volume: %w", err)
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			Total int64 `bson:"total"`
		}
		if err := cursor.Decode(&result); err != nil {
			return wallet_vo.NewAmount(0), err
		}
		return wallet_vo.NewAmountFromCents(result.Total), nil
	}

	return wallet_vo.NewAmount(0), nil
}

// FindByUserAndDateRange for tax reporting
func (r *LedgerRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, from time.Time, to time.Time) ([]*wallet_entities.LedgerEntry, error) {
	collection := r.db.Collection(ledgerEntriesCollection)

	filter := bson.M{
		"created_by": userID,
		"created_at": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*wallet_entities.LedgerEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}
