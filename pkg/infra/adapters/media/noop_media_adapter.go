package media

import (
	"context"
	"fmt"
	"log/slog"
)

// NoopMediaAdapter is a no-op implementation of MediaWriter
// It can be replaced with a real implementation (S3, GCS, etc.) later
type NoopMediaAdapter struct{}

// NewNoopMediaAdapter creates a new no-op media adapter
func NewNoopMediaAdapter() *NoopMediaAdapter {
	return &NoopMediaAdapter{}
}

// Create stores media (currently a no-op that returns a placeholder URI)
func (a *NoopMediaAdapter) Create(ctx context.Context, media []byte, name string, extension string) (string, error) {
	slog.WarnContext(ctx, "NoopMediaAdapter.Create called - media storage not implemented",
		"name", name,
		"extension", extension,
		"size", len(media))

	// Return a placeholder URI
	return fmt.Sprintf("noop://media/%s.%s", name, extension), nil
}
