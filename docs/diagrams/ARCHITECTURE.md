# System Architecture Diagrams

> Visual representations of the ESAP backend architecture

---

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Domain Layer](#domain-layer)
3. [Data Flow](#data-flow)
4. [Infrastructure](#infrastructure)

---

## High-Level Architecture

### System Overview

```mermaid
graph TB
    subgraph "Frontend"
        WEB[Next.js Web App]
        MOBILE[Mobile Apps]
    end

    subgraph "API Gateway"
        GW[Nginx / K8s Ingress]
    end

    subgraph "Backend Services"
        API[Replay API<br/>Go]
    end

    subgraph "Data Layer"
        MONGO[(MongoDB)]
        REDIS[(Redis)]
        S3[(MinIO/S3)]
    end

    subgraph "External Services"
        STRIPE[Stripe]
        STEAM[Steam API]
        GOOGLE[Google OAuth]
    end

    WEB --> GW
    MOBILE --> GW
    GW --> API
    API --> MONGO
    API --> REDIS
    API --> S3
    API --> STRIPE
    API --> STEAM
    API --> GOOGLE
```

### Request Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant GW as API Gateway
    participant H as Handler
    participant UC as Use Case
    participant R as Repository
    participant DB as MongoDB

    C->>GW: HTTP Request
    GW->>H: Route to Handler
    H->>H: Validate Input
    H->>UC: Execute Use Case
    UC->>R: Repository Call
    R->>DB: Database Query
    DB-->>R: Result
    R-->>UC: Domain Entity
    UC-->>H: Response DTO
    H-->>C: HTTP Response
```

---

## Domain Layer

### Core Domains

```mermaid
graph TB
    subgraph "Identity & Access"
        IAM[IAM Domain]
        AUTH[Auth Domain]
    end

    subgraph "Gaming"
        MM[Matchmaking]
        TOURN[Tournament]
        REPLAY[Replay]
        SQUAD[Squad]
    end

    subgraph "Financial"
        WALLET[Wallet]
        PAYMENT[Payment]
        BILLING[Billing]
    end

    subgraph "Supporting"
        MEDIA[Media]
        LEGAL[Legal]
    end

    IAM --> MM
    IAM --> TOURN
    IAM --> SQUAD
    IAM --> WALLET

    WALLET --> MM
    WALLET --> TOURN
    WALLET --> PAYMENT

    MM --> REPLAY
    TOURN --> REPLAY
```

### Domain Structure

```mermaid
graph LR
    subgraph "Each Domain"
        E[entities/] --> P[ports/]
        P --> A[adapters/]
        P --> APP[application/]
        APP --> CMD[commands/]
        APP --> QRY[queries/]
    end
```

### Matchmaking Entities

```mermaid
erDiagram
    MatchmakingLobby ||--o{ LobbyPlayer : contains
    MatchmakingLobby ||--|| PrizePool : has
    MatchmakingLobby }|--|| LobbySettings : configured_by
    PrizePool ||--o{ PrizeDistribution : distributes
    PrizePool ||--|| PrizePoolEscrow : escrows_in

    MatchmakingLobby {
        UUID id PK
        string status
        string game_id
        string mode
        string region
        datetime created_at
    }

    LobbyPlayer {
        UUID id PK
        UUID lobby_id FK
        UUID player_id FK
        string team
        boolean ready
        datetime joined_at
    }

    PrizePool {
        UUID id PK
        UUID lobby_id FK
        decimal total_amount
        string currency
        string distribution_type
    }

    PrizeDistribution {
        int placement
        decimal percentage
        decimal amount
    }

    PrizePoolEscrow {
        UUID id PK
        UUID prize_pool_id FK
        decimal amount
        string status
    }

    LobbySettings {
        int min_players
        int max_players
        decimal entry_fee
        int ready_timeout_seconds
    }
```

### Wallet Entities

```mermaid
erDiagram
    UserWallet ||--o{ WalletEntry : contains
    UserWallet ||--o{ AssetBalance : tracks
    WalletEntry }|--|| Transaction : part_of
    
    UserWallet {
        UUID id PK
        UUID user_id FK
        string status
        datetime created_at
    }

    AssetBalance {
        UUID wallet_id FK
        string asset_type
        decimal available
        decimal locked
        decimal pending
    }

    WalletEntry {
        UUID id PK
        UUID wallet_id FK
        string entry_type
        decimal amount
        string asset_type
        UUID transaction_id FK
        datetime created_at
    }

    Transaction {
        UUID id PK
        string type
        string status
        decimal amount
        UUID source_wallet FK
        UUID destination_wallet FK
    }
```

### Tournament Entities

```mermaid
erDiagram
    Tournament ||--o{ TournamentParticipant : has
    Tournament ||--o{ TournamentMatch : contains
    Tournament ||--|| TournamentBracket : uses
    TournamentMatch ||--o{ MatchResult : produces

    Tournament {
        UUID id PK
        string name
        string game_id
        string format
        string status
        datetime start_date
        datetime end_date
    }

    TournamentParticipant {
        UUID id PK
        UUID tournament_id FK
        UUID player_id FK
        UUID squad_id FK
        int seed
        string status
    }

    TournamentMatch {
        UUID id PK
        UUID tournament_id FK
        int round
        int match_number
        UUID participant_a FK
        UUID participant_b FK
        UUID winner FK
    }

    TournamentBracket {
        UUID id PK
        string format
        int rounds
        json structure
    }

    MatchResult {
        UUID match_id FK
        UUID participant_id FK
        int score
        string stats
    }
```

---

## Data Flow

### Authentication Flow

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant API as Replay API
    participant Steam as Steam OAuth
    participant DB as MongoDB

    U->>FE: Click "Login with Steam"
    FE->>API: GET /onboarding/steam
    API->>Steam: Redirect to Steam Login
    Steam-->>U: Steam Login Page
    U->>Steam: Enter Credentials
    Steam->>API: Callback with token
    API->>Steam: Validate token
    Steam-->>API: User profile
    API->>DB: Create/Update user
    API-->>FE: JWT + Session
    FE-->>U: Logged In
```

### Lobby Creation Flow

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant API as Replay API
    participant WS as WebSocket
    participant DB as MongoDB

    U->>FE: Create Lobby
    FE->>API: POST /api/lobbies
    API->>API: Validate request
    API->>DB: Create Lobby
    API->>DB: Create PrizePool
    API-->>FE: Lobby created
    FE->>WS: Subscribe to lobby
    WS-->>FE: Connection established
    
    loop Player joins
        FE->>API: POST /api/lobbies/{id}/join
        API->>DB: Add player
        API->>WS: Broadcast update
        WS-->>FE: Player joined event
    end
```

### Payment Flow

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant API as Replay API
    participant Stripe as Stripe
    participant DB as MongoDB

    U->>FE: Add funds $50
    FE->>API: POST /payments/checkout
    API->>Stripe: Create PaymentIntent
    Stripe-->>API: client_secret
    API-->>FE: client_secret
    FE->>Stripe: Confirm payment
    Stripe-->>FE: Payment success
    
    Note over Stripe,API: Webhook (async)
    Stripe->>API: POST /webhooks/stripe
    API->>API: Verify signature
    API->>DB: Create wallet entry
    API->>DB: Update balance
```

### Prize Distribution Flow

```mermaid
sequenceDiagram
    participant MM as Matchmaking Service
    participant WALLET as Wallet Service
    participant ESCROW as Escrow Account
    participant DB as MongoDB

    Note over MM: Match completed
    MM->>WALLET: Request distribution
    
    loop For each winner
        WALLET->>ESCROW: Release funds
        ESCROW->>DB: Create debit entry
        WALLET->>DB: Create credit entry
        WALLET->>DB: Update winner balance
    end
    
    WALLET->>DB: Mark escrow settled
    WALLET-->>MM: Distribution complete
```

---

## Infrastructure

### Kubernetes Deployment

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "Ingress"
            ING[Nginx Ingress]
        end

        subgraph "Application"
            BLUE[Blue Deployment<br/>replicas: 3]
            GREEN[Green Deployment<br/>replicas: 0]
            SVC[Service]
        end

        subgraph "Data"
            MONGO[MongoDB<br/>StatefulSet]
            REDIS[Redis<br/>Deployment]
        end

        subgraph "Monitoring"
            PROM[Prometheus]
            GRAF[Grafana]
        end
    end

    ING --> SVC
    SVC --> BLUE
    SVC -.-> GREEN
    BLUE --> MONGO
    BLUE --> REDIS
    GREEN --> MONGO
    GREEN --> REDIS
    PROM --> BLUE
    PROM --> GREEN
    GRAF --> PROM
```

### Local Development

```mermaid
graph LR
    subgraph "Local Machine"
        DEV[Go Application]
        DC[Docker Compose]
    end

    subgraph "Docker Containers"
        MONGO[(MongoDB)]
        REDIS[(Redis)]
        MINIO[(MinIO)]
    end

    DEV --> MONGO
    DEV --> REDIS
    DEV --> MINIO
    DC --> MONGO
    DC --> REDIS
    DC --> MINIO
```

### CI/CD Pipeline

```mermaid
graph LR
    subgraph "Development"
        CODE[Code Push]
        PR[Pull Request]
    end

    subgraph "CI"
        LINT[Lint]
        TEST[Test]
        BUILD[Build Image]
    end

    subgraph "CD"
        PUSH[Push to Registry]
        DEPLOY[Deploy to K8s]
        VERIFY[Health Check]
    end

    CODE --> PR
    PR --> LINT
    LINT --> TEST
    TEST --> BUILD
    BUILD --> PUSH
    PUSH --> DEPLOY
    DEPLOY --> VERIFY
```

---

## Component Interaction

### Handler to Repository

```mermaid
flowchart LR
    subgraph "Presentation"
        H[Handler]
    end

    subgraph "Application"
        CMD[Command/Query]
        UC[Use Case]
    end

    subgraph "Domain"
        E[Entity]
        P[Port Interface]
    end

    subgraph "Infrastructure"
        A[Adapter]
        DB[(Database)]
    end

    H --> CMD
    CMD --> UC
    UC --> P
    P --> A
    A --> DB
    A --> E
    E --> UC
    UC --> H
```

### Event-Driven Communication

```mermaid
graph TB
    subgraph "Publishers"
        MM[Matchmaking]
        PAY[Payment]
        TOURN[Tournament]
    end

    subgraph "Event Bus"
        KAFKA[Kafka Topics]
    end

    subgraph "Subscribers"
        WALLET[Wallet]
        NOTIF[Notifications]
        ANALYTICS[Analytics]
    end

    MM --> KAFKA
    PAY --> KAFKA
    TOURN --> KAFKA
    KAFKA --> WALLET
    KAFKA --> NOTIF
    KAFKA --> ANALYTICS
```

---

**Last Updated**: November 2025
