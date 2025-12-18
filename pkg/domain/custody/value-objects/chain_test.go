package custody_vo

import (
	"testing"
)

// TestChainID_GetChainType verifies chain type classification
func TestChainID_GetChainType(t *testing.T) {
	tests := []struct {
		name     string
		chain    ChainID
		expected ChainType
	}{
		{"Solana Mainnet", ChainSolanaMainnet, ChainTypeSolana},
		{"Solana Devnet", ChainSolanaDevnet, ChainTypeSolana},
		{"Solana Testnet", ChainSolanaTestnet, ChainTypeSolana},
		{"Local Solana", ChainLocalSolana, ChainTypeSolana},
		{"Ethereum Mainnet", ChainEthereumMainnet, ChainTypeEVM},
		{"Polygon", ChainPolygon, ChainTypeEVM},
		{"Base", ChainBase, ChainTypeEVM},
		{"Arbitrum", ChainArbitrum, ChainTypeEVM},
		{"Optimism", ChainOptimism, ChainTypeEVM},
		{"Sepolia", ChainSepolia, ChainTypeEVM},
		{"Local EVM", ChainLocalEVM, ChainTypeEVM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.chain.GetChainType()
			if got != tt.expected {
				t.Errorf("GetChainType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestChainID_IsSolana verifies Solana chain detection
func TestChainID_IsSolana(t *testing.T) {
	solanaChains := []ChainID{ChainSolanaMainnet, ChainSolanaDevnet, ChainSolanaTestnet, ChainLocalSolana}
	evmChains := []ChainID{ChainEthereumMainnet, ChainPolygon, ChainBase, ChainArbitrum}

	for _, chain := range solanaChains {
		if !chain.IsSolana() {
			t.Errorf("Expected %s to be Solana chain", chain)
		}
	}

	for _, chain := range evmChains {
		if chain.IsSolana() {
			t.Errorf("Expected %s to NOT be Solana chain", chain)
		}
	}
}

// TestChainID_IsEVM verifies EVM chain detection
func TestChainID_IsEVM(t *testing.T) {
	evmChains := []ChainID{ChainEthereumMainnet, ChainPolygon, ChainBase, ChainArbitrum, ChainOptimism, ChainSepolia}
	solanaChains := []ChainID{ChainSolanaMainnet, ChainSolanaDevnet, ChainLocalSolana}

	for _, chain := range evmChains {
		if !chain.IsEVM() {
			t.Errorf("Expected %s to be EVM chain", chain)
		}
	}

	for _, chain := range solanaChains {
		if chain.IsEVM() {
			t.Errorf("Expected %s to NOT be EVM chain", chain)
		}
	}
}

// TestChainID_GetEVMChainID verifies numeric chain ID extraction
func TestChainID_GetEVMChainID(t *testing.T) {
	tests := []struct {
		name        string
		chain       ChainID
		expectedID  uint64
		expectError bool
	}{
		{"Ethereum Mainnet", ChainEthereumMainnet, 1, false},
		{"Polygon", ChainPolygon, 137, false},
		{"Base", ChainBase, 8453, false},
		{"Arbitrum", ChainArbitrum, 42161, false},
		{"Sepolia", ChainSepolia, 11155111, false},
		{"Hardhat Local", ChainLocalEVM, 31337, false},
		{"Solana (invalid)", ChainSolanaMainnet, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.chain.GetEVMChainID()
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if got != tt.expectedID {
				t.Errorf("GetEVMChainID() = %v, want %v", got, tt.expectedID)
			}
		})
	}
}

// TestChainID_String verifies human-readable chain names
func TestChainID_String(t *testing.T) {
	tests := []struct {
		chain    ChainID
		expected string
	}{
		{ChainSolanaMainnet, "Solana Mainnet"},
		{ChainEthereumMainnet, "Ethereum"},
		{ChainPolygon, "Polygon"},
		{ChainBase, "Base"},
		{ChainArbitrum, "Arbitrum One"},
		{ChainID("unknown:chain"), "unknown:chain"},
	}

	for _, tt := range tests {
		t.Run(string(tt.chain), func(t *testing.T) {
			got := tt.chain.String()
			if got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSupportedChains verifies production chains list
func TestSupportedChains(t *testing.T) {
	chains := SupportedChains()

	if len(chains) == 0 {
		t.Error("Expected non-empty supported chains list")
	}

	// Verify Solana is primary
	if chains[0] != ChainSolanaMainnet {
		t.Errorf("Expected Solana Mainnet as first chain, got %v", chains[0])
	}

	// Verify all production chains are present
	expectedChains := map[ChainID]bool{
		ChainSolanaMainnet:   true,
		ChainPolygon:         true,
		ChainBase:            true,
		ChainArbitrum:        true,
		ChainEthereumMainnet: true,
	}

	for _, chain := range chains {
		if !expectedChains[chain] {
			t.Errorf("Unexpected chain in supported list: %v", chain)
		}
		delete(expectedChains, chain)
	}

	if len(expectedChains) > 0 {
		t.Errorf("Missing expected chains: %v", expectedChains)
	}
}

// TestPrimaryChain verifies Solana is primary
func TestPrimaryChain(t *testing.T) {
	primary := PrimaryChain()
	if primary != ChainSolanaMainnet {
		t.Errorf("Expected Solana Mainnet as primary chain, got %v", primary)
	}
}

// TestAssetID_GetChainID verifies CAIP-19 parsing
func TestAssetID_GetChainID(t *testing.T) {
	tests := []struct {
		asset    AssetID
		expected ChainID
	}{
		{AssetUSDCSolana, ChainSolanaMainnet},
		{AssetUSDTSolana, ChainSolanaMainnet},
		{AssetUSDCPolygon, ChainPolygon},
		{AssetUSDCBase, ChainBase},
		{AssetUSDCArbitrum, ChainArbitrum},
	}

	for _, tt := range tests {
		t.Run(string(tt.asset), func(t *testing.T) {
			got := tt.asset.GetChainID()
			if got != tt.expected {
				t.Errorf("GetChainID() = %v, want %v", got, tt.expected)
			}
		})
	}
}
