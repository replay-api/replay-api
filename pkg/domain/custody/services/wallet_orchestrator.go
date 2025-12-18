package custody_services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
	custody_entities "github.com/replay-api/replay-api/pkg/domain/custody/entities"
	custody_in "github.com/replay-api/replay-api/pkg/domain/custody/ports/in"
	custody_out "github.com/replay-api/replay-api/pkg/domain/custody/ports/out"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// WalletOrchestrator coordinates smart wallet operations across multiple chains
type WalletOrchestrator struct {
	// MPC Provider
	mpcProvider custody_out.MPCProvider

	// HSM Provider
	hsmProvider custody_out.HSMProvider

	// Chain Clients
	solanaClient  custody_out.SolanaClient
	evmClients    map[custody_vo.ChainID]custody_out.EVMClient

	// Repositories
	walletRepo     custody_out.SmartWalletRepository
	txRepo         custody_out.TransactionRepository
	keyRepo        custody_out.KeyRepository
	signingRepo    custody_out.SigningSessionRepository

	// Configuration
	config *OrchestratorConfig

	mu sync.RWMutex
}

// OrchestratorConfig contains orchestrator configuration
type OrchestratorConfig struct {
	DefaultThreshold    custody_vo.ThresholdConfig
	DefaultRecoveryDelay time.Duration
	DefaultDailyLimit   *big.Int
	SupportedChains     []custody_vo.ChainID
	EntryPoints         map[custody_vo.ChainID]string
	PaymasterAddresses  map[custody_vo.ChainID]string
	WalletFactories     map[custody_vo.ChainID]string
}

// NewWalletOrchestrator creates a new wallet orchestrator
func NewWalletOrchestrator(
	mpcProvider custody_out.MPCProvider,
	hsmProvider custody_out.HSMProvider,
	solanaClient custody_out.SolanaClient,
	evmClients map[custody_vo.ChainID]custody_out.EVMClient,
	walletRepo custody_out.SmartWalletRepository,
	txRepo custody_out.TransactionRepository,
	keyRepo custody_out.KeyRepository,
	signingRepo custody_out.SigningSessionRepository,
	config *OrchestratorConfig,
) *WalletOrchestrator {
	return &WalletOrchestrator{
		mpcProvider:   mpcProvider,
		hsmProvider:   hsmProvider,
		solanaClient:  solanaClient,
		evmClients:    evmClients,
		walletRepo:    walletRepo,
		txRepo:        txRepo,
		keyRepo:       keyRepo,
		signingRepo:   signingRepo,
		config:        config,
	}
}

// CreateWallet creates a new smart wallet with MPC keys
func (o *WalletOrchestrator) CreateWallet(
	ctx context.Context,
	req *custody_in.CreateWalletRequest,
) (*custody_in.CreateWalletResult, error) {
	walletID := uuid.New()

	// Determine key curves needed based on chains
	needsSecp256k1 := false
	needsEd25519 := false

	for _, chainID := range append([]custody_vo.ChainID{req.PrimaryChain}, req.Chains...) {
		if chainID.IsSolana() {
			needsEd25519 = true
		} else if chainID.IsEVM() {
			needsSecp256k1 = true
		}
	}

	// Generate MPC keys
	var evmKeyResult, solanaKeyResult *custody_out.GeneratedKey

	// Generate secp256k1 key for EVM chains (GG20 or CMP scheme)
	if needsSecp256k1 {
		keyGenReq := &custody_out.KeyGenRequest{
			KeyID:     generateKeyID(walletID, "evm"),
			WalletID:  walletID.String(),
			Curve:     custody_vo.CurveSecp256k1,
			Scheme:    custody_vo.MPCSchemeCMP, // Faster than GG20
			Threshold: o.config.DefaultThreshold,
			Metadata: map[string]string{
				"wallet_id": walletID.String(),
				"purpose":   string(custody_vo.KeyPurposeTransaction),
			},
		}

		session, err := o.mpcProvider.InitiateKeyGeneration(ctx, keyGenReq)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate EVM key generation: %w", err)
		}

		// Wait for key generation to complete
		evmKeyResult, err = o.waitForKeyGeneration(ctx, session.SessionID)
		if err != nil {
			return nil, fmt.Errorf("EVM key generation failed: %w", err)
		}
	}

	// Generate Ed25519 key for Solana (FROST scheme)
	if needsEd25519 {
		keyGenReq := &custody_out.KeyGenRequest{
			KeyID:     generateKeyID(walletID, "solana"),
			WalletID:  walletID.String(),
			Curve:     custody_vo.CurveEd25519,
			Scheme:    custody_vo.MPCSchemeFROSTEd25519,
			Threshold: o.config.DefaultThreshold,
			Metadata: map[string]string{
				"wallet_id": walletID.String(),
				"purpose":   string(custody_vo.KeyPurposeTransaction),
			},
		}

		session, err := o.mpcProvider.InitiateKeyGeneration(ctx, keyGenReq)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate Solana key generation: %w", err)
		}

		solanaKeyResult, err = o.waitForKeyGeneration(ctx, session.SessionID)
		if err != nil {
			return nil, fmt.Errorf("Solana key generation failed: %w", err)
		}
	}

	// Build addresses map
	addresses := make(map[custody_vo.ChainID]string)

	if evmKeyResult != nil {
		for _, chainID := range req.Chains {
			if chainID.IsEVM() {
				addresses[chainID] = evmKeyResult.EVMAddress
			}
		}
		if req.PrimaryChain.IsEVM() {
			addresses[req.PrimaryChain] = evmKeyResult.EVMAddress
		}
	}

	if solanaKeyResult != nil {
		for _, chainID := range req.Chains {
			if chainID.IsSolana() {
				addresses[chainID] = solanaKeyResult.SolanaAddress
			}
		}
		if req.PrimaryChain.IsSolana() {
			addresses[req.PrimaryChain] = solanaKeyResult.SolanaAddress
		}
	}

	// Create wallet entity using the constructor
	wallet := custody_entities.NewSmartWallet(
		req.ResourceOwner,
		req.UserID,
		req.Label,
		req.WalletType,
		req.PrimaryChain,
	)
	// Override the auto-generated ID with our specific wallet ID
	wallet.BaseEntity.ID = walletID
	wallet.Addresses = addresses
	wallet.KYCStatus = req.KYCStatus

	// Convert metadata from map[string]string to map[string]interface{}
	if req.Metadata != nil {
		metadata := make(map[string]interface{})
		for k, v := range req.Metadata {
			metadata[k] = v
		}
		wallet.Metadata = metadata
	}

	// Set MPC key info
	var masterKeyID string
	if evmKeyResult != nil {
		masterKeyID = evmKeyResult.KeyID
		wallet.PublicKey = hex.EncodeToString(evmKeyResult.PublicKey)
	} else if solanaKeyResult != nil {
		masterKeyID = solanaKeyResult.KeyID
		wallet.PublicKey = hex.EncodeToString(solanaKeyResult.PublicKey)
	}

	// Store Solana key info in metadata if available
	if solanaKeyResult != nil {
		wallet.Metadata["solana_key_id"] = solanaKeyResult.KeyID
		wallet.Metadata["solana_public_key"] = hex.EncodeToString(solanaKeyResult.PublicKey)
	}

	// Set transaction limits
	if req.Limits != nil {
		wallet.Limits = *req.Limits
	} else {
		now := time.Now()
		defaultDailyLimit := uint64(10000_00) // $10,000 in cents
		if o.config.DefaultDailyLimit != nil {
			defaultDailyLimit = o.config.DefaultDailyLimit.Uint64()
		}
		wallet.Limits = custody_entities.TransactionLimits{
			DailyLimit:       defaultDailyLimit,
			WeeklyLimit:      defaultDailyLimit * 7,
			MonthlyLimit:     defaultDailyLimit * 30,
			SingleTxLimit:    defaultDailyLimit,
			DailyUsed:        0,
			WeeklyUsed:       0,
			MonthlyUsed:      0,
			LastResetDaily:   now,
			LastResetWeekly:  now,
			LastResetMonthly: now,
		}
	}

	// Set AA config for EVM chains
	if needsSecp256k1 && evmKeyResult != nil {
		wallet.AAConfig = &custody_entities.AccountAbstractionConfig{
			IsDeployed:       make(map[custody_vo.ChainID]bool),
			EntryPointAddr:   o.config.EntryPoints[req.PrimaryChain],
			PaymasterEnabled: o.config.PaymasterAddresses[req.PrimaryChain] != "",
			PaymasterAddress: o.config.PaymasterAddresses[req.PrimaryChain],
		}

		// Set entry points and compute counterfactual addresses
		for chainID := range addresses {
			if chainID.IsEVM() {
				wallet.AAConfig.IsDeployed[chainID] = false
			}
		}
	}

	// Set recovery config
	wallet.RecoveryConfig = &custody_entities.WalletRecoveryConfig{
		IsEnabled:         true,
		RecoveryDelay:     o.config.DefaultRecoveryDelay,
		GuardianThreshold: o.config.DefaultThreshold.Threshold,
	}

	// Save wallet
	if err := o.walletRepo.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to save wallet: %w", err)
	}

	// Save key records
	keyCreatedAt := time.Now()
	if evmKeyResult != nil {
		keyRecord := &custody_out.KeyRecord{
			KeyID:         evmKeyResult.KeyID,
			WalletID:      walletID,
			PublicKey:     evmKeyResult.PublicKey,
			PublicKeyHex:  evmKeyResult.PublicKeyHex,
			Curve:         evmKeyResult.Curve,
			Scheme:        evmKeyResult.Scheme,
			Threshold:     evmKeyResult.Threshold,
			Purpose:       custody_vo.KeyPurposeTransaction,
			EVMAddress:    evmKeyResult.EVMAddress,
			ShareMetadata: evmKeyResult.ShareMetadata,
			IsActive:      true,
			CreatedAt:     keyCreatedAt,
		}
		if err := o.keyRepo.Create(ctx, keyRecord); err != nil {
			return nil, fmt.Errorf("failed to save EVM key record: %w", err)
		}
	}

	if solanaKeyResult != nil {
		keyRecord := &custody_out.KeyRecord{
			KeyID:          solanaKeyResult.KeyID,
			WalletID:       walletID,
			PublicKey:      solanaKeyResult.PublicKey,
			PublicKeyHex:   solanaKeyResult.PublicKeyHex,
			Curve:          solanaKeyResult.Curve,
			Scheme:         solanaKeyResult.Scheme,
			Threshold:      solanaKeyResult.Threshold,
			Purpose:        custody_vo.KeyPurposeTransaction,
			SolanaAddress:  solanaKeyResult.SolanaAddress,
			ShareMetadata:  solanaKeyResult.ShareMetadata,
			IsActive:       true,
			CreatedAt:      keyCreatedAt,
		}
		if err := o.keyRepo.Create(ctx, keyRecord); err != nil {
			return nil, fmt.Errorf("failed to save Solana key record: %w", err)
		}
	}

	// Set master key ID using the MPC key config
	wallet.SetMPCKeyConfig(masterKeyID, custody_entities.MPCKeyConfiguration{
		Scheme:         custody_vo.MPCSchemeCMP,
		Curve:          custody_vo.CurveSecp256k1,
		Threshold:      o.config.DefaultThreshold,
		KeyGeneratedAt: time.Now(),
	}, nil)

	return &custody_in.CreateWalletResult{
		Wallet:    wallet,
		MPCKeyID:  masterKeyID,
		Addresses: addresses,
	}, nil
}

// DeployWallet deploys the smart wallet on a specific chain
func (o *WalletOrchestrator) DeployWallet(
	ctx context.Context,
	walletID uuid.UUID,
	chainID custody_vo.ChainID,
) (*custody_in.DeployWalletResult, error) {
	wallet, err := o.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	if chainID.IsSolana() {
		return o.deploySolanaWallet(ctx, wallet, chainID)
	} else if chainID.IsEVM() {
		return o.deployEVMWallet(ctx, wallet, chainID)
	}

	return nil, fmt.Errorf("unsupported chain: %s", chainID)
}

// deploySolanaWallet deploys wallet on Solana using PDA
func (o *WalletOrchestrator) deploySolanaWallet(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	chainID custody_vo.ChainID,
) (*custody_in.DeployWalletResult, error) {
	// For Solana, the wallet PDA is derived deterministically
	// No actual deployment transaction needed - just initialize the account

	// Build initialize_wallet instruction
	walletIDBytes := sha256.Sum256([]byte(wallet.BaseEntity.ID.String()))

	// Get Solana public key from metadata
	solanaPublicKey, _ := wallet.Metadata["solana_public_key"].(string)
	if solanaPublicKey == "" {
		solanaPublicKey = wallet.PublicKey
	}

	// Get guardian threshold and recovery delay
	guardianThreshold := uint8(2)
	recoveryDelay := int64(24 * 3600) // 24 hours default
	if wallet.RecoveryConfig != nil {
		guardianThreshold = wallet.RecoveryConfig.GuardianThreshold
		recoveryDelay = int64(wallet.RecoveryConfig.RecoveryDelay.Seconds())
	}

	instrReq := &custody_out.ProgramInstructionRequest{
		ProgramID: "LeetWa11etPr0gram1111111111111111111111111", // Program ID
		Accounts: []custody_out.AccountMeta{
			{Address: wallet.Addresses[chainID], IsSigner: false, IsWritable: true}, // Wallet PDA
			{Address: solanaPublicKey, IsSigner: false, IsWritable: false},          // Owner
			{Address: solanaPublicKey, IsSigner: true, IsWritable: false},           // Authority
			{Address: "", IsSigner: true, IsWritable: true},                          // Payer (to be filled by signer)
			{Address: "11111111111111111111111111111111", IsSigner: false, IsWritable: false}, // System program
		},
		Data: buildInitializeWalletData(
			walletIDBytes[:],
			guardianThreshold,
			wallet.Limits.DailyLimit,
			recoveryDelay,
		),
	}

	unsignedTx, err := o.solanaClient.BuildProgramInstruction(ctx, instrReq)
	if err != nil {
		return nil, fmt.Errorf("failed to build Solana init instruction: %w", err)
	}

	// Sign and submit
	signedTx, err := o.signAndSubmitSolana(ctx, wallet, unsignedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Solana wallet: %w", err)
	}

	// Wait for confirmation
	receipt, err := o.solanaClient.WaitForConfirmation(ctx, signedTx.TxHash, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm Solana deployment: %w", err)
	}

	// Update wallet status
	wallet.Status = custody_entities.WalletStatusActive
	if wallet.AAConfig == nil {
		wallet.AAConfig = &custody_entities.AccountAbstractionConfig{
			IsDeployed: make(map[custody_vo.ChainID]bool),
		}
	}
	wallet.AAConfig.IsDeployed[chainID] = true
	now := time.Now()
	wallet.AAConfig.DeployedAt = &now
	wallet.UpdatedAt = now

	// Store deployment tx in metadata
	if wallet.Metadata == nil {
		wallet.Metadata = make(map[string]interface{})
	}
	wallet.Metadata["deployment_tx_"+string(chainID)] = signedTx.TxHash

	if err := o.walletRepo.Update(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return &custody_in.DeployWalletResult{
		ChainID:     chainID,
		Address:     wallet.Addresses[chainID],
		TxHash:      signedTx.TxHash,
		BlockNumber: receipt.BlockNumber,
		GasUsed:     receipt.GasUsed,
	}, nil
}

// deployEVMWallet deploys wallet on an EVM chain using ERC-4337
func (o *WalletOrchestrator) deployEVMWallet(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	chainID custody_vo.ChainID,
) (*custody_in.DeployWalletResult, error) {
	evmClient, ok := o.evmClients[chainID]
	if !ok {
		return nil, fmt.Errorf("EVM client not found for chain: %s", chainID)
	}

	// Get recovery delay
	recoveryDelay := 24 * time.Hour
	if wallet.RecoveryConfig != nil {
		recoveryDelay = wallet.RecoveryConfig.RecoveryDelay
	}

	// Build UserOperation for wallet deployment
	// The initCode contains the factory address + create2 salt + initialization data
	initCode := buildWalletInitCode(
		o.config.WalletFactories[chainID],
		wallet.BaseEntity.ID,
		wallet.PublicKey,
		o.config.EntryPoints[chainID],
		big.NewInt(int64(wallet.Limits.DailyLimit)),
		recoveryDelay,
	)

	userOpReq := &custody_out.UserOpRequest{
		Sender:        wallet.Addresses[chainID],
		Target:        wallet.Addresses[chainID], // Self-call for deployment
		Value:         big.NewInt(0),
		CallData:      []byte{}, // No call data for deployment
		Paymaster:     stringPtr(o.config.PaymasterAddresses[chainID]),
		PaymasterData: []byte{0x00}, // Sponsored mode
	}

	userOp, err := evmClient.BuildUserOperation(ctx, userOpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to build UserOperation: %w", err)
	}

	userOp.InitCode = initCode

	// Estimate gas
	gasEstimate, err := evmClient.EstimateUserOperationGas(ctx, userOp)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Apply gas estimates
	userOp.PreVerificationGas = new(big.Int).SetUint64(gasEstimate.PreVerificationGas)
	// userOp.VerificationGasLimit would be set in AccountGasLimits

	// Sign UserOperation
	signedUserOp, err := o.signUserOperation(ctx, wallet, chainID, userOp)
	if err != nil {
		return nil, fmt.Errorf("failed to sign UserOperation: %w", err)
	}

	// Submit UserOperation
	result, err := evmClient.SubmitUserOperation(ctx, signedUserOp)
	if err != nil {
		return nil, fmt.Errorf("failed to submit UserOperation: %w", err)
	}

	// Wait for receipt
	receipt, err := evmClient.GetUserOperationReceipt(ctx, result.UserOpHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get UserOp receipt: %w", err)
	}

	if !receipt.Success {
		return nil, fmt.Errorf("wallet deployment failed")
	}

	// Update wallet status
	wallet.Status = custody_entities.WalletStatusActive
	if wallet.AAConfig == nil {
		wallet.AAConfig = &custody_entities.AccountAbstractionConfig{
			IsDeployed: make(map[custody_vo.ChainID]bool),
		}
	}
	wallet.AAConfig.IsDeployed[chainID] = true
	now := time.Now()
	wallet.AAConfig.DeployedAt = &now
	wallet.UpdatedAt = now

	// Store deployment tx in metadata
	if wallet.Metadata == nil {
		wallet.Metadata = make(map[string]interface{})
	}
	wallet.Metadata["deployment_tx_"+string(chainID)] = receipt.TxHash

	if err := o.walletRepo.Update(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return &custody_in.DeployWalletResult{
		ChainID:     chainID,
		Address:     wallet.Addresses[chainID],
		TxHash:      receipt.TxHash,
		BlockNumber: receipt.BlockNumber,
		GasUsed:     receipt.ActualGasUsed,
	}, nil
}

// Transfer executes a native token transfer
func (o *WalletOrchestrator) Transfer(
	ctx context.Context,
	req *custody_in.TransferRequest,
) (*custody_in.TransferResult, error) {
	wallet, err := o.walletRepo.GetByID(ctx, req.WalletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	if wallet.IsFrozen {
		return nil, fmt.Errorf("wallet is frozen")
	}

	// Check spending limits
	if err := o.checkSpendingLimits(ctx, wallet, req.Amount); err != nil {
		return nil, err
	}

	// Create transaction record
	txID := uuid.New()
	txRecord := &custody_out.CustodyTransaction{
		ID:        txID,
		WalletID:  wallet.BaseEntity.ID,
		ChainID:   req.ChainID,
		TxType:    custody_out.TxTypeTransfer,
		Status:    custody_out.TxRecordStatusPending,
		From:      wallet.Addresses[req.ChainID],
		To:        req.To,
		Value:     req.Amount.String(),
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := o.txRepo.Create(ctx, txRecord); err != nil {
		return nil, fmt.Errorf("failed to create tx record: %w", err)
	}

	// Execute transfer based on chain
	var txHash string
	var gasUsed uint64

	if req.ChainID.IsSolana() {
		result, err := o.executeSolanaTransfer(ctx, wallet, req)
		if err != nil {
			txRecord.Status = custody_out.TxRecordStatusFailed
			txRecord.FailureReason = stringPtr(err.Error())
			if updateErr := o.txRepo.Update(ctx, txRecord); updateErr != nil {
				slog.WarnContext(ctx, "Failed to update failed tx record", "error", updateErr, "tx_id", txRecord.ID)
			}
			return nil, err
		}
		txHash = result.TxHash
		gasUsed = result.GasUsed
	} else if req.ChainID.IsEVM() {
		result, err := o.executeEVMTransfer(ctx, wallet, req)
		if err != nil {
			txRecord.Status = custody_out.TxRecordStatusFailed
			txRecord.FailureReason = stringPtr(err.Error())
			if updateErr := o.txRepo.Update(ctx, txRecord); updateErr != nil {
				slog.WarnContext(ctx, "Failed to update failed tx record", "error", updateErr, "tx_id", txRecord.ID)
			}
			return nil, err
		}
		txHash = result.TxHash
		gasUsed = result.GasUsed
	} else {
		return nil, fmt.Errorf("unsupported chain: %s", req.ChainID)
	}

	// Update transaction record
	now := time.Now()
	txRecord.Status = custody_out.TxRecordStatusConfirmed
	txRecord.TxHash = txHash
	txRecord.GasUsed = &gasUsed
	txRecord.ConfirmedAt = &now
	txRecord.UpdatedAt = now

	if err := o.txRepo.Update(ctx, txRecord); err != nil {
		slog.WarnContext(ctx, "Failed to update tx record after confirmation", "error", err, "tx_id", txRecord.ID)
	}

	// Update spending
	o.updateSpending(ctx, wallet, req.Amount)

	return &custody_in.TransferResult{
		TxID:        txID,
		TxHash:      txHash,
		ChainID:     req.ChainID,
		From:        wallet.Addresses[req.ChainID],
		To:          req.To,
		Amount:      req.Amount,
		Status:      "confirmed",
		GasUsed:     &gasUsed,
		SubmittedAt: txRecord.CreatedAt,
		ConfirmedAt: txRecord.ConfirmedAt,
	}, nil
}

// Helper functions

func (o *WalletOrchestrator) waitForKeyGeneration(
	ctx context.Context,
	sessionID string,
) (*custody_out.GeneratedKey, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(60 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("key generation timeout")
		case <-ticker.C:
			status, err := o.mpcProvider.GetKeyGenStatus(ctx, sessionID)
			if err != nil {
				return nil, err
			}

			switch status.State {
			case custody_out.KeyGenStateCompleted:
				return o.mpcProvider.FinalizeKeyGeneration(ctx, sessionID)
			case custody_out.KeyGenStateFailed:
				return nil, fmt.Errorf("key generation failed: %s", status.Error)
			}
		}
	}
}

func (o *WalletOrchestrator) signAndSubmitSolana(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	unsignedTx *custody_out.UnsignedTransaction,
) (*custody_out.TxSubmitResult, error) {
	// Get Solana key ID from metadata or fallback to master key
	solanaKeyID, _ := wallet.Metadata["solana_key_id"].(string)
	if solanaKeyID == "" {
		solanaKeyID = wallet.MasterKeyID
	}

	// Create signing request
	signingReq := &custody_out.SigningRequest{
		SessionID:   uuid.New().String(),
		KeyID:       solanaKeyID,
		MessageHash: unsignedTx.MessageHash,
		MessageType: custody_vo.MessageTypeSolanaMessage,
		ChainID:     custody_vo.ChainSolanaMainnet,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	// Initiate signing
	session, err := o.mpcProvider.InitiateSigning(ctx, signingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate signing: %w", err)
	}

	// Wait for signature
	signature, err := o.waitForSignature(ctx, session.SessionID)
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	// Attach signature to transaction
	signedTx := append(unsignedTx.RawTx, signature.Signature...)

	// Submit transaction
	return o.solanaClient.SubmitTransaction(ctx, signedTx)
}

func (o *WalletOrchestrator) signUserOperation(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	chainID custody_vo.ChainID,
	userOp *custody_out.UserOperation,
) (*custody_out.UserOperation, error) {
	// Create signing request
	signingReq := &custody_out.SigningRequest{
		SessionID:   uuid.New().String(),
		KeyID:       wallet.MasterKeyID,
		MessageHash: userOp.UserOpHash,
		MessageType: custody_vo.MessageTypeTransaction,
		ChainID:     chainID,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	// Initiate signing
	session, err := o.mpcProvider.InitiateSigning(ctx, signingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate signing: %w", err)
	}

	// Wait for signature
	signature, err := o.waitForSignature(ctx, session.SessionID)
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	// Attach signature to UserOperation
	userOp.Signature = signature.Signature

	return userOp, nil
}

func (o *WalletOrchestrator) waitForSignature(
	ctx context.Context,
	sessionID string,
) (*custody_out.SignatureResult, error) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("signing timeout")
		case <-ticker.C:
			status, err := o.mpcProvider.GetSigningStatus(ctx, sessionID)
			if err != nil {
				return nil, err
			}

			switch status.State {
			case custody_out.SigningStateCompleted:
				return o.mpcProvider.GetSignature(ctx, sessionID)
			case custody_out.SigningStateFailed:
				return nil, fmt.Errorf("signing failed: %s", status.Error)
			case custody_out.SigningStateExpired:
				return nil, fmt.Errorf("signing session expired")
			}
		}
	}
}

func (o *WalletOrchestrator) checkSpendingLimits(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	amount *big.Int,
) error {
	// Reset limits if needed
	now := time.Now()
	if now.YearDay() != wallet.Limits.LastResetDaily.YearDay() || now.Year() != wallet.Limits.LastResetDaily.Year() {
		wallet.Limits.DailyUsed = 0
		wallet.Limits.LastResetDaily = now
	}

	amountUint := amount.Uint64()

	// Check per-tx limit
	if amountUint > wallet.Limits.SingleTxLimit {
		return fmt.Errorf("amount exceeds per-transaction limit")
	}

	// Check daily limit
	if wallet.Limits.DailyUsed+amountUint > wallet.Limits.DailyLimit {
		return fmt.Errorf("daily spending limit exceeded")
	}

	return nil
}

func (o *WalletOrchestrator) updateSpending(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	amount *big.Int,
) {
	amountUint := amount.Uint64()
	wallet.Limits.DailyUsed += amountUint
	wallet.Limits.WeeklyUsed += amountUint
	wallet.Limits.MonthlyUsed += amountUint
	wallet.UpdatedAt = time.Now()

	if err := o.walletRepo.Update(ctx, wallet); err != nil {
		slog.WarnContext(ctx, "Failed to update wallet limits", "error", err, "wallet_id", wallet.ID)
	}
}

func (o *WalletOrchestrator) executeSolanaTransfer(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	req *custody_in.TransferRequest,
) (*transferResult, error) {
	// Build transfer instruction
	txReq := &custody_out.TxBuildRequest{
		From:  wallet.Addresses[req.ChainID],
		To:    req.To,
		Value: req.Amount,
	}

	unsignedTx, err := o.solanaClient.BuildTransaction(ctx, txReq)
	if err != nil {
		return nil, err
	}

	result, err := o.signAndSubmitSolana(ctx, wallet, unsignedTx)
	if err != nil {
		return nil, err
	}

	receipt, err := o.solanaClient.WaitForConfirmation(ctx, result.TxHash, 1)
	if err != nil {
		return nil, err
	}

	return &transferResult{
		TxHash:  result.TxHash,
		GasUsed: receipt.GasUsed,
	}, nil
}

func (o *WalletOrchestrator) executeEVMTransfer(
	ctx context.Context,
	wallet *custody_entities.SmartWallet,
	req *custody_in.TransferRequest,
) (*transferResult, error) {
	evmClient, ok := o.evmClients[req.ChainID]
	if !ok {
		return nil, fmt.Errorf("EVM client not found for chain: %s", req.ChainID)
	}

	// Build execute calldata for smart wallet
	executeCalldata := buildExecuteCalldata(req.To, req.Amount, []byte{})

	// Build UserOperation
	userOpReq := &custody_out.UserOpRequest{
		Sender:        wallet.Addresses[req.ChainID],
		Target:        wallet.Addresses[req.ChainID],
		Value:         big.NewInt(0),
		CallData:      executeCalldata,
		Paymaster:     stringPtr(o.config.PaymasterAddresses[req.ChainID]),
		PaymasterData: []byte{0x00}, // Sponsored mode
	}

	userOp, err := evmClient.BuildUserOperation(ctx, userOpReq)
	if err != nil {
		return nil, err
	}

	// Sign and submit
	signedUserOp, err := o.signUserOperation(ctx, wallet, req.ChainID, userOp)
	if err != nil {
		return nil, err
	}

	result, err := evmClient.SubmitUserOperation(ctx, signedUserOp)
	if err != nil {
		return nil, err
	}

	receipt, err := evmClient.GetUserOperationReceipt(ctx, result.UserOpHash)
	if err != nil {
		return nil, err
	}

	if !receipt.Success {
		return nil, fmt.Errorf("transfer failed")
	}

	return &transferResult{
		TxHash:  receipt.TxHash,
		GasUsed: receipt.ActualGasUsed,
	}, nil
}

type transferResult struct {
	TxHash  string
	GasUsed uint64
}

// Helper functions for building calldata

func generateKeyID(walletID uuid.UUID, suffix string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", walletID.String(), suffix)))
	return hex.EncodeToString(hash[:16])
}

func buildInitializeWalletData(walletID []byte, threshold uint8, dailyLimit uint64, recoveryDelay int64) []byte {
	// Build Anchor instruction data
	// In production, use proper Borsh serialization
	data := make([]byte, 0, 50)
	data = append(data, walletID...)
	data = append(data, threshold)
	// Add other parameters...
	return data
}

func buildWalletInitCode(factory string, walletID uuid.UUID, publicKey string, entryPoint string, dailyLimit *big.Int, recoveryDelay time.Duration) []byte {
	// Build initCode for ERC-4337 wallet deployment
	// Format: factory address (20 bytes) + create function calldata
	// In production, use proper ABI encoding
	return []byte{}
}

func buildExecuteCalldata(to string, value *big.Int, data []byte) []byte {
	// Build execute(address,uint256,bytes) calldata
	// In production, use proper ABI encoding
	return []byte{}
}

func stringPtr(s string) *string {
	return &s
}
