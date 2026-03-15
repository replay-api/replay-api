package oracle_services

import (
	"sync"
	"time"

	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
)

// ProviderReliability tracks the reliability metrics of a single provider
type ProviderReliability struct {
	SourceType       oracle_vo.OracleSourceType `json:"source_type" bson:"source_type"`
	TotalSubmissions int64                       `json:"total_submissions" bson:"total_submissions"`
	AccurateCount    int64                       `json:"accurate_count" bson:"accurate_count"`
	InaccurateCount  int64                       `json:"inaccurate_count" bson:"inaccurate_count"`
	AnomalyCount     int64                       `json:"anomaly_count" bson:"anomaly_count"`
	AverageLatencyMs int64                       `json:"average_latency_ms" bson:"average_latency_ms"`
	LastUpdated      time.Time                   `json:"last_updated" bson:"last_updated"`
}

// AccuracyRate returns the ratio of accurate submissions
func (r *ProviderReliability) AccuracyRate() float64 {
	if r.TotalSubmissions == 0 {
		return 1.0 // Assume reliable until proven otherwise
	}
	return float64(r.AccurateCount) / float64(r.TotalSubmissions)
}

// AnomalyRate returns the ratio of anomalous submissions
func (r *ProviderReliability) AnomalyRate() float64 {
	if r.TotalSubmissions == 0 {
		return 0.0
	}
	return float64(r.AnomalyCount) / float64(r.TotalSubmissions)
}

// ProviderReliabilityTracker tracks reliability across all providers
type ProviderReliabilityTracker struct {
	mu          sync.RWMutex
	reliability map[oracle_vo.OracleSourceType]*ProviderReliability
}

// NewProviderReliabilityTracker creates a new tracker
func NewProviderReliabilityTracker() *ProviderReliabilityTracker {
	return &ProviderReliabilityTracker{
		reliability: make(map[oracle_vo.OracleSourceType]*ProviderReliability),
	}
}

// GetReliability returns the reliability metrics for a given source
func (t *ProviderReliabilityTracker) GetReliability(source oracle_vo.OracleSourceType) *ProviderReliability {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.reliability[source]
}

// RecordAccurate records an accurate submission from a provider
func (t *ProviderReliabilityTracker) RecordAccurate(source oracle_vo.OracleSourceType, latencyMs int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	r := t.getOrCreate(source)
	r.TotalSubmissions++
	r.AccurateCount++
	r.AverageLatencyMs = (r.AverageLatencyMs*(r.TotalSubmissions-1) + latencyMs) / r.TotalSubmissions
	r.LastUpdated = time.Now().UTC()
}

// RecordInaccurate records an inaccurate submission from a provider
func (t *ProviderReliabilityTracker) RecordInaccurate(source oracle_vo.OracleSourceType) {
	t.mu.Lock()
	defer t.mu.Unlock()

	r := t.getOrCreate(source)
	r.TotalSubmissions++
	r.InaccurateCount++
	r.LastUpdated = time.Now().UTC()
}

// RecordAnomaly records an anomalous submission from a provider
func (t *ProviderReliabilityTracker) RecordAnomaly(source oracle_vo.OracleSourceType) {
	t.mu.Lock()
	defer t.mu.Unlock()

	r := t.getOrCreate(source)
	r.AnomalyCount++
	r.LastUpdated = time.Now().UTC()
}

// GetAllReliability returns a snapshot of all provider reliability metrics
func (t *ProviderReliabilityTracker) GetAllReliability() map[oracle_vo.OracleSourceType]*ProviderReliability {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[oracle_vo.OracleSourceType]*ProviderReliability, len(t.reliability))
	for k, v := range t.reliability {
		copy := *v
		result[k] = &copy
	}
	return result
}

func (t *ProviderReliabilityTracker) getOrCreate(source oracle_vo.OracleSourceType) *ProviderReliability {
	r, ok := t.reliability[source]
	if !ok {
		r = &ProviderReliability{
			SourceType:  source,
			LastUpdated: time.Now().UTC(),
		}
		t.reliability[source] = r
	}
	return r
}
