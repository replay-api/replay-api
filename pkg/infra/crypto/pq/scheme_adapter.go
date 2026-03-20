// Package pq provides post-quantum cryptographic adapters.
//
// schemeAdapter is a generic PostQuantumSigner backed by any circl sign.Scheme.
// Used by both MlDsaAdapter (ML-DSA-65, FIPS 204) and SlhDsaAdapter (SLH-DSA-SHA2-256s, FIPS 205).
package pq

import (
	"context"
	"fmt"

	"github.com/cloudflare/circl/sign"
	security_out "github.com/replay-api/replay-api/pkg/domain/security/ports/out"
)

// schemeAdapter wraps any circl sign.Scheme to implement PostQuantumSigner.
// All operations are stateless and safe for concurrent use.
type schemeAdapter struct {
	scheme sign.Scheme
}

func (a *schemeAdapter) GenerateKeyPair(ctx context.Context) ([]byte, []byte, error) {
	pk, sk, err := a.scheme.GenerateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("%s GenerateKeyPair: %w", a.scheme.Name(), err)
	}
	pkBytes, err := pk.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("%s GenerateKeyPair marshal pub: %w", a.scheme.Name(), err)
	}
	skBytes, err := sk.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("%s GenerateKeyPair marshal priv: %w", a.scheme.Name(), err)
	}
	return pkBytes, skBytes, nil
}

func (a *schemeAdapter) Sign(ctx context.Context, payload, signingKeyBytes []byte) ([]byte, error) {
	sk, err := a.scheme.UnmarshalBinaryPrivateKey(signingKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("%s Sign: parse private key: %w", a.scheme.Name(), err)
	}
	sig := a.scheme.Sign(sk, payload, nil)
	return sig, nil
}

func (a *schemeAdapter) Verify(ctx context.Context, payload, signature, verifyKeyBytes []byte) bool {
	pk, err := a.scheme.UnmarshalBinaryPublicKey(verifyKeyBytes)
	if err != nil {
		return false
	}
	return a.scheme.Verify(pk, payload, signature, nil)
}

func (a *schemeAdapter) Algorithm() string { return a.scheme.Name() }

// Ensure schemeAdapter satisfies the port interface at compile time.
var _ security_out.PostQuantumSigner = (*schemeAdapter)(nil)
