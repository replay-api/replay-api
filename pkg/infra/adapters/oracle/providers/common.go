package oracle_providers

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

// deterministicUUID generates a deterministic UUID from a namespace and external ID.
// This ensures consistent mapping between external provider IDs and internal UUIDs.
func deterministicUUID(namespace, externalID string) uuid.UUID {
	data := []byte(namespace + ":" + externalID)
	hash := sha256.Sum256(data)
	// Use first 16 bytes as UUID, setting version 5 (SHA-1) and variant bits
	id, _ := uuid.FromBytes(hash[:16])
	// Set version 5
	id[6] = (id[6] & 0x0f) | 0x50
	// Set variant to RFC 4122
	id[8] = (id[8] & 0x3f) | 0x80
	return id
}

// hashBytes computes SHA-256 of raw bytes
func hashBytes(data []byte) [32]byte {
	return sha256.Sum256(data)
}
