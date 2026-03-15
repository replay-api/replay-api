package oracle_entities

import (
	"testing"

	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOCRStreamConfig(t *testing.T) {
	config := NewOCRStreamConfig(
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		replay_common.CS2_GAME_ID,
		"ext-match-123",
	)

	require.NotNil(t, config)
	assert.Equal(t, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", config.StreamURL)
	assert.Equal(t, replay_common.CS2_GAME_ID, config.GameID)
	assert.Equal(t, "ext-match-123", config.ExternalMatchID)
	assert.Equal(t, StreamPlatformYouTube, config.StreamPlatform)
	assert.Equal(t, "dQw4w9WgXcQ", config.VideoID)
	assert.Equal(t, OCRStreamStatusPending, config.Status)
	assert.Equal(t, 10, config.CaptureIntervalSeconds)
}

func TestOCRStreamConfig_DetectPlatform(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected StreamPlatform
	}{
		{"YouTube standard", "https://www.youtube.com/watch?v=abc123", StreamPlatformYouTube},
		{"YouTube short", "https://youtu.be/abc123", StreamPlatformYouTube},
		{"Twitch", "https://www.twitch.tv/esl_csgo", StreamPlatformTwitch},
		{"Other", "https://example.com/stream", StreamPlatformOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewOCRStreamConfig(tt.url, replay_common.CS2_GAME_ID, "match-1")
			assert.Equal(t, tt.expected, config.StreamPlatform)
		})
	}
}

func TestOCRStreamConfig_ExtractVideoID(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"YouTube standard", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"YouTube with params", "https://www.youtube.com/watch?v=abc123&t=42", "abc123"},
		{"YouTube short", "https://youtu.be/abc123", "abc123"},
		{"YouTube short with params", "https://youtu.be/abc123?t=42", "abc123"},
		{"Non-YouTube", "https://twitch.tv/stream", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewOCRStreamConfig(tt.url, replay_common.CS2_GAME_ID, "match-1")
			assert.Equal(t, tt.expected, config.VideoID)
		})
	}
}

func TestOCRStreamConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
		assert.NoError(t, config.Validate())
	})

	t.Run("missing stream URL", func(t *testing.T) {
		config := NewOCRStreamConfig("", replay_common.CS2_GAME_ID, "match-1")
		assert.Error(t, config.Validate())
	})

	t.Run("missing game ID", func(t *testing.T) {
		config := &OCRStreamConfig{
			StreamURL:       "https://youtube.com/watch?v=test",
			ExternalMatchID: "match-1",
			CaptureIntervalSeconds: 10,
		}
		assert.Error(t, config.Validate())
	})

	t.Run("missing external match ID", func(t *testing.T) {
		config := &OCRStreamConfig{
			StreamURL:       "https://youtube.com/watch?v=test",
			GameID:          replay_common.CS2_GAME_ID,
			CaptureIntervalSeconds: 10,
		}
		assert.Error(t, config.Validate())
	})

	t.Run("interval too low", func(t *testing.T) {
		config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
		config.CaptureIntervalSeconds = 1
		assert.Error(t, config.Validate())
	})
}

func TestOCRStreamConfig_Activate(t *testing.T) {
	t.Run("from pending", func(t *testing.T) {
		config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
		assert.Equal(t, OCRStreamStatusPending, config.Status)
		assert.NoError(t, config.Activate())
		assert.Equal(t, OCRStreamStatusActive, config.Status)
	})

	t.Run("from paused", func(t *testing.T) {
		config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
		config.Status = OCRStreamStatusPaused
		assert.NoError(t, config.Activate())
		assert.Equal(t, OCRStreamStatusActive, config.Status)
	})

	t.Run("from active fails", func(t *testing.T) {
		config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
		config.Status = OCRStreamStatusActive
		assert.Error(t, config.Activate())
	})

	t.Run("from completed fails", func(t *testing.T) {
		config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
		config.Status = OCRStreamStatusComplete
		assert.Error(t, config.Activate())
	})
}

func TestOCRStreamConfig_Complete(t *testing.T) {
	config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
	config.Complete()
	assert.Equal(t, OCRStreamStatusComplete, config.Status)
}

func TestOCRStreamConfig_RecordError(t *testing.T) {
	config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")

	// Record a few errors
	config.RecordError("connection timeout")
	assert.Equal(t, 1, config.ErrorCount)
	assert.Equal(t, "connection timeout", config.LastError)
	assert.NotEqual(t, OCRStreamStatusFailed, config.Status)

	// Record enough to trigger failure (50)
	for i := 0; i < 49; i++ {
		config.RecordError("repeated error")
	}
	assert.Equal(t, OCRStreamStatusFailed, config.Status)
	assert.Equal(t, 50, config.ErrorCount)
}

func TestOCRStreamConfig_RecordCapture(t *testing.T) {
	config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
	assert.Nil(t, config.LastCaptureAt)

	config.RecordCapture()
	assert.NotNil(t, config.LastCaptureAt)
}

func TestOCRStreamConfig_ScoreboardRegion(t *testing.T) {
	config := NewOCRStreamConfig("https://youtube.com/watch?v=test", replay_common.CS2_GAME_ID, "match-1")
	config.ScoreboardRegion = &ScoreboardRegion{
		X:      100,
		Y:      50,
		Width:  600,
		Height: 80,
	}

	assert.NotNil(t, config.ScoreboardRegion)
	assert.Equal(t, 100, config.ScoreboardRegion.X)
	assert.Equal(t, 50, config.ScoreboardRegion.Y)
	assert.Equal(t, 600, config.ScoreboardRegion.Width)
	assert.Equal(t, 80, config.ScoreboardRegion.Height)
}
