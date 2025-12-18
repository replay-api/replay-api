package blockchain_vo

import (
	"testing"
)

// TestChainID_IsSupported verifies chain support detection
func TestChainID_IsSupported(t *testing.T) {
	supportedChains := []ChainID{
		ChainIDEthereum,
		ChainIDPolygon,
		ChainIDBase,
		ChainIDArbitrum,
		ChainIDSepolia,
		ChainIDHardhat,
	}

	for _, chain := range supportedChains {
		if !chain.IsSupported() {
			t.Errorf("Expected chain %d to be supported", chain)
		}
	}

	unsupported := ChainID(999999)
	if unsupported.IsSupported() {
		t.Error("Expected unknown chain to not be supported")
	}
}

// TestChainID_IsMainnet verifies mainnet detection
func TestChainID_IsMainnet(t *testing.T) {
	tests := []struct {
		chain     ChainID
		isMainnet bool
	}{
		{ChainIDEthereum, true},
		{ChainIDPolygon, true},
		{ChainIDBase, true},
		{ChainIDArbitrum, true},
		{ChainIDSepolia, false},
		{ChainIDPolygonMumbai, false},
		{ChainIDHardhat, false},
	}

	for _, tt := range tests {
		t.Run(tt.chain.String(), func(t *testing.T) {
			if got := tt.chain.IsMainnet(); got != tt.isMainnet {
				t.Errorf("IsMainnet() = %v, want %v", got, tt.isMainnet)
			}
		})
	}
}

// TestChainID_String verifies chain name formatting
func TestChainID_String(t *testing.T) {
	tests := []struct {
		chain    ChainID
		expected string
	}{
		{ChainIDEthereum, "Ethereum Mainnet"},
		{ChainIDPolygon, "Polygon"},
		{ChainIDBase, "Base"},
		{ChainIDArbitrum, "Arbitrum One"},
		{ChainIDSepolia, "Sepolia"},
		{ChainIDHardhat, "Hardhat Local"},
		{ChainID(999999), "Unknown(999999)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.chain.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGetChainConfig verifies chain configuration retrieval
func TestGetChainConfig(t *testing.T) {
	tests := []struct {
		chain       ChainID
		expectError bool
	}{
		{ChainIDEthereum, false},
		{ChainIDPolygon, false},
		{ChainIDBase, false},
		{ChainIDArbitrum, false},
		{ChainID(999999), true},
	}

	for _, tt := range tests {
		t.Run(tt.chain.String(), func(t *testing.T) {
			config, err := GetChainConfig(tt.chain)
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
			if config.ChainID != tt.chain {
				t.Errorf("ChainID mismatch: got %d, want %d", config.ChainID, tt.chain)
			}
		})
	}
}

// TestChainConfig_Properties verifies chain config properties
func TestChainConfig_Properties(t *testing.T) {
	tests := []struct {
		chain       ChainID
		nativeCoin  string
		isTestnet   bool
		hasExplorer bool
	}{
		{ChainIDEthereum, "ETH", false, true},
		{ChainIDPolygon, "MATIC", false, true},
		{ChainIDBase, "ETH", false, true},
		{ChainIDArbitrum, "ETH", false, true},
		{ChainIDSepolia, "ETH", true, true},
		{ChainIDHardhat, "ETH", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.chain.String(), func(t *testing.T) {
			config, err := GetChainConfig(tt.chain)
			if err != nil {
				t.Fatalf("Failed to get config: %v", err)
			}

			if config.NativeCoin != tt.nativeCoin {
				t.Errorf("NativeCoin = %s, want %s", config.NativeCoin, tt.nativeCoin)
			}
			if config.IsTestnet != tt.isTestnet {
				t.Errorf("IsTestnet = %v, want %v", config.IsTestnet, tt.isTestnet)
			}
			hasExplorer := config.ExplorerURL != ""
			if hasExplorer != tt.hasExplorer {
				t.Errorf("HasExplorer = %v, want %v", hasExplorer, tt.hasExplorer)
			}
		})
	}
}

// TestSupportedMainnets verifies mainnet list
func TestSupportedMainnets(t *testing.T) {
	mainnets := SupportedMainnets()

	if len(mainnets) == 0 {
		t.Error("Expected non-empty mainnet list")
	}

	for _, chain := range mainnets {
		if !chain.IsMainnet() {
			t.Errorf("Expected %s to be mainnet", chain.String())
		}
	}

	// Verify expected chains are present
	expected := map[ChainID]bool{
		ChainIDPolygon:  false,
		ChainIDBase:     false,
		ChainIDArbitrum: false,
	}

	for _, chain := range mainnets {
		expected[chain] = true
	}

	for chain, found := range expected {
		if !found {
			t.Errorf("Expected %s in supported mainnets", chain.String())
		}
	}
}

// TestPrimaryChain_Blockchain verifies primary chain is Polygon
func TestPrimaryChain_Blockchain(t *testing.T) {
	primary := PrimaryChain()
	if primary != ChainIDPolygon {
		t.Errorf("Expected Polygon as primary chain, got %s", primary.String())
	}
}

// TestChainConfig_BlockTime verifies block time configuration
func TestChainConfig_BlockTime(t *testing.T) {
	tests := []struct {
		chain         ChainID
		minBlockTime  uint64
		maxBlockTime  uint64
	}{
		{ChainIDEthereum, 10, 15},
		{ChainIDPolygon, 1, 3},
		{ChainIDBase, 1, 3},
		{ChainIDArbitrum, 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.chain.String(), func(t *testing.T) {
			config, err := GetChainConfig(tt.chain)
			if err != nil {
				t.Fatalf("Failed to get config: %v", err)
			}

			if config.BlockTime < tt.minBlockTime || config.BlockTime > tt.maxBlockTime {
				t.Errorf("BlockTime = %d, want between %d and %d",
					config.BlockTime, tt.minBlockTime, tt.maxBlockTime)
			}
		})
	}
}
