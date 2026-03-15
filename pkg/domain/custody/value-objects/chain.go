package custody_vo

import "fmt"

// ChainType represents the blockchain protocol type
type ChainType string

const (
	ChainTypeSolana   ChainType = "Solana"
	ChainTypeEVM      ChainType = "EVM"
	ChainTypeBitcoin  ChainType = "Bitcoin"
	ChainTypeCosmos   ChainType = "Cosmos"
)

// ChainID represents a blockchain network
type ChainID string

const (
	// Solana Networks (Primary)
	ChainSolanaMainnet ChainID = "solana:mainnet"
	ChainSolanaDevnet  ChainID = "solana:devnet"
	ChainSolanaTestnet ChainID = "solana:testnet"

	// EVM Networks
	ChainEthereumMainnet ChainID = "eip155:1"
	ChainPolygon         ChainID = "eip155:137"
	ChainBase            ChainID = "eip155:8453"
	ChainArbitrum        ChainID = "eip155:42161"
	ChainOptimism        ChainID = "eip155:10"
	ChainAvalanche       ChainID = "eip155:43114"
	ChainBSC             ChainID = "eip155:56"

	// Testnets
	ChainSepolia        ChainID = "eip155:11155111"
	ChainPolygonMumbai  ChainID = "eip155:80001"
	ChainBaseSepolia    ChainID = "eip155:84532"

	// Bitcoin Networks (BIP-122 CAIP-2)
	ChainBitcoinMainnet ChainID = "bip122:000000000019d6689c085ae165831e93"
	ChainBitcoinTestnet ChainID = "bip122:000000000933ea01ad0ee984209779ba"
	ChainBitcoinSignet  ChainID = "bip122:00000008819873e925422c1ff0f99f7c"
	ChainLightning      ChainID = "lightning:mainnet"
	ChainLightningTestnet ChainID = "lightning:testnet"

	// Local
	ChainLocalSolana  ChainID = "solana:localnet"
	ChainLocalEVM     ChainID = "eip155:31337"
	ChainLocalBitcoin ChainID = "bip122:regtest"
)

// ChainConfig holds comprehensive chain configuration
type ChainConfig struct {
	ChainID          ChainID
	ChainType        ChainType
	Name             string
	NativeCurrency   string
	NativeDecimals   uint8
	RPCURL           string
	WSURL            string
	ExplorerURL      string
	ExplorerAPIURL   string
	IsTestnet        bool
	SupportsAA       bool     // ERC-4337 Account Abstraction
	EntryPointAddr   string   // ERC-4337 EntryPoint contract
	PaymasterAddr    string   // Platform Paymaster
	BlockTime        uint64   // Average block time in milliseconds
	Confirmations    uint64   // Required confirmations for finality
	MaxGasPrice      uint64   // Safety cap in native units
	PriorityFeeRange [2]uint64 // Min/Max priority fee
}

// SupportedChains returns all production-ready chains
func SupportedChains() []ChainID {
	return []ChainID{
		ChainSolanaMainnet,
		ChainPolygon,
		ChainBase,
		ChainArbitrum,
		ChainEthereumMainnet,
		ChainBitcoinMainnet,
		ChainLightning,
	}
}

// PrimaryChain returns Solana as the primary chain
func PrimaryChain() ChainID {
	return ChainSolanaMainnet
}

// GetChainType returns the protocol type for a chain
func (c ChainID) GetChainType() ChainType {
	switch c {
	case ChainSolanaMainnet, ChainSolanaDevnet, ChainSolanaTestnet, ChainLocalSolana:
		return ChainTypeSolana
	case ChainBitcoinMainnet, ChainBitcoinTestnet, ChainBitcoinSignet, ChainLocalBitcoin:
		return ChainTypeBitcoin
	case ChainLightning, ChainLightningTestnet:
		return ChainTypeBitcoin
	default:
		return ChainTypeEVM
	}
}

// IsBitcoin checks if chain is Bitcoin-based (includes Lightning)
func (c ChainID) IsBitcoin() bool {
	return c.GetChainType() == ChainTypeBitcoin
}

// IsLightning checks if chain is Lightning Network
func (c ChainID) IsLightning() bool {
	return c == ChainLightning || c == ChainLightningTestnet
}

// IsSolana checks if chain is Solana-based
func (c ChainID) IsSolana() bool {
	return c.GetChainType() == ChainTypeSolana
}

// IsEVM checks if chain is EVM-compatible
func (c ChainID) IsEVM() bool {
	return c.GetChainType() == ChainTypeEVM
}

// GetEVMChainID extracts the numeric chain ID for EVM chains
func (c ChainID) GetEVMChainID() (uint64, error) {
	if !c.IsEVM() {
		return 0, fmt.Errorf("not an EVM chain: %s", c)
	}
	var id uint64
	_, err := fmt.Sscanf(string(c), "eip155:%d", &id)
	return id, err
}

// String returns chain name
func (c ChainID) String() string {
	names := map[ChainID]string{
		ChainSolanaMainnet:    "Solana Mainnet",
		ChainSolanaDevnet:     "Solana Devnet",
		ChainEthereumMainnet:  "Ethereum",
		ChainPolygon:          "Polygon",
		ChainBase:             "Base",
		ChainArbitrum:         "Arbitrum One",
		ChainOptimism:         "Optimism",
		ChainBitcoinMainnet:   "Bitcoin Mainnet",
		ChainBitcoinTestnet:   "Bitcoin Testnet",
		ChainBitcoinSignet:    "Bitcoin Signet",
		ChainLightning:        "Lightning Network",
		ChainLightningTestnet: "Lightning Testnet",
	}
	if name, ok := names[c]; ok {
		return name
	}
	return string(c)
}

// TokenStandard represents the token standard on each chain
type TokenStandard string

const (
	TokenStandardSPL     TokenStandard = "SPL"      // Solana Program Library
	TokenStandardERC20   TokenStandard = "ERC20"    // Ethereum ERC-20
	TokenStandardERC721  TokenStandard = "ERC721"   // Ethereum NFT
	TokenStandardERC1155 TokenStandard = "ERC1155"  // Ethereum Multi-token
)

// AssetID represents a unique asset across chains (CAIP-19)
type AssetID string

// Common stablecoin asset IDs
const (
	AssetUSDCSolana   AssetID = "solana:mainnet/spl:EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	AssetUSDTSolana   AssetID = "solana:mainnet/spl:Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
	AssetUSDCPolygon  AssetID = "eip155:137/erc20:0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359"
	AssetUSDCBase     AssetID = "eip155:8453/erc20:0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
	AssetUSDCArbitrum AssetID = "eip155:42161/erc20:0xaf88d065e77c8cC2239327C5EDb3A432268e5831"
)

func (a AssetID) GetChainID() ChainID {
	// Parse CAIP-19 format: chain:network/standard:address
	// Example: "solana:mainnet/spl:EPjF..." -> "solana:mainnet"
	assetStr := string(a)
	slashIdx := -1
	for i, c := range assetStr {
		if c == '/' {
			slashIdx = i
			break
		}
	}
	if slashIdx == -1 {
		return ChainID(assetStr)
	}
	return ChainID(assetStr[:slashIdx])
}
