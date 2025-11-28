# Replay API - Domain Model

> Entity relationships and domain structure for the LeetGaming Replay API

## Domain Overview

The Replay API follows Clean Architecture (Hexagonal) with Domain-Driven Design principles.

```
pkg/domain/
├── auth/           # Authentication (MFA, email verification)
├── billing/        # Subscriptions, plans, billable entries
├── challenge/      # Challenge system (future)
├── cs/             # Counter-Strike specific domain
├── google/         # Google OAuth integration
├── iam/            # Identity & Access Management
├── legal/          # Legal agreements, terms
├── matchmaking/    # Lobbies, prize pools, sessions
├── media/          # Media assets
├── payment/        # Payment processing (Stripe)
├── replay/         # Replay files, matches, events
├── squad/          # Teams/squads
├── steam/          # Steam OAuth integration
├── tournament/     # Tournament management
└── wallet/         # Multi-asset wallet system
```

---

## Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          CORE DOMAIN MODEL                                   │
│                                                                              │
│  ┌───────────────┐         ┌───────────────┐         ┌──────────────────┐  │
│  │     User      │ 1────M  │ PlayerProfile │ 1────M  │      Squad       │  │
│  │               │         │               │         │                  │  │
│  │  • id         │         │  • id         │         │  • id            │  │
│  │  • steam_id   │         │  • user_id    │         │  • name          │  │
│  │  • google_id  │         │  • nickname   │         │  • tag           │  │
│  │  • email      │         │  • avatar_url │         │  • game_id       │  │
│  │  • created_at │         │  • game_id    │         │  • members[]     │  │
│  └───────┬───────┘         │  • stats      │         │  • visibility    │  │
│          │                 └───────────────┘         └────────┬─────────┘  │
│          │                                                     │            │
│          │ 1                                                   │ M          │
│          │                                                     │            │
│  ┌───────▼───────┐         ┌───────────────┐         ┌────────▼─────────┐  │
│  │  UserWallet   │ 1────M  │  Transaction  │         │  SquadMember     │  │
│  │               │         │               │         │                  │  │
│  │  • id         │         │  • id         │         │  • player_id     │  │
│  │  • user_id    │         │  • wallet_id  │         │  • role          │  │
│  │  • evm_addr   │         │  • type       │         │  • joined_at     │  │
│  │  • balances   │         │  • amount     │         │  • status        │  │
│  │  • is_locked  │         │  • currency   │         └──────────────────┘  │
│  └───────────────┘         │  • status     │                               │
│                            └───────────────┘                               │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                        MATCHMAKING DOMAIN                                    │
│                                                                              │
│  ┌───────────────────┐     ┌───────────────────┐     ┌──────────────────┐  │
│  │ MatchmakingLobby  │ 1─1 │    PrizePool      │ 1─1 │     Match        │  │
│  │                   │     │                   │     │                  │  │
│  │  • id             │     │  • id             │     │  • id            │  │
│  │  • creator_id     │     │  • match_id       │     │  • game_id       │  │
│  │  • game_id        │     │  • total_amount   │     │  • map           │  │
│  │  • region         │     │  • currency       │     │  • duration      │  │
│  │  • tier           │     │  • status         │     │  • players[]     │  │
│  │  • player_slots[] │     │  • distribution   │     │  • rounds[]      │  │
│  │  • status         │     │  • winners[]      │     │  • winner_team   │  │
│  │  • max_players    │     │  • escrow_end     │     │  • replay_id     │  │
│  └─────────┬─────────┘     └───────────────────┘     └────────┬─────────┘  │
│            │                                                   │            │
│            │ M                                                 │ 1          │
│            │                                                   │            │
│  ┌─────────▼─────────┐                               ┌────────▼─────────┐  │
│  │    PlayerSlot     │                               │   ReplayFile     │  │
│  │                   │                               │                  │  │
│  │  • slot_number    │                               │  • id            │  │
│  │  • player_id      │                               │  • match_id      │  │
│  │  • is_ready       │                               │  • file_uri      │  │
│  │  • joined_at      │                               │  • file_size     │  │
│  │  • mmr            │                               │  • status        │  │
│  └───────────────────┘                               │  • visibility    │  │
│                                                       └──────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                        TOURNAMENT DOMAIN                                     │
│                                                                              │
│  ┌───────────────────┐     ┌───────────────────┐     ┌──────────────────┐  │
│  │    Tournament     │ 1─M │ TournamentMatch   │ M─1 │TournamentPlayer  │  │
│  │                   │     │                   │     │                  │  │
│  │  • id             │     │  • match_id       │     │  • player_id     │  │
│  │  • name           │     │  • round          │     │  • display_name  │  │
│  │  • description    │     │  • bracket_pos    │     │  • registered_at │  │
│  │  • game_id        │     │  • player1_id     │     │  • seed          │  │
│  │  • format         │     │  • player2_id     │     │  • status        │  │
│  │  • entry_fee      │     │  • winner_id      │     └──────────────────┘  │
│  │  • prize_pool     │     │  • status         │                           │
│  │  • status         │     └───────────────────┘                           │
│  │  • participants[] │                                                      │
│  │  • winners[]      │     ┌───────────────────┐                           │
│  └───────────────────┘     │ TournamentWinner  │                           │
│                            │                   │                           │
│                            │  • player_id      │                           │
│                            │  • placement      │                           │
│                            │  • prize          │                           │
│                            │  • paid_at        │                           │
│                            └───────────────────┘                           │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          BILLING DOMAIN                                      │
│                                                                              │
│  ┌───────────────────┐     ┌───────────────────┐     ┌──────────────────┐  │
│  │       Plan        │ 1─M │   Subscription    │ M─1 │  BillableEntry   │  │
│  │                   │     │                   │     │                  │  │
│  │  • id             │     │  • id             │     │  • id            │  │
│  │  • name           │     │  • user_id        │     │  • subscription  │  │
│  │  • tier           │     │  • plan_id        │     │  • operation_id  │  │
│  │  • price          │     │  • status         │     │  • amount        │  │
│  │  • features[]     │     │  • period         │     │  • created_at    │  │
│  │  • limits[]       │     │  • expires_at     │     └──────────────────┘  │
│  └───────────────────┘     │  • history[]      │                           │
│                            └───────────────────┘                           │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          PAYMENT DOMAIN                                      │
│                                                                              │
│  ┌───────────────────┐     ┌───────────────────┐     ┌──────────────────┐  │
│  │  PaymentIntent    │ 1─1 │    Payment        │ 1─M │   LedgerEntry    │  │
│  │                   │     │                   │     │                  │  │
│  │  • id             │     │  • id             │     │  • id            │  │
│  │  • user_id        │     │  • intent_id      │     │  • transaction_id│  │
│  │  • amount         │     │  • stripe_id      │     │  • account_id    │  │
│  │  • currency       │     │  • status         │     │  • entry_type    │  │
│  │  • method         │     │  • confirmed_at   │     │  • amount        │  │
│  │  • status         │     │  • refunded_at    │     │  • balance_after │  │
│  │  • client_secret  │     └───────────────────┘     │  • is_reversed   │  │
│  └───────────────────┘                               └──────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Domain Details

### IAM (Identity & Access Management)

**Entities:**
- `User` - Core user account
- `PlayerProfile` - Game-specific player profile
- `Group` - Organization/group (for teams)
- `GroupAccount` - Group membership
- `Membership` - Role-based membership
- `RIDToken` - Resource ID token for auth

**Value Objects:**
- `Role` (Owner, Admin, Member)
- `MembershipStatus` (Active, Inactive, Pending)

---

### Wallet

**Entities:**
- `UserWallet` - Multi-currency wallet aggregate
- `LedgerEntry` - Immutable transaction record
- `IdempotentOperation` - Duplicate prevention

**Value Objects:**
- `Currency` (USD, USDC, USDT)
- `Amount` - Monetary amount (cents)
- `EVMAddress` - Ethereum address
- `EntryType` (Debit, Credit)
- `TransactionStatus` (Pending, Completed, Failed, RolledBack)

**Domain Services:**
- `LedgerService` - Double-entry accounting
- `TransactionCoordinator` - Saga pattern orchestration
- `ReconciliationService` - Balance verification

---

### Matchmaking

**Entities:**
- `MatchmakingLobby` - Lobby aggregate root
- `MatchmakingSession` - Queue session
- `MatchmakingPool` - Player pool by tier/region
- `PrizePool` - Prize money aggregate

**Value Objects:**
- `LobbyStatus` (open, ready_check, starting, started, cancelled)
- `PrizePoolStatus` (accumulating, locked, in_escrow, distributed, cancelled)
- `DistributionRule` (winner_takes_all, top_3, split_even)
- `PlayerSlot` - Lobby player slot

---

### Tournament

**Entities:**
- `Tournament` - Tournament aggregate root
- `TournamentMatch` - Match within tournament
- `TournamentPlayer` - Registered participant
- `TournamentWinner` - Prize recipient

**Value Objects:**
- `TournamentStatus` (draft, registration, ready, in_progress, completed, cancelled)
- `TournamentFormat` (single_elimination, double_elimination, round_robin, swiss)
- `MatchStatus` (scheduled, in_progress, completed, cancelled)

---

### Replay

**Entities:**
- `ReplayFile` - Uploaded replay file
- `Match` - Parsed match data
- `GameEvent` - In-game event (kill, death, etc.)
- `Round` - Match round

**Value Objects:**
- `ReplayStatus` (pending, processing, completed, failed)
- `Visibility` (public, private, unlisted)
- `GameID` (cs2, valorant, etc.)

---

### Billing

**Entities:**
- `Plan` - Subscription plan
- `Subscription` - User subscription
- `BillableEntry` - Usage record
- `Payable` - Pending payment

**Value Objects:**
- `PlanTier` (free, pro, team)
- `BillingPeriod` (monthly, yearly, lifetime)
- `SubscriptionStatus` (active, inactive, cancelled, expired)

---

### Payment

**Entities:**
- `PaymentIntent` - Payment request
- `Payment` - Completed payment

**Value Objects:**
- `PaymentStatus` (pending, processing, completed, failed, refunded)
- `PaymentMethod` (stripe, paypal, crypto)

---

## Aggregate Boundaries

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        AGGREGATE ROOTS                                   │
│                                                                          │
│  Each aggregate has a single entry point (root) that controls access    │
│  to all entities within the aggregate boundary.                          │
│                                                                          │
│  ┌─────────────────┐                                                    │
│  │   UserWallet    │  Controls: LedgerEntry[], IdempotentOperation     │
│  │   (Root)        │  Invariants:                                       │
│  │                 │    • Balance never negative                        │
│  │                 │    • Ledger always balanced                        │
│  │                 │    • Daily limits enforced                         │
│  └─────────────────┘                                                    │
│                                                                          │
│  ┌─────────────────┐                                                    │
│  │MatchmakingLobby │  Controls: PlayerSlot[]                           │
│  │   (Root)        │  Invariants:                                       │
│  │                 │    • Max players respected                         │
│  │                 │    • One player per slot                           │
│  │                 │    • Status transitions valid                      │
│  └─────────────────┘                                                    │
│                                                                          │
│  ┌─────────────────┐                                                    │
│  │   PrizePool     │  Controls: PlayerContributions, Winners[]         │
│  │   (Root)        │  Invariants:                                       │
│  │                 │    • Total = Platform + Sum(Players)              │
│  │                 │    • Distribution = Total                          │
│  │                 │    • Escrow period respected                       │
│  └─────────────────┘                                                    │
│                                                                          │
│  ┌─────────────────┐                                                    │
│  │   Tournament    │  Controls: TournamentMatch[], TournamentPlayer[]  │
│  │   (Root)        │  Invariants:                                       │
│  │                 │    • Min/max participants                          │
│  │                 │    • Registration window                           │
│  │                 │    • Bracket integrity                             │
│  └─────────────────┘                                                    │
│                                                                          │
│  ┌─────────────────┐                                                    │
│  │     Squad       │  Controls: SquadMember[]                          │
│  │   (Root)        │  Invariants:                                       │
│  │                 │    • One owner per squad                           │
│  │                 │    • Unique player per squad                       │
│  │                 │    • Max members limit                             │
│  └─────────────────┘                                                    │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Ports & Adapters

Each domain follows the Hexagonal Architecture pattern:

```
pkg/domain/{domain}/
├── entities/          # Domain entities
├── value-objects/     # Immutable value objects
├── ports/
│   ├── in/           # Input ports (use cases, queries, commands)
│   └── out/          # Output ports (repository interfaces)
├── services/         # Domain services
└── usecases/         # Use case implementations
```

**Input Ports** define what the domain can do:
- Commands (mutations): `CreateLobby`, `JoinLobby`, `SetPlayerReady`
- Queries (reads): `GetLobby`, `SearchLobbies`, `GetLobbyStats`

**Output Ports** define what the domain needs:
- Repositories: `LobbyWriter`, `LobbyReader`
- External services: `PaymentGateway`, `NotificationService`

---

**Last Updated**: November 2025
