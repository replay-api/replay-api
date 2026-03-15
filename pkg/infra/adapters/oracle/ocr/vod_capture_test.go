package oracle_ocr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVodCapture(t *testing.T) {
	vc := NewVodCapture()
	require.NotNil(t, vc)
	assert.NotEmpty(t, vc.ffmpegPath)
	assert.NotEmpty(t, vc.ytdlpPath)
	assert.True(t, vc.captureTimeout > 0)
}

func TestVodCapture_IsStreamLive(t *testing.T) {
	vc := NewVodCapture()
	live, err := vc.IsStreamLive(context.Background(), "https://www.youtube.com/watch?v=test")
	require.NoError(t, err)
	assert.False(t, live, "VODs are never live")
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		wantErr  bool
	}{
		{"1:23:45", 5025, false},
		{"0:05:30", 330, false},
		{"45:30", 2730, false},
		{"5:00", 300, false},
		{"120", 120, false},
		{"0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
