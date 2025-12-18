# Production-Grade Multi-Asset Wallet System

## Overview

This document describes the financial-grade wallet infrastructure for the LeetGaming platform. The system implements double-entry accounting with immutable ledger, supporting multiple asset types (Fiat, Crypto, NFTs, Game Credits) with complete audit compliance.

## Architecture

### Core Principles

1. **Financial Integrity First**: Every transaction creates immutable ledger entries
2. **Double-Entry Accounting**: Accounting equation always balances (Assets = Liabilities + Equity)
3. **Atomic Operations**: Saga pattern with automatic rollback on failure
4. **Zero Mocks**: E2E tests use real MongoDB and Hardhat EVM
5. **Production-Ready**: Ready for regulatory compliance (SOX, PCI-DSS)

### System Components

```
┌─────────────────────────────────────────────────────────┐
│                    Wallet Service                        │
│  (Orchestrates wallet operations with transaction       │
│   coordinator for atomic execution)                     │
└──────────────────┬──────────────────────────────────────┘
                   │
    ┌──────────────┴──────────────┐
    │                             │
┌───▼──────────────┐    ┌────────▼─────────────┐
│   Transaction    │    │   Reconciliation     │
│   Coordinator    │    │      Service         │
│  (Saga Pattern)  │    │  (Balance Verify)    │
└───┬──────────────┘    └──────────────────────┘
    │
    │  Atomic Execution
    │
┌───▼──────────────────────────────────────────────┐
│           Ledger Service                         │
│  (Double-Entry Accounting Logic)                 │
│  • RecordDeposit                                 │
│  • RecordWithdrawal                              │
│  • RecordEntryFee                                │
│  • RecordPrizeWinning                            │
│  • RecordRefund                                  │
└───┬──────────────────────────────────────────────┘
    │
    │  Persistence
    │
┌───▼──────────────┐    ┌──────────────────┐
│  Ledger          │    │  Idempotency     │
│  Repository      │    │  Repository      │
│  (MongoDB)       │    │  (MongoDB TTL)   │
└──────────────────┘    └──────────────────┘
```

## Double-Entry Accounting

### Transaction Types

#### 1. Deposit (User receives money)
```
Entry 1: DEBIT  User's Asset Account    (+$100)
Entry 2: CREDIT Platform Liability      (+$100)

Effect: User has more money, platform owes user more
```

#### 2. Withdrawal (User withdraws money)
```
Entry 1: CREDIT User's Asset Account    (-$50)
Entry 2: DEBIT  Platform Liability      (-$50)

Effect: User has less money, platform owes user less
```

#### 3. Entry Fee (User pays to join match/tournament)
```
Entry 1: CREDIT User's Asset Account    (-$10)
Entry 2: DEBIT  Platform Revenue        (+$10)

Effect: User pays fee, platform earns revenue
```

#### 4. Prize Winning (User wins prize)
```
Entry 1: DEBIT  User's Asset Account    (+$50)
Entry 2: CREDIT Platform Expense        (+$50)

Effect: User receives prize, platform incurs expense
```

#### 5. Refund (Reverse original transaction)
```
Creates opposite entries of the original transaction
Marks original entries as reversed
```

### System Accounts

```go
SystemLiabilityAccountID = "00000000-0000-0000-0000-000000000001"  // Platform owes users
SystemRevenueAccountID   = "00000000-0000-0000-0000-000000000002"  // Platform earnings
SystemExpenseAccountID   = "00000000-0000-0000-0000-000000000003"  // Platform costs
```

## Transaction Coordinator (Saga Pattern)

### Atomic Execution with Automatic Rollback

```go
coordinator.ExecuteDeposit(ctx, wallet, currency, amount, paymentID, metadata)

// Saga steps:
// 1. Record in ledger
//    → Rollback: Reverse ledger entry
// 2. Update wallet balance
//    → Rollback: Reverse wallet balance
// 3. Persist wallet to database
//    → Rollback: N/A (DB already rolled back)
```

### Rollback Guarantees

- **Automatic**: Failures trigger immediate rollback
- **Complete**: All executed steps reversed in reverse order
- **Logged**: Critical failures logged for manual intervention
- **Idempotent**: Rollback operations are safe to retry

### Example Rollback Scenario

```go
// Deposit attempt: Ledger succeeds, wallet update fails
ledgerTxID, err := coordinator.ExecuteDeposit(...)

// System automatically:
// 1. Detects wallet update failure
// 2. Calls ledgerService.RecordRefund(ledgerTxID, "Automatic rollback")
// 3. Reverses wallet balance change
// 4. Returns error to caller
//
// Result: NO money created or lost, complete data integrity maintained
```

## Idempotency Protection

### Implementation

Every transaction has a unique idempotency key:
```go
idempotencyKey := fmt.Sprintf("deposit_%s_%s", paymentID.String(), walletID.String())
```

### TTL Auto-Cleanup

```go
type IdempotentOperation struct {
    Key       string              // Primary key
    Status    OperationStatus     // Processing, Completed, Failed
    ExpiresAt time.Time          // Auto-cleanup after 24 hours
    ResultID  *uuid.UUID         // Transaction ID for completed operations
}
```

MongoDB TTL index automatically deletes expired operations after 24 hours.

### Duplicate Detection

```go
// First request
deposit("payment_123", $100) → Success, TxID: abc-def

// Duplicate request (same payment ID)
deposit("payment_123", $100) → Returns existing TxID: abc-def
// Balance unchanged, no duplicate money created
```

## Reconciliation Service

### Daily Balance Verification

```go
result := reconciliationService.ReconcileWallet(ctx, walletID)

// Compares:
// - Wallet.Balances[USD] = $150.00
// - LedgerBalance = SUM(debits) - SUM(credits) = $150.00
//
// Status: ReconciliationStatusMatched
```

### Discrepancy Detection

```go
type BalanceDiscrepancy struct {
    Currency      Currency
    WalletBalance Amount       // $150.00
    LedgerBalance Amount       // $149.50  ← DISCREPANCY!
    Difference    Amount       // $0.50
    Severity      DiscrepancySeverity  // Low, Medium, High, Critical
}

// Severities:
// Low:      < $1
// Medium:   $1 - $100
// High:     $100 - $1000
// Critical: > $1000 (requires manual review)
```

### Auto-Correction

```go
// For low-severity discrepancies, automatically correct wallet to match ledger
reconciliationService.AutoCorrectWallet(ctx, walletID, approverID)

// IMPORTANT: Ledger is ALWAYS the source of truth
```

## MongoDB Schema

### Ledger Entries Collection

```javascript
{
  _id: UUID,
  transaction_id: UUID,         // Groups double-entries
  account_id: UUID,              // Wallet ID or system account
  account_type: "Asset"|"Liability"|"Revenue"|"Expense",
  entry_type: "Debit"|"Credit",
  asset_type: "Fiat"|"Crypto"|"NFT"|"GameCredit",
  currency: "USD"|"USDC"|"USDT",
  amount: { cents: 10000 },      // $100.00
  balance_after: { cents: 15000 },
  description: "Deposit via payment abc-123",
  idempotency_key: "deposit_abc-123_wallet-456",
  created_at: ISODate(),
  created_by: UUID,
  is_reversed: false,
  metadata: {
    operation_type: "Deposit",
    payment_id: UUID,
    source_ip: "192.168.1.1",
    risk_score: 0.05,
    approval_status: "AutoApproved"
  }
}
```

### Critical Indexes

```javascript
// Prevent duplicate transactions
db.ledger_entries.createIndex({ idempotency_key: 1 }, { unique: true })

// Fast account history queries
db.ledger_entries.createIndex({ account_id: 1, created_at: -1 })

// Transaction lookup
db.ledger_entries.createIndex({ transaction_id: 1 })

// Balance calculation
db.ledger_entries.createIndex({ account_id: 1, currency: 1 })

// Fraud detection
db.ledger_entries.createIndex({ created_at: 1 })
db.ledger_entries.createIndex({ "metadata.source_ip": 1, created_at: -1 })
```

### Idempotency Collection

```javascript
{
  _id: "deposit_abc-123_wallet-456",  // Idempotency key as primary key
  operation_type: "Deposit",
  status: "Completed",
  result_id: UUID,                     // Transaction ID
  created_at: ISODate(),
  expires_at: ISODate(),               // TTL index - auto-delete after 24h
  attempt_count: 1
}

// TTL index for auto-cleanup
db.idempotent_operations.createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 })
```

## Testing Infrastructure

### NO MOCKS - Real Services Only

#### E2E Test Setup

```bash
# Start test infrastructure
make -f Makefile.test test-setup

# Starts:
# - MongoDB on port 27018
# - Hardhat EVM node on port 8545
# - Deploys smart contracts (USDC, USDT, GameNFT)

# Run E2E tests
make -f Makefile.test test-e2e

# Run with coverage
make -f Makefile.test test-ci
```

#### Test Coverage

```go
✓ Deposit with double-entry ledger validation
✓ Withdrawal with balance checks
✓ Reconciliation (wallet matches ledger)
✓ Idempotency (duplicate detection)
✓ Entry fee with insufficient balance rollback
✓ Prize winning with daily limit enforcement
✓ Transaction history with pagination
✓ Ledger balance calculation accuracy
✓ Benchmark: Deposit throughput
```

### Smart Contracts for Testing

#### MockUSDC (ERC-20)
- 6 decimals (matches real USDC)
- Faucet for easy testing
- Owner can mint

#### MockUSDT (ERC-20)
- 6 decimals (matches real USDT)
- Same interface as USDC

#### GameNFT (ERC-721)
- Rarity levels (Common, Rare, Epic, Legendary)
- Metadata URIs
- Batch minting support

## Production Deployment

### Environment Variables

```bash
# MongoDB
MONGO_URI=mongodb://user:pass@mongo:27017/replay_prod?authSource=admin

# Blockchain RPC
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_API_KEY
POLYGON_RPC_URL=https://polygon-mainnet.g.alchemy.com/v2/YOUR_API_KEY

# Contract Addresses
USDC_CONTRACT_ADDRESS=0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
USDT_CONTRACT_ADDRESS=0xdAC17F958D2ee523a2206206994597C13D831ec7
```

### Monitoring

- **Ledger Balance Verification**: Daily reconciliation cron job
- **Discrepancy Alerts**: Slack/PagerDuty for critical discrepancies
- **Transaction Metrics**: Prometheus + Grafana
- **Audit Logs**: All transactions logged with full metadata

### Security

- **Idempotency Keys**: Prevent duplicate transactions
- **Risk Scoring**: Fraud detection on every transaction
- **Approval Workflows**: Manual review for high-risk transactions
- **Encryption**: All sensitive data encrypted at rest
- **Rate Limiting**: Prevent abuse

## Compliance

### SOX (Sarbanes-Oxley)

- ✅ Immutable audit trail (ledger entries never deleted)
- ✅ Complete transaction history with timestamps
- ✅ User attribution (created_by field)
- ✅ Reconciliation reports

### PCI-DSS

- ✅ No credit card data stored (payment IDs only)
- ✅ Encryption at rest and in transit
- ✅ Access controls and authentication
- ✅ Regular security audits

### AML/KYC

- ✅ Transaction metadata (IP, geolocation)
- ✅ Risk scoring
- ✅ Daily transaction limits
- ✅ Suspicious activity alerts

### Tax Reporting

- ✅ Complete transaction history
- ✅ 1099-K generation support
- ✅ Date range queries for tax periods

## Performance

### Benchmarks

```
BenchmarkDeposit-8    500 ops    2847 ns/op    1456 B/op    23 allocs/op
```

### Optimization Strategies

1. **Indexes**: Compound indexes for common query patterns
2. **Connection Pooling**: MongoDB connection pool (min 10, max 100)
3. **Batch Operations**: Bulk insert for ledger entries
4. **Caching**: Redis for frequent balance lookups (TTL 60s)

## Disaster Recovery

### Backup Strategy

- **MongoDB**: Point-in-time backups every 6 hours
- **Retention**: 30 days
- **Restore Time Objective (RTO)**: < 1 hour
- **Recovery Point Objective (RPO)**: < 6 hours

### Data Integrity

- **Ledger**: Immutable, never deleted
- **Checksums**: SHA-256 for ledger entry verification
- **Reconciliation**: Automated daily verification

## Banking-Grade Custody Infrastructure

### Overview

The platform implements professional-grade custodial wallet infrastructure with:
- **MPC (Multi-Party Computation)** for secure key management
- **Account Abstraction (ERC-4337)** for smart contract wallets
- **Solana Program** for native Solana smart wallets
- **Social Recovery** for trustless wallet recovery
- **Fee Sponsorship** via Paymaster for gasless transactions

### Multi-Chain Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                    Wallet Orchestrator                          │
│    (Coordinates MPC signing, multi-chain deployment)            │
└──────────────────────┬─────────────────────────────────────────┘
                       │
       ┌───────────────┼───────────────┐
       │               │               │
┌──────▼───────┐ ┌─────▼─────┐ ┌───────▼──────┐
│ MPC Provider │ │    HSM    │ │   Secure     │
│  (TSS/FROST) │ │ Provider  │ │   Enclave    │
│              │ │           │ │ (AWS Nitro)  │
└──────┬───────┘ └─────┬─────┘ └───────┬──────┘
       │               │               │
       └───────────────┼───────────────┘
                       │
       ┌───────────────┴───────────────┐
       │          Key Shares           │
       └───────────────────────────────┘
                       │
    ┌──────────────────┼──────────────────┐
    │                  │                  │
┌───▼────┐       ┌─────▼─────┐      ┌─────▼─────┐
│ Solana │       │  Polygon  │      │   Base    │
│ (Ed25519│       │ (secp256k1)│      │(secp256k1)│
│ FROST)  │       │  ERC-4337 │      │  ERC-4337 │
└─────────┘       └───────────┘      └───────────┘
```

### MPC Key Management

**Supported Schemes:**
| Scheme | Curve | Use Case |
|--------|-------|----------|
| CMP | secp256k1 | EVM chains (fast ECDSA) |
| GG20 | secp256k1 | EVM chains (standard ECDSA) |
| FROST | Ed25519 | Solana (Schnorr-based) |
| Lindell17 | secp256k1 | 2-party ECDSA |

**Threshold Configurations:**
- Personal Wallets: 2-of-3 (user device + platform + backup)
- Business Wallets: 3-of-5 (multi-approver)
- Treasury: 4-of-7 (governance + cold storage)

**Key Storage Locations:**
- HSM (AWS CloudHSM, Azure HSM, Thales Luna)
- Secure Enclaves (AWS Nitro, Azure SGX)
- Cloud KMS (wrapped keys)
- User Devices (for recovery)
- Cold Storage (air-gapped)

### Smart Wallet Contracts

#### Solana Smart Wallet (Rust/Anchor)
Location: `programs/solana-wallet/src/lib.rs`

```rust
// PDA-derived deterministic addresses
pub fn initialize_wallet(
    ctx: Context<InitializeWallet>,
    wallet_id: [u8; 32],
    guardian_threshold: u8,
    daily_limit: u64,
    recovery_delay: i64,
) -> Result<()>

// SPL token transfers with spending limits
pub fn transfer_spl(ctx: Context<TransferSPL>, amount: u64) -> Result<()>

// Social recovery flow
pub fn initiate_recovery(ctx: Context<InitiateRecovery>, new_authority: Pubkey) -> Result<()>
pub fn approve_recovery(ctx: Context<ApproveRecovery>) -> Result<()>
pub fn execute_recovery(ctx: Context<ExecuteRecovery>) -> Result<()>
```

#### ERC-4337 Smart Wallet (Solidity)
Location: `test/blockchain/contracts/aa/LeetSmartWallet.sol`

Features:
- ERC-4337 compliant (validateUserOp)
- Social recovery with time-locked execution
- Session keys for delegated signing
- Daily spending limits
- ERC-1271 signature validation
- UUPS upgradeable

```solidity
// ERC-4337 validation
function validateUserOp(
    PackedUserOperation calldata userOp,
    bytes32 userOpHash,
    uint256 missingAccountFunds
) external onlyEntryPoint returns (uint256 validationData)

// Batch execution
function executeBatch(
    address[] calldata targets,
    uint256[] calldata values,
    bytes[] calldata datas
) external onlyOwnerOrEntryPoint notFrozen
```

#### Paymaster (Gas Sponsorship)
Location: `test/blockchain/contracts/paymaster/LeetPaymaster.sol`

Payment Modes:
1. **Sponsored**: Platform-sponsored (whitelisted wallets)
2. **GasCredits**: Pre-purchased gas credits
3. **TokenPayment**: Pay in USDC/USDT
4. **VerifiedFree**: Free for verified users (platform-signed)

### Social Recovery System

```
┌──────────────────────────────────────────────────────────────┐
│                    Recovery Flow                              │
└──────────────────────────────────────────────────────────────┘

Guardian                    Smart Wallet                Platform
   │                             │                          │
   │  1. initiateRecovery()      │                          │
   ├────────────────────────────►│                          │
   │                             │  Freeze wallet           │
   │                             │  Start delay timer       │
   │                             │                          │
   │  2. Other guardians         │                          │
   │     approveRecovery()       │                          │
   ├────────────────────────────►│                          │
   │                             │                          │
   │  [Wait for delay period]    │                          │
   │                             │                          │
   │  3. executeRecovery()       │                          │
   ├────────────────────────────►│                          │
   │                             │  Change owner            │
   │                             │  Revoke session keys     │
   │                             │  Unfreeze wallet         │
```

Guardian Types:
- **Wallet**: Another blockchain address
- **Email**: Verified email (hash-derived address)
- **Phone**: Verified phone (hash-derived address)
- **Hardware**: YubiKey or similar
- **Institution**: Trusted third party

### Custody Service Architecture

**Value Objects** (`pkg/domain/custody/value-objects/`):
- `chain.go`: CAIP-standard chain IDs, Solana-first multichain support
- `mpc.go`: MPC schemes, key curves, threshold configs, HSM/enclave configs

**Entities** (`pkg/domain/custody/entities/`):
- `smart_wallet.go`: SmartWallet with MPC keys, AA config, recovery, limits

**Ports** (`pkg/domain/custody/ports/`):
- `out/mpc_provider.go`: MPC key generation and signing interface
- `out/hsm_provider.go`: HSM and Secure Enclave operations
- `out/chain_client.go`: Multi-chain blockchain client interface
- `out/wallet_repository.go`: Wallet persistence interface
- `in/wallet_service.go`: Wallet operations interface
- `in/signing_service.go`: MPC signing interface
- `in/recovery_service.go`: Social recovery interface

**Services** (`pkg/domain/custody/services/`):
- `wallet_orchestrator.go`: Multi-chain wallet coordination
- `recovery_service.go`: Social recovery implementation

### Chain Support Matrix

| Chain | Primary | MPC Scheme | Wallet Type | Features |
|-------|---------|------------|-------------|----------|
| Solana | ✅ | FROST-Ed25519 | PDA Program | SPL tokens, Priority fees |
| Polygon | ✅ | CMP | ERC-4337 | Paymaster, Session keys |
| Base | ✅ | CMP | ERC-4337 | Paymaster, Session keys |
| Arbitrum | ✅ | CMP | ERC-4337 | Paymaster, Session keys |
| Ethereum | - | CMP | ERC-4337 | High-value only |

### Transaction Limits

```go
type TransactionLimits struct {
    DailyLimit    *big.Int  // Max per day
    WeeklyLimit   *big.Int  // Max per week
    MonthlyLimit  *big.Int  // Max per month
    PerTxLimit    *big.Int  // Max single transaction
    WhitelistOnly bool      // Only whitelisted addresses
}
```

Default Limits (Personal Wallet):
- Daily: $10,000
- Weekly: $50,000
- Monthly: $100,000
- Per Transaction: $5,000

### Security Levels

| Level | MPC Config | Features |
|-------|------------|----------|
| Basic | 2-of-3 | Single approval |
| Standard | 2-of-3 | Time delay for large tx |
| High | 3-of-5 | Multi-party + HSM |
| Critical | 4-of-7 | Governance + cold storage |

### API Examples

**Create Wallet:**
```go
result, err := walletService.CreateWallet(ctx, &custody_in.CreateWalletRequest{
    UserID:       userID,
    TenantID:     tenantID,
    WalletType:   custody_entities.WalletTypePersonal,
    PrimaryChain: custody_vo.ChainSolanaMainnet,
    Chains:       []custody_vo.ChainID{custody_vo.ChainPolygon, custody_vo.ChainBase},
    Label:        "Main Wallet",
})
```

**Transfer:**
```go
result, err := walletService.Transfer(ctx, &custody_in.TransferRequest{
    WalletID: walletID,
    ChainID:  custody_vo.ChainSolanaMainnet,
    To:       recipientAddress,
    Amount:   big.NewInt(1000000), // 1 USDC
})
```

**Add Guardian:**
```go
result, err := recoveryService.AddGuardian(ctx, &custody_in.AddGuardianRequest{
    WalletID:     walletID,
    GuardianType: custody_entities.GuardianTypeEmail,
    Email:        "guardian@example.com",
    Label:        "Recovery Email",
})
```

## Future Enhancements

- [x] Multi-chain custody (Solana, Polygon, Base, Arbitrum)
- [x] MPC key management (GG20, CMP, FROST)
- [x] ERC-4337 Account Abstraction
- [x] Social recovery with guardians
- [x] Fee sponsorship (Paymaster)
- [x] Session keys for delegated signing
- [ ] Multi-currency support (EUR, GBP, JPY)
- [ ] Crypto withdrawals (direct to user wallets)
- [ ] NFT marketplace integration
- [ ] Machine learning fraud detection
- [ ] Real-time balance streaming via WebSocket
- [ ] Cross-chain atomic swaps
- [ ] Hardware wallet integration (Ledger, Trezor)

## Contact

For questions about the wallet system, contact the platform engineering team.
