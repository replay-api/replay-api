package media_out

import (
	"context"
)

type MediaWriter interface {
	Create(ctx context.Context, media []byte, name string, extension string) (string, error)
}
