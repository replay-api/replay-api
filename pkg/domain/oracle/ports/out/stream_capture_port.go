package oracle_out

import (
	"context"
)

// StreamCapturePort defines the contract for capturing frames from live video streams
type StreamCapturePort interface {
	// CaptureFrame captures a single frame from the given stream URL and returns PNG bytes
	CaptureFrame(ctx context.Context, streamURL string) ([]byte, error)

	// IsStreamLive checks whether the given stream URL is currently broadcasting live
	IsStreamLive(ctx context.Context, streamURL string) (bool, error)
}
