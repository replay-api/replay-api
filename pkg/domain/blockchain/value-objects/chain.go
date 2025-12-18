package blockchain_vo

import "fmt"

// ChainID represents a blockchain network identifier
type ChainID uint64

const (
	ChainIDEthereum ChainID = 1
	ChainIDPolygon  ChainID = 137
	ChainIDBase     ChainID = 8453
	ChainIDArbitrum ChainID = 42161

	// Testnets
	ChainIDSepolia       ChainID = 11155111
	ChainIDPolygonMumbai ChainID = 80001
	ChainIDBaseSepolia   ChainID = 84532

	// Local
	ChainIDHardhat ChainID = 31337
)

// ChainConfig holds configuration for a specific chain
type ChainConfig struct {
	ChainID     ChainID
	Name        string
	RPCURL      string
	ExplorerURL string
	NativeCoin  string
	IsTestnet   bool
	BlockTime   uint64 // Average block time in seconds
}

var chainConfigs = map[ChainID]ChainConfig{
	ChainIDEthereum: {
		ChainID:     ChainIDEthereum,
		Name:        "Ethereum Mainnet",
		NativeCoin:  "ETH",
		ExplorerURL: "https://etherscan.io",
		IsTestnet:   false,
		BlockTime:   12,
	},
	ChainIDPolygon: {
		ChainID:     ChainIDPolygon,
		Name:        "Polygon",
		NativeCoin:  "MATIC",
		ExplorerURL: "https://polygonscan.com",
		IsTestnet:   false,
		BlockTime:   2,
	},
	ChainIDBase: {
		ChainID:     ChainIDBase,
		Name:        "Base",
		NativeCoin:  "ETH",
		ExplorerURL: "https://basescan.org",
		IsTestnet:   false,
		BlockTime:   2,
	},
	ChainIDArbitrum: {
		ChainID:     ChainIDArbitrum,
		Name:        "Arbitrum One",
		NativeCoin:  "ETH",
		ExplorerURL: "https://arbiscan.io",
		IsTestnet:   false,
		BlockTime:   1,
	},
	ChainIDSepolia: {
		ChainID:     ChainIDSepolia,
		Name:        "Sepolia",
		NativeCoin:  "ETH",
		ExplorerURL: "https://sepolia.etherscan.io",
		IsTestnet:   true,
		BlockTime:   12,
	},
	ChainIDHardhat: {
		ChainID:     ChainIDHardhat,
		Name:        "Hardhat Local",
		NativeCoin:  "ETH",
		ExplorerURL: "",
		IsTestnet:   true,
		BlockTime:   1,
	},
}

// GetChainConfig returns config for a chain ID
func GetChainConfig(chainID ChainID) (ChainConfig, error) {
	config, ok := chainConfigs[chainID]
	if !ok {
		return ChainConfig{}, fmt.Errorf("unknown chain ID: %d", chainID)
	}
	return config, nil
}

// IsSupported checks if a chain is supported
func (c ChainID) IsSupported() bool {
	_, ok := chainConfigs[c]
	return ok
}

// IsMainnet returns true if this is a mainnet chain
func (c ChainID) IsMainnet() bool {
	config, ok := chainConfigs[c]
	return ok && !config.IsTestnet
}

// String returns the chain name
func (c ChainID) String() string {
	config, ok := chainConfigs[c]
	if !ok {
		return fmt.Sprintf("Unknown(%d)", c)
	}
	return config.Name
}

// SupportedMainnets returns all supported mainnet chain IDs
func SupportedMainnets() []ChainID {
	return []ChainID{ChainIDPolygon, ChainIDBase, ChainIDArbitrum}
}

// PrimaryChain returns the primary chain for transactions (Polygon)
func PrimaryChain() ChainID {
	return ChainIDPolygon
}
