# Replay API - REST Endpoints Reference

> Complete API endpoint documentation for the LeetGaming Replay API

## Base URL

- **Local Development**: `http://localhost:30800`
- **Staging**: `https://api-staging.leetgaming.pro`
- **Production**: `https://api.leetgaming.pro`

---

## Authentication

All endpoints (except `/health` and `/onboarding/*`) require authentication via JWT token.

```http
Authorization: Bearer <jwt_token>
```

---

## Endpoints

### Health & Monitoring

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |
| GET | `/metrics` | Prometheus metrics |
| GET | `/coverage` | Code coverage report (CI only) |

---

### Authentication & Onboarding

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/onboarding/steam` | Authenticate via Steam OAuth |
| POST | `/onboarding/google` | Authenticate via Google OAuth |

#### POST `/onboarding/steam`

Authenticates a user using Steam OpenID and creates/retrieves their account.

**Request Body:**
```json
{
  "steam_id": "76561198012345678",
  "persona_name": "PlayerOne",
  "avatar_url": "https://steamcdn-a.akamaihd.net/..."
}
```

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "rid_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-11-29T00:00:00Z"
}
```

---

### Match API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/games/{game_id}/match` | List matches |
| GET | `/games/{game_id}/match/{match_id}` | Get match details |

#### GET `/games/{game_id}/match`

List matches with optional filters.

**Query Parameters:**
- `limit` (int): Max results (default: 20)
- `offset` (int): Pagination offset
- `player_id` (uuid): Filter by player
- `map` (string): Filter by map name

**Response:**
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "game_id": "cs2",
      "map": "de_dust2",
      "duration": 2400,
      "created_at": "2025-11-28T12:00:00Z"
    }
  ],
  "total_count": 150,
  "limit": 20,
  "offset": 0
}
```

---

### Replay File API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/games/{game_id}/replays` | Upload replay file |
| GET | `/games/{game_id}/replays/{id}` | Get replay metadata |
| PUT | `/games/{game_id}/replays/{id}` | Update replay metadata |
| DELETE | `/games/{game_id}/replays/{id}` | Delete replay file |
| GET | `/games/{game_id}/replays/{id}/download` | Download replay file |
| GET | `/games/{game_id}/replays/{id}/status` | Get processing status |

#### POST `/games/{game_id}/replays`

Upload a replay file for processing.

**Request:** `multipart/form-data`
- `file`: The replay file (.dem, .replay, etc.)
- `visibility`: `public` | `private` | `unlisted`

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "file_name": "match_2025.dem",
  "file_size": 15728640,
  "created_at": "2025-11-28T12:00:00Z"
}
```

---

### Game Events API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/games/{game_id}/events` | List game events |

---

### Player Profile API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/players` | Create player profile |
| GET | `/players` | Search player profiles |
| GET | `/players/{id}` | Get player profile |
| PUT | `/players/{id}` | Update player profile |
| DELETE | `/players/{id}` | Delete player profile |

#### GET `/players`

Search player profiles.

**Query Parameters:**
- `q` (string): Search query
- `game_id` (string): Filter by game
- `limit` (int): Max results
- `offset` (int): Pagination offset

---

### Squad API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/squads` | Create squad |
| GET | `/squads/{id}` | Get squad details |
| PUT | `/squads/{id}` | Update squad |
| DELETE | `/squads/{id}` | Delete squad |
| POST | `/squads/{id}/members` | Add member |
| DELETE | `/squads/{id}/members/{player_id}` | Remove member |
| PUT | `/squads/{id}/members/{player_id}/role` | Update member role |

#### POST `/squads`

Create a new squad/team.

**Request Body:**
```json
{
  "name": "Team Alpha",
  "tag": "ALPHA",
  "game_id": "cs2",
  "visibility": "public"
}
```

---

### Share Token API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/share-tokens` | Create share token |
| GET | `/share-tokens` | List user's share tokens |
| GET | `/share-tokens/{token}` | Get shared resource |
| DELETE | `/share-tokens/{token}` | Revoke share token |

---

### Matchmaking API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/match-making/queue` | Join matchmaking queue |
| DELETE | `/match-making/queue/{session_id}` | Leave queue |
| GET | `/match-making/session/{session_id}` | Get session status |
| GET | `/match-making/pools/{game_id}` | Get pool statistics |

---

### Prize Pool Lobby API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/lobbies` | Create lobby |
| POST | `/api/lobbies/{lobby_id}/join` | Join lobby |
| DELETE | `/api/lobbies/{lobby_id}/leave` | Leave lobby |
| PUT | `/api/lobbies/{lobby_id}/ready` | Set player ready |
| POST | `/api/lobbies/{lobby_id}/start` | Start match |
| DELETE | `/api/lobbies/{lobby_id}` | Cancel lobby |
| WS | `/ws/lobby/{lobby_id}` | Real-time lobby updates |

#### POST `/api/lobbies`

Create a new prize pool lobby.

**Request Body:**
```json
{
  "game_id": "cs2",
  "region": "na-east",
  "tier": "gold",
  "distribution_rule": "winner_takes_all",
  "max_players": 10,
  "entry_fee": {
    "currency": "USD",
    "amount": "10.00"
  }
}
```

---

### Prize Pool API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/prize-pools/{id}` | Get prize pool details |
| GET | `/prize-pools/{id}/history` | Get prize pool history |
| GET | `/matches/{match_id}/prize-pool` | Get match prize pool |
| GET | `/prize-pools/pending-distributions` | Get pending distributions |

---

### Tournament API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/tournaments` | Create tournament |
| GET | `/tournaments` | List tournaments |
| GET | `/tournaments/upcoming` | Get upcoming tournaments |
| GET | `/tournaments/{id}` | Get tournament details |
| PUT | `/tournaments/{id}` | Update tournament |
| DELETE | `/tournaments/{id}` | Delete tournament |
| POST | `/tournaments/{id}/register` | Register for tournament |
| DELETE | `/tournaments/{id}/register` | Unregister from tournament |
| POST | `/tournaments/{id}/start` | Start tournament |
| GET | `/players/{player_id}/tournaments` | Get player's tournaments |
| GET | `/organizers/{organizer_id}/tournaments` | Get organizer's tournaments |

#### POST `/tournaments`

Create a new tournament.

**Request Body:**
```json
{
  "name": "Weekend Showdown",
  "description": "CS2 1v1 tournament",
  "game_id": "cs2",
  "game_mode": "1v1",
  "region": "na-east",
  "format": "single_elimination",
  "max_participants": 32,
  "min_participants": 8,
  "entry_fee": {
    "currency": "USD",
    "amount": "5.00"
  },
  "start_time": "2025-11-30T18:00:00Z",
  "registration_open": "2025-11-28T00:00:00Z",
  "registration_close": "2025-11-30T17:00:00Z"
}
```

---

### Wallet API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/wallet/balance` | Get wallet balance |
| GET | `/wallet/transactions` | Get transaction history |

#### GET `/wallet/balance`

Get current wallet balance for all currencies.

**Response:**
```json
{
  "evm_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f1e001",
  "balances": {
    "USD": { "cents": 123456 },
    "USDC": { "cents": 50000 },
    "USDT": { "cents": 25000 }
  },
  "total_deposited": { "cents": 500000 },
  "total_withdrawn": { "cents": 300000 },
  "total_prizes_won": { "cents": 78456 },
  "is_locked": false
}
```

---

### Payment API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/payments` | Create payment intent |
| GET | `/payments` | Get user's payments |
| GET | `/payments/{payment_id}` | Get payment details |
| POST | `/payments/{payment_id}/confirm` | Confirm payment |
| POST | `/payments/{payment_id}/cancel` | Cancel payment |
| POST | `/payments/{payment_id}/refund` | Refund payment |
| POST | `/webhooks/stripe` | Stripe webhook (no auth) |

#### POST `/payments`

Create a new payment intent.

**Request Body:**
```json
{
  "amount": {
    "currency": "USD",
    "cents": 5000
  },
  "payment_method": "stripe",
  "description": "Wallet deposit"
}
```

---

### Search API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/search/{query}` | Unified search across all entities |

#### GET `/search/{query}`

Search across matches, players, teams, and replays.

**Query Parameters:**
- `type`: Filter by entity type (`match`, `player`, `squad`, `replay`)
- `limit`: Max results
- `offset`: Pagination offset

---

### IAM API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/groups` | List user's group memberships |

---

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body",
    "details": [
      {
        "field": "entry_fee",
        "message": "must be positive"
      }
    ]
  }
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Validation error |
| 401 | Unauthorized - Missing/invalid token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Business rule violation |
| 429 | Too Many Requests - Rate limited |
| 500 | Internal Server Error |

---

## Rate Limiting

- **Default**: 100 requests per minute per IP
- **Authenticated**: 200 requests per minute per user
- **Premium**: 500 requests per minute per user

Rate limit headers:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1701187200
```

---

## WebSocket Endpoints

### `/ws/lobby/{lobby_id}`

Real-time lobby updates.

**Events:**
- `player_joined` - A player joined the lobby
- `player_left` - A player left the lobby
- `player_ready` - A player marked ready
- `lobby_starting` - Match is starting
- `lobby_cancelled` - Lobby was cancelled

**Example Message:**
```json
{
  "type": "player_joined",
  "data": {
    "player_id": "550e8400-e29b-41d4-a716-446655440000",
    "display_name": "PlayerOne",
    "slot": 3
  },
  "timestamp": "2025-11-28T12:00:00Z"
}
```

---

**Last Updated**: November 2025
