package oracle_vo

// ChainID represents a blockchain network identifier
type ChainID int64

const (
	// ChainIDPolygon is the Polygon PoS mainnet chain ID
	ChainIDPolygon ChainID = 137

	// ChainIDPolygonAmoy is the Polygon Amoy testnet chain ID
	ChainIDPolygonAmoy ChainID = 80002

	// ChainIDSolanaMainnet represents Solana mainnet (internal identifier, no EVM chain ID)
	ChainIDSolanaMainnet ChainID = 900901

	// ChainIDSolanaDevnet represents Solana devnet
	ChainIDSolanaDevnet ChainID = 900902
)

// CAIP2 returns the CAIP-2 chain identifier string
func (c ChainID) CAIP2() string {
	switch c {
	case ChainIDPolygon:
		return "eip155:137"
	case ChainIDPolygonAmoy:
		return "eip155:80002"
	case ChainIDSolanaMainnet:
		return "solana:mainnet"
	case ChainIDSolanaDevnet:
		return "solana:devnet"
	default:
		return "unknown"
	}
}

// IsEVM returns true if this is an EVM-compatible chain
func (c ChainID) IsEVM() bool {
	return c == ChainIDPolygon || c == ChainIDPolygonAmoy
}

// IsSolana returns true if this is a Solana chain
func (c ChainID) IsSolana() bool {
	return c == ChainIDSolanaMainnet || c == ChainIDSolanaDevnet
}

// IsMainnet returns true if this is a mainnet chain
func (c ChainID) IsMainnet() bool {
	return c == ChainIDPolygon || c == ChainIDSolanaMainnet
}
