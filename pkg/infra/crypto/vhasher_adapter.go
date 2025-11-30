package crypto

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	steam_out "github.com/replay-api/replay-api/pkg/domain/steam/ports/out"
)

type SHA256VHasherAdapter struct {
	VHashSource string
}

func NewSHA256VHasherAdapter(vHashSource string) steam_out.VHashWriter {
	return &SHA256VHasherAdapter{VHashSource: vHashSource}
}

func (v *SHA256VHasherAdapter) CreateVHash(ctx context.Context, steamID string) string {
	sha256 := sha256.New()

	sha256.Write([]byte(steamID + v.VHashSource))

	return hex.EncodeToString(sha256.Sum(nil))
}
