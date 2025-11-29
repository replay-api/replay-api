package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	common "github.com/replay-api/replay-api/pkg/domain"
	db "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_vo "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// ==========================================
// SYSTEM CONSTANTS (Well-Known IDs)
// ==========================================

var (
	// System Tenant - LeetGaming PRO Platform
	SystemTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	// System Client - Web Application
	SystemClientID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	// System User - Platform Admin
	SystemUserID = uuid.MustParse("00000000-0000-0000-0000-000000000003")
)

// ==========================================
// SEED DATA DEFINITIONS
// ==========================================

// Professional esports team seed data
var seedSquads = []struct {
	Name        string
	Symbol      string
	GameID      common.GameIDKey
	Description string
	LogoURI     string
	SlugURI     string
	Region      string
	Members     []seedMember
}{
	{
		Name:        "n1ty",
		Symbol:      "N1TY",
		GameID:      common.CS2_GAME_ID,
		Description: "Our Featured Elite Counter-Strike players. The dream team sponsored by LeetGamingPRO. 2024 World Champions with an unmatched record.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "n1ty",
		Region:      "Global",
		Members: []seedMember{
			{Nickname: "Ace", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=11", MMR: 2450},
			{Nickname: "Sniper", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=12", MMR: 2500},
			{Nickname: "Clutch", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=13", MMR: 2380},
			{Nickname: "Mind", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=14", MMR: 2420},
			{Nickname: "Ghost", Role: "Lurker", Avatar: "https://i.pravatar.cc/150?img=15", MMR: 2400},
		},
	},
	{
		Name:        "et3rn1ty",
		Symbol:      "ET3R",
		GameID:      common.CS2_GAME_ID,
		Description: "Legends never die. EU powerhouse with a strategic approach. Winners of 5 major LAN events. Known for their incredible AWP plays and flawless executes.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "et3rn1ty",
		Region:      "EU",
		Members: []seedMember{
			{Nickname: "Eternal", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=21", MMR: 2520},
			{Nickname: "Forever", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=22", MMR: 2480},
			{Nickname: "Infinite", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=23", MMR: 2400},
			{Nickname: "Timeless", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=24", MMR: 2440},
			{Nickname: "Immortal", Role: "Rifler", Avatar: "https://i.pravatar.cc/150?img=25", MMR: 2420},
		},
	},
	{
		Name:        "1337gg",
		Symbol:      "1337",
		GameID:      common.CS2_GAME_ID,
		Description: "Elite gaming at its finest. The original esports legends. Multiple Major championships and unparalleled mechanical skill. Sponsored by LeetGamingPRO.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "1337gg",
		Region:      "Global",
		Members: []seedMember{
			{Nickname: "Leet", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=31", MMR: 2600},
			{Nickname: "Hax0r", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=32", MMR: 2580},
			{Nickname: "Root", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=33", MMR: 2520},
			{Nickname: "Sudo", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=34", MMR: 2560},
			{Nickname: "Pwn", Role: "Lurker", Avatar: "https://i.pravatar.cc/150?img=35", MMR: 2540},
		},
	},
	{
		Name:        "M14UZ",
		Symbol:      "M14Z",
		GameID:      common.VLRNT_GAME_ID,
		Description: "Lets have some fun and play some Valorant. Rising stars redefining the meta. Known for innovative agent compositions and clutch performances.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "m14uz",
		Region:      "Americas",
		Members: []seedMember{
			{Nickname: "M14U", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=41", MMR: 2350},
			{Nickname: "Zeus", Role: "Controller", Avatar: "https://i.pravatar.cc/150?img=42", MMR: 2300},
			{Nickname: "Apex", Role: "Sentinel", Avatar: "https://i.pravatar.cc/150?img=43", MMR: 2280},
			{Nickname: "Fury", Role: "Initiator", Avatar: "https://i.pravatar.cc/150?img=44", MMR: 2320},
			{Nickname: "Storm", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=45", MMR: 2340},
		},
	},
	{
		Name:        "Crimson Tide",
		Symbol:      "CRMS",
		GameID:      common.CS2_GAME_ID,
		Description: "LATAM's pride. Explosive gameplay and passionate fanbase. 2024 South American Championship winners with an undefeated streak.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "crimson-tide",
		Region:      "SA",
		Members: []seedMember{
			{Nickname: "Inferno", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=51", MMR: 2100},
			{Nickname: "Ember", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=53", MMR: 2150},
			{Nickname: "Blitz", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=54", MMR: 2050},
			{Nickname: "Thunder", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=55", MMR: 2080},
			{Nickname: "Riot", Role: "Rifler", Avatar: "https://i.pravatar.cc/150?img=56", MMR: 2060},
		},
	},
	{
		Name:        "Quantum Force",
		Symbol:      "QNTM",
		GameID:      common.VLRNT_GAME_ID,
		Description: "Precision meets creativity. This EMEA squad is known for unconventional strategies that catch opponents off-guard every time.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "quantum-force",
		Region:      "EMEA",
		Members: []seedMember{
			{Nickname: "Neutrino", Role: "Controller", Avatar: "https://i.pravatar.cc/150?img=61", MMR: 2200},
			{Nickname: "Quark", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=62", MMR: 2250},
			{Nickname: "Photon", Role: "Initiator", Avatar: "https://i.pravatar.cc/150?img=63", MMR: 2180},
			{Nickname: "Ion", Role: "Sentinel", Avatar: "https://i.pravatar.cc/150?img=64", MMR: 2150},
			{Nickname: "Proton", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=66", MMR: 2220},
		},
	},
	{
		Name:        "Digital Storm",
		Symbol:      "DSTM",
		GameID:      common.CS2_GAME_ID,
		Description: "CIS region veterans. Known for disciplined executes and flawless utility usage. 2x Major finalists.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "digital-storm",
		Region:      "CIS",
		Members: []seedMember{
			{Nickname: "Binary", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=1", MMR: 2280},
			{Nickname: "Cipher", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=2", MMR: 2240},
			{Nickname: "Matrix", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=3", MMR: 2200},
			{Nickname: "Vector", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=4", MMR: 2160},
			{Nickname: "Pixel", Role: "Lurker", Avatar: "https://i.pravatar.cc/150?img=5", MMR: 2180},
		},
	},
	{
		Name:        "Omega Protocol",
		Symbol:      "OMGA",
		GameID:      common.VLRNT_GAME_ID,
		Description: "Korean precision gaming. Famous for their mechanical excellence and innovative agent plays. VCT Pacific contenders.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "omega-protocol",
		Region:      "KR",
		Members: []seedMember{
			{Nickname: "Alpha", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=6", MMR: 2350},
			{Nickname: "Beta", Role: "Controller", Avatar: "https://i.pravatar.cc/150?img=7", MMR: 2300},
			{Nickname: "Gamma", Role: "Sentinel", Avatar: "https://i.pravatar.cc/150?img=8", MMR: 2280},
			{Nickname: "Delta", Role: "Initiator", Avatar: "https://i.pravatar.cc/150?img=9", MMR: 2320},
			{Nickname: "Epsilon", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=10", MMR: 2340},
		},
	},
}

// Individual player profile seed data (free agents / LFT players)
var seedPlayers = []struct {
	Nickname    string
	GameID      common.GameIDKey
	SlugURI     string
	Avatar      string
	Roles       []string
	Description string
	MMR         int
	Region      string
}{
	{
		Nickname:    "ProGamer2024",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "progamer2024",
		Avatar:      "https://i.pravatar.cc/150?img=68",
		Roles:       []string{"AWPer", "Entry Fragger"},
		Description: "Professional CS2 player with 5+ years competitive experience. Former ESL Pro League player. Looking for a top tier team.",
		MMR:         2100,
		Region:      "NA",
	},
	{
		Nickname:    "TacticalMind",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "tacticalmind",
		Avatar:      "https://i.pravatar.cc/150?img=69",
		Roles:       []string{"IGL", "Controller"},
		Description: "Strategic mastermind with exceptional game sense. Led 3 teams to regional finals. Available for tryouts.",
		MMR:         2050,
		Region:      "EU",
	},
	{
		Nickname:    "QuickScope",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "quickscope",
		Avatar:      "https://i.pravatar.cc/150?img=60",
		Roles:       []string{"AWPer"},
		Description: "Dedicated AWPer with 4000+ hours in competitive play. Known for clutch plays and aggressive positioning.",
		MMR:         1950,
		Region:      "NA",
	},
	{
		Nickname:    "SilentAssassin",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "silentassassin",
		Avatar:      "https://i.pravatar.cc/150?img=52",
		Roles:       []string{"Duelist", "Initiator"},
		Description: "Aggressive duelist main with exceptional first blood percentage. Multiple Radiant achievements.",
		MMR:         2200,
		Region:      "APAC",
	},
	{
		Nickname:    "StrategyKing",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "strategyking",
		Avatar:      "https://i.pravatar.cc/150?img=57",
		Roles:       []string{"IGL", "Support"},
		Description: "Veteran in-game leader with championship experience. Known for innovative strategies and team coordination.",
		MMR:         2150,
		Region:      "EU",
	},
	{
		Nickname:    "FlashPoint",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "flashpoint",
		Avatar:      "https://i.pravatar.cc/150?img=16",
		Roles:       []string{"Entry Fragger", "Rifler"},
		Description: "High-impact entry fragger with 85% flash success rate. Masters in peek timing and trade fragging.",
		MMR:         1800,
		Region:      "SA",
	},
	{
		Nickname:    "NightHawk",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "nighthawk",
		Avatar:      "https://i.pravatar.cc/150?img=17",
		Roles:       []string{"Sentinel", "Controller"},
		Description: "Flexible support player specializing in post-plant situations. Known for creative smoke lineups.",
		MMR:         1750,
		Region:      "NA",
	},
	{
		Nickname:    "ZeroGravity",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "zerogravity",
		Avatar:      "https://i.pravatar.cc/150?img=18",
		Roles:       []string{"Lurker"},
		Description: "Silent lurker with excellent map awareness. Specializes in flanking and information gathering.",
		MMR:         1900,
		Region:      "CIS",
	},
	{
		Nickname:    "ThunderStrike",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "thunderstrike",
		Avatar:      "https://i.pravatar.cc/150?img=19",
		Roles:       []string{"Duelist"},
		Description: "Mechanical prodigy with inhuman reaction times. Jett/Raze specialist with aggressive playstyle.",
		MMR:         2300,
		Region:      "KR",
	},
	{
		Nickname:    "IceQueen",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "icequeen",
		Avatar:      "https://i.pravatar.cc/150?img=20",
		Roles:       []string{"Support", "AWPer"},
		Description: "Versatile support player with secondary AWP skills. Excellent at reading opponents and calling rotations.",
		MMR:         2000,
		Region:      "EU",
	},
}

// Tournament seed data
var seedTournaments = []struct {
	Name            string
	Description     string
	GameID          common.GameIDKey
	GameMode        string
	Region          string
	Format          tournament_entities.TournamentFormat
	MaxParticipants int
	MinParticipants int
	EntryFee        int64
	Currency        wallet_vo.Currency
	Status          tournament_entities.TournamentStatus
	DaysFromNow     int // Negative = past, Positive = future
}{
	{
		Name:            "Weekly CS2 Showdown #47",
		Description:     "Weekly competitive CS2 tournament. Open to all skill levels. Great prizes for top performers!",
		GameID:          common.CS2_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "NA",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 32,
		MinParticipants: 8,
		EntryFee:        500,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusRegistration,
		DaysFromNow:     3,
	},
	{
		Name:            "Valorant Rising Stars Cup",
		Description:     "Tournament for emerging talent. Showcase your skills and get noticed by pro scouts!",
		GameID:          common.VLRNT_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "EU",
		Format:          tournament_entities.TournamentFormatDoubleElimination,
		MaxParticipants: 64,
		MinParticipants: 16,
		EntryFee:        1000,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusRegistration,
		DaysFromNow:     7,
	},
	{
		Name:            "Pro League Qualifier - Season 4",
		Description:     "Qualify for the Pro League! Top 4 teams advance to the main event with $50,000 prize pool.",
		GameID:          common.CS2_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "Global",
		Format:          tournament_entities.TournamentFormatSwiss,
		MaxParticipants: 128,
		MinParticipants: 32,
		EntryFee:        2500,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusDraft,
		DaysFromNow:     14,
	},
	{
		Name:            "Community Cup #23",
		Description:     "Free-to-enter community tournament. Just for fun with small prizes!",
		GameID:          common.VLRNT_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "APAC",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 16,
		MinParticipants: 4,
		EntryFee:        0,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusInProgress,
		DaysFromNow:     0,
	},
	{
		Name:            "Winter Championship 2024",
		Description:     "The biggest tournament of the season! Compete against the best for glory and prizes.",
		GameID:          common.CS2_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "Global",
		Format:          tournament_entities.TournamentFormatDoubleElimination,
		MaxParticipants: 256,
		MinParticipants: 64,
		EntryFee:        5000,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusCompleted,
		DaysFromNow:     -30,
	},
}

// Matchmaking pool seed data
var seedMatchmakingPools = []struct {
	GameID       string
	GameMode     string
	Region       string
	TotalPlayers int
	AvgWaitTime  int
	MatchesLast24h int
}{
	{GameID: "cs2", GameMode: "5v5 Competitive", Region: "NA", TotalPlayers: 156, AvgWaitTime: 45, MatchesLast24h: 892},
	{GameID: "cs2", GameMode: "5v5 Competitive", Region: "EU", TotalPlayers: 234, AvgWaitTime: 30, MatchesLast24h: 1247},
	{GameID: "cs2", GameMode: "5v5 Competitive", Region: "APAC", TotalPlayers: 89, AvgWaitTime: 75, MatchesLast24h: 456},
	{GameID: "cs2", GameMode: "5v5 Competitive", Region: "SA", TotalPlayers: 67, AvgWaitTime: 90, MatchesLast24h: 312},
	{GameID: "vlrnt", GameMode: "5v5 Competitive", Region: "NA", TotalPlayers: 198, AvgWaitTime: 35, MatchesLast24h: 1023},
	{GameID: "vlrnt", GameMode: "5v5 Competitive", Region: "EU", TotalPlayers: 267, AvgWaitTime: 28, MatchesLast24h: 1456},
	{GameID: "vlrnt", GameMode: "5v5 Competitive", Region: "APAC", TotalPlayers: 312, AvgWaitTime: 25, MatchesLast24h: 1678},
	{GameID: "vlrnt", GameMode: "5v5 Competitive", Region: "KR", TotalPlayers: 445, AvgWaitTime: 20, MatchesLast24h: 2134},
}

type seedMember struct {
	Nickname string
	Role     string
	Avatar   string
	MMR      int
}

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load environment variables
	if os.Getenv("DEV_ENV") == "true" || os.Getenv("MONGO_URI") == "" {
		if err := godotenv.Load(); err != nil {
			slog.Warn("No .env file found, using environment variables")
		}
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:dev-mongo-password-change-me@localhost:27017"
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = os.Getenv("MONGO_DB_NAME") // fallback for legacy
	}
	if dbName == "" {
		dbName = "replay_api"
	}

	slog.Info("Connecting to MongoDB", "uri", mongoURI[:30]+"...", "db", dbName)

	// Connect to MongoDB with UUID registry for proper serialization
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI).SetRegistry(db.MongoRegistry))
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		slog.Error("Failed to ping MongoDB", "error", err)
		os.Exit(1)
	}

	slog.Info("Connected to MongoDB successfully")

	// ==========================================
	// SEED ALL DATA
	// ==========================================

	// 1. System entities (Tenant, Client)
	slog.Info("Step 1/6: Seeding system entities...")
	if err := seedSystemEntities(ctx, client, dbName); err != nil {
		slog.Error("Failed to seed system entities", "error", err)
		os.Exit(1)
	}

	// 2. Squads
	slog.Info("Step 2/6: Seeding squads...")
	if err := seedSquadsData(ctx, client, dbName); err != nil {
		slog.Error("Failed to seed squads", "error", err)
		os.Exit(1)
	}

	// 3. Player profiles
	slog.Info("Step 3/6: Seeding player profiles...")
	if err := seedPlayerProfilesData(ctx, client, dbName); err != nil {
		slog.Error("Failed to seed player profiles", "error", err)
		os.Exit(1)
	}

	// 4. Tournaments
	slog.Info("Step 4/6: Seeding tournaments...")
	if err := seedTournamentsData(ctx, client, dbName); err != nil {
		slog.Error("Failed to seed tournaments", "error", err)
		os.Exit(1)
	}

	// 5. Matchmaking pools
	slog.Info("Step 5/6: Seeding matchmaking pools...")
	if err := seedMatchmakingPoolsData(ctx, client, dbName); err != nil {
		slog.Error("Failed to seed matchmaking pools", "error", err)
		os.Exit(1)
	}

	// 6. Wallets
	slog.Info("Step 6/6: Seeding wallets...")
	if err := seedWalletsData(ctx, client, dbName); err != nil {
		slog.Error("Failed to seed wallets", "error", err)
		os.Exit(1)
	}

	slog.Info("Seed completed successfully!")
	fmt.Println("")
	fmt.Println("===========================================")
	fmt.Println("  SEED SUMMARY")
	fmt.Println("===========================================")
	fmt.Printf("  Squads:       %d teams\n", len(seedSquads))
	fmt.Printf("  Players:      %d profiles\n", len(seedPlayers))
	fmt.Printf("  Tournaments:  %d events\n", len(seedTournaments))
	fmt.Printf("  MM Pools:     %d queues\n", len(seedMatchmakingPools))
	fmt.Println("===========================================")
}

// ==========================================
// SEED FUNCTIONS
// ==========================================

func seedSystemEntities(ctx context.Context, client *mongo.Client, dbName string) error {
	// Seed Tenant
	tenantsCol := client.Database(dbName).Collection("tenants")
	count, err := tenantsCol.CountDocuments(ctx, map[string]interface{}{"_id": SystemTenantID})
	if err != nil {
		return fmt.Errorf("failed to check tenant: %w", err)
	}

	if count == 0 {
		tenant := &iam_entities.Tenant{
			ID:          SystemTenantID,
			Name:        "LeetGaming PRO",
			Slug:        "leetgaming-pro",
			Description: "Professional esports platform for competitive gaming",
			VHash:       "system-vhash-do-not-use-in-production",
			VHashSalt:   "system-salt",
			Status:      iam_entities.TenantStatusActive,
			Domain:      "leetgaming.pro",
			AllowedURLs: []string{
				"http://localhost:3000",
				"http://localhost:3030",
				"https://leetgaming.pro",
				"https://app.leetgaming.pro",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if _, err := tenantsCol.InsertOne(ctx, tenant); err != nil {
			return fmt.Errorf("failed to insert tenant: %w", err)
		}
		slog.Info("Created system tenant", "id", SystemTenantID)
	} else {
		slog.Info("System tenant already exists", "id", SystemTenantID)
	}

	// Seed Client
	clientsCol := client.Database(dbName).Collection("clients")
	count, err = clientsCol.CountDocuments(ctx, map[string]interface{}{"_id": SystemClientID})
	if err != nil {
		return fmt.Errorf("failed to check client: %w", err)
	}

	if count == 0 {
		appClient := &iam_entities.Client{
			ID:          SystemClientID,
			TenantID:    SystemTenantID,
			Name:        "LeetGaming PRO Web",
			Slug:        "leetgaming-pro-web",
			Description: "Main web application for LeetGaming PRO platform",
			Type:        iam_entities.ClientTypeWeb,
			Status:      iam_entities.ClientStatusActive,
			ClientSecret: "dev-client-secret-do-not-use-in-production",
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://localhost:3030",
				"https://leetgaming.pro",
				"https://app.leetgaming.pro",
			},
			AllowedCallbacks: []string{
				"http://localhost:3000/api/auth/callback",
				"http://localhost:3030/api/auth/callback",
				"https://leetgaming.pro/api/auth/callback",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if _, err := clientsCol.InsertOne(ctx, appClient); err != nil {
			return fmt.Errorf("failed to insert client: %w", err)
		}
		slog.Info("Created system client", "id", SystemClientID)
	} else {
		slog.Info("System client already exists", "id", SystemClientID)
	}

	return nil
}

func seedSquadsData(ctx context.Context, client *mongo.Client, dbName string) error {
	collection := client.Database(dbName).Collection("squads")

	for _, seedData := range seedSquads {
		// Check if squad already exists by slug
		count, err := collection.CountDocuments(ctx, map[string]interface{}{"slug_uri": seedData.SlugURI})
		if err != nil {
			return fmt.Errorf("failed to check existing squad: %w", err)
		}

		if count > 0 {
			slog.Info("Squad already exists, skipping", "slug", seedData.SlugURI)
			continue
		}

		// Create membership for each member
		membership := make([]squad_vo.SquadMembership, len(seedData.Members))
		for i, member := range seedData.Members {
			playerProfileID := uuid.New()
			membership[i] = squad_vo.SquadMembership{
				UserID:          uuid.New(),
				PlayerProfileID: playerProfileID,
				Type:            squad_vo.SquadMembershipTypeMember,
				Roles:           []string{member.Role},
				Status:          map[time.Time]squad_vo.SquadMembershipStatus{time.Now(): squad_vo.SquadMembershipStatusActive},
				History:         map[time.Time]squad_vo.SquadMembershipType{time.Now(): squad_vo.SquadMembershipTypeMember},
			}
		}

		// Create squad
		squad := &squad_entities.Squad{
			BaseEntity: common.BaseEntity{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				VisibilityLevel: common.ClientApplicationAudienceIDKey,
				VisibilityType:  common.PublicVisibilityTypeKey,
				ResourceOwner: common.ResourceOwner{
					TenantID: SystemTenantID,
					ClientID: SystemClientID,
				},
			},
			GameID:      seedData.GameID,
			Name:        seedData.Name,
			Symbol:      seedData.Symbol,
			Description: seedData.Description,
			LogoURI:     seedData.LogoURI,
			SlugURI:     seedData.SlugURI,
			Membership:  membership,
		}

		_, err = collection.InsertOne(ctx, squad)
		if err != nil {
			return fmt.Errorf("failed to insert squad %s: %w", seedData.Name, err)
		}

		slog.Info("Created squad", "name", seedData.Name, "game", seedData.GameID, "members", len(seedData.Members))
	}

	return nil
}

func seedPlayerProfilesData(ctx context.Context, client *mongo.Client, dbName string) error {
	collection := client.Database(dbName).Collection("player_profiles")

	for _, seedData := range seedPlayers {
		// Check if player already exists by slug
		count, err := collection.CountDocuments(ctx, map[string]interface{}{"slug_uri": seedData.SlugURI})
		if err != nil {
			return fmt.Errorf("failed to check existing player: %w", err)
		}

		if count > 0 {
			slog.Info("Player profile already exists, skipping", "slug", seedData.SlugURI)
			continue
		}

		// Create player profile
		profile := &squad_entities.PlayerProfile{
			BaseEntity: common.BaseEntity{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				VisibilityLevel: common.ClientApplicationAudienceIDKey,
				VisibilityType:  common.PublicVisibilityTypeKey,
				ResourceOwner: common.ResourceOwner{
					TenantID: SystemTenantID,
					ClientID: SystemClientID,
				},
			},
			GameID:      seedData.GameID,
			Nickname:    seedData.Nickname,
			SlugURI:     seedData.SlugURI,
			Avatar:      seedData.Avatar,
			Roles:       seedData.Roles,
			Description: seedData.Description,
		}

		_, err = collection.InsertOne(ctx, profile)
		if err != nil {
			return fmt.Errorf("failed to insert player profile %s: %w", seedData.Nickname, err)
		}

		slog.Info("Created player profile", "nickname", seedData.Nickname, "game", seedData.GameID, "mmr", seedData.MMR)
	}

	return nil
}

func seedTournamentsData(ctx context.Context, client *mongo.Client, dbName string) error {
	collection := client.Database(dbName).Collection("tournaments")

	for _, seedData := range seedTournaments {
		// Check if tournament already exists by name
		count, err := collection.CountDocuments(ctx, map[string]interface{}{"name": seedData.Name})
		if err != nil {
			return fmt.Errorf("failed to check existing tournament: %w", err)
		}

		if count > 0 {
			slog.Info("Tournament already exists, skipping", "name", seedData.Name)
			continue
		}

		// Calculate dates
		now := time.Now()
		startTime := now.AddDate(0, 0, seedData.DaysFromNow)
		registrationOpen := startTime.AddDate(0, 0, -7) // Opens 7 days before
		registrationClose := startTime.AddDate(0, 0, -1) // Closes 1 day before

		// For completed tournaments, set end time
		var endTime *time.Time
		if seedData.Status == tournament_entities.TournamentStatusCompleted {
			end := startTime.AddDate(0, 0, 2)
			endTime = &end
		}

		tournament := &tournament_entities.Tournament{
			BaseEntity: common.BaseEntity{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				VisibilityLevel: common.ClientApplicationAudienceIDKey,
				VisibilityType:  common.PublicVisibilityTypeKey,
				ResourceOwner: common.ResourceOwner{
					TenantID: SystemTenantID,
					ClientID: SystemClientID,
				},
			},
			Name:              seedData.Name,
			Description:       seedData.Description,
			GameID:            seedData.GameID,
			GameMode:          seedData.GameMode,
			Region:            seedData.Region,
			Format:            seedData.Format,
			MaxParticipants:   seedData.MaxParticipants,
			MinParticipants:   seedData.MinParticipants,
			EntryFee:          wallet_vo.NewAmountFromCents(seedData.EntryFee),
			Currency:          seedData.Currency,
			PrizePool:         wallet_vo.NewAmountFromCents(seedData.EntryFee * int64(seedData.MinParticipants)),
			Status:            seedData.Status,
			StartTime:         startTime,
			EndTime:           endTime,
			RegistrationOpen:  registrationOpen,
			RegistrationClose: registrationClose,
			Participants:      []tournament_entities.TournamentPlayer{},
			Matches:           []tournament_entities.TournamentMatch{},
			Winners:           []tournament_entities.TournamentWinner{},
			Rules: tournament_entities.TournamentRules{
				BestOf:              3,
				MapPool:             []string{"de_mirage", "de_inferno", "de_nuke", "de_ancient", "de_anubis"},
				BanPickEnabled:      true,
				CheckInRequired:     true,
				CheckInWindowMins:   30,
				MatchTimeoutMins:    90,
				DisconnectGraceMins: 5,
			},
			OrganizerID: SystemUserID,
		}

		_, err = collection.InsertOne(ctx, tournament)
		if err != nil {
			return fmt.Errorf("failed to insert tournament %s: %w", seedData.Name, err)
		}

		slog.Info("Created tournament", "name", seedData.Name, "game", seedData.GameID, "status", seedData.Status)
	}

	return nil
}

func seedMatchmakingPoolsData(ctx context.Context, client *mongo.Client, dbName string) error {
	collection := client.Database(dbName).Collection("matchmaking_pools")

	for _, seedData := range seedMatchmakingPools {
		// Check if pool already exists
		filter := map[string]interface{}{
			"game_id":   seedData.GameID,
			"game_mode": seedData.GameMode,
			"region":    seedData.Region,
		}
		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			return fmt.Errorf("failed to check existing pool: %w", err)
		}

		if count > 0 {
			slog.Info("Matchmaking pool already exists, skipping", "game", seedData.GameID, "region", seedData.Region)
			continue
		}

		pool := &matchmaking_entities.MatchmakingPool{
			ID:             uuid.New(),
			GameID:         seedData.GameID,
			GameMode:       seedData.GameMode,
			Region:         seedData.Region,
			ActiveSessions: []uuid.UUID{},
			PoolStats: matchmaking_entities.PoolStatistics{
				TotalPlayers:      seedData.TotalPlayers,
				AverageWaitTime:   seedData.AvgWaitTime,
				PlayersByTier: map[matchmaking_entities.MatchmakingTier]int{
					matchmaking_entities.TierFree:    int(float64(seedData.TotalPlayers) * 0.5),
					matchmaking_entities.TierPremium: int(float64(seedData.TotalPlayers) * 0.3),
					matchmaking_entities.TierPro:     int(float64(seedData.TotalPlayers) * 0.15),
					matchmaking_entities.TierElite:   int(float64(seedData.TotalPlayers) * 0.05),
				},
				PlayersBySkill: map[string]int{
					"0-1000":    int(float64(seedData.TotalPlayers) * 0.10),
					"1000-1500": int(float64(seedData.TotalPlayers) * 0.25),
					"1500-2000": int(float64(seedData.TotalPlayers) * 0.40),
					"2000-2500": int(float64(seedData.TotalPlayers) * 0.20),
					"2500+":     int(float64(seedData.TotalPlayers) * 0.05),
				},
				EstimatedMatchTime: seedData.AvgWaitTime + 10,
				MatchesLast24h:     seedData.MatchesLast24h,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = collection.InsertOne(ctx, pool)
		if err != nil {
			return fmt.Errorf("failed to insert matchmaking pool: %w", err)
		}

		slog.Info("Created matchmaking pool", "game", seedData.GameID, "region", seedData.Region, "players", seedData.TotalPlayers)
	}

	return nil
}

func seedWalletsData(ctx context.Context, client *mongo.Client, dbName string) error {
	collection := client.Database(dbName).Collection("wallets")

	// Create sample wallets for demo purposes
	sampleWallets := []struct {
		EVMAddress string
		USDBalance int64
	}{
		{EVMAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f8B2e1", USDBalance: 15000},
		{EVMAddress: "0x8ba1f109551bD432803012645Ac136ddd64DBA72", USDBalance: 25000},
		{EVMAddress: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", USDBalance: 50000},
	}

	for _, w := range sampleWallets {
		count, err := collection.CountDocuments(ctx, map[string]interface{}{"evm_address.address": w.EVMAddress})
		if err != nil {
			return fmt.Errorf("failed to check existing wallet: %w", err)
		}

		if count > 0 {
			slog.Info("Wallet already exists, skipping", "address", w.EVMAddress[:10]+"...")
			continue
		}

		evmAddr, err := wallet_vo.NewEVMAddress(w.EVMAddress)
		if err != nil {
			slog.Warn("Invalid EVM address, skipping", "address", w.EVMAddress[:10]+"...", "error", err)
			continue
		}

		wallet := &wallet_entities.UserWallet{
			BaseEntity: common.BaseEntity{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				VisibilityLevel: common.UserAudienceIDKey,
				VisibilityType:  common.PrivateVisibilityTypeKey,
				ResourceOwner: common.ResourceOwner{
					TenantID: SystemTenantID,
					ClientID: SystemClientID,
				},
			},
			EVMAddress: evmAddr,
			Balances: map[wallet_vo.Currency]wallet_vo.Amount{
				wallet_vo.CurrencyUSD:  wallet_vo.NewAmountFromCents(w.USDBalance),
				wallet_vo.CurrencyUSDC: wallet_vo.NewAmountFromCents(0),
				wallet_vo.CurrencyUSDT: wallet_vo.NewAmountFromCents(0),
			},
			PendingTransactions: []uuid.UUID{},
			TotalDeposited:      wallet_vo.NewAmountFromCents(w.USDBalance),
			TotalWithdrawn:      wallet_vo.NewAmountFromCents(0),
			TotalPrizesWon:      wallet_vo.NewAmountFromCents(0),
			DailyPrizeWinnings:  wallet_vo.NewAmountFromCents(0),
			LastPrizeWinDate:    time.Now(),
			IsLocked:            false,
		}

		_, err = collection.InsertOne(ctx, wallet)
		if err != nil {
			return fmt.Errorf("failed to insert wallet: %w", err)
		}

		slog.Info("Created wallet", "address", w.EVMAddress[:10]+"...", "balance", w.USDBalance)
	}

	return nil
}
