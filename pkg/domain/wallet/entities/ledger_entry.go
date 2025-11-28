package wallet_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// LedgerEntry represents an immutable double-entry accounting record
// Every transaction creates TWO entries: one debit and one credit
// This ensures balance integrity and provides complete audit trail
type LedgerEntry struct {
	ID             uuid.UUID          `bson:"_id" json:"id"`
	TransactionID  uuid.UUID          `bson:"transaction_id" json:"transaction_id"` // Groups double-entries together
	AccountID      uuid.UUID          `bson:"account_id" json:"account_id"`         // Wallet ID or system account ID
	AccountType    AccountType        `bson:"account_type" json:"account_type"`
	EntryType      EntryType          `bson:"entry_type" json:"entry_type"` // Debit or Credit
	AssetType      AssetType          `bson:"asset_type" json:"asset_type"` // Fiat, Crypto, NFT, GameCredit
	Currency       *wallet_vo.Currency `bson:"currency,omitempty" json:"currency,omitempty"`
	Amount         wallet_vo.Amount   `bson:"amount" json:"amount"`
	BalanceAfter   wallet_vo.Amount   `bson:"balance_after" json:"balance_after"` // Balance snapshot after this entry
	NFTAssetID     *uuid.UUID         `bson:"nft_asset_id,omitempty" json:"nft_asset_id,omitempty"`
	GameCredits    *int64             `bson:"game_credits,omitempty" json:"game_credits,omitempty"`
	Description    string             `bson:"description" json:"description"`
	Metadata       LedgerMetadata     `bson:"metadata" json:"metadata"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	CreatedBy      uuid.UUID          `bson:"created_by" json:"created_by"`
	IdempotencyKey string             `bson:"idempotency_key" json:"idempotency_key"` // Prevents duplicate transactions
	ParentEntryID  *uuid.UUID         `bson:"parent_entry_id,omitempty" json:"parent_entry_id,omitempty"`
	IsReversed     bool               `bson:"is_reversed" json:"is_reversed"`
	ReversedBy     *uuid.UUID         `bson:"reversed_by,omitempty" json:"reversed_by,omitempty"`
}

// EntryType defines whether this is a debit or credit entry
type EntryType string

const (
	EntryTypeDebit  EntryType = "Debit"  // Increases assets, decreases liabilities
	EntryTypeCredit EntryType = "Credit" // Decreases assets, increases liabilities
)

// AccountType categorizes the account for accounting purposes
type AccountType string

const (
	AccountTypeAsset     AccountType = "Asset"     // User wallet (debit increases balance)
	AccountTypeLiability AccountType = "Liability" // Platform owes user
	AccountTypeRevenue   AccountType = "Revenue"   // Platform earnings
	AccountTypeExpense   AccountType = "Expense"   // Platform costs
)

// AssetType defines the type of asset being transacted
type AssetType string

const (
	AssetTypeFiat       AssetType = "Fiat"
	AssetTypeCrypto     AssetType = "Crypto"
	AssetTypeNFT        AssetType = "NFT"
	AssetTypeGameCredit AssetType = "GameCredit"
)

// LedgerMetadata stores additional context about the transaction
type LedgerMetadata struct {
	OperationType  string                 `bson:"operation_type" json:"operation_type"` // "Deposit", "Withdrawal", "Transfer", "Prize"
	TxHash         string                 `bson:"tx_hash,omitempty" json:"tx_hash,omitempty"` // Blockchain transaction hash
	PaymentID      *uuid.UUID             `bson:"payment_id,omitempty" json:"payment_id,omitempty"`
	MatchID        *uuid.UUID             `bson:"match_id,omitempty" json:"match_id,omitempty"`
	TournamentID   *uuid.UUID             `bson:"tournament_id,omitempty" json:"tournament_id,omitempty"`
	SourceIP       string                 `bson:"source_ip,omitempty" json:"source_ip,omitempty"`
	UserAgent      string                 `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	GeolocationCtry string                `bson:"geolocation_country,omitempty" json:"geolocation_country,omitempty"`
	RiskScore      float64                `bson:"risk_score" json:"risk_score"` // 0.0 - 1.0
	ApprovalStatus ApprovalStatus         `bson:"approval_status" json:"approval_status"`
	ApproverID     *uuid.UUID             `bson:"approver_id,omitempty" json:"approver_id,omitempty"`
	Notes          string                 `bson:"notes,omitempty" json:"notes,omitempty"`
	CustomData     map[string]interface{} `bson:"custom_data,omitempty" json:"custom_data,omitempty"`
}

// ApprovalStatus indicates whether the transaction was auto-approved or requires review
type ApprovalStatus string

const (
	ApprovalStatusAutoApproved     ApprovalStatus = "AutoApproved"
	ApprovalStatusPendingReview    ApprovalStatus = "PendingReview"
	ApprovalStatusManuallyApproved ApprovalStatus = "ManuallyApproved"
	ApprovalStatusRejected         ApprovalStatus = "Rejected"
)

// NewLedgerEntry creates a new ledger entry with validation
func NewLedgerEntry(
	txID uuid.UUID,
	accountID uuid.UUID,
	accountType AccountType,
	entryType EntryType,
	assetType AssetType,
	amount wallet_vo.Amount,
	description string,
	idempotencyKey string,
	createdBy uuid.UUID,
) *LedgerEntry {
	now := time.Now().UTC()

	entry := &LedgerEntry{
		ID:             uuid.New(),
		TransactionID:  txID,
		AccountID:      accountID,
		AccountType:    accountType,
		EntryType:      entryType,
		AssetType:      assetType,
		Amount:         amount,
		Description:    description,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		CreatedBy:      createdBy,
		IsReversed:     false,
		Metadata: LedgerMetadata{
			RiskScore:      0.0,
			ApprovalStatus: ApprovalStatusAutoApproved,
		},
	}

	return entry
}

// Reverse creates a reversal entry for this ledger entry
// Used for refunds, cancellations, corrections
func (l *LedgerEntry) Reverse(reversalReason string, reversedBy uuid.UUID) *LedgerEntry {
	reversalEntry := NewLedgerEntry(
		uuid.New(), // New transaction ID for reversal
		l.AccountID,
		l.AccountType,
		l.oppositeEntryType(),
		l.AssetType,
		l.Amount,
		"Reversal: "+l.Description+" - "+reversalReason,
		l.IdempotencyKey+"_reversal_"+uuid.New().String(),
		reversedBy,
	)

	reversalEntry.ParentEntryID = &l.ID
	reversalEntry.Currency = l.Currency
	reversalEntry.NFTAssetID = l.NFTAssetID
	reversalEntry.GameCredits = l.GameCredits

	// Mark original as reversed
	l.IsReversed = true
	reversalID := reversalEntry.ID
	l.ReversedBy = &reversalID

	return reversalEntry
}

// oppositeEntryType returns the opposite entry type (for reversals)
func (l *LedgerEntry) oppositeEntryType() EntryType {
	if l.EntryType == EntryTypeDebit {
		return EntryTypeCredit
	}
	return EntryTypeDebit
}

// Validate ensures the ledger entry is valid
func (l *LedgerEntry) Validate() error {
	if l.TransactionID == uuid.Nil {
		return common.NewErrBadRequest("transaction_id is required")
	}

	if l.AccountID == uuid.Nil {
		return common.NewErrBadRequest("account_id is required")
	}

	if l.IdempotencyKey == "" {
		return common.NewErrBadRequest("idempotency_key is required")
	}

	if l.Amount.IsZero() {
		return common.NewErrBadRequest("amount must be greater than zero")
	}

	if l.Description == "" {
		return common.NewErrBadRequest("description is required")
	}

	// Validate asset-specific fields
	switch l.AssetType {
	case AssetTypeFiat, AssetTypeCrypto:
		if l.Currency == nil {
			return common.NewErrBadRequest("currency is required for fiat/crypto entries")
		}
	case AssetTypeNFT:
		if l.NFTAssetID == nil {
			return common.NewErrBadRequest("nft_asset_id is required for NFT entries")
		}
	case AssetTypeGameCredit:
		if l.GameCredits == nil {
			return common.NewErrBadRequest("game_credits is required for credit entries")
		}
	}

	return nil
}

// GetResourceOwner returns the resource owner for authorization
func (l *LedgerEntry) GetResourceOwner() common.ResourceOwner {
	// Ledger entries are owned by the account holder
	return common.ResourceOwner{
		UserID: l.CreatedBy,
	}
}

// SetBalanceAfter sets the balance snapshot after this entry
func (l *LedgerEntry) SetBalanceAfter(balance wallet_vo.Amount) {
	l.BalanceAfter = balance
}

// SetCurrency sets the currency for fiat/crypto entries
func (l *LedgerEntry) SetCurrency(currency wallet_vo.Currency) {
	l.Currency = &currency
}

// SetNFTAssetID sets the NFT asset ID for NFT entries
func (l *LedgerEntry) SetNFTAssetID(nftID uuid.UUID) {
	l.NFTAssetID = &nftID
}

// SetGameCredits sets the credit amount for game credit entries
func (l *LedgerEntry) SetGameCredits(credits int64) {
	l.GameCredits = &credits
}

// SetMetadata updates the metadata for this entry
func (l *LedgerEntry) SetMetadata(metadata LedgerMetadata) {
	l.Metadata = metadata
}

// IsDebit returns true if this is a debit entry
func (l *LedgerEntry) IsDebit() bool {
	return l.EntryType == EntryTypeDebit
}

// IsCredit returns true if this is a credit entry
func (l *LedgerEntry) IsCredit() bool {
	return l.EntryType == EntryTypeCredit
}
