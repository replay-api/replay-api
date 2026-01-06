package challenge_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestResourceOwner() shared.ResourceOwner {
	return shared.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
	}
}

func TestNewChallenge_Success(t *testing.T) {
	resourceOwner := createTestResourceOwner()
	matchID := uuid.New()
	challengerID := uuid.New()

	challenge, err := NewChallenge(
		resourceOwner,
		matchID,
		challengerID,
		"cs2",
		ChallengeTypeBugReport,
		"Game Breaking Bug",
		"Player clipped through wall",
		ChallengePriorityHigh,
	)

	require.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, matchID, challenge.MatchID)
	assert.Equal(t, challengerID, challenge.ChallengerID)
	assert.Equal(t, "cs2", challenge.GameID)
	assert.Equal(t, ChallengeTypeBugReport, challenge.Type)
	assert.Equal(t, "Game Breaking Bug", challenge.Title)
	assert.Equal(t, ChallengePriorityHigh, challenge.Priority)
	assert.Equal(t, ChallengeStatusPending, challenge.Status)
	assert.Equal(t, ChallengeResolutionNone, challenge.Resolution)
	assert.Empty(t, challenge.Evidence)
	assert.Empty(t, challenge.Votes)
	assert.NotNil(t, challenge.ExpiresAt)
}

func TestNewChallenge_EmptyTitle_ReturnsError(t *testing.T) {
	resourceOwner := createTestResourceOwner()

	challenge, err := NewChallenge(
		resourceOwner,
		uuid.New(),
		uuid.New(),
		"cs2",
		ChallengeTypeBugReport,
		"", // Empty title
		"Description",
		ChallengePriorityNormal,
	)

	assert.Error(t, err)
	assert.Nil(t, challenge)
	assert.Contains(t, err.Error(), "title is required")
}

func TestNewChallenge_EmptyDescription_ReturnsError(t *testing.T) {
	resourceOwner := createTestResourceOwner()

	challenge, err := NewChallenge(
		resourceOwner,
		uuid.New(),
		uuid.New(),
		"cs2",
		ChallengeTypeBugReport,
		"Title",
		"", // Empty description
		ChallengePriorityNormal,
	)

	assert.Error(t, err)
	assert.Nil(t, challenge)
	assert.Contains(t, err.Error(), "description is required")
}

func TestNewChallenge_InvalidType_ReturnsError(t *testing.T) {
	resourceOwner := createTestResourceOwner()

	challenge, err := NewChallenge(
		resourceOwner,
		uuid.New(),
		uuid.New(),
		"cs2",
		ChallengeType("invalid_type"),
		"Title",
		"Description",
		ChallengePriorityNormal,
	)

	assert.Error(t, err)
	assert.Nil(t, challenge)
	assert.Contains(t, err.Error(), "invalid challenge type")
}

func TestNewChallenge_CriticalPriority_ShortExpiration(t *testing.T) {
	resourceOwner := createTestResourceOwner()

	challenge, err := NewChallenge(
		resourceOwner,
		uuid.New(),
		uuid.New(),
		"cs2",
		ChallengeTypeVAR,
		"Critical Issue",
		"Match needs to be paused",
		ChallengePriorityCritical,
	)

	require.NoError(t, err)
	assert.NotNil(t, challenge.ExpiresAt)
	// Critical challenges expire in ~2 hours
	expectedExpiry := time.Now().UTC().Add(2*time.Hour + time.Minute)
	assert.True(t, challenge.ExpiresAt.Before(expectedExpiry))
}

func TestChallenge_AddEvidence_Success(t *testing.T) {
	challenge := createTestChallenge(t)

	tickRange := &TickRange{StartTick: 1000, EndTick: 1500}
	err := challenge.AddEvidence("screenshot", "https://cdn.example.com/evidence.png", "Bug visible here", tickRange)

	require.NoError(t, err)
	assert.Len(t, challenge.Evidence, 1)
	assert.Equal(t, "screenshot", challenge.Evidence[0].Type)
	assert.Equal(t, "https://cdn.example.com/evidence.png", challenge.Evidence[0].URL)
	assert.Equal(t, int64(1000), challenge.Evidence[0].TickRange.StartTick)
}

func TestChallenge_AddEvidence_ResolvedChallenge_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusResolved

	err := challenge.AddEvidence("screenshot", "https://example.com/img.png", "Description", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot add evidence to closed challenge")
}

func TestChallenge_AddVote_Success(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusVotePending
	playerID := uuid.New()
	teamID := uuid.New()

	err := challenge.AddVote(playerID, &teamID, "approve", "Valid bug report")

	require.NoError(t, err)
	assert.Len(t, challenge.Votes, 1)
	assert.Equal(t, playerID, challenge.Votes[0].PlayerID)
	assert.Equal(t, "approve", challenge.Votes[0].VoteType)
}

func TestChallenge_AddVote_DuplicateVote_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusVotePending
	playerID := uuid.New()

	err := challenge.AddVote(playerID, nil, "approve", "")
	require.NoError(t, err)

	err = challenge.AddVote(playerID, nil, "reject", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already voted")
}

func TestChallenge_AddVote_InvalidVoteType_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusVotePending

	err := challenge.AddVote(uuid.New(), nil, "invalid_vote", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid vote type")
}

func TestChallenge_AddVote_WrongStatus_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	// Status is Pending, not VotePending

	err := challenge.AddVote(uuid.New(), nil, "approve", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not accepting votes")
}

func TestChallenge_StartReview_Success(t *testing.T) {
	challenge := createTestChallenge(t)
	adminID := uuid.New()

	err := challenge.StartReview(adminID)

	require.NoError(t, err)
	assert.Equal(t, ChallengeStatusInReview, challenge.Status)
	assert.NotNil(t, challenge.ReviewDeadline)
	assert.Len(t, challenge.AdminActions, 1)
	assert.Equal(t, "review_started", challenge.AdminActions[0].Action)
}

func TestChallenge_StartVoting_Success(t *testing.T) {
	challenge := createTestChallenge(t)

	err := challenge.StartVoting(5 * time.Minute)

	require.NoError(t, err)
	assert.Equal(t, ChallengeStatusVotePending, challenge.Status)
	assert.NotNil(t, challenge.VotingDeadline)
}

func TestChallenge_Approve_Success(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusInReview
	adminID := uuid.New()

	err := challenge.Approve(adminID, ChallengeResolutionRoundRestarted, "Bug confirmed, round restarted")

	require.NoError(t, err)
	assert.Equal(t, ChallengeStatusApproved, challenge.Status)
	assert.Equal(t, ChallengeResolutionRoundRestarted, challenge.Resolution)
	assert.Equal(t, &adminID, challenge.ResolvedByID)
	assert.NotNil(t, challenge.ResolvedAt)
}

func TestChallenge_Approve_NoResolution_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusInReview

	err := challenge.Approve(uuid.New(), ChallengeResolutionNone, "Notes")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "valid resolution required")
}

func TestChallenge_Reject_Success(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusInReview
	adminID := uuid.New()

	err := challenge.Reject(adminID, "No evidence of bug")

	require.NoError(t, err)
	assert.Equal(t, ChallengeStatusRejected, challenge.Status)
	assert.Equal(t, ChallengeResolutionNoAction, challenge.Resolution)
	assert.Equal(t, "No evidence of bug", challenge.ResolutionNotes)
}

func TestChallenge_Resolve_Success(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusApproved

	err := challenge.Resolve()

	require.NoError(t, err)
	assert.Equal(t, ChallengeStatusResolved, challenge.Status)
}

func TestChallenge_Cancel_Success(t *testing.T) {
	challenge := createTestChallenge(t)

	err := challenge.Cancel("Challenger withdrew claim")

	require.NoError(t, err)
	assert.Equal(t, ChallengeStatusCancelled, challenge.Status)
	assert.Equal(t, "Challenger withdrew claim", challenge.ResolutionNotes)
}

func TestChallenge_Cancel_ResolvedChallenge_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusResolved

	err := challenge.Cancel("Reason")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel resolved challenge")
}

func TestChallenge_PauseMatch_Success(t *testing.T) {
	challenge := createTestChallenge(t)
	score := map[string]int{"team1": 5, "team2": 3}

	err := challenge.PauseMatch(score)

	require.NoError(t, err)
	assert.NotNil(t, challenge.MatchPausedAt)
	assert.Equal(t, score, challenge.ScoreBeforeChallenge)
}

func TestChallenge_PauseMatch_AlreadyPaused_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	now := time.Now().UTC()
	challenge.MatchPausedAt = &now

	err := challenge.PauseMatch(map[string]int{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already paused")
}

func TestChallenge_ResumeMatch_Success(t *testing.T) {
	challenge := createTestChallenge(t)
	now := time.Now().UTC()
	challenge.MatchPausedAt = &now
	challenge.ScoreBeforeChallenge = map[string]int{"team1": 5, "team2": 3}
	newScore := map[string]int{"team1": 4, "team2": 3} // Score adjusted after restart

	err := challenge.ResumeMatch(newScore)

	require.NoError(t, err)
	assert.NotNil(t, challenge.MatchResumedAt)
	assert.Equal(t, newScore, challenge.ScoreAfterResolution)
}

func TestChallenge_GetVoteCounts(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Votes = []Vote{
		{PlayerID: uuid.New(), VoteType: "approve", VotedAt: time.Now()},
		{PlayerID: uuid.New(), VoteType: "approve", VotedAt: time.Now()},
		{PlayerID: uuid.New(), VoteType: "reject", VotedAt: time.Now()},
		{PlayerID: uuid.New(), VoteType: "abstain", VotedAt: time.Now()},
	}

	approve, reject, abstain := challenge.GetVoteCounts()

	assert.Equal(t, 2, approve)
	assert.Equal(t, 1, reject)
	assert.Equal(t, 1, abstain)
}

func TestChallenge_IsExpired(t *testing.T) {
	challenge := createTestChallenge(t)

	// Not expired with future expiry
	futureTime := time.Now().UTC().Add(1 * time.Hour)
	challenge.ExpiresAt = &futureTime
	assert.False(t, challenge.IsExpired())

	// Expired with past expiry
	pastTime := time.Now().UTC().Add(-1 * time.Hour)
	challenge.ExpiresAt = &pastTime
	assert.True(t, challenge.IsExpired())

	// No expiry set
	challenge.ExpiresAt = nil
	assert.False(t, challenge.IsExpired())
}

func TestChallenge_CanBeReviewedBy(t *testing.T) {
	challenge := createTestChallenge(t)
	challengerID := challenge.ChallengerID
	adminID := uuid.New()

	// Admin can review pending challenge
	assert.True(t, challenge.CanBeReviewedBy(adminID))

	// Challenger cannot review their own challenge
	assert.False(t, challenge.CanBeReviewedBy(challengerID))

	// Cannot review resolved challenge
	challenge.Status = ChallengeStatusResolved
	assert.False(t, challenge.CanBeReviewedBy(adminID))
}

func TestChallenge_Validate_Success(t *testing.T) {
	challenge := createTestChallenge(t)

	err := challenge.Validate()

	assert.NoError(t, err)
}

func TestChallenge_Validate_MissingMatchID_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.MatchID = uuid.Nil

	err := challenge.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "match ID is required")
}

func TestChallenge_Validate_ResolvedWithoutResolver_ReturnsError(t *testing.T) {
	challenge := createTestChallenge(t)
	challenge.Status = ChallengeStatusResolved
	challenge.Resolution = ChallengeResolutionRoundRestarted
	challenge.ResolvedByID = nil

	err := challenge.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolved challenge must have resolver ID")
}

func TestChallengeType_IsValid(t *testing.T) {
	validTypes := []ChallengeType{
		ChallengeTypeBugReport,
		ChallengeTypeVAR,
		ChallengeTypeRoundRestart,
		ChallengeTypeMatchRestart,
		ChallengeTypeTechnicalIssue,
		ChallengeTypeRuleViolation,
		ChallengeTypeScoreDispute,
	}

	for _, ct := range validTypes {
		assert.True(t, ct.IsValid(), "Expected %s to be valid", ct)
	}

	invalidType := ChallengeType("invalid")
	assert.False(t, invalidType.IsValid())
}

func TestChallengeStatus_IsValid(t *testing.T) {
	validStatuses := []ChallengeStatus{
		ChallengeStatusPending,
		ChallengeStatusInReview,
		ChallengeStatusVotePending,
		ChallengeStatusApproved,
		ChallengeStatusRejected,
		ChallengeStatusResolved,
		ChallengeStatusExpired,
		ChallengeStatusCancelled,
	}

	for _, s := range validStatuses {
		assert.True(t, s.IsValid(), "Expected %s to be valid", s)
	}

	invalidStatus := ChallengeStatus("invalid")
	assert.False(t, invalidStatus.IsValid())
}

func TestChallengePriority_IsValid(t *testing.T) {
	validPriorities := []ChallengePriority{
		ChallengePriorityLow,
		ChallengePriorityNormal,
		ChallengePriorityHigh,
		ChallengePriorityCritical,
	}

	for _, p := range validPriorities {
		assert.True(t, p.IsValid(), "Expected %s to be valid", p)
	}

	invalidPriority := ChallengePriority("invalid")
	assert.False(t, invalidPriority.IsValid())
}

func TestChallengeResolution_IsValid(t *testing.T) {
	validResolutions := []ChallengeResolution{
		ChallengeResolutionNone,
		ChallengeResolutionRoundRestarted,
		ChallengeResolutionMatchRestarted,
		ChallengeResolutionScoreAdjusted,
		ChallengeResolutionPenaltyApplied,
		ChallengeResolutionNoAction,
		ChallengeResolutionMatchVoided,
		ChallengeResolutionCompensation,
	}

	for _, r := range validResolutions {
		assert.True(t, r.IsValid(), "Expected %s to be valid", r)
	}

	invalidResolution := ChallengeResolution("invalid")
	assert.False(t, invalidResolution.IsValid())
}

// Helper function to create a test challenge
func createTestChallenge(t *testing.T) *Challenge {
	resourceOwner := createTestResourceOwner()
	challenge, err := NewChallenge(
		resourceOwner,
		uuid.New(),
		uuid.New(),
		"cs2",
		ChallengeTypeBugReport,
		"Test Challenge",
		"Test description for challenge",
		ChallengePriorityNormal,
	)
	require.NoError(t, err)
	return challenge
}

