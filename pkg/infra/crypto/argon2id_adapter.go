// Package crypto provides cryptographic adapters for password hashing and other security operations.
package crypto

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	email_out "github.com/replay-api/replay-api/pkg/domain/email/ports/out"
	"golang.org/x/crypto/argon2"
)

// Argon2idParams contains the parameters for the Argon2id hashing algorithm.
// These defaults follow OWASP recommendations for password hashing.
type Argon2idParams struct {
	Memory      uint32 // Memory usage in KiB
	Iterations  uint32 // Number of iterations
	Parallelism uint8  // Number of parallel threads
	SaltLength  uint32 // Length of the random salt
	KeyLength   uint32 // Length of the derived key
}

// DefaultArgon2idParams returns secure default parameters following OWASP recommendations.
// Memory: 64 MiB, Iterations: 3, Parallelism: 4, Salt: 16 bytes, Key: 32 bytes
func DefaultArgon2idParams() *Argon2idParams {
	return &Argon2idParams{
		Memory:      64 * 1024, // 64 MiB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// Argon2idPasswordHasherAdapter implements PasswordHasher using the Argon2id algorithm.
// Argon2id is the recommended password hashing algorithm as it provides:
// - Memory-hardness (resistant to GPU/ASIC attacks)
// - CPU-hardness (adjustable computation time)
// - Data-dependent memory access (side-channel resistance)
type Argon2idPasswordHasherAdapter struct {
	params *Argon2idParams
}

// NewArgon2idPasswordHasherAdapter creates a new Argon2id password hasher with default secure parameters.
func NewArgon2idPasswordHasherAdapter() email_out.PasswordHasher {
	return &Argon2idPasswordHasherAdapter{
		params: DefaultArgon2idParams(),
	}
}

// NewArgon2idPasswordHasherAdapterWithParams creates a new Argon2id password hasher with custom parameters.
func NewArgon2idPasswordHasherAdapterWithParams(params *Argon2idParams) email_out.PasswordHasher {
	if params == nil {
		params = DefaultArgon2idParams()
	}
	return &Argon2idPasswordHasherAdapter{params: params}
}

// HashPassword hashes a password using Argon2id and returns the encoded hash string.
// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
func (a *Argon2idPasswordHasherAdapter) HashPassword(ctx context.Context, password string) (string, error) {
	// Generate cryptographically secure random salt
	salt := make([]byte, a.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	// Hash the password using Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		a.params.Iterations,
		a.params.Memory,
		a.params.Parallelism,
		a.params.KeyLength,
	)

	// Encode to PHC string format
	encoded := encodeArgon2idHash(a.params, salt, hash)
	return encoded, nil
}

// ComparePassword compares a password with a hashed password.
// Returns nil if they match, error otherwise.
// Uses constant-time comparison to prevent timing attacks.
func (a *Argon2idPasswordHasherAdapter) ComparePassword(ctx context.Context, hashedPassword string, password string) error {
	// Decode the encoded hash
	params, salt, storedHash, err := decodeArgon2idHash(hashedPassword)
	if err != nil {
		return fmt.Errorf("invalid hash format: %w", err)
	}

	// Compute hash of provided password with same parameters
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(storedHash, computedHash) != 1 {
		return ErrPasswordMismatch
	}

	return nil
}

// ErrPasswordMismatch is returned when password comparison fails.
var ErrPasswordMismatch = errors.New("password does not match")

// ErrInvalidHashFormat is returned when the hash string format is invalid.
var ErrInvalidHashFormat = errors.New("invalid argon2id hash format")

// encodeArgon2idHash encodes parameters, salt, and hash to PHC string format.
// Format: $argon2id$v=19$m=65536,t=3,p=4$<base64-salt>$<base64-hash>
func encodeArgon2idHash(params *Argon2idParams, salt, hash []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		b64Salt,
		b64Hash,
	)
}

// decodeArgon2idHash decodes a PHC string format hash to its components.
func decodeArgon2idHash(encodedHash string) (*Argon2idParams, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHashFormat
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid version: %w", err)
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible argon2 version: %d", version)
	}

	params := &Argon2idParams{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid hash: %w", err)
	}

	params.KeyLength = uint32(len(hash))
	params.SaltLength = uint32(len(salt))

	return params, salt, hash, nil
}

