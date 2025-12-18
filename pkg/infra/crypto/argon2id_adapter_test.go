package crypto

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgon2idPasswordHasher_HashPassword(t *testing.T) {
	ctx := context.Background()
	hasher := NewArgon2idPasswordHasherAdapter()

	t.Run("should hash password successfully", func(t *testing.T) {
		password := "SecureP@ssw0rd123!"

		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Verify PHC format
		assert.True(t, strings.HasPrefix(hash, "$argon2id$"))
		parts := strings.Split(hash, "$")
		assert.Len(t, parts, 6, "should have 6 parts in PHC format")
	})

	t.Run("should generate different hashes for same password", func(t *testing.T) {
		password := "SamePassword123!"

		hash1, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)

		hash2, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2, "hashes should differ due to random salt")
	})

	t.Run("should hash empty password", func(t *testing.T) {
		hash, err := hasher.HashPassword(ctx, "")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("should hash very long password", func(t *testing.T) {
		password := strings.Repeat("A", 10000)
		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("should hash unicode password", func(t *testing.T) {
		password := "пароль密码🔐"
		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})
}

func TestArgon2idPasswordHasher_ComparePassword(t *testing.T) {
	ctx := context.Background()
	hasher := NewArgon2idPasswordHasherAdapter()

	t.Run("should verify correct password", func(t *testing.T) {
		password := "SecureP@ssw0rd123!"

		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)

		err = hasher.ComparePassword(ctx, hash, password)
		assert.NoError(t, err)
	})

	t.Run("should reject incorrect password", func(t *testing.T) {
		password := "CorrectPassword!"
		wrongPassword := "WrongPassword!"

		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)

		err = hasher.ComparePassword(ctx, hash, wrongPassword)
		assert.Error(t, err)
		assert.Equal(t, ErrPasswordMismatch, err)
	})

	t.Run("should reject with case-sensitive mismatch", func(t *testing.T) {
		password := "CaseSensitive!"

		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)

		err = hasher.ComparePassword(ctx, hash, "casesensitive!")
		assert.Error(t, err)
	})

	t.Run("should reject invalid hash format", func(t *testing.T) {
		err := hasher.ComparePassword(ctx, "invalid-hash", "password")
		assert.Error(t, err)
	})

	t.Run("should reject truncated hash", func(t *testing.T) {
		password := "Password123!"

		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)

		truncated := hash[:len(hash)-10]
		err = hasher.ComparePassword(ctx, truncated, password)
		assert.Error(t, err)
	})

	t.Run("should reject wrong algorithm identifier", func(t *testing.T) {
		wrongHash := "$argon2i$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA"
		err := hasher.ComparePassword(ctx, wrongHash, "password")
		assert.Error(t, err)
	})
}

func TestArgon2idPasswordHasher_CustomParams(t *testing.T) {
	ctx := context.Background()

	t.Run("should work with custom parameters", func(t *testing.T) {
		params := &Argon2idParams{
			Memory:      32 * 1024, // 32 MiB
			Iterations:  2,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32,
		}
		hasher := NewArgon2idPasswordHasherAdapterWithParams(params)

		password := "TestPassword!"
		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)

		err = hasher.ComparePassword(ctx, hash, password)
		assert.NoError(t, err)

		// Verify custom params are in hash
		assert.Contains(t, hash, "m=32768")
		assert.Contains(t, hash, "t=2")
		assert.Contains(t, hash, "p=2")
	})

	t.Run("should use defaults when nil params provided", func(t *testing.T) {
		hasher := NewArgon2idPasswordHasherAdapterWithParams(nil)

		password := "TestPassword!"
		hash, err := hasher.HashPassword(ctx, password)
		require.NoError(t, err)

		// Should use default params
		assert.Contains(t, hash, "m=65536")
		assert.Contains(t, hash, "t=3")
		assert.Contains(t, hash, "p=4")
	})
}

func TestDefaultArgon2idParams(t *testing.T) {
	params := DefaultArgon2idParams()

	assert.Equal(t, uint32(64*1024), params.Memory, "should use 64 MiB memory")
	assert.Equal(t, uint32(3), params.Iterations, "should use 3 iterations")
	assert.Equal(t, uint8(4), params.Parallelism, "should use 4 threads")
	assert.Equal(t, uint32(16), params.SaltLength, "should use 16-byte salt")
	assert.Equal(t, uint32(32), params.KeyLength, "should use 32-byte key")
}

func BenchmarkArgon2idHashPassword(b *testing.B) {
	ctx := context.Background()
	hasher := NewArgon2idPasswordHasherAdapter()
	password := "BenchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.HashPassword(ctx, password)
	}
}

func BenchmarkArgon2idComparePassword(b *testing.B) {
	ctx := context.Background()
	hasher := NewArgon2idPasswordHasherAdapter()
	password := "BenchmarkPassword123!"

	hash, _ := hasher.HashPassword(ctx, password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasher.ComparePassword(ctx, hash, password)
	}
}

