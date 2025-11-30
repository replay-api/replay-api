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

// Professional esports team seed data - Fictional teams inspired by competitive scene
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
	// CS2 Teams
	{
		Name:        "Natus Victoria",
		Symbol:      "NVIC",
		GameID:      common.CS2_GAME_ID,
		Description: "Born Victorious. Eastern European esports powerhouse and one of the most decorated CS organizations in history. Multiple Major champions with legendary players.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "natus-victoria",
		Region:      "CIS",
		Members: []seedMember{
			{Nickname: "s2mpler", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=11", MMR: 2850},
			{Nickname: "b2t", Role: "Rifler", Avatar: "https://i.pravatar.cc/150?img=12", MMR: 2780},
			{Nickname: "jK", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=13", MMR: 2720},
			{Nickname: "Alexic", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=14", MMR: 2680},
			{Nickname: "aM", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=15", MMR: 2650},
		},
	},
	{
		Name:        "N3 Gaming",
		Symbol:      "N3",
		GameID:      common.CS2_GAME_ID,
		Description: "European esports giants. Known for their entertainment value and competitive excellence across multiple titles. Major Champions with world-class talent.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "n3-gaming",
		Region:      "EU",
		Members: []seedMember{
			{Nickname: "NiKa", Role: "Rifler", Avatar: "https://i.pravatar.cc/150?img=21", MMR: 2820},
			{Nickname: "m0NYE", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=22", MMR: 2800},
			{Nickname: "tr4cker", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=23", MMR: 2750},
			{Nickname: "nexx", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=24", MMR: 2700},
			{Nickname: "Snix", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=25", MMR: 2680},
		},
	},
	{
		Name:        "Blaze Syndicate",
		Symbol:      "BLZE",
		GameID:      common.CS2_GAME_ID,
		Description: "Lifestyle and esports brand with global recognition. Multiple Major champions known for their star-studded international roster.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "blaze-syndicate",
		Region:      "Global",
		Members: []seedMember{
			{Nickname: "ropx", Role: "Lurker", Avatar: "https://i.pravatar.cc/150?img=31", MMR: 2790},
			{Nickname: "brokz", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=32", MMR: 2760},
			{Nickname: "storm", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=33", MMR: 2720},
			{Nickname: "karragan", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=34", MMR: 2650},
			{Nickname: "frosted", Role: "Rifler", Avatar: "https://i.pravatar.cc/150?img=35", MMR: 2730},
		},
	},
	{
		Name:        "Team Velocity",
		Symbol:      "VLCT",
		GameID:      common.CS2_GAME_ID,
		Description: "French esports organization with world-class CS roster. Premier World Final champions and consistent top-tier performers.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "team-velocity",
		Region:      "EU",
		Members: []seedMember{
			{Nickname: "ZivOx", Role: "AWPer", Avatar: "https://i.pravatar.cc/150?img=41", MMR: 2870},
			{Nickname: "apXE", Role: "IGL", Avatar: "https://i.pravatar.cc/150?img=42", MMR: 2680},
			{Nickname: "Sphinx", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=43", MMR: 2740},
			{Nickname: "blazeZ", Role: "Rifler", Avatar: "https://i.pravatar.cc/150?img=44", MMR: 2760},
			{Nickname: "mizzi", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=45", MMR: 2700},
		},
	},
	{
		Name:        "Team Fluid",
		Symbol:      "FLD",
		GameID:      common.CS2_GAME_ID,
		Description: "North American esports institution. Premier organization competing at the highest level across multiple titles with a dedicated fanbase.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "team-fluid",
		Region:      "NA",
		Members: []seedMember{
			{Nickname: "NAG", Role: "Rifler", Avatar: "https://i.pravatar.cc/150?img=51", MMR: 2720},
			{Nickname: "YEKINAR", Role: "Entry Fragger", Avatar: "https://i.pravatar.cc/150?img=52", MMR: 2750},
			{Nickname: "jkz", Role: "Support", Avatar: "https://i.pravatar.cc/150?img=53", MMR: 2680},
			{Nickname: "cadianN", Role: "AWPer/IGL", Avatar: "https://i.pravatar.cc/150?img=54", MMR: 2700},
			{Nickname: "Twizztz", Role: "Rifler", Avatar: "https://i.pravatar.cc/150?img=55", MMR: 2740},
		},
	},
	// Valorant Teams
	{
		Name:        "Wardens",
		Symbol:      "WRD",
		GameID:      common.VLRNT_GAME_ID,
		Description: "Iconic North American organization. International Masters champions and home to legendary tactical shooter players. Known for creating superstars.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "wardens",
		Region:      "Americas",
		Members: []seedMember{
			{Nickname: "TenX", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=61", MMR: 2800},
			{Nickname: "zekker", Role: "Flex", Avatar: "https://i.pravatar.cc/150?img=62", MMR: 2750},
			{Nickname: "Sazy", Role: "Initiator", Avatar: "https://i.pravatar.cc/150?img=63", MMR: 2720},
			{Nickname: "johnqr", Role: "IGL/Controller", Avatar: "https://i.pravatar.cc/150?img=64", MMR: 2700},
			{Nickname: "pANzada", Role: "Controller", Avatar: "https://i.pravatar.cc/150?img=65", MMR: 2680},
		},
	},
	{
		Name:        "Cardboard Tiger",
		Symbol:      "CTGR",
		GameID:      common.VLRNT_GAME_ID,
		Description: "Singapore-based APAC powerhouse. Famous for their aggressive W-key playstyle that revolutionized competitive gaming. International Masters champions.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "cardboard-tiger",
		Region:      "APAC",
		Members: []seedMember{
			{Nickname: "f0rgiveN", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=1", MMR: 2780},
			{Nickname: "Jingg", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=2", MMR: 2760},
			{Nickname: "d5v51", Role: "Initiator", Avatar: "https://i.pravatar.cc/150?img=3", MMR: 2720},
			{Nickname: "Monket", Role: "Controller", Avatar: "https://i.pravatar.cc/150?img=4", MMR: 2680},
			{Nickname: "mindbreak", Role: "Sentinel/IGL", Avatar: "https://i.pravatar.cc/150?img=5", MMR: 2700},
		},
	},
	{
		Name:        "Fanatic Gaming",
		Symbol:      "FNTC",
		GameID:      common.VLRNT_GAME_ID,
		Description: "Legendary EMEA organization with a rich esports history. International champions and consistent top performers in global competition.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "fanatic-gaming",
		Region:      "EMEA",
		Members: []seedMember{
			{Nickname: "Betajer", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=6", MMR: 2820},
			{Nickname: "Darke", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=7", MMR: 2800},
			{Nickname: "Chronicler", Role: "Sentinel", Avatar: "https://i.pravatar.cc/150?img=8", MMR: 2750},
			{Nickname: "Leon", Role: "Initiator", Avatar: "https://i.pravatar.cc/150?img=9", MMR: 2720},
			{Nickname: "Toaster", Role: "IGL/Controller", Avatar: "https://i.pravatar.cc/150?img=10", MMR: 2680},
		},
	},
	{
		Name:        "DRZ Gaming",
		Symbol:      "DRZ",
		GameID:      common.VLRNT_GAME_ID,
		Description: "Korean esports dynasty. World Champions and perennial favorites. Known for precision gameplay and clutch performances.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "drz-gaming",
		Region:      "KR",
		Members: []seedMember{
			{Nickname: "BuZy", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=16", MMR: 2780},
			{Nickname: "MaKa", Role: "Controller", Avatar: "https://i.pravatar.cc/150?img=17", MMR: 2750},
			{Nickname: "Rc", Role: "Flex", Avatar: "https://i.pravatar.cc/150?img=18", MMR: 2720},
			{Nickname: "staxx", Role: "IGL/Initiator", Avatar: "https://i.pravatar.cc/150?img=19", MMR: 2700},
			{Nickname: "Foxy8", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=20", MMR: 2760},
		},
	},
	{
		Name:        "ROAR Esports",
		Symbol:      "ROAR",
		GameID:      common.VLRNT_GAME_ID,
		Description: "Brazilian powerhouse and World Champions runner-up. Massive fanbase and known for their passionate gameplay and team synergy.",
		LogoURI:     "https://avatars.githubusercontent.com/u/168373383",
		SlugURI:     "roar-esports",
		Region:      "Americas",
		Members: []seedMember{
			{Nickname: "aspax", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=26", MMR: 2850},
			{Nickname: "Lass", Role: "Initiator", Avatar: "https://i.pravatar.cc/150?img=27", MMR: 2780},
			{Nickname: "tuyx", Role: "Controller", Avatar: "https://i.pravatar.cc/150?img=28", MMR: 2720},
			{Nickname: "cauanzer", Role: "Duelist", Avatar: "https://i.pravatar.cc/150?img=29", MMR: 2740},
			{Nickname: "qcx", Role: "Sentinel", Avatar: "https://i.pravatar.cc/150?img=30", MMR: 2700},
		},
	},
}

// Individual player profile seed data (free agents / LFT players) - Fictional players
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
		Nickname:    "devize",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "devize",
		Avatar:      "https://i.pravatar.cc/150?img=68",
		Roles:       []string{"AWPer", "Rifler"},
		Description: "Multiple Major champion and one of the greatest CS players of all time. 4x Major winner seeking new challenges. Precision AWPing and clutch performances.",
		MMR:         2750,
		Region:      "EU",
	},
	{
		Nickname:    "nitr1",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "nitr1",
		Avatar:      "https://i.pravatar.cc/150?img=69",
		Roles:       []string{"IGL", "Controller"},
		Description: "Former CS Major champion turned tactical shooter pro. Exceptional leadership and game sense. Grand Slam winner seeking new opportunities.",
		MMR:         2650,
		Region:      "NA",
	},
	{
		Nickname:    "bennyK",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "bennyk",
		Avatar:      "https://i.pravatar.cc/150?img=60",
		Roles:       []string{"AWPer"},
		Description: "The Magic Stick. One of the most aggressive AWPers in CS history. Former world #1 player with legendary highlight reels.",
		MMR:         2600,
		Region:      "EU",
	},
	{
		Nickname:    "Wardel",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "wardel",
		Avatar:      "https://i.pravatar.cc/150?img=52",
		Roles:       []string{"Duelist", "Op Specialist"},
		Description: "NA Operator specialist known for aggressive peeks and clutch rounds. Star player with incredible mechanical skill.",
		MMR:         2550,
		Region:      "NA",
	},
	{
		Nickname:    "gl4ive",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "gl4ive",
		Avatar:      "https://i.pravatar.cc/150?img=57",
		Roles:       []string{"IGL", "Support"},
		Description: "Legendary IGL and 4x Major champion. Mastermind behind the most dominant CS era. Seeking new roster opportunities.",
		MMR:         2680,
		Region:      "EU",
	},
	{
		Nickname:    "KSCERADO",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "kscerado",
		Avatar:      "https://i.pravatar.cc/150?img=36",
		Roles:       []string{"Entry Fragger", "Rifler"},
		Description: "Brazilian prodigy and team cornerstone. Known for explosive entry fragging and mechanical perfection. Top 10 ranked player.",
		MMR:         2720,
		Region:      "SA",
	},
	{
		Nickname:    "ShreaK",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "shreak",
		Avatar:      "https://i.pravatar.cc/150?img=37",
		Roles:       []string{"Duelist", "Entry"},
		Description: "The Headshot Machine. CS legend turned tactical shooter star. Known for inhuman aim and one-tap kills. Team foundation player.",
		MMR:         2620,
		Region:      "EMEA",
	},
	{
		Nickname:    "elektronic",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "elektronic",
		Avatar:      "https://i.pravatar.cc/150?img=38",
		Roles:       []string{"Rifler", "Entry Fragger"},
		Description: "CIS legend and Major champion. One of the best riflers in CS history. Explosive gameplay and clutch factor.",
		MMR:         2780,
		Region:      "CIS",
	},
	{
		Nickname:    "yey",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "yey",
		Avatar:      "https://i.pravatar.cc/150?img=39",
		Roles:       []string{"Sentinel", "Operator"},
		Description: "El Diablo. International Masters champion. Sentinel specialist with unmatched operator precision. LFT for international team.",
		MMR:         2750,
		Region:      "NA",
	},
	{
		Nickname:    "frostzero",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "frostzero",
		Avatar:      "https://i.pravatar.cc/150?img=40",
		Roles:       []string{"Lurker", "AWPer"},
		Description: "Back-to-back world #1 player. 2x Major champion and Brazilian CS icon. Versatile player with legendary clutch ability.",
		MMR:         2700,
		Region:      "SA",
	},
	{
		Nickname:    "cMed",
		GameID:      common.VLRNT_GAME_ID,
		SlugURI:     "cmed",
		Avatar:      "https://i.pravatar.cc/150?img=46",
		Roles:       []string{"Duelist", "Jett"},
		Description: "Turkish tactical shooter superstar. World Champions winner. Known for aggressive plays and insane mechanical skill.",
		MMR:         2680,
		Region:      "EMEA",
	},
	{
		Nickname:    "Stevie3K",
		GameID:      common.CS2_GAME_ID,
		SlugURI:     "stevie3k",
		Avatar:      "https://i.pravatar.cc/150?img=47",
		Roles:       []string{"Entry Fragger", "IGL"},
		Description: "NA Major champion and smoke criminal. Aggressive playstyle that defined a generation. Looking for redemption arc.",
		MMR:         2580,
		Region:      "NA",
	},
}

// Tournament seed data - Professional esports events
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
		Name:            "BLAST Premier Spring Open Qualifier",
		Description:     "Open qualifier for BLAST Premier Spring Groups. Top 2 teams advance to the closed qualifier with $25,000 prize pool.",
		GameID:          common.CS2_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "EU",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 64,
		MinParticipants: 16,
		EntryFee:        1500,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusRegistration,
		DaysFromNow:     5,
	},
	{
		Name:            "VCT Challengers League Open Qualifier",
		Description:     "Your path to VCT Challengers starts here. Top 8 teams earn spots in the Challengers League main event.",
		GameID:          common.VLRNT_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "Americas",
		Format:          tournament_entities.TournamentFormatDoubleElimination,
		MaxParticipants: 128,
		MinParticipants: 32,
		EntryFee:        2000,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusRegistration,
		DaysFromNow:     7,
	},
	{
		Name:            "ESL Pro League Season 20 - Open Qualifier",
		Description:     "Compete for a spot in ESL Pro League! Swiss format with $50,000 prize pool and 4 main event spots up for grabs.",
		GameID:          common.CS2_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "Global",
		Format:          tournament_entities.TournamentFormatSwiss,
		MaxParticipants: 256,
		MinParticipants: 64,
		EntryFee:        5000,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusDraft,
		DaysFromNow:     14,
	},
	{
		Name:            "APAC Premier Invitational",
		Description:     "Elite APAC teams battle for regional supremacy. Live broadcast on all major platforms. $15,000 prize pool.",
		GameID:          common.VLRNT_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "APAC",
		Format:          tournament_entities.TournamentFormatDoubleElimination,
		MaxParticipants: 16,
		MinParticipants: 8,
		EntryFee:        0,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusInProgress,
		DaysFromNow:     0,
	},
	{
		Name:            "IEM Katowice 2025 - Play-In",
		Description:     "The legendary IEM Katowice returns! Play-in stage determines final main event bracket. $1,000,000 total prize pool.",
		GameID:          common.CS2_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "Global",
		Format:          tournament_entities.TournamentFormatDoubleElimination,
		MaxParticipants: 24,
		MinParticipants: 16,
		EntryFee:        0,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusCompleted,
		DaysFromNow:     -45,
	},
	{
		Name:            "NA Ranked Cup - Weekly Series #12",
		Description:     "Weekly grassroots tournament for NA players. Perfect for grinding ranking points and small prizes.",
		GameID:          common.CS2_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "NA",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 32,
		MinParticipants: 8,
		EntryFee:        500,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusRegistration,
		DaysFromNow:     2,
	},
	{
		Name:            "Red Bull Campus Clutch - Regional Finals",
		Description:     "University teams compete for national glory! Winners advance to the World Final in Germany.",
		GameID:          common.VLRNT_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "EU",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 16,
		MinParticipants: 8,
		EntryFee:        0,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusRegistration,
		DaysFromNow:     10,
	},
	{
		Name:            "FACEIT Pro League Showmatch",
		Description:     "Exhibition matches featuring top FPL players. $10,000 showmatch prize plus bragging rights.",
		GameID:          common.CS2_GAME_ID,
		GameMode:        "5v5 Competitive",
		Region:          "EU",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 8,
		MinParticipants: 4,
		EntryFee:        0,
		Currency:        wallet_vo.CurrencyUSD,
		Status:          tournament_entities.TournamentStatusCompleted,
		DaysFromNow:     -7,
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
	defer func() { _ = client.Disconnect(ctx) }()

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
