package challenge_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// ChallengeType represents the type of challenge/dispute
type ChallengeType string

const (
	ChallengeTypeBugReport      ChallengeType = "bug_report"       // Game bug affecting outcome
	ChallengeTypeVAR            ChallengeType = "var"              // Video Assistant Review request
	ChallengeTypeRoundRestart   ChallengeType = "round_restart"    // Request to restart current round
	ChallengeTypeMatchRestart   ChallengeType = "match_restart"    // Request to restart entire match
	ChallengeTypeTechnicalIssue ChallengeType = "technical_issue"  // Server/connection problems
	ChallengeTypeRuleViolation  ChallengeType = "rule_violation"   // Alleged rule violation by opponent
	ChallengeTypeScoreDispute   ChallengeType = "score_dispute"    // Dispute about match score
)

// ChallengeStatus represents the lifecycle status of a challenge
type ChallengeStatus string

const (
	ChallengeStatusPending    ChallengeStatus = "pending"     // Awaiting review
	ChallengeStatusInReview   ChallengeStatus = "in_review"   // Being reviewed by admin/system
	ChallengeStatusVotePending ChallengeStatus = "vote_pending" // Waiting for player votes
	ChallengeStatusApproved   ChallengeStatus = "approved"    // Challenge accepted
	ChallengeStatusRejected   ChallengeStatus = "rejected"    // Challenge denied
	ChallengeStatusResolved   ChallengeStatus = "resolved"    // Final resolution applied
	ChallengeStatusExpired    ChallengeStatus = "expired"     // Timeout without resolution
	ChallengeStatusCancelled  ChallengeStatus = "cancelled"   // Cancelled by challenger
)

// ChallengePriority indicates urgency level
type ChallengePriority string

const (
	ChallengePriorityLow      ChallengePriority = "low"      // Non-urgent, review later
	ChallengePriorityNormal   ChallengePriority = "normal"   // Standard processing
	ChallengePriorityHigh     ChallengePriority = "high"     // Urgent, affects ongoing match
	ChallengePriorityCritical ChallengePriority = "critical" // Match-stopping issue
)

// ChallengeResolution represents the outcome of a challenge
type ChallengeResolution string

const (
	ChallengeResolutionNone           ChallengeResolution = "none"            // Not yet resolved
	ChallengeResolutionRoundRestarted ChallengeResolution = "round_restarted" // Round was restarted
	ChallengeResolutionMatchRestarted ChallengeResolution = "match_restarted" // Match was restarted
	ChallengeResolutionScoreAdjusted  ChallengeResolution = "score_adjusted"  // Score was corrected
	ChallengeResolutionPenaltyApplied ChallengeResolution = "penalty_applied" // Penalty given to violator
	ChallengeResolutionNoAction       ChallengeResolution = "no_action"       // No action warranted
	ChallengeResolutionMatchVoided    ChallengeResolution = "match_voided"    // Match result invalidated
	ChallengeResolutionCompensation   ChallengeResolution = "compensation"    // Financial compensation
)

// Evidence represents supporting evidence for a challenge
type Evidence struct {
	ID          uuid.UUID `json:"id" bson:"id"`
	Type        string    `json:"type" bson:"type"`                 // screenshot, replay_clip, log, video
	URL         string    `json:"url" bson:"url"`                   // Storage URL
	Description string    `json:"description" bson:"description"`   // Evidence description
	Timestamp   time.Time `json:"timestamp" bson:"timestamp"`       // When evidence was captured
	TickRange   *TickRange `json:"tick_range,omitempty" bson:"tick_range,omitempty"` // Replay tick range
	UploadedAt  time.Time `json:"uploaded_at" bson:"uploaded_at"`
}

// TickRange represents a range in replay ticks
type TickRange struct {
	StartTick int64 `json:"start_tick" bson:"start_tick"`
	EndTick   int64 `json:"end_tick" bson:"end_tick"`
}

// Vote represents a player vote on a challenge
type Vote struct {
	PlayerID  uuid.UUID `json:"player_id" bson:"player_id"`
	TeamID    *uuid.UUID `json:"team_id,omitempty" bson:"team_id,omitempty"`
	VoteType  string    `json:"vote_type" bson:"vote_type"` // approve, reject, abstain
	Reason    string    `json:"reason,omitempty" bson:"reason,omitempty"`
	VotedAt   time.Time `json:"voted_at" bson:"voted_at"`
}

// AdminAction represents an action taken by an admin/referee
type AdminAction struct {
	AdminID     uuid.UUID `json:"admin_id" bson:"admin_id"`
	Action      string    `json:"action" bson:"action"` // review_started, decision_made, escalated
	Notes       string    `json:"notes" bson:"notes"`
	PerformedAt time.Time `json:"performed_at" bson:"performed_at"`
}

// Challenge is the aggregate root for game challenges/disputes
type Challenge struct {
	shared.BaseEntity `bson:",inline"`

	// Match Context
	MatchID     uuid.UUID  `json:"match_id" bson:"match_id"`
	RoundNumber *int       `json:"round_number,omitempty" bson:"round_number,omitempty"`
	GameID      string     `json:"game_id" bson:"game_id"`
	LobbyID     *uuid.UUID `json:"lobby_id,omitempty" bson:"lobby_id,omitempty"`
	TournamentID *uuid.UUID `json:"tournament_id,omitempty" bson:"tournament_id,omitempty"`

	// Challenger Info
	ChallengerID uuid.UUID  `json:"challenger_id" bson:"challenger_id"`
	ChallengerTeamID *uuid.UUID `json:"challenger_team_id,omitempty" bson:"challenger_team_id,omitempty"`

	// Challenge Details
	Type        ChallengeType       `json:"type" bson:"type"`
	Title       string              `json:"title" bson:"title"`
	Description string              `json:"description" bson:"description"`
	Priority    ChallengePriority   `json:"priority" bson:"priority"`
	Status      ChallengeStatus     `json:"status" bson:"status"`
	Resolution  ChallengeResolution `json:"resolution" bson:"resolution"`

	// Evidence & Voting
	Evidence    []Evidence    `json:"evidence" bson:"evidence"`
	Votes       []Vote        `json:"votes" bson:"votes"`
	AdminActions []AdminAction `json:"admin_actions" bson:"admin_actions"`

	// Resolution Details
	ResolvedByID    *uuid.UUID `json:"resolved_by_id,omitempty" bson:"resolved_by_id,omitempty"`
	ResolutionNotes string     `json:"resolution_notes,omitempty" bson:"resolution_notes,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`

	// Timeouts
	VotingDeadline  *time.Time `json:"voting_deadline,omitempty" bson:"voting_deadline,omitempty"`
	ReviewDeadline  *time.Time `json:"review_deadline,omitempty" bson:"review_deadline,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`

	// Impact Tracking
	MatchPausedAt    *time.Time `json:"match_paused_at,omitempty" bson:"match_paused_at,omitempty"`
	MatchResumedAt   *time.Time `json:"match_resumed_at,omitempty" bson:"match_resumed_at,omitempty"`
	ScoreBeforeChallenge map[string]int `json:"score_before_challenge,omitempty" bson:"score_before_challenge,omitempty"`
	ScoreAfterResolution map[string]int `json:"score_after_resolution,omitempty" bson:"score_after_resolution,omitempty"`
}

// NewChallenge creates a new challenge
func NewChallenge(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	challengerID uuid.UUID,
	gameID string,
	challengeType ChallengeType,
	title string,
	description string,
	priority ChallengePriority,
) (*Challenge, error) {
	if title == "" {
		return nil, fmt.Errorf("challenge title is required")
	}
	if description == "" {
		return nil, fmt.Errorf("challenge description is required")
	}
	if !challengeType.IsValid() {
		return nil, fmt.Errorf("invalid challenge type: %s", challengeType)
	}
	if !priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}

	// Set default expiration (24 hours for normal challenges)
	expiration := time.Now().UTC().Add(24 * time.Hour)
	if priority == ChallengePriorityCritical || priority == ChallengePriorityHigh {
		// High priority challenges expire in 2 hours
		expiration = time.Now().UTC().Add(2 * time.Hour)
	}

	challenge := &Challenge{
		BaseEntity:   shared.NewEntity(resourceOwner),
		MatchID:      matchID,
		ChallengerID: challengerID,
		GameID:       gameID,
		Type:         challengeType,
		Title:        title,
		Description:  description,
		Priority:     priority,
		Status:       ChallengeStatusPending,
		Resolution:   ChallengeResolutionNone,
		Evidence:     make([]Evidence, 0),
		Votes:        make([]Vote, 0),
		AdminActions: make([]AdminAction, 0),
		ExpiresAt:    &expiration,
	}

	return challenge, nil
}

// GetID implements shared.Entity interface
func (c *Challenge) GetID() uuid.UUID {
	return c.ID
}

// AddEvidence adds supporting evidence to the challenge
func (c *Challenge) AddEvidence(evidenceType string, url string, description string, tickRange *TickRange) error {
	if c.Status == ChallengeStatusResolved || c.Status == ChallengeStatusCancelled {
		return fmt.Errorf("cannot add evidence to closed challenge")
	}

	evidence := Evidence{
		ID:          uuid.New(),
		Type:        evidenceType,
		URL:         url,
		Description: description,
		Timestamp:   time.Now().UTC(),
		TickRange:   tickRange,
		UploadedAt:  time.Now().UTC(),
	}

	c.Evidence = append(c.Evidence, evidence)
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// AddVote records a player vote on the challenge
func (c *Challenge) AddVote(playerID uuid.UUID, teamID *uuid.UUID, voteType string, reason string) error {
	if c.Status != ChallengeStatusVotePending {
		return fmt.Errorf("challenge is not accepting votes (status: %s)", c.Status)
	}

	// Check if player already voted
	for _, v := range c.Votes {
		if v.PlayerID == playerID {
			return fmt.Errorf("player has already voted")
		}
	}

	// Validate vote type
	if voteType != "approve" && voteType != "reject" && voteType != "abstain" {
		return fmt.Errorf("invalid vote type: %s", voteType)
	}

	vote := Vote{
		PlayerID: playerID,
		TeamID:   teamID,
		VoteType: voteType,
		Reason:   reason,
		VotedAt:  time.Now().UTC(),
	}

	c.Votes = append(c.Votes, vote)
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// StartReview transitions challenge to review status
func (c *Challenge) StartReview(adminID uuid.UUID) error {
	if c.Status != ChallengeStatusPending && c.Status != ChallengeStatusVotePending {
		return fmt.Errorf("challenge cannot be reviewed (status: %s)", c.Status)
	}

	c.Status = ChallengeStatusInReview
	c.UpdatedAt = time.Now().UTC()

	// Set review deadline (4 hours for normal, 30 min for critical)
	deadline := time.Now().UTC().Add(4 * time.Hour)
	if c.Priority == ChallengePriorityCritical {
		deadline = time.Now().UTC().Add(30 * time.Minute)
	}
	c.ReviewDeadline = &deadline

	c.AdminActions = append(c.AdminActions, AdminAction{
		AdminID:     adminID,
		Action:      "review_started",
		Notes:       "Review initiated",
		PerformedAt: time.Now().UTC(),
	})

	return nil
}

// StartVoting transitions challenge to voting status
func (c *Challenge) StartVoting(votingDuration time.Duration) error {
	if c.Status != ChallengeStatusPending {
		return fmt.Errorf("challenge cannot start voting (status: %s)", c.Status)
	}

	c.Status = ChallengeStatusVotePending
	deadline := time.Now().UTC().Add(votingDuration)
	c.VotingDeadline = &deadline
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// Approve approves the challenge
func (c *Challenge) Approve(adminID uuid.UUID, resolution ChallengeResolution, notes string) error {
	if c.Status != ChallengeStatusInReview && c.Status != ChallengeStatusVotePending {
		return fmt.Errorf("challenge cannot be approved (status: %s)", c.Status)
	}

	if !resolution.IsValid() || resolution == ChallengeResolutionNone {
		return fmt.Errorf("valid resolution required for approval")
	}

	now := time.Now().UTC()
	c.Status = ChallengeStatusApproved
	c.Resolution = resolution
	c.ResolvedByID = &adminID
	c.ResolutionNotes = notes
	c.ResolvedAt = &now
	c.UpdatedAt = now

	c.AdminActions = append(c.AdminActions, AdminAction{
		AdminID:     adminID,
		Action:      "approved",
		Notes:       notes,
		PerformedAt: now,
	})

	return nil
}

// Reject rejects the challenge
func (c *Challenge) Reject(adminID uuid.UUID, reason string) error {
	if c.Status != ChallengeStatusInReview && c.Status != ChallengeStatusVotePending {
		return fmt.Errorf("challenge cannot be rejected (status: %s)", c.Status)
	}

	now := time.Now().UTC()
	c.Status = ChallengeStatusRejected
	c.Resolution = ChallengeResolutionNoAction
	c.ResolvedByID = &adminID
	c.ResolutionNotes = reason
	c.ResolvedAt = &now
	c.UpdatedAt = now

	c.AdminActions = append(c.AdminActions, AdminAction{
		AdminID:     adminID,
		Action:      "rejected",
		Notes:       reason,
		PerformedAt: now,
	})

	return nil
}

// Resolve marks the challenge as fully resolved
func (c *Challenge) Resolve() error {
	if c.Status != ChallengeStatusApproved {
		return fmt.Errorf("only approved challenges can be resolved (status: %s)", c.Status)
	}

	c.Status = ChallengeStatusResolved
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// Cancel cancels the challenge
func (c *Challenge) Cancel(reason string) error {
	if c.Status == ChallengeStatusResolved {
		return fmt.Errorf("cannot cancel resolved challenge")
	}

	c.Status = ChallengeStatusCancelled
	c.ResolutionNotes = reason
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// MarkExpired marks the challenge as expired
func (c *Challenge) MarkExpired() error {
	if c.Status == ChallengeStatusResolved || c.Status == ChallengeStatusCancelled {
		return fmt.Errorf("cannot expire closed challenge")
	}

	c.Status = ChallengeStatusExpired
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// PauseMatch records that the match was paused for this challenge
func (c *Challenge) PauseMatch(currentScore map[string]int) error {
	if c.MatchPausedAt != nil {
		return fmt.Errorf("match already paused for this challenge")
	}

	now := time.Now().UTC()
	c.MatchPausedAt = &now
	c.ScoreBeforeChallenge = currentScore
	c.UpdatedAt = now

	return nil
}

// ResumeMatch records that the match was resumed
func (c *Challenge) ResumeMatch(newScore map[string]int) error {
	if c.MatchPausedAt == nil {
		return fmt.Errorf("match was not paused")
	}
	if c.MatchResumedAt != nil {
		return fmt.Errorf("match already resumed")
	}

	now := time.Now().UTC()
	c.MatchResumedAt = &now
	c.ScoreAfterResolution = newScore
	c.UpdatedAt = now

	return nil
}

// GetVoteCounts returns counts of approve/reject/abstain votes
func (c *Challenge) GetVoteCounts() (approve int, reject int, abstain int) {
	for _, v := range c.Votes {
		switch v.VoteType {
		case "approve":
			approve++
		case "reject":
			reject++
		case "abstain":
			abstain++
		}
	}
	return
}

// IsExpired returns true if the challenge has expired
func (c *Challenge) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*c.ExpiresAt)
}

// IsVotingExpired returns true if the voting deadline has passed
func (c *Challenge) IsVotingExpired() bool {
	if c.VotingDeadline == nil {
		return false
	}
	return time.Now().UTC().After(*c.VotingDeadline)
}

// IsReviewExpired returns true if the review deadline has passed
func (c *Challenge) IsReviewExpired() bool {
	if c.ReviewDeadline == nil {
		return false
	}
	return time.Now().UTC().After(*c.ReviewDeadline)
}

// CanBeReviewedBy checks if an admin can review this challenge
func (c *Challenge) CanBeReviewedBy(adminID uuid.UUID) bool {
	// Admin cannot review their own challenge
	if c.ChallengerID == adminID {
		return false
	}
	return c.Status == ChallengeStatusPending || c.Status == ChallengeStatusVotePending
}

// Validate ensures the challenge state is consistent
func (c *Challenge) Validate() error {
	if c.MatchID == uuid.Nil {
		return fmt.Errorf("match ID is required")
	}
	if c.ChallengerID == uuid.Nil {
		return fmt.Errorf("challenger ID is required")
	}
	if c.Title == "" {
		return fmt.Errorf("title is required")
	}
	if c.Description == "" {
		return fmt.Errorf("description is required")
	}
	if !c.Type.IsValid() {
		return fmt.Errorf("invalid challenge type")
	}
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid status")
	}
	if !c.Priority.IsValid() {
		return fmt.Errorf("invalid priority")
	}

	// Resolved challenges must have resolution details
	if c.Status == ChallengeStatusResolved || c.Status == ChallengeStatusApproved {
		if c.Resolution == ChallengeResolutionNone {
			return fmt.Errorf("resolved challenge must have a resolution")
		}
		if c.ResolvedByID == nil {
			return fmt.Errorf("resolved challenge must have resolver ID")
		}
	}

	return nil
}

// IsValid returns true if the challenge type is valid
func (t ChallengeType) IsValid() bool {
	switch t {
	case ChallengeTypeBugReport, ChallengeTypeVAR, ChallengeTypeRoundRestart,
		ChallengeTypeMatchRestart, ChallengeTypeTechnicalIssue,
		ChallengeTypeRuleViolation, ChallengeTypeScoreDispute:
		return true
	}
	return false
}

// IsValid returns true if the status is valid
func (s ChallengeStatus) IsValid() bool {
	switch s {
	case ChallengeStatusPending, ChallengeStatusInReview, ChallengeStatusVotePending,
		ChallengeStatusApproved, ChallengeStatusRejected, ChallengeStatusResolved,
		ChallengeStatusExpired, ChallengeStatusCancelled:
		return true
	}
	return false
}

// IsValid returns true if the priority is valid
func (p ChallengePriority) IsValid() bool {
	switch p {
	case ChallengePriorityLow, ChallengePriorityNormal, ChallengePriorityHigh, ChallengePriorityCritical:
		return true
	}
	return false
}

// IsValid returns true if the resolution is valid
func (r ChallengeResolution) IsValid() bool {
	switch r {
	case ChallengeResolutionNone, ChallengeResolutionRoundRestarted, ChallengeResolutionMatchRestarted,
		ChallengeResolutionScoreAdjusted, ChallengeResolutionPenaltyApplied, ChallengeResolutionNoAction,
		ChallengeResolutionMatchVoided, ChallengeResolutionCompensation:
		return true
	}
	return false
}

