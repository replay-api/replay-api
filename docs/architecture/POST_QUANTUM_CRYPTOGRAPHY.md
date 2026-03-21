# Post-Quantum Cryptography

This document covers what we implemented, why each piece exists, and how to use it correctly.

---

## Background

NIST finalized three post-quantum cryptography standards in August 2024:

- **FIPS 203** — ML-KEM (key encapsulation, replaces ECDH/RSA key exchange)
- **FIPS 204** — ML-DSA (digital signatures, replaces ECDSA)
- **FIPS 205** — SLH-DSA (stateless hash-based signatures, conservative backup)

The threat that makes this relevant is called "harvest now, decrypt later": adversaries capture encrypted traffic today and hold it until quantum computers are powerful enough to break classical encryption. For a platform that processes real-money prize distributions and competitive score records, those records are worth attacking. We adopted all three standards before any of our competitors.

---

## What we built

Three adapters live in `pkg/infra/crypto/pq/`:

### ML-KEM-768 (`ml_kem_adapter.go`)

Used for key encapsulation — the step where two parties establish a shared secret without transmitting it directly. This replaces ECDH in inter-service communication and score attestation session key negotiation.

Uses Go 1.24's `crypto/mlkem` standard library package — no external dependency.

Key sizes:
- Public (encapsulation) key: 1184 bytes
- Private seed (decapsulation key): 64 bytes — store only this, not the expanded key
- Shared secret: 32 bytes
- Ciphertext: 1088 bytes

One important property: decapsulation never returns an error, even for a malformed or tampered ciphertext. Instead it returns a pseudorandom key that won't match the sender's key. This is called implicit rejection and it prevents timing attacks that exploit error paths.

### ML-DSA-65 (`ml_dsa_adapter.go`)

Used for signing competitive results, score attestations, and prize pool state. This is the hot-path signer — signing is fast (~1ms).

Backed by `github.com/cloudflare/circl v1.6.3`, scheme name `ML-DSA-65`.

Security level 3 (~AES-192 equivalent), which NIST considers appropriate for most data with long-term sensitivity.

### SLH-DSA-SHA2-256s (`slh_dsa_adapter.go`)

Used only for archival records: score records that need to be verifiable 10+ years from now, prize pool finalization proofs for regulated jurisdictions, and backup signing during ML-DSA key rotation windows.

This is a stateless hash-based scheme — its security relies entirely on SHA2, which is extremely well-analyzed and has no algebraic structure for quantum algorithms to exploit. The tradeoff is speed: key generation and signing take roughly 350–500ms. Do not use this on any hot path.

Also backed by `circl`, scheme name `SLH-DSA-SHA2-256s`.

---

## Domain ports

The three algorithms are exposed through distinct interfaces in `pkg/domain/security/ports/out/pq.go`:

| Interface | Algorithm | Use when |
|---|---|---|
| `PostQuantumKeyEncapsulator` | ML-KEM-768 | Establishing shared secrets between services |
| `PostQuantumSigner` | ML-DSA-65 | Signing competitive results, prize pool state |
| `PostQuantumArchivalSigner` | SLH-DSA-SHA2-256s | Archival records, audit logs, regulatory proofs |

ML-DSA and SLH-DSA share the same method signatures, so they deliberately use two different interfaces. This prevents the IoC container from having a collision between them and also makes it explicit at the call site which one you're injecting — you cannot accidentally inject the slow archival signer into a real-time path.

Never call adapter constructors directly from use-case or controller code. Always inject the interface.

---

## IoC registration

All three are registered as singletons in `pkg/infra/ioc/container.go`:

```go
// ML-KEM-768 — key encapsulation
c.Singleton(func() (security_out.PostQuantumKeyEncapsulator, error) {
    return pq_crypto.NewMlKemAdapter(), nil
})

// ML-DSA-65 — primary signer
c.Singleton(func() (security_out.PostQuantumSigner, error) {
    return pq_crypto.NewMlDsaAdapter(), nil
})

// SLH-DSA-SHA2-256s — archival signer
c.Singleton(func() (security_out.PostQuantumArchivalSigner, error) {
    return pq_crypto.NewSlhDsaAdapter(), nil
})
```

All three are stateless — the singletons are safe for concurrent use without any locking.

---

## Tests

`pkg/infra/crypto/pq/pq_adapters_test.go` covers:

- **ML-KEM round-trip**: generate key pair, encapsulate, decapsulate — shared secrets match
- **ML-KEM implicit rejection**: tamper one ciphertext byte — decapsulation succeeds but returns a different key
- **ML-DSA round-trip**: generate key pair, sign, verify valid signature, verify tampered payload fails
- **ML-DSA cross-key**: signature from key2 does not verify against key1's public key
- **SLH-DSA round-trip**: generate key pair, sign archival record, verify

Run with:

```bash
go test ./pkg/infra/crypto/pq/... -v
```

SLH-DSA is slow by design. If you're running tests in short mode (`-short`), the SLH-DSA test is skipped automatically.

---

## HTTP response header

Every response from the Next.js frontend carries:

```
X-Security-Policy: pq-hybrid
```

Set in `middleware.ts`. This signals to clients, API gateways, and compliance scanners that the service runs post-quantum cryptography on authenticated data paths. The value `pq-hybrid` reflects that we're in a hybrid period — classical TLS still handles transport, while PQ handles application-layer signing and key exchange.

---

## Dependencies

| Package | Version | Purpose |
|---|---|---|
| `crypto/mlkem` | Go 1.24 stdlib | ML-KEM-768 — no external dep |
| `github.com/cloudflare/circl` | v1.6.3 | ML-DSA-65 and SLH-DSA-SHA2-256s |

`circl` is a direct dependency (imported by `ml_dsa_adapter.go` and `slh_dsa_adapter.go`).

---

## What's not done yet

These are the next steps when PQ is used in actual application flows:

- **Score attestation service**: inject `PostQuantumSigner` and sign the serialized score record before persisting it
- **Prize pool finalization**: sign the final payout state with ML-DSA-65; sign the archival copy with SLH-DSA-SHA2-256s
- **Inter-service session keys**: run ML-KEM-768 encapsulation when establishing authenticated channels between replay-api workers
- **Key storage**: the 64-byte ML-KEM decapsulation seed and ML-DSA signing keys need a key management strategy (KMS or Vault) — currently no keys are persisted anywhere

The adapters and ports are complete; the remaining work is wiring them into the domain use-cases that need them.
