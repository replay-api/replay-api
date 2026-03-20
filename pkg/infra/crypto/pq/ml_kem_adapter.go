// Package pq provides post-quantum cryptographic adapters.
//
// This file implements ML-KEM-768 (NIST FIPS 203) using Go 1.24's built-in
// crypto/mlkem package - no external dependency required.
//
// ML-KEM-768 provides NIST security level 3 (~AES-192), replacing classical
// ECDH/RSA key exchange in inter-service communication and score attestation
// session key negotiation.
package pq

import (
	"context"
	"crypto/mlkem"
	"fmt"

	security_out "github.com/replay-api/replay-api/pkg/domain/security/ports/out"
)

const algorithmMLKEM768 = "ML-KEM-768"

// MlKemAdapter implements PostQuantumKeyEncapsulator using ML-KEM-768 (NIST FIPS 203).
// All operations are stateless and safe for concurrent use.
type MlKemAdapter struct{}

// NewMlKemAdapter returns a ready-to-use ML-KEM-768 key encapsulator.
func NewMlKemAdapter() security_out.PostQuantumKeyEncapsulator {
	return &MlKemAdapter{}
}

// GenerateKeyPair produces a fresh ML-KEM-768 key pair.
// encapsKey (1184 bytes) is the public key; decapsKey (64-byte seed) is private.
func (a *MlKemAdapter) GenerateKeyPair(ctx context.Context) ([]byte, []byte, error) {
	dk, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, nil, fmt.Errorf("ml-kem-768 GenerateKeyPair: %w", err)
	}
	return dk.EncapsulationKey().Bytes(), dk.Bytes(), nil
}

// Encapsulate uses the recipient's public key to derive a fresh shared secret.
// Returns (ciphertext, sharedSecret). Ciphertext is transmitted to the recipient.
//
// Note: Go 1.24's ek.Encapsulate() follows FIPS 203 output ordering (K, c),
// returning (sharedKey, ciphertext). We swap to match our interface convention.
func (a *MlKemAdapter) Encapsulate(ctx context.Context, encapsKeyBytes []byte) ([]byte, []byte, error) {
	ek, err := mlkem.NewEncapsulationKey768(encapsKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("ml-kem-768 Encapsulate: %w", err)
	}
	// FIPS 203 ordering: (sharedKey K, ciphertext c) — swap to our (ciphertext, sharedKey) convention.
	sharedKey, ciphertext := ek.Encapsulate()
	return ciphertext, sharedKey, nil
}

// Decapsulate recovers the shared secret from a ciphertext using the private key.
func (a *MlKemAdapter) Decapsulate(ctx context.Context, ciphertext, decapsKeyBytes []byte) ([]byte, error) {
	dk, err := mlkem.NewDecapsulationKey768(decapsKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("ml-kem-768 Decapsulate parse key: %w", err)
	}
	ss, err := dk.Decapsulate(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("ml-kem-768 Decapsulate: %w", err)
	}
	return ss, nil
}

// Algorithm returns the identifier for this key encapsulation mechanism.
func (a *MlKemAdapter) Algorithm() string { return algorithmMLKEM768 }
