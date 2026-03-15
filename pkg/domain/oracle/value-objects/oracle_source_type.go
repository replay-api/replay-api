package oracle_vo

import "fmt"

// OracleSourceType identifies the external data provider or OCR source
type OracleSourceType string

const (
	OracleSourcePandaScore   OracleSourceType = "pandascore"
	OracleSourceSteamWebAPI  OracleSourceType = "steam_web_api"
	OracleSourceFACEIT       OracleSourceType = "faceit_data_api"
	OracleSourceSportsDataIO OracleSourceType = "sportsdataio"
	OracleSourceGRID         OracleSourceType = "grid"
	OracleSourceAbios        OracleSourceType = "abios"
	OracleSourceOCRStream    OracleSourceType = "ocr_stream"
	OracleSourceOCRUpload    OracleSourceType = "ocr_screenshot"
)

// SourceConfidenceWeights maps each source to its default confidence weight
var SourceConfidenceWeights = map[OracleSourceType]float64{
	OracleSourceSteamWebAPI:  0.95, // Authoritative for Valve games
	OracleSourceGRID:         0.92, // Official tournament data
	OracleSourceFACEIT:       0.90, // Major platform
	OracleSourceSportsDataIO: 0.90, // Established data provider
	OracleSourcePandaScore:   0.85, // Broad coverage
	OracleSourceAbios:        0.85, // Broad coverage
	OracleSourceOCRStream:    0.70, // Live stream OCR
	OracleSourceOCRUpload:    0.50, // User-uploaded screenshots (lower trust)
}

// IsValid returns true if the source type is known
func (s OracleSourceType) IsValid() bool {
	_, ok := SourceConfidenceWeights[s]
	return ok
}

// String returns the string representation
func (s OracleSourceType) String() string {
	return string(s)
}

// ConfidenceWeight returns the default confidence weight for this source
func (s OracleSourceType) ConfidenceWeight() float64 {
	if w, ok := SourceConfidenceWeights[s]; ok {
		return w
	}
	return 0.0
}

// IsAutomated returns true if the source is programmatic
func (s OracleSourceType) IsAutomated() bool {
	return s != OracleSourceOCRUpload
}

// IsOCR returns true if the source is OCR-based
func (s OracleSourceType) IsOCR() bool {
	return s == OracleSourceOCRStream || s == OracleSourceOCRUpload
}

// Validate returns an error if the source type is invalid
func (s OracleSourceType) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid oracle source type: %s", s)
	}
	return nil
}
