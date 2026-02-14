package scores_vo

import "fmt"

// VerificationMethod represents how a match result was verified
type VerificationMethod string

const (
	// VerificationMethodAutomatic indicates the result was verified programmatically (e.g., replay parsing)
	VerificationMethodAutomatic VerificationMethod = "automatic"
	// VerificationMethodManual indicates the result was verified by an administrator
	VerificationMethodManual VerificationMethod = "manual"
	// VerificationMethodVARReview indicates the result was reviewed by a VAR-style review panel
	VerificationMethodVARReview VerificationMethod = "var_review"
	// VerificationMethodConsensus indicates the result was verified by team consensus
	VerificationMethodConsensus VerificationMethod = "consensus"
)

func (v VerificationMethod) IsValid() bool {
	switch v {
	case VerificationMethodAutomatic, VerificationMethodManual, VerificationMethodVARReview, VerificationMethodConsensus:
		return true
	default:
		return false
	}
}

func (v VerificationMethod) String() string {
	return string(v)
}

// Validate returns an error if the verification method is invalid
func (v VerificationMethod) Validate() error {
	if !v.IsValid() {
		return fmt.Errorf("invalid verification method: %s", v)
	}
	return nil
}
