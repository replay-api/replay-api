// ML-DSA adapter - implements PostQuantumSigner using ML-DSA-65 (NIST FIPS 204).
//
// ML-DSA-65 (CRYSTALS-Dilithium, Mode3) provides NIST security level 3 (~AES-192).
//
// Primary platform use cases:
//   - Signing serialised PrizePool state (tamper-evident prize distribution proofs)
//   - Signing finalized score attestations before blockchain settlement
//   - Generating portable, verifiable proofs of match outcomes for external consumers
package pq

import (
	"github.com/cloudflare/circl/sign/schemes"
	security_out "github.com/replay-api/replay-api/pkg/domain/security/ports/out"
)

// NewMlDsaAdapter returns a ready-to-use ML-DSA-65 (FIPS 204) signer.
// Backed by cloudflare/circl mldsa65.Scheme().
func NewMlDsaAdapter() security_out.PostQuantumSigner {
	sch := schemes.ByName("ML-DSA-65")
	if sch == nil {
		panic("pq: ML-DSA-65 scheme not found in circl registry - upgrade github.com/cloudflare/circl")
	}
	return &schemeAdapter{scheme: sch}
}
