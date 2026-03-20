// SLH-DSA adapter - implements PostQuantumSigner for cold-path archival use cases.
//
// Uses SLH-DSA-SHA2-256s (NIST FIPS 205, category 5, ~AES-256 security).
// SLH-DSA is a stateless hash-based signature scheme: conservative, well-analysed,
// and requires no long-term secret state — ideal for archival records.
//
// Platform use cases:
//   - Archival score records requiring 10+ year tamper-evidence
//   - Cold-path prize pool finalization proofs for regulated jurisdictions
//   - Backup signing during ML-DSA key rotation windows
package pq

import (
	"github.com/cloudflare/circl/sign/schemes"
	security_out "github.com/replay-api/replay-api/pkg/domain/security/ports/out"
)

// NewSlhDsaAdapter returns a high-security post-quantum signer using SLH-DSA-SHA2-256s (FIPS 205).
// Backed by cloudflare/circl slhdsa.SHA2_256s.Scheme().
func NewSlhDsaAdapter() security_out.PostQuantumSigner {
	sch := schemes.ByName("SLH-DSA-SHA2-256s")
	if sch == nil {
		panic("pq: SLH-DSA-SHA2-256s scheme not found in circl registry - upgrade github.com/cloudflare/circl")
	}
	return &schemeAdapter{scheme: sch}
}
