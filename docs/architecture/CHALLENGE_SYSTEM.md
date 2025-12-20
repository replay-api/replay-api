# Challenge System Architecture

## Overview

The Challenge System provides a comprehensive framework for match integrity verification, including VAR (Video Assistant Referee) reviews, round restarts, bug reports, and player disputes. It enables transparent dispute resolution with community voting and admin oversight.

## Domain Model

### Core Entities

#### Challenge

The main entity representing a challenge against a match or round.

```go
type Challenge struct {
    ID                 string
    MatchID            string             // Reference to the match
    RoundNumber        *int               // Optional specific round
    ChallengerID       string             // Player who submitted
    Type               ChallengeType      // var_review, round_restart, etc.
    Title              string
    Description        string
    Status             ChallengeStatus    // pending, under_review, voting, resolved, etc.
    Priority           ChallengePriority  // low, normal, high, critical
    Evidence           []Evidence         // Supporting evidence
    Votes              []Vote             // Community votes
    Resolution         *ChallengeResolution
    AdminNotes         string
    AffectedPlayerIDs  []string
    PausedMatch        bool               // Whether match was paused
    CreatedAt          time.Time
    UpdatedAt          time.Time
}
```

#### Evidence

Supporting evidence attached to a challenge.

```go
type Evidence struct {
    ID          string
    Type        string    // screenshot, replay_clip, log, video
    URL         string
    Description string
    Timestamp   time.Time
    TickRange   *TickRange
    UploadedAt  time.Time
}
```

#### Vote

Community vote on a challenge.

```go
type Vote struct {
    VoterID   string
    Approved  bool
    Reason    string
    VotedAt   time.Time
}
```

### Enums

- **ChallengeType**: `var_review`, `round_restart`, `bug_report`, `admin_decision`, `player_dispute`
- **ChallengeStatus**: `pending`, `under_review`, `voting`, `resolved`, `cancelled`, `escalated`
- **ChallengePriority**: `low`, `normal`, `high`, `critical`
- **ChallengeResolution**: `upheld`, `rejected`, `partial`, `penalty_applied`, `no_action`, `match_voided`, `compensation`

## Use Cases

### 1. Create Challenge (CreateChallengeUseCase)

**Actor**: Authenticated Player

**Flow**:
1. Player submits challenge with type, description, and optional evidence
2. System validates player is part of the match
3. Challenge is created with `pending` status
4. For critical priority: match may be paused automatically
5. Affected parties are notified

**Business Rules**:
- Only match participants can create challenges
- Maximum 3 active challenges per match per player
- Critical challenges require additional verification

### 2. Add Evidence (AddEvidenceUseCase)

**Actor**: Challenger or Admin

**Flow**:
1. User uploads evidence (screenshot, video, replay clip)
2. System validates evidence format and size
3. Evidence is attached to challenge
4. Timestamp and tick range are recorded

**Business Rules**:
- Evidence must be uploaded within 24 hours of challenge creation
- Maximum 10 pieces of evidence per challenge

### 3. Vote on Challenge (VoteOnChallengeUseCase)

**Actor**: Eligible Voter (match participant or community member)

**Flow**:
1. Challenge must be in `voting` status
2. Voter submits approval/rejection with reason
3. Vote is recorded
4. Vote threshold checked for auto-resolution

**Business Rules**:
- One vote per player per challenge
- Voters cannot vote on their own challenges
- Minimum vote threshold: 3 votes
- Super-majority (66%) required for resolution

### 4. Resolve Challenge (ResolveChallengeUseCase)

**Actor**: Admin or Automatic (vote threshold)

**Flow**:
1. Admin reviews evidence and votes
2. Resolution decision is made
3. Challenge status set to `resolved`
4. Affected players notified
5. Match adjustments applied if needed

**Business Rules**:
- Only admins can override community votes
- Resolution must include justification
- Penalties require separate workflow

### 5. Cancel Challenge (CancelChallengeUseCase)

**Actor**: Challenger or Admin

**Flow**:
1. Challenge must be in `pending` or `under_review` status
2. Cancellation reason provided
3. Challenge status set to `cancelled`

**Business Rules**:
- Cannot cancel challenges in `voting` or `resolved` status
- Admins can cancel any challenge

## API Endpoints

### Commands (Write Operations)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/challenges` | Create new challenge |
| POST | `/challenges/{id}/evidence` | Add evidence to challenge |
| POST | `/challenges/{id}/vote` | Submit vote on challenge |
| PUT | `/challenges/{id}/resolve` | Resolve challenge (admin) |
| DELETE | `/challenges/{id}` | Cancel challenge |

### Queries (Read Operations)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/challenges` | List/search challenges |
| GET | `/challenges/{id}` | Get challenge details |
| GET | `/matches/{id}/challenges` | Get challenges for a match |

## State Machine

```
                    ┌─────────────┐
                    │   pending   │
                    └──────┬──────┘
                           │ start_review()
                    ┌──────▼──────┐
       cancel() ───►│ under_review│◄─── escalate()
                    └──────┬──────┘         │
                           │ start_voting() │
                    ┌──────▼──────┐    ┌────▼────┐
                    │   voting    │    │escalated│
                    └──────┬──────┘    └────┬────┘
                           │ resolve()      │
                    ┌──────▼──────┐         │
                    │  resolved   │◄────────┘
                    └─────────────┘

       ┌─────────────┐
       │  cancelled  │ (terminal state from pending/under_review)
       └─────────────┘
```

## Resource Ownership

The Challenge domain implements full resource ownership validation:

- **Challenger**: Owner of the challenge, can add evidence and cancel (while pending)
- **Match Participants**: Can vote on challenges (except their own)
- **Admins**: Full access to all operations

```go
// Resource ownership check in use case
resourceOwner := common.GetResourceOwner(ctx)
if resourceOwner == nil {
    return nil, common.NewErrUnauthenticated()
}

if challenge.ChallengerID != resourceOwner.UserID {
    return nil, common.NewErrForbidden("only challenger can add evidence")
}
```

## Integration Points

### Match Domain
- Challenge creation validates match existence
- Match can be paused when critical challenge is filed
- Match results can be voided upon challenge resolution

### Player Domain
- Player eligibility for voting
- Player reputation tracking based on challenge outcomes

### Notification Domain
- Real-time notifications for status changes
- Email alerts for resolution

### Replay Domain
- Evidence can reference specific replay ticks
- Replay clips automatically generated for VAR reviews

## MongoDB Schema

```javascript
// challenges collection
{
  _id: ObjectId,
  match_id: ObjectId,
  round_number: Number,
  challenger_id: ObjectId,
  type: String,
  title: String,
  description: String,
  status: String,
  priority: String,
  evidence: [
    {
      id: String,
      type: String,
      url: String,
      description: String,
      timestamp: Date,
      tick_range: { start_tick: Number, end_tick: Number },
      uploaded_at: Date
    }
  ],
  votes: [
    {
      voter_id: ObjectId,
      approved: Boolean,
      reason: String,
      voted_at: Date
    }
  ],
  resolution: String,
  admin_notes: String,
  admin_actions: [
    {
      admin_id: ObjectId,
      action: String,
      notes: String,
      timestamp: Date
    }
  ],
  affected_player_ids: [ObjectId],
  paused_match: Boolean,
  created_at: Date,
  updated_at: Date
}

// Indexes
db.challenges.createIndex({ match_id: 1 })
db.challenges.createIndex({ challenger_id: 1 })
db.challenges.createIndex({ status: 1 })
db.challenges.createIndex({ type: 1 })
db.challenges.createIndex({ created_at: -1 })
db.challenges.createIndex({ "votes.voter_id": 1 })
```

## Testing Strategy

### Unit Tests
- Entity business logic (state transitions, validations)
- Use case logic with mocked repository
- Coverage target: ≥85%

### Integration Tests
- Repository operations against test MongoDB
- Full use case flow with real dependencies

### E2E Tests
- API endpoint testing
- Authentication and authorization flows
- Concurrency scenarios (simultaneous votes)

## Security Considerations

1. **Authentication**: All endpoints require valid JWT
2. **Authorization**: Resource ownership validated per operation
3. **Rate Limiting**: 10 challenges per hour per player
4. **Input Validation**: Title, description length limits
5. **Evidence Upload**: File type validation, size limits (50MB max)
6. **Vote Integrity**: One vote per player, cryptographic verification

## Observability

### Metrics
- `challenge_created_total` - Counter by type
- `challenge_resolved_total` - Counter by resolution
- `challenge_duration_seconds` - Histogram of resolution time
- `challenge_votes_total` - Counter of votes cast

### Logs
- Challenge creation with match context
- State transitions with actor
- Resolution decisions with evidence summary

### Tracing
- Span for each use case execution
- Correlation with match domain operations

