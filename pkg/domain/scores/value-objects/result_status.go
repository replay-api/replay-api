package scores_vo

import "fmt"

// ResultStatus represents the lifecycle state of a match result
type ResultStatus string

const (
	// ResultStatusSubmitted indicates the result has been submitted but not yet reviewed
	ResultStatusSubmitted ResultStatus = "submitted"

	// ResultStatusUnderReview indicates the result is being reviewed
	ResultStatusUnderReview ResultStatus = "under_review"

	// ResultStatusVerified indicates the result has been verified as accurate
	ResultStatusVerified ResultStatus = "verified"

	// ResultStatusDisputed indicates a team or player has disputed the result
	ResultStatusDisputed ResultStatus = "disputed"

	// ResultStatusConciliated indicates the dispute has been resolved
	ResultStatusConciliated ResultStatus = "conciliated"

	// ResultStatusFinalized indicates the result is final and prize distribution can proceed
	ResultStatusFinalized ResultStatus = "finalized"

	// ResultStatusCancelled indicates the result has been cancelled/voided
	ResultStatusCancelled ResultStatus = "cancelled"
)

// IsValid returns true if the status is a known valid value
func (s ResultStatus) IsValid() bool {
	switch s {
	case ResultStatusSubmitted, ResultStatusUnderReview, ResultStatusVerified,
		ResultStatusDisputed, ResultStatusConciliated, ResultStatusFinalized,
		ResultStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (s ResultStatus) String() string {
	return string(s)
}

// CanTransitionTo validates whether a status transition is allowed
func (s ResultStatus) CanTransitionTo(target ResultStatus) bool {
	switch s {
	case ResultStatusSubmitted:
		return target == ResultStatusUnderReview || target == ResultStatusVerified || target == ResultStatusCancelled
	case ResultStatusUnderReview:
		return target == ResultStatusVerified || target == ResultStatusCancelled
	case ResultStatusVerified:
		return target == ResultStatusDisputed || target == ResultStatusFinalized
	case ResultStatusDisputed:
		return target == ResultStatusConciliated || target == ResultStatusCancelled
	case ResultStatusConciliated:
		return target == ResultStatusDisputed || target == ResultStatusFinalized
	case ResultStatusFinalized:
		return false // Terminal state
	case ResultStatusCancelled:
		return false // Terminal state
	default:
		return false
	}
}

// IsTerminal returns true if no further transitions are possible
func (s ResultStatus) IsTerminal() bool {
	return s == ResultStatusFinalized || s == ResultStatusCancelled
}

// IsPrizeDistributable returns true if the result status allows prize distribution
func (s ResultStatus) IsPrizeDistributable() bool {
	return s == ResultStatusFinalized
}

// IsDisputable returns true if the result can be disputed in its current state
func (s ResultStatus) IsDisputable() bool {
	return s == ResultStatusVerified || s == ResultStatusConciliated
}

// Validate returns an error if the result status is invalid
func (s ResultStatus) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid result status: %s", s)
	}
	return nil
}

// ValidateTransition returns an error if the transition is invalid
func (s ResultStatus) ValidateTransition(target ResultStatus) error {
	if err := target.Validate(); err != nil {
		return err
	}
	if !s.CanTransitionTo(target) {
		return fmt.Errorf("invalid status transition from %s to %s", s, target)
	}
	return nil
}
