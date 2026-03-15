package oracle

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
)

// ScoreOracle Solidity ABI (publishScore function only)
const scoreOracleABI = `[{
	"inputs": [
		{"name": "oracleResultId", "type": "bytes32"},
		{"name": "externalMatchId", "type": "bytes32"},
		{"name": "teamAId", "type": "bytes32"},
		{"name": "teamBId", "type": "bytes32"},
		{"name": "teamAScore", "type": "uint16"},
		{"name": "teamBScore", "type": "uint16"},
		{"name": "winnerId", "type": "bytes32"},
		{"name": "isDraw", "type": "bool"},
		{"name": "roundsPlayed", "type": "uint32"},
		{"name": "gameId", "type": "string"},
		{"name": "sourceHash", "type": "bytes32"}
	],
	"name": "publishScore",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"inputs": [{"name": "oracleResultId", "type": "bytes32"}],
	"name": "getScore",
	"outputs": [{
		"components": [
			{"name": "externalMatchId", "type": "bytes32"},
			{"name": "teamAId", "type": "bytes32"},
			{"name": "teamBId", "type": "bytes32"},
			{"name": "teamAScore", "type": "uint16"},
			{"name": "teamBScore", "type": "uint16"},
			{"name": "winnerId", "type": "bytes32"},
			{"name": "isDraw", "type": "bool"},
			{"name": "roundsPlayed", "type": "uint32"},
			{"name": "gameId", "type": "string"},
			{"name": "sourceHash", "type": "bytes32"},
			{"name": "publishedAt", "type": "uint256"},
			{"name": "finalized", "type": "bool"},
			{"name": "disputed", "type": "bool"}
		],
		"name": "",
		"type": "tuple"
	}],
	"stateMutability": "view",
	"type": "function"
}, {
	"inputs": [{"name": "oracleResultId", "type": "bytes32"}],
	"name": "isFinalized",
	"outputs": [{"name": "", "type": "bool"}],
	"stateMutability": "view",
	"type": "function"
}]`

// RealEVMChainScoreGateway implements ChainScoreGateway with real on-chain transactions
type RealEVMChainScoreGateway struct {
	clients        map[oracle_vo.ChainID]*ethclient.Client
	contractAddrs  map[oracle_vo.ChainID]common.Address
	privateKey     *ecdsa.PrivateKey
	senderAddress  common.Address
	parsedABI      abi.ABI
	supportedChains []oracle_vo.ChainID
}

var _ oracle_out.ChainScoreGateway = (*RealEVMChainScoreGateway)(nil)

// RealEVMConfig holds configuration for real EVM chain connections
type RealEVMConfig struct {
	AmoyRPCURL       string
	AmoyContractAddr string
	PrivateKeyHex    string // hex-encoded private key (without 0x prefix)
}

// NewRealEVMChainScoreGateway creates a gateway that sends real transactions
func NewRealEVMChainScoreGateway(cfg RealEVMConfig) (*RealEVMChainScoreGateway, error) {
	// Parse private key
	keyHex := strings.TrimPrefix(cfg.PrivateKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	senderAddress := crypto.PubkeyToAddress(*publicKey)

	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(scoreOracleABI))
	if err != nil {
		return nil, fmt.Errorf("parse ScoreOracle ABI: %w", err)
	}

	gw := &RealEVMChainScoreGateway{
		clients:        make(map[oracle_vo.ChainID]*ethclient.Client),
		contractAddrs:  make(map[oracle_vo.ChainID]common.Address),
		privateKey:     privateKey,
		senderAddress:  senderAddress,
		parsedABI:      parsedABI,
	}

	// Connect to Amoy testnet
	if cfg.AmoyRPCURL != "" {
		client, err := ethclient.Dial(cfg.AmoyRPCURL)
		if err != nil {
			return nil, fmt.Errorf("connect to Amoy RPC: %w", err)
		}

		gw.clients[oracle_vo.ChainIDPolygonAmoy] = client
		gw.contractAddrs[oracle_vo.ChainIDPolygonAmoy] = common.HexToAddress(cfg.AmoyContractAddr)
		gw.supportedChains = append(gw.supportedChains, oracle_vo.ChainIDPolygonAmoy)

		slog.Info("real EVM gateway connected",
			slog.String("chain", "amoy"),
			slog.String("rpc", cfg.AmoyRPCURL),
			slog.String("contract", cfg.AmoyContractAddr),
			slog.String("sender", senderAddress.Hex()),
		)
	}

	if len(gw.supportedChains) == 0 {
		return nil, fmt.Errorf("no EVM chains configured")
	}

	return gw, nil
}

// PublishScore sends a real transaction to publish a score on-chain
func (gw *RealEVMChainScoreGateway) PublishScore(ctx context.Context, chainID oracle_vo.ChainID, result *oracle_entities.OracleResult) (*oracle_entities.ChainPublication, error) {
	client, ok := gw.clients[chainID]
	if !ok {
		return nil, fmt.Errorf("chain %d not configured", chainID)
	}

	contractAddr, ok := gw.contractAddrs[chainID]
	if !ok {
		return nil, fmt.Errorf("no contract for chain %d", chainID)
	}

	// Build the publishScore calldata
	oracleResultID := uuidToBytes32(result.ID)
	externalMatchID := hashToBytes32(func() string {
		if result.ExternalMatchID != nil {
			return *result.ExternalMatchID
		}
		return ""
	}())

	var teamAID, teamBID [32]byte
	if result.ConsensusResult != nil && len(result.ConsensusResult.TeamScores) >= 2 {
		teamAID = uuidToBytes32(result.ConsensusResult.TeamScores[0].TeamID)
		teamBID = uuidToBytes32(result.ConsensusResult.TeamScores[1].TeamID)
	}

	var teamAScore, teamBScore uint16
	var roundsPlayed uint32
	var isDraw bool
	var winnerID [32]byte

	if result.ConsensusResult != nil {
		if len(result.ConsensusResult.TeamScores) >= 2 {
			teamAScore = uint16(result.ConsensusResult.TeamScores[0].Score)
			teamBScore = uint16(result.ConsensusResult.TeamScores[1].Score)
		}
		isDraw = result.ConsensusResult.IsDraw

		if result.ConsensusResult.WinnerTeamID != nil {
			winnerID = uuidToBytes32(*result.ConsensusResult.WinnerTeamID)
		}
	}

	// Compute source hash
	sourceHashStr := computeSourceHash(result)
	var sourceHash [32]byte
	if sourceHashStr != "" {
		decoded, _ := hex.DecodeString(sourceHashStr)
		copy(sourceHash[:], decoded)
	}

	gameID := string(result.GameID)

	// Encode transaction data
	txData, err := gw.parsedABI.Pack("publishScore",
		oracleResultID,
		externalMatchID,
		teamAID,
		teamBID,
		teamAScore,
		teamBScore,
		winnerID,
		isDraw,
		roundsPlayed,
		gameID,
		sourceHash,
	)
	if err != nil {
		return nil, fmt.Errorf("pack publishScore: %w", err)
	}

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, gw.senderAddress)
	if err != nil {
		return nil, fmt.Errorf("get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas price: %w", err)
	}

	// Estimate gas
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: gw.senderAddress,
		To:   &contractAddr,
		Data: txData,
	})
	if err != nil {
		slog.WarnContext(ctx, "gas estimation failed, using default",
			slog.String("error", err.Error()),
		)
		gasLimit = 200000 // Default gas limit for publishScore
	}

	// Build transaction
	evmChainID := big.NewInt(int64(chainID))
	tx := types.NewTransaction(nonce, contractAddr, big.NewInt(0), gasLimit, gasPrice, txData)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(evmChainID), gw.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign transaction: %w", err)
	}

	// Send transaction
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()

	slog.InfoContext(ctx, "score published to EVM chain (real tx)",
		slog.Int64("chain_id", int64(chainID)),
		slog.String("tx_hash", txHash),
		slog.String("contract", contractAddr.Hex()),
		slog.String("oracle_result_id", result.ID.String()),
		slog.Uint64("nonce", nonce),
		slog.Uint64("gas_limit", gasLimit),
	)

	pub := &oracle_entities.ChainPublication{
		ChainID:         chainID,
		CAIP2:           chainID.CAIP2(),
		ContractAddress: contractAddr.Hex(),
		TxHash:          txHash,
		BlockNumber:     0, // Will be available after mining
		Status:          "pending",
		PublishedAt:     time.Now().UTC(),
	}

	return pub, nil
}

// GetPublishedScore retrieves a published score from the chain via contract call
func (gw *RealEVMChainScoreGateway) GetPublishedScore(ctx context.Context, chainID oracle_vo.ChainID, matchID uuid.UUID) (*oracle_out.OnChainScore, error) {
	client, ok := gw.clients[chainID]
	if !ok {
		return nil, fmt.Errorf("chain %d not configured", chainID)
	}

	contractAddr := gw.contractAddrs[chainID]
	oracleResultID := uuidToBytes32(matchID)

	callData, err := gw.parsedABI.Pack("getScore", oracleResultID)
	if err != nil {
		return nil, fmt.Errorf("pack getScore: %w", err)
	}

	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("call getScore: %w", err)
	}

	// Unpack the result
	outputs, err := gw.parsedABI.Unpack("getScore", result)
	if err != nil {
		return nil, fmt.Errorf("unpack getScore: %w", err)
	}

	if len(outputs) == 0 {
		return nil, fmt.Errorf("empty result from getScore")
	}

	// The result is a struct, returned as an anonymous struct-like type
	// For now, return a simplified version
	slog.InfoContext(ctx, "retrieved on-chain score",
		slog.Int64("chain_id", int64(chainID)),
		slog.String("oracle_result_id", matchID.String()),
	)

	return &oracle_out.OnChainScore{
		ChainID:   chainID,
		IsFinalized: false,
	}, nil
}

// IsScoreFinalized checks the finalization status on-chain
func (gw *RealEVMChainScoreGateway) IsScoreFinalized(ctx context.Context, chainID oracle_vo.ChainID, matchID uuid.UUID) (bool, error) {
	client, ok := gw.clients[chainID]
	if !ok {
		return false, fmt.Errorf("chain %d not configured", chainID)
	}

	contractAddr := gw.contractAddrs[chainID]
	oracleResultID := uuidToBytes32(matchID)

	callData, err := gw.parsedABI.Pack("isFinalized", oracleResultID)
	if err != nil {
		return false, fmt.Errorf("pack isFinalized: %w", err)
	}

	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: callData,
	}, nil)
	if err != nil {
		return false, fmt.Errorf("call isFinalized: %w", err)
	}

	outputs, err := gw.parsedABI.Unpack("isFinalized", result)
	if err != nil {
		return false, fmt.Errorf("unpack isFinalized: %w", err)
	}

	if len(outputs) == 0 {
		return false, nil
	}

	finalized, ok := outputs[0].(bool)
	return ok && finalized, nil
}

// SupportedChains returns configured chains
func (gw *RealEVMChainScoreGateway) SupportedChains() []oracle_vo.ChainID {
	return gw.supportedChains
}

// --- Helpers ---

func uuidToBytes32(id uuid.UUID) [32]byte {
	var b32 [32]byte
	copy(b32[:16], id[:])
	return b32
}

func hashToBytes32(s string) [32]byte {
	hash := sha256.Sum256([]byte(s))
	return hash
}
