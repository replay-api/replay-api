package pq_test

import (
	"context"
	"testing"

	pq "github.com/replay-api/replay-api/pkg/infra/crypto/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMlKemAdapter_RoundTrip(t *testing.T) {
	ctx := context.Background()
	adapter := pq.NewMlKemAdapter()

	encapsKey, decapsKey, err := adapter.GenerateKeyPair(ctx)
	require.NoError(t, err)
	assert.Len(t, encapsKey, 1184, "ML-KEM-768 public key must be 1184 bytes")
	assert.Len(t, decapsKey, 64, "ML-KEM-768 private seed must be 64 bytes")

	ciphertext, sharedSecretSender, err := adapter.Encapsulate(ctx, encapsKey)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.Len(t, sharedSecretSender, 32)

	sharedSecretRecipient, err := adapter.Decapsulate(ctx, ciphertext, decapsKey)
	require.NoError(t, err)
	assert.Equal(t, sharedSecretSender, sharedSecretRecipient, "shared secrets must match")
	assert.Equal(t, "ML-KEM-768", adapter.Algorithm())
}

func TestMlKemAdapter_WrongCiphertextImplicitRejection(t *testing.T) {
	// ML-KEM uses "implicit rejection": decapsulation never returns an error,
	// even for a malformed or wrong ciphertext. Instead, it returns a
	// pseudorandom key that will not match the sender's shared key.
	// This property prevents timing attacks that exploit error paths.
	ctx := context.Background()
	adapter := pq.NewMlKemAdapter()

	encapsKey, decapsKey, err := adapter.GenerateKeyPair(ctx)
	require.NoError(t, err)

	// Produce a legitimate ciphertext and sender's shared secret.
	ciphertext, senderSecret, err := adapter.Encapsulate(ctx, encapsKey)
	require.NoError(t, err)

	// Tamper with one byte — decapsulation must succeed but produce a wrong key.
	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[0] ^= 0xFF

	recipientSecret, err := adapter.Decapsulate(ctx, tampered, decapsKey)
	// No error expected — implicit rejection.
	require.NoError(t, err, "ML-KEM implicit rejection: decapsulation must not surface an error")
	assert.NotEqual(t, senderSecret, recipientSecret, "tampered ciphertext must produce a different shared key")
}

func testSignerRoundTrip(t *testing.T, name string, newSigner func() interface {
	GenerateKeyPair(context.Context) ([]byte, []byte, error)
	Sign(context.Context, []byte, []byte) ([]byte, error)
	Verify(context.Context, []byte, []byte, []byte) bool
	Algorithm() string
}) {
	t.Helper()
	ctx := context.Background()
	signer := newSigner()

	verifyKey, signingKey, err := signer.GenerateKeyPair(ctx)
	require.NoError(t, err, "%s: GenerateKeyPair", name)
	assert.NotEmpty(t, verifyKey)
	assert.NotEmpty(t, signingKey)

	payload := []byte("prize-pool-state:tournament-abc123:payout-hash:deadbeef")
	sig, err := signer.Sign(ctx, payload, signingKey)
	require.NoError(t, err, "%s: Sign", name)
	assert.NotEmpty(t, sig)

	assert.True(t, signer.Verify(ctx, payload, sig, verifyKey), "%s: Verify valid sig", name)

	tampered := []byte("prize-pool-state:tournament-abc123:payout-hash:TAMPERED")
	assert.False(t, signer.Verify(ctx, tampered, sig, verifyKey), "%s: Verify tampered payload", name)

	assert.Contains(t, signer.Algorithm(), name[:4]) // at least first 4 chars match
}

func TestMlDsaAdapter_RoundTrip(t *testing.T) {
	ctx := context.Background()
	adapter := pq.NewMlDsaAdapter()

	verifyKey, signingKey, err := adapter.GenerateKeyPair(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, verifyKey)
	assert.NotEmpty(t, signingKey)

	payload := []byte("score-attestation:match-abc:2026-03-20")
	sig, err := adapter.Sign(ctx, payload, signingKey)
	require.NoError(t, err)

	assert.True(t, adapter.Verify(ctx, payload, sig, verifyKey))
	assert.False(t, adapter.Verify(ctx, append(payload, '!'), sig, verifyKey))
	assert.Equal(t, "ML-DSA-65", adapter.Algorithm())
}

func TestMlDsaAdapter_CrossKeyFails(t *testing.T) {
	ctx := context.Background()
	adapter := pq.NewMlDsaAdapter()

	vk1, _, err := adapter.GenerateKeyPair(ctx)
	require.NoError(t, err)
	_, sk2, err := adapter.GenerateKeyPair(ctx)
	require.NoError(t, err)

	payload := []byte("score-attestation:match-xyz")
	sig, err := adapter.Sign(ctx, payload, sk2)
	require.NoError(t, err)

	assert.False(t, adapter.Verify(ctx, payload, sig, vk1), "key2 sig must not verify against key1 pubkey")
}

func TestSlhDsaAdapter_RoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("SLH-DSA key generation is slow; skipping in short mode")
	}
	ctx := context.Background()
	adapter := pq.NewSlhDsaAdapter()

	verifyKey, signingKey, err := adapter.GenerateKeyPair(ctx)
	require.NoError(t, err)

	archivalRecord := []byte("archival-score:player-uuid:2026-03-20:verified-outcome-hash")
	sig, err := adapter.Sign(ctx, archivalRecord, signingKey)
	require.NoError(t, err)

	assert.True(t, adapter.Verify(ctx, archivalRecord, sig, verifyKey))
	assert.Contains(t, adapter.Algorithm(), "SLH-DSA")
}
