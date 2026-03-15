package oracle_vo

// ConsensusPolicy defines the parameters for consensus evaluation
type ConsensusPolicy struct {
	MinSources        int     `json:"min_sources" bson:"min_sources"`               // Minimum number of sources required
	MinConfidence     float64 `json:"min_confidence" bson:"min_confidence"`         // Minimum agreement ratio [0.0, 1.0]
	WinnerWeight      float64 `json:"winner_weight" bson:"winner_weight"`           // Weight for winner consensus (default 0.60)
	SeriesScoreWeight float64 `json:"series_score_weight" bson:"series_score_weight"` // Weight for series score consensus (default 0.30)
	GameScoreWeight   float64 `json:"game_score_weight" bson:"game_score_weight"`   // Weight for per-game consensus (default 0.10)
}

// StrictPolicy requires 3 sources with 90% agreement — for tournament finals
func StrictPolicy() ConsensusPolicy {
	return ConsensusPolicy{
		MinSources:        3,
		MinConfidence:     0.90,
		WinnerWeight:      0.60,
		SeriesScoreWeight: 0.30,
		GameScoreWeight:   0.10,
	}
}

// StandardPolicy requires 3 sources with 75% agreement — default for ranked
func StandardPolicy() ConsensusPolicy {
	return ConsensusPolicy{
		MinSources:        3,
		MinConfidence:     0.75,
		WinnerWeight:      0.60,
		SeriesScoreWeight: 0.30,
		GameScoreWeight:   0.10,
	}
}

// RelaxedPolicy requires 2 sources with 60% agreement — for casual matches
func RelaxedPolicy() ConsensusPolicy {
	return ConsensusPolicy{
		MinSources:        2,
		MinConfidence:     0.60,
		WinnerWeight:      0.60,
		SeriesScoreWeight: 0.30,
		GameScoreWeight:   0.10,
	}
}

// OCROnlyPolicy requires 1 source with 50% confidence — for OCR-only ingestion
// from YouTube tournament streams where no API providers are available.
// Consensus is trivially reached with a single submission; the resulting
// ConfidenceLevel will be gated by the source's weight (0.70 for stream OCR).
func OCROnlyPolicy() ConsensusPolicy {
	return ConsensusPolicy{
		MinSources:        1,
		MinConfidence:     0.50,
		WinnerWeight:      0.60,
		SeriesScoreWeight: 0.30,
		GameScoreWeight:   0.10,
	}
}
