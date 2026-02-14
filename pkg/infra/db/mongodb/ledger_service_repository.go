package db

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_services "github.com/replay-api/replay-api/pkg/domain/wallet/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ledgerAccountsCollection = "ledger_accounts"
	ledgerJournalsCollection = "ledger_journals"
	ledgerWalletsCollection  = "ledger_wallets"
)

// LedgerServiceRepository implements wallet_services.LedgerRepository
// This is the internal repository interface used by LedgerService for double-entry accounting.
// It manages LedgerAccounts, JournalEntries, and LedgerWallets (NOT LedgerEntry items).
type LedgerServiceRepository struct {
	db     *mongo.Database
	client *mongo.Client
}

// NewLedgerServiceRepository creates a new MongoDB-backed ledger service repository
func NewLedgerServiceRepository(client *mongo.Client, dbName string) wallet_services.LedgerRepository {
	repo := &LedgerServiceRepository{
		db:     client.Database(dbName),
		client: client,
	}
	repo.ensureIndexes()
	return repo
}

func (r *LedgerServiceRepository) ensureIndexes() {
	ctx := context.Background()

	// Accounts indexes
	accountsColl := r.db.Collection(ledgerAccountsCollection)
	accountIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "currency", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
	}
	if _, err := accountsColl.Indexes().CreateMany(ctx, accountIndexes); err != nil {
		slog.Error("failed to create ledger account indexes", "error", err)
	}

	// Journals indexes
	journalsColl := r.db.Collection(ledgerJournalsCollection)
	journalIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "reference", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "transaction_type", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "created_by", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}
	if _, err := journalsColl.Indexes().CreateMany(ctx, journalIndexes); err != nil {
		slog.Error("failed to create ledger journal indexes", "error", err)
	}

	// Wallets indexes
	walletsColl := r.db.Collection(ledgerWalletsCollection)
	walletIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "currency", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "ledger_account_id", Value: 1}},
		},
	}
	if _, err := walletsColl.Indexes().CreateMany(ctx, walletIndexes); err != nil {
		slog.Error("failed to create ledger wallet indexes", "error", err)
	}

	slog.Info("ledger service repository indexes created successfully")
}

// --- Accounts ---

// CreateAccount creates a new ledger account
func (r *LedgerServiceRepository) CreateAccount(ctx context.Context, account *wallet_entities.LedgerAccount) error {
	collection := r.db.Collection(ledgerAccountsCollection)

	_, err := collection.InsertOne(ctx, account)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			slog.DebugContext(ctx, "Ledger account already exists (idempotent)", "code", account.Code)
			return nil // Idempotent — account already exists
		}
		return fmt.Errorf("failed to create ledger account: %w", err)
	}

	slog.InfoContext(ctx, "Ledger account created",
		"account_id", account.ID,
		"code", account.Code,
		"name", account.Name,
		"type", account.Type,
	)

	return nil
}

// GetAccountByID retrieves a ledger account by ID
func (r *LedgerServiceRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*wallet_entities.LedgerAccount, error) {
	collection := r.db.Collection(ledgerAccountsCollection)

	filter := bson.M{"_id": id}
	var account wallet_entities.LedgerAccount

	err := collection.FindOne(ctx, filter).Decode(&account)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("ledger account not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ledger account: %w", err)
	}

	return &account, nil
}

// GetAccountByCode retrieves a ledger account by code (e.g., "1001")
func (r *LedgerServiceRepository) GetAccountByCode(ctx context.Context, code string) (*wallet_entities.LedgerAccount, error) {
	collection := r.db.Collection(ledgerAccountsCollection)

	filter := bson.M{"code": code}
	var account wallet_entities.LedgerAccount

	err := collection.FindOne(ctx, filter).Decode(&account)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("ledger account not found for code: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ledger account by code: %w", err)
	}

	return &account, nil
}

// GetAccountByUserID retrieves a user-specific ledger account
func (r *LedgerServiceRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID, currency string) (*wallet_entities.LedgerAccount, error) {
	collection := r.db.Collection(ledgerAccountsCollection)

	filter := bson.M{
		"user_id":  userID,
		"currency": currency,
	}
	var account wallet_entities.LedgerAccount

	err := collection.FindOne(ctx, filter).Decode(&account)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("ledger account not found for user %s with currency %s", userID, currency)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ledger account by user: %w", err)
	}

	return &account, nil
}

// UpdateAccountBalance updates balance with optimistic locking via version
func (r *LedgerServiceRepository) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, balance, available, held *big.Float, version int) error {
	collection := r.db.Collection(ledgerAccountsCollection)

	// Optimistic locking: only update if version matches
	filter := bson.M{
		"_id":     accountID,
		"version": version,
	}

	update := bson.M{
		"$set": bson.M{
			"balance":           balance,
			"available_balance": available,
			"held_balance":      held,
			"updated_at":        time.Now().UTC(),
		},
		"$inc": bson.M{
			"version": 1,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("optimistic locking conflict: account %s version %d not found (concurrent modification)", accountID, version)
	}

	return nil
}

// --- Journals ---

// CreateJournal creates a new journal entry (atomically with all its ledger entries)
func (r *LedgerServiceRepository) CreateJournal(ctx context.Context, journal *wallet_entities.JournalEntry) error {
	collection := r.db.Collection(ledgerJournalsCollection)

	_, err := collection.InsertOne(ctx, journal)
	if err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	slog.InfoContext(ctx, "Journal entry created",
		"journal_id", journal.ID,
		"type", journal.TransactionType,
		"reference", journal.Reference,
		"status", journal.Status,
		"total_debit", journal.TotalDebit.Text('f', 2),
		"total_credit", journal.TotalCredit.Text('f', 2),
	)

	return nil
}

// GetJournalByID retrieves a journal entry by ID
func (r *LedgerServiceRepository) GetJournalByID(ctx context.Context, id uuid.UUID) (*wallet_entities.JournalEntry, error) {
	collection := r.db.Collection(ledgerJournalsCollection)

	filter := bson.M{"_id": id}
	var journal wallet_entities.JournalEntry

	err := collection.FindOne(ctx, filter).Decode(&journal)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("journal entry not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find journal entry: %w", err)
	}

	return &journal, nil
}

// GetLastJournalHash retrieves the hash of the most recent journal for hash-chain integrity
func (r *LedgerServiceRepository) GetLastJournalHash(ctx context.Context) (string, error) {
	collection := r.db.Collection(ledgerJournalsCollection)

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var journal wallet_entities.JournalEntry

	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&journal)
	if err == mongo.ErrNoDocuments {
		return "genesis", nil // Genesis block — first entry in the chain
	}
	if err != nil {
		return "", fmt.Errorf("failed to get last journal hash: %w", err)
	}

	return journal.Hash, nil
}

// UpdateJournalStatus updates the status of a journal entry
func (r *LedgerServiceRepository) UpdateJournalStatus(ctx context.Context, id uuid.UUID, status wallet_entities.JournalStatus) error {
	collection := r.db.Collection(ledgerJournalsCollection)

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	// Add posted_at timestamp for POSTED status
	if status == wallet_entities.JournalStatusPosted {
		now := time.Now().UTC()
		update["$set"].(bson.M)["posted_at"] = now
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update journal status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("journal entry not found: %s", id)
	}

	return nil
}

// --- Wallets (LedgerWallet) ---

// CreateWallet creates a new ledger-integrated wallet
func (r *LedgerServiceRepository) CreateWallet(ctx context.Context, wallet *wallet_entities.LedgerWallet) error {
	collection := r.db.Collection(ledgerWalletsCollection)

	_, err := collection.InsertOne(ctx, wallet)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			slog.DebugContext(ctx, "Ledger wallet already exists (idempotent)",
				"user_id", wallet.UserID,
				"currency", wallet.Currency,
			)
			return nil
		}
		return fmt.Errorf("failed to create ledger wallet: %w", err)
	}

	slog.InfoContext(ctx, "Ledger wallet created",
		"wallet_id", wallet.ID,
		"user_id", wallet.UserID,
		"currency", wallet.Currency,
	)

	return nil
}

// GetWalletByUserID retrieves a ledger wallet by user ID and currency
func (r *LedgerServiceRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID, currency string) (*wallet_entities.LedgerWallet, error) {
	collection := r.db.Collection(ledgerWalletsCollection)

	filter := bson.M{
		"user_id":  userID,
		"currency": currency,
	}
	var wallet wallet_entities.LedgerWallet

	err := collection.FindOne(ctx, filter).Decode(&wallet)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("ledger wallet not found for user %s with currency %s", userID, currency)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find ledger wallet: %w", err)
	}

	return &wallet, nil
}

// UpdateWallet updates a ledger wallet with optimistic locking
func (r *LedgerServiceRepository) UpdateWallet(ctx context.Context, wallet *wallet_entities.LedgerWallet) error {
	collection := r.db.Collection(ledgerWalletsCollection)

	filter := bson.M{
		"_id":     wallet.ID,
		"version": wallet.Version,
	}

	wallet.UpdatedAt = time.Now().UTC()
	newVersion := wallet.Version + 1

	update := bson.M{
		"$set": bson.M{
			"balance":           wallet.Balance,
			"available_balance": wallet.AvailableBalance,
			"held_balance":      wallet.HeldBalance,
			"updated_at":        wallet.UpdatedAt,
			"version":           newVersion,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update ledger wallet: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("optimistic locking conflict: wallet %s version %d (concurrent modification)", wallet.ID, wallet.Version)
	}

	wallet.Version = newVersion
	return nil
}

// --- Reporting ---

// GetJournalsByDateRange retrieves journals within a date range
func (r *LedgerServiceRepository) GetJournalsByDateRange(ctx context.Context, from, to time.Time) ([]wallet_entities.JournalEntry, error) {
	collection := r.db.Collection(ledgerJournalsCollection)

	filter := bson.M{
		"created_at": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find journals: %w", err)
	}
	defer cursor.Close(ctx)

	var journals []wallet_entities.JournalEntry
	if err := cursor.All(ctx, &journals); err != nil {
		return nil, fmt.Errorf("failed to decode journals: %w", err)
	}

	return journals, nil
}

// GetAccountBalances retrieves all active account balances
func (r *LedgerServiceRepository) GetAccountBalances(ctx context.Context) ([]wallet_entities.LedgerAccount, error) {
	collection := r.db.Collection(ledgerAccountsCollection)

	filter := bson.M{"is_active": true}
	opts := options.Find().SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts: %w", err)
	}
	defer cursor.Close(ctx)

	var accounts []wallet_entities.LedgerAccount
	if err := cursor.All(ctx, &accounts); err != nil {
		return nil, fmt.Errorf("failed to decode accounts: %w", err)
	}

	return accounts, nil
}

// Compile-time interface compliance check
var _ wallet_services.LedgerRepository = (*LedgerServiceRepository)(nil)
