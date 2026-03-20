// Package security_out defines output port interfaces for post-quantum cryptography.
// Concrete adapters live under pkg/infra/crypto/pq/.
//
// NIST FIPS post-quantum standards implemented:
//   - ML-KEM  (CRYSTALS-Kyber)     FIPS 203, key encapsulation
//   - ML-DSA  (CRYSTALS-Dilithium) FIPS 204, digital signatures
//   - SLH-DSA (SPHINCS+)           FIPS 205, stateless hash-based signatures
package security_out

import "context"

// PostQuantumKeyEncapsulator defines the port for ML-KEM key encapsulation (NIST FIPS 203).
// Replaces classical ECDH/RSA key exchange for inter-service session key negotiation
// and score attestation session key establishment.
type PostQuantumKeyEncapsulator interface {
	// GenerateKeyPair produces a fresh ML-KEM key pair.
	// Returns (encapsulationKeyBytes, decapsulationKeyBytes) i.e. (pubkey, privkey).
	GenerateKeyPair(ctx context.Context) (encapsKey, decapsKey []byte, err error)

	// Encapsulate derives a shared secret using the recipient's encapsulation key.
	// Returns (ciphertext, sharedSecret). The ciphertext is transmitted to the
	// recipient; the sharedSecret is used locally to derive a symmetric key.
	Encapsulate(ctx context.Context, encapsKey []byte) (ciphertext, sharedSecret []byte, err error)

	// Decapsulate recovers the shared secret from a ciphertext using the private key.
	Decapsulate(ctx context.Context, ciphertext, decapsKey []byte) (sharedSecret []byte, err error)

	// Algorithm returns the human-readable algorithm identifier (e.g. "ML-KEM-768").
	Algorithm() string
}

// PostQuantumSigner defines the port for post-quantum digital signatures.
//
// Two concrete algorithms are provided:
//   - ML-DSA-65  (mldsa65, FIPS 204) primary signer for score attestations & prize pools
//   - SLH-DSA    (sha2-256s, FIPS 205) stateless hash-based signer for archival records
type PostQuantumSigner interface {
	// GenerateKeyPair produces a fresh signing key pair.
	// Returns (verificationKeyBytes, signingKeyBytes) i.e. (pubkey, privkey).
	GenerateKeyPair(ctx context.Context) (verifyKey, signingKey []byte, err error)

	// Sign produces a post-quantum signature over the payload using the private key.
	Sign(ctx context.Context, payload, signingKey []byte) (signature []byte, err error)

	// Verify returns true iff the signature was produced by the holder of the
	// private key corresponding to verifyKey over the exact payload.
	Verify(ctx context.Context, payload, signature, verifyKey []byte) bool

	// Algorithm returns the human-readable algorithm identifier.
	Algorithm() string
}
