package oracle_entities

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// ScoreboardRegion defines a rectangular region for fixed-region scoreboard cropping
type ScoreboardRegion struct {
	X      int `json:"x" bson:"x"`
	Y      int `json:"y" bson:"y"`
	Width  int `json:"width" bson:"width"`
	Height int `json:"height" bson:"height"`
}

// OCRStreamConfig configures monitoring of a live stream for score extraction
type OCRStreamConfig struct {
	ID uuid.UUID `json:"_id" bson:"_id"`

	// Stream identification
	StreamURL       string                 `json:"stream_url" bson:"stream_url"`
	StreamPlatform  StreamPlatform         `json:"stream_platform" bson:"stream_platform"`
	VideoID         string                 `json:"video_id,omitempty" bson:"video_id,omitempty"` // YouTube video ID
	GameID          replay_common.GameIDKey `json:"game_id" bson:"game_id"`
	ExternalMatchID string                 `json:"external_match_id" bson:"external_match_id"`

	// Scoreboard region (fixed-region crop)
	ScoreboardRegion *ScoreboardRegion `json:"scoreboard_region,omitempty" bson:"scoreboard_region,omitempty"`

	// Capture settings
	CaptureIntervalSeconds int    `json:"capture_interval_seconds" bson:"capture_interval_seconds"`
	StreamQuality          string `json:"stream_quality,omitempty" bson:"stream_quality,omitempty"` // e.g. "720p", "480p"

	// Team hints — optional names to help OCR resolution
	TeamAHint string `json:"team_a_hint,omitempty" bson:"team_a_hint,omitempty"`
	TeamBHint string `json:"team_b_hint,omitempty" bson:"team_b_hint,omitempty"`

	// State management
	Status        OCRStreamStatus `json:"status" bson:"status"`
	LastCaptureAt *time.Time      `json:"last_capture_at,omitempty" bson:"last_capture_at,omitempty"`
	ErrorCount    int             `json:"error_count" bson:"error_count"`
	LastError     string          `json:"last_error,omitempty" bson:"last_error,omitempty"`

	// Linked oracle result
	OracleResultID *uuid.UUID `json:"oracle_result_id,omitempty" bson:"oracle_result_id,omitempty"`
}

// StreamPlatform identifies the streaming platform
type StreamPlatform string

const (
	StreamPlatformYouTube StreamPlatform = "youtube"
	StreamPlatformTwitch  StreamPlatform = "twitch"
	StreamPlatformOther   StreamPlatform = "other"
)

// OCRStreamStatus represents the lifecycle state of a stream config
type OCRStreamStatus string

const (
	OCRStreamStatusPending  OCRStreamStatus = "pending"   // Created but not started
	OCRStreamStatusActive   OCRStreamStatus = "active"    // Currently monitoring
	OCRStreamStatusPaused   OCRStreamStatus = "paused"    // Temporarily paused
	OCRStreamStatusComplete OCRStreamStatus = "completed" // Stream ended / match complete
	OCRStreamStatusFailed   OCRStreamStatus = "failed"    // Too many errors
)

// NewOCRStreamConfig creates a new OCR stream configuration
func NewOCRStreamConfig(
	streamURL string,
	gameID replay_common.GameIDKey,
	externalMatchID string,
) *OCRStreamConfig {
	config := &OCRStreamConfig{
		StreamURL:              streamURL,
		GameID:                 gameID,
		ExternalMatchID:        externalMatchID,
		CaptureIntervalSeconds: 10,
		Status:                 OCRStreamStatusPending,
	}

	// Auto-detect platform
	config.StreamPlatform = detectPlatform(streamURL)
	config.VideoID = extractVideoID(streamURL)

	return config
}

func (c *OCRStreamConfig) Validate() error {
	if c.StreamURL == "" {
		return fmt.Errorf("stream_url is required")
	}
	if c.GameID == "" {
		return fmt.Errorf("game_id is required")
	}
	if c.ExternalMatchID == "" {
		return fmt.Errorf("external_match_id is required")
	}
	if c.CaptureIntervalSeconds < 3 {
		return fmt.Errorf("capture_interval_seconds must be >= 3")
	}
	return nil
}

// Activate transitions the config to active monitoring
func (c *OCRStreamConfig) Activate() error {
	if c.Status != OCRStreamStatusPending && c.Status != OCRStreamStatusPaused {
		return fmt.Errorf("cannot activate from status %s", c.Status)
	}
	c.Status = OCRStreamStatusActive
	return nil
}

// Complete marks the stream as done
func (c *OCRStreamConfig) Complete() {
	c.Status = OCRStreamStatusComplete
}

// RecordError records an error and potentially marks as failed
func (c *OCRStreamConfig) RecordError(errMsg string) {
	c.ErrorCount++
	c.LastError = errMsg
	if c.ErrorCount >= 50 {
		c.Status = OCRStreamStatusFailed
	}
}

// RecordCapture updates the last capture timestamp
func (c *OCRStreamConfig) RecordCapture() {
	now := time.Now()
	c.LastCaptureAt = &now
}

// detectPlatform detects the streaming platform from the URL
func detectPlatform(url string) StreamPlatform {
	lower := strings.ToLower(url)
	switch {
	case strings.Contains(lower, "youtube.com") || strings.Contains(lower, "youtu.be"):
		return StreamPlatformYouTube
	case strings.Contains(lower, "twitch.tv"):
		return StreamPlatformTwitch
	default:
		return StreamPlatformOther
	}
}

// extractVideoID extracts the YouTube video ID from a URL
func extractVideoID(url string) string {
	// Handle youtube.com/watch?v=VIDEO_ID
	if idx := strings.Index(url, "v="); idx >= 0 {
		id := url[idx+2:]
		if ampIdx := strings.Index(id, "&"); ampIdx >= 0 {
			id = id[:ampIdx]
		}
		return id
	}
	// Handle youtu.be/VIDEO_ID
	if strings.Contains(url, "youtu.be/") {
		parts := strings.Split(url, "youtu.be/")
		if len(parts) == 2 {
			id := parts[1]
			if qIdx := strings.Index(id, "?"); qIdx >= 0 {
				id = id[:qIdx]
			}
			return id
		}
	}
	return ""
}
