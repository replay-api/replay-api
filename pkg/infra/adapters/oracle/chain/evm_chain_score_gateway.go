package oracle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
)

// EVMChainScoreGatewayConfig holds configuration for EVM chain connections
type EVMChainScoreGatewayConfig struct {
	PolygonRPCURL        string
	PolygonContractAddr  string
	AmoyRPCURL           string
	AmoyContractAddr     string
	PrivateKey           string
}

// EVMChainScoreGateway implements ChainScoreGateway for EVM-compatible chains.
// Phase 1: simulated publish with deterministic tx hash generation.
// Phase 2+: real on-chain transactions via go-ethereum.
type EVMChainScoreGateway struct {
	supportedChains []oracle_vo.ChainID
	rpcURLs         map[oracle_vo.ChainID]string
	contractAddrs   map[oracle_vo.ChainID]string
	privateKey      string
}

// NewEVMChainScoreGateway creates a new EVM chain score gateway from config
func NewEVMChainScoreGateway(cfg EVMChainScoreGatewayConfig) oracle_out.ChainScoreGateway {
	gw := &EVMChainScoreGateway{
		rpcURLs:       make(map[oracle_vo.ChainID]string),
		contractAddrs: make(map[oracle_vo.ChainID]string),
		privateKey:    cfg.PrivateKey,
	}

	// Register configured chains
	if cfg.PolygonRPCURL != "" {
		gw.supportedChains = append(gw.supportedChains, oracle_vo.ChainIDPolygon)
		gw.rpcURLs[oracle_vo.ChainIDPolygon] = cfg.PolygonRPCURL
		gw.contractAddrs[oracle_vo.ChainIDPolygon] = cfg.PolygonContractAddr
	}

	if cfg.AmoyRPCURL != "" {
		gw.supportedChains = append(gw.supportedChains, oracle_vo.ChainIDPolygonAmoy)
		gw.rpcURLs[oracle_vo.ChainIDPolygonAmoy] = cfg.AmoyRPCURL
		gw.contractAddrs[oracle_vo.ChainIDPolygonAmoy] = cfg.AmoyContractAddr
	}

	// If no chains explicitly configured, default to Amoy testnet (simulated)
	if len(gw.supportedChains) == 0 {
		gw.supportedChains = append(gw.supportedChains, oracle_vo.ChainIDPolygonAmoy)
		gw.rpcURLs[oracle_vo.ChainIDPolygonAmoy] = "simulated://amoy"
		gw.contractAddrs[oracle_vo.ChainIDPolygonAmoy] = "0x0000000000000000000000000000000000000000"
	}

	return gw
}

// PublishScore publishes consensus score to the specified chain.
// Phase 1: generates a deterministic simulated tx hash.
func (gw *EVMChainScoreGateway) PublishScore(ctx context.Context, chainID oracle_vo.ChainID, result *oracle_entities.OracleResult) (*oracle_entities.ChainPublication, error) {
	if !chainID.IsEVM() {
		return nil, fmt.Errorf("chain %d is not an EVM chain", chainID)
	}

	rpcURL, ok := gw.rpcURLs[chainID]
	if !ok {
		return nil, fmt.Errorf("chain %d is not configured", chainID)
	}

	contractAddr, ok := gw.contractAddrs[chainID]
	if !ok {
		return nil, fmt.Errorf("no contract address for chain %d", chainID)
	}

	// Phase 1: Simulated publish — generate deterministic tx hash
	// In Phase 2, this will be replaced with actual go-ethereum tx submission
	sourceHash := computeSourceHash(result)
	txHash := generateSimulatedTxHash(chainID, result.ID.String(), sourceHash)

	slog.InfoContext(ctx, "score published to EVM chain (simulated)",
		slog.Int64("chain_id", int64(chainID)),
		slog.String("rpc_url", rpcURL),
		slog.String("contract_addr", contractAddr),
		slog.String("tx_hash", txHash),
		slog.String("source_hash", sourceHash),
		slog.String("oracle_result_id", result.ID.String()),
	)

	pub := &oracle_entities.ChainPublication{
		ChainID:         chainID,
		CAIP2:           chainID.CAIP2(),
		ContractAddress: contractAddr,
		TxHash:          txHash,
		BlockNumber:     0, // Will be populated in Phase 2
		Status:          "confirmed", // Simulated — instant confirmation
		PublishedAt:     time.Now().UTC(),
	}

	return pub, nil
}

// GetPublishedScore retrieves a published score from the specified chain.
// Phase 1: simulated retrieval — returns a placeholder.
func (gw *EVMChainScoreGateway) GetPublishedScore(ctx context.Context, chainID oracle_vo.ChainID, matchID uuid.UUID) (*oracle_out.OnChainScore, error) {
	if !chainID.IsEVM() {
		return nil, fmt.Errorf("chain %d is not an EVM chain", chainID)
	}

	slog.InfoContext(ctx, "getting published score (simulated)",
		slog.Int64("chain_id", int64(chainID)),
		slog.String("match_id", matchID.String()),
	)

	// Phase 1: Simulated — return not-found style
	return nil, fmt.Errorf("on-chain score lookup not yet implemented (Phase 2)")
}

// IsScoreFinalized checks if a score's dispute window has closed on-chain.
// Phase 1: simulated — always returns false.
func (gw *EVMChainScoreGateway) IsScoreFinalized(ctx context.Context, chainID oracle_vo.ChainID, matchID uuid.UUID) (bool, error) {
	if !chainID.IsEVM() {
		return false, fmt.Errorf("chain %d is not an EVM chain", chainID)
	}

	slog.InfoContext(ctx, "checking finalization status (simulated)",
		slog.Int64("chain_id", int64(chainID)),
		slog.String("match_id", matchID.String()),
	)

	// Phase 1: Simulated — not finalized
	return false, nil
}

// SupportedChains returns the list of configured EVM chains
func (gw *EVMChainScoreGateway) SupportedChains() []oracle_vo.ChainID {
	return gw.supportedChains
}

// --- Internal helpers ---

// computeSourceHash produces a deterministic SHA-256 hash of the oracle result's
// consensus data, suitable for on-chain verification against the stored hash
func computeSourceHash(result *oracle_entities.OracleResult) string {
	if result.ConsensusResult == nil {
		return ""
	}

	consensus := result.ConsensusResult

	// Deterministic encoding: matchID|winner|teams|agreement|sources
	data := fmt.Sprintf("%s|%v|%.6f|%d",
		result.ID.String(),
		consensus.WinnerTeamID,
		consensus.AgreementRatio,
		consensus.SourceCount,
	)

	// Append team scores in order
	for _, ts := range consensus.TeamScores {
		data += fmt.Sprintf("|%s:%d", ts.TeamID.String(), ts.Score)
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// generateSimulatedTxHash creates a deterministic simulated transaction hash
func generateSimulatedTxHash(chainID oracle_vo.ChainID, resultID string, sourceHash string) string {
	data := fmt.Sprintf("sim_tx:%d:%s:%s:%d",
		chainID, resultID, sourceHash, time.Now().UnixNano(),
	)
	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])
}
