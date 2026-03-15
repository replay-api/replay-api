package oracle_vo

import "fmt"

// OracleStatus represents the lifecycle state of an oracle result
type OracleStatus string

const (
	// OracleStatusPending indicates the oracle result is awaiting provider submissions
	OracleStatusPending OracleStatus = "pending"

	// OracleStatusConsensusReached indicates consensus has been computed from submissions
	OracleStatusConsensusReached OracleStatus = "consensus_reached"

	// OracleStatusPublishing indicates the result is being published on-chain
	OracleStatusPublishing OracleStatus = "publishing"

	// OracleStatusPublished indicates the result has been published on-chain
	OracleStatusPublished OracleStatus = "published"

	// OracleStatusFinalized indicates the result is final (dispute window closed)
	OracleStatusFinalized OracleStatus = "finalized"

	// OracleStatusDisputed indicates a published score has been disputed
	OracleStatusDisputed OracleStatus = "disputed"

	// OracleStatusFailed indicates the oracle process failed
	OracleStatusFailed OracleStatus = "failed"
)

// IsValid returns true if the status is a known valid value
func (s OracleStatus) IsValid() bool {
	switch s {
	case OracleStatusPending, OracleStatusConsensusReached, OracleStatusPublishing,
		OracleStatusPublished, OracleStatusFinalized, OracleStatusDisputed, OracleStatusFailed:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (s OracleStatus) String() string {
	return string(s)
}

// CanTransitionTo validates whether a status transition is allowed
func (s OracleStatus) CanTransitionTo(target OracleStatus) bool {
	switch s {
	case OracleStatusPending:
		return target == OracleStatusConsensusReached || target == OracleStatusFailed
	case OracleStatusConsensusReached:
		return target == OracleStatusPublishing || target == OracleStatusFailed
	case OracleStatusPublishing:
		return target == OracleStatusPublished || target == OracleStatusFailed
	case OracleStatusPublished:
		return target == OracleStatusFinalized || target == OracleStatusDisputed
	case OracleStatusDisputed:
		return target == OracleStatusPending || target == OracleStatusFailed // Reset to pending for re-consensus
	case OracleStatusFinalized:
		return false // Terminal state
	case OracleStatusFailed:
		return target == OracleStatusPending // Allow retry
	default:
		return false
	}
}

// IsTerminal returns true if no further transitions are possible (except retry from failed)
func (s OracleStatus) IsTerminal() bool {
	return s == OracleStatusFinalized
}

// IsPublishable returns true if the result can be published on-chain
func (s OracleStatus) IsPublishable() bool {
	return s == OracleStatusConsensusReached
}

// IsDisputable returns true if the result can be disputed
func (s OracleStatus) IsDisputable() bool {
	return s == OracleStatusPublished
}

// Validate returns an error if the status is invalid
func (s OracleStatus) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid oracle status: %s", s)
	}
	return nil
}

// ValidateTransition returns an error if the transition is invalid
func (s OracleStatus) ValidateTransition(target OracleStatus) error {
	if err := target.Validate(); err != nil {
		return err
	}
	if !s.CanTransitionTo(target) {
		return fmt.Errorf("invalid oracle status transition from %s to %s", s, target)
	}
	return nil
}
