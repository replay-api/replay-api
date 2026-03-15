package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_out "github.com/replay-api/replay-api/pkg/domain/analytics/ports/out"
	"github.com/replay-api/replay-api/pkg/infra/metrics"
	shared "github.com/resource-ownership/go-common/pkg/common"
	segkafka "github.com/segmentio/kafka-go"
)

// AnalyticsEventConsumer processes analytics events to compute aggregations.
// It consumes EntityViewed events from Kafka, updates ViewStatistics and ViewerInsight
// documents in MongoDB, and records Prometheus metrics.
type AnalyticsEventConsumer struct {
	consumer      *Consumer
	statsWriter   analytics_out.ViewStatisticsWriter
	statsReader   analytics_out.ViewStatisticsReader
	insightWriter analytics_out.ViewerInsightWriter
	insightReader analytics_out.ViewerInsightReader
}

// AnalyticsEventConsumerConfig holds configuration for the analytics consumer
type AnalyticsEventConsumerConfig struct {
	GroupID       string
	StatsWriter   analytics_out.ViewStatisticsWriter
	StatsReader   analytics_out.ViewStatisticsReader
	InsightWriter analytics_out.ViewerInsightWriter
	InsightReader analytics_out.ViewerInsightReader
}

// NewAnalyticsEventConsumer creates a consumer that processes entity.viewed events
// and computes aggregations into view_statistics and viewer_insights collections.
func NewAnalyticsEventConsumer(client *Client, config *AnalyticsEventConsumerConfig) *AnalyticsEventConsumer {
	topics := []string{TopicAnalyticsEntityViewed}
	consumerConfig := DefaultConsumerConfig(config.GroupID, topics)
	consumer := NewConsumer(client, consumerConfig)

	aec := &AnalyticsEventConsumer{
		consumer:      consumer,
		statsWriter:   config.StatsWriter,
		statsReader:   config.StatsReader,
		insightWriter: config.InsightWriter,
		insightReader: config.InsightReader,
	}

	consumer.RegisterHandler(TopicAnalyticsEntityViewed, aec.handleEntityViewed)

	return aec
}

func (aec *AnalyticsEventConsumer) handleEntityViewed(ctx context.Context, msg *segkafka.Message) error {
	start := time.Now()

	var event EntityViewedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal entity viewed event",
			"error", err,
			"offset", msg.Offset,
			"partition", msg.Partition,
		)
		metrics.AnalyticsEventsProcessedTotal.WithLabelValues("unmarshal_error", "unknown").Inc()
		return fmt.Errorf("failed to unmarshal entity viewed event: %w", err)
	}

	entityType := analytics_entities.EntityTypeKey(event.EntityType)

	slog.DebugContext(ctx, "Processing entity viewed event",
		"entity_id", event.EntityID,
		"entity_type", event.EntityType,
		"viewer_id", event.ViewerID,
		"geo_region", event.GeoRegion,
	)

	// Update ViewStatistics aggregation
	if err := aec.updateViewStatistics(ctx, &event, entityType); err != nil {
		slog.ErrorContext(ctx, "Failed to update view statistics",
			"error", err,
			"entity_id", event.EntityID,
		)
		metrics.AnalyticsEventsProcessedTotal.WithLabelValues("stats_error", event.EntityType).Inc()
		return fmt.Errorf("failed to update view statistics: %w", err)
	}

	// Update ViewerInsight if authenticated viewer
	if event.ViewerID != nil && event.ViewerType != "anonymous" {
		if err := aec.updateViewerInsight(ctx, &event, entityType); err != nil {
			slog.ErrorContext(ctx, "Failed to update viewer insight",
				"error", err,
				"entity_id", event.EntityID,
				"viewer_id", event.ViewerID,
			)
			// Non-fatal: stats were updated, viewer insight is supplementary
			metrics.AnalyticsEventsProcessedTotal.WithLabelValues("insight_error", event.EntityType).Inc()
		}
	}

	// Record Prometheus metrics
	duration := time.Since(start)
	metrics.AnalyticsEventsProcessedTotal.WithLabelValues("success", event.EntityType).Inc()
	metrics.AnalyticsEventProcessingDuration.WithLabelValues(event.EntityType).Observe(duration.Seconds())
	metrics.EntityViewsRecordedTotal.WithLabelValues(event.EntityType, event.ReferrerType).Inc()

	if event.ViewerID != nil {
		metrics.EntityUniqueViewersTotal.WithLabelValues(event.EntityType).Inc()
	}

	slog.DebugContext(ctx, "Entity viewed event processed",
		"entity_id", event.EntityID,
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}

// updateViewStatistics reads-modify-writes the aggregated statistics for an entity.
func (aec *AnalyticsEventConsumer) updateViewStatistics(ctx context.Context, event *EntityViewedEvent, entityType analytics_entities.EntityTypeKey) error {
	// Try to fetch existing stats
	existing, err := aec.statsReader.GetByEntity(ctx, event.EntityID, entityType)
	if err != nil {
		return fmt.Errorf("failed to read existing stats: %w", err)
	}

	var stats *analytics_entities.ViewStatistics
	if existing != nil {
		stats = existing
	} else {
		stats = analytics_entities.NewViewStatistics(event.EntityID, entityType)
	}

	// Increment counters
	stats.TotalViews++

	// Track viewer type
	if event.ViewerType == "authenticated" {
		stats.AuthenticatedViews++
	} else {
		stats.AnonymousViews++
	}

	// Views by day
	dayKey := event.ViewedAt.Format("2006-01-02")
	if stats.ViewsByDay == nil {
		stats.ViewsByDay = make(map[string]int64)
	}
	stats.ViewsByDay[dayKey]++

	// Views by region
	if event.GeoRegion != "" {
		if stats.ViewsByRegion == nil {
			stats.ViewsByRegion = make(map[string]int64)
		}
		stats.ViewsByRegion[event.GeoRegion]++
	}

	// Views by device
	if event.DeviceType != "" {
		if stats.ViewsByDevice == nil {
			stats.ViewsByDevice = make(map[string]int64)
		}
		stats.ViewsByDevice[event.DeviceType]++
	}

	// Views by referrer
	if event.ReferrerType != "" {
		if stats.ViewsByReferrer == nil {
			stats.ViewsByReferrer = make(map[string]int64)
		}
		stats.ViewsByReferrer[event.ReferrerType]++
	}

	// Compute trend (compare last 7 days vs previous 7 days)
	stats.TrendDirection, stats.TrendPercentage = computeTrend(stats.ViewsByDay)

	// Track peak day
	var peakCount int64
	for day, count := range stats.ViewsByDay {
		if count > peakCount {
			peakCount = count
			stats.PeakViewsDay = day
		}
	}

	// Estimate unique viewers from authenticated views
	if event.ViewerID != nil {
		stats.UniqueViewers = stats.AuthenticatedViews + (stats.AnonymousViews / 3)
	}

	stats.LastComputedAt = time.Now()
	stats.UpdatedAt = time.Now()

	return aec.statsWriter.Upsert(ctx, stats)
}

// updateViewerInsight creates or updates a per-viewer insight record.
func (aec *AnalyticsEventConsumer) updateViewerInsight(ctx context.Context, event *EntityViewedEvent, entityType analytics_entities.EntityTypeKey) error {
	if event.ViewerID == nil {
		return nil
	}

	// Build a minimal search to check for existing insight
	search := shared.Search{
		ResultOptions: shared.SearchResultOptions{
			Skip:  0,
			Limit: 50,
		},
		SortOptions: []shared.SortableField{
			{Field: "last_viewed_at", Direction: shared.DescendingIDKey},
		},
	}

	// Note: GetByEntity already filters by entity_id + entity_type + is_anonymous=false
	insights, _, err := aec.insightReader.GetByEntity(ctx, event.EntityID, entityType, search)
	if err != nil {
		return fmt.Errorf("failed to read existing insight: %w", err)
	}

	// Find existing insight for this specific viewer
	var existingInsight *analytics_entities.ViewerInsight
	for i := range insights {
		if insights[i].ViewerID == *event.ViewerID {
			existingInsight = &insights[i]
			break
		}
	}

	if existingInsight != nil {
		existingInsight.IncrementView()
		return aec.insightWriter.Upsert(ctx, existingInsight)
	}

	// New viewer
	insight := analytics_entities.NewViewerInsight(
		event.EntityID,
		entityType,
		*event.ViewerID,
		"", // nickname — enriched asynchronously via player lookup
		"", // avatar — enriched asynchronously
		"", // game_id
		false,
	)

	return aec.insightWriter.Upsert(ctx, insight)
}

// Start begins consuming analytics events
func (aec *AnalyticsEventConsumer) Start(ctx context.Context) error {
	return aec.consumer.Start(ctx)
}

// Close closes the consumer
func (aec *AnalyticsEventConsumer) Close() error {
	return aec.consumer.Close()
}

// computeTrend calculates the trend direction and percentage change
// by comparing the last 7 days vs the previous 7 days.
func computeTrend(viewsByDay map[string]int64) (analytics_entities.TrendKey, float64) {
	now := time.Now()
	var recentTotal, previousTotal int64

	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		recentTotal += viewsByDay[day]

		prevDay := now.AddDate(0, 0, -(i + 7)).Format("2006-01-02")
		previousTotal += viewsByDay[prevDay]
	}

	if previousTotal == 0 {
		if recentTotal > 0 {
			return analytics_entities.TrendUp, 100.0
		}
		return analytics_entities.TrendStable, 0.0
	}

	change := float64(recentTotal-previousTotal) / float64(previousTotal) * 100.0

	if change > 5 {
		return analytics_entities.TrendUp, change
	} else if change < -5 {
		return analytics_entities.TrendDown, change
	}
	return analytics_entities.TrendStable, change
}
