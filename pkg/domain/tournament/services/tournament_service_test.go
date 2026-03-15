package tournament_services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
	tournament_services "github.com/replay-api/replay-api/pkg/domain/tournament/services"
	tournament_usecases "github.com/replay-api/replay-api/pkg/domain/tournament/usecases"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- MockTournamentRepository ---

type MockTournamentRepo struct {
	mock.Mock
}

func (m *MockTournamentRepo) Save(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return m.Called(ctx, tournament).Error(0)
}

func (m *MockTournamentRepo) Update(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return m.Called(ctx, tournament).Error(0)
}

func (m *MockTournamentRepo) FindByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) FindByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, organizerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) FindByGameAndRegion(ctx context.Context, gameID, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, gameID, region, status, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) FindUpcoming(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, gameID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) FindInProgress(ctx context.Context, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) FindPlayerTournaments(ctx context.Context, playerID uuid.UUID, statusFilter []tournament_entities.TournamentStatus) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, playerID, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*tournament_entities.Tournament, error) {
	args := m.Called(ctx, matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockTournamentRepo) GetByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) Search(ctx context.Context, s shared.Search) ([]tournament_entities.Tournament, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepo) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, searchParams, resultOptions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}

// --- MockWalletCommand ---

type MockWalletCmd struct {
	mock.Mock
}

func (m *MockWalletCmd) CreateWallet(ctx context.Context, cmd wallet_in.CreateWalletCommand) (*wallet_entities.UserWallet, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.UserWallet), args.Error(1)
}

func (m *MockWalletCmd) Deposit(ctx context.Context, cmd wallet_in.DepositCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockWalletCmd) Withdraw(ctx context.Context, cmd wallet_in.WithdrawCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockWalletCmd) DeductEntryFee(ctx context.Context, cmd wallet_in.DeductEntryFeeCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockWalletCmd) AddPrize(ctx context.Context, cmd wallet_in.AddPrizeCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockWalletCmd) Refund(ctx context.Context, cmd wallet_in.RefundCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

func (m *MockWalletCmd) DebitWallet(ctx context.Context, cmd wallet_in.DebitWalletCommand) (*wallet_entities.WalletTransaction, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.WalletTransaction), args.Error(1)
}

func (m *MockWalletCmd) CreditWallet(ctx context.Context, cmd wallet_in.CreditWalletCommand) (*wallet_entities.WalletTransaction, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet_entities.WalletTransaction), args.Error(1)
}

// --- MockBillingHandler ---

type MockBillingHandler struct {
	mock.Mock
}

func (m *MockBillingHandler) Exec(ctx context.Context, cmd billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
	args := m.Called(ctx, cmd)
	var entry *billing_entities.BillableEntry
	var sub *billing_entities.Subscription
	if args.Get(0) != nil {
		entry = args.Get(0).(*billing_entities.BillableEntry)
	}
	if args.Get(1) != nil {
		sub = args.Get(1).(*billing_entities.Subscription)
	}
	return entry, sub, args.Error(2)
}

func (m *MockBillingHandler) Validate(ctx context.Context, cmd billing_in.BillableOperationCommand) error {
	return m.Called(ctx, cmd).Error(0)
}

// --- MockTournamentAuthorization ---

type MockTournamentAuth struct {
	mock.Mock
}

func (m *MockTournamentAuth) IsOrganizer(ctx context.Context, userID uuid.UUID, tournamentID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, tournamentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTournamentAuth) IsPlatformAdmin(ctx context.Context) bool {
	return m.Called(ctx).Bool(0)
}

func (m *MockTournamentAuth) IsParticipant(ctx context.Context, userID uuid.UUID, tournamentID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, tournamentID)
	return args.Bool(0), args.Error(1)
}

// --- MockTournamentEventPublisher ---

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishTournamentCreated(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return m.Called(ctx, tournament).Error(0)
}

func (m *MockEventPublisher) PublishRegistrationOpened(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return m.Called(ctx, tournament).Error(0)
}

func (m *MockEventPublisher) PublishRegistrationClosed(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return m.Called(ctx, tournament).Error(0)
}

func (m *MockEventPublisher) PublishTournamentStarted(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return m.Called(ctx, tournament).Error(0)
}

func (m *MockEventPublisher) PublishTournamentCompleted(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return m.Called(ctx, tournament).Error(0)
}

func (m *MockEventPublisher) PublishTournamentCancelled(ctx context.Context, tournament *tournament_entities.Tournament) error {
	return m.Called(ctx, tournament).Error(0)
}

func (m *MockEventPublisher) PublishPlayerRegistered(ctx context.Context, tournament *tournament_entities.Tournament, playerID uuid.UUID) error {
	return m.Called(ctx, tournament, playerID).Error(0)
}

func (m *MockEventPublisher) PublishPlayerUnregistered(ctx context.Context, tournament *tournament_entities.Tournament, playerID uuid.UUID) error {
	return m.Called(ctx, tournament, playerID).Error(0)
}

func (m *MockEventPublisher) PublishMatchResultRecorded(ctx context.Context, tournament *tournament_entities.Tournament, matchID uuid.UUID, winnerID uuid.UUID) error {
	return m.Called(ctx, tournament, matchID, winnerID).Error(0)
}

func (m *MockEventPublisher) PublishBracketAdvanced(ctx context.Context, tournament *tournament_entities.Tournament, newMatches []tournament_entities.TournamentMatch) error {
	return m.Called(ctx, tournament, newMatches).Error(0)
}

// --- Helpers ---

func testOrganizerContext(organizerID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, organizerID)
	return ctx
}

func testAdminContext(adminID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, adminID)
	ctx = context.WithValue(ctx, shared.AudienceKey, shared.TenantAudienceIDKey)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	return ctx
}

func createTestTournamentEntity(organizerID uuid.UUID) *tournament_entities.Tournament {
	now := time.Now()
	ro := shared.ResourceOwner{UserID: organizerID}
	t, _ := tournament_entities.NewTournament(
		ro,
		"Test Championship",
		"A competitive tournament",
		replay_common.GameIDKey("cs2"),
		"5v5", "NA",
		tournament_entities.TournamentFormatSingleElimination,
		8, 2,
		wallet_vo.NewAmountFromCents(1000),
		wallet_vo.Currency("USD"),
		now.Add(24*time.Hour),
		now.Add(-time.Hour),
		now.Add(12*time.Hour),
		tournament_entities.TournamentRules{BestOf: 3, CheckInRequired: true},
		organizerID,
	)
	return t
}

func createServiceWithMocks() (tournament_in.TournamentCommand, *MockTournamentRepo, *MockWalletCmd, *MockTournamentAuth, *MockEventPublisher) {
	repo := new(MockTournamentRepo)
	wallet := new(MockWalletCmd)
	auth := new(MockTournamentAuth)
	events := new(MockEventPublisher)
	billing := new(MockBillingHandler)
	bracketGen := tournament_usecases.NewGenerateBracketsUseCase(billing, repo)
	svc := tournament_services.NewTournamentService(repo, wallet, bracketGen, auth, events)
	return svc, repo, wallet, auth, events
}

// ============================================================================
// RBAC Authorization Tests
// ============================================================================

func TestTournamentService_RBAC_OrganizerCanOpenRegistration(t *testing.T) {
	svc, repo, _, auth, events := createServiceWithMocks()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	ctx := testOrganizerContext(organizerID)

	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, organizerID, tournament.ID).Return(true, nil)
	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	events.On("PublishRegistrationOpened", ctx, mock.Anything).Return(nil)

	err := svc.OpenRegistration(ctx, tournament.ID)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Update", ctx, mock.Anything)
	events.AssertCalled(t, "PublishRegistrationOpened", ctx, mock.Anything)
}

func TestTournamentService_RBAC_AdminCanOpenRegistration(t *testing.T) {
	svc, repo, _, auth, events := createServiceWithMocks()
	adminID := uuid.New()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	ctx := testAdminContext(adminID)

	auth.On("IsPlatformAdmin", ctx).Return(true)
	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	events.On("PublishRegistrationOpened", ctx, mock.Anything).Return(nil)

	err := svc.OpenRegistration(ctx, tournament.ID)
	assert.NoError(t, err)
}

func TestTournamentService_RBAC_UnauthorizedCannotOpenRegistration(t *testing.T) {
	svc, _, _, auth, _ := createServiceWithMocks()
	unauthorizedID := uuid.New()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	ctx := testOrganizerContext(unauthorizedID)

	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, unauthorizedID, tournament.ID).Return(false, nil)

	err := svc.OpenRegistration(ctx, tournament.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTournamentService_RBAC_UnauthorizedCannotCancelTournament(t *testing.T) {
	svc, _, _, auth, _ := createServiceWithMocks()
	unauthorizedID := uuid.New()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	ctx := testOrganizerContext(unauthorizedID)

	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, unauthorizedID, tournament.ID).Return(false, nil)

	cmd := tournament_in.CancelTournamentCommand{TournamentID: tournament.ID, Reason: "test"}
	err := svc.CancelTournament(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTournamentService_RBAC_UnauthorizedCannotRecordMatchResult(t *testing.T) {
	svc, _, _, auth, _ := createServiceWithMocks()
	unauthorizedID := uuid.New()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	ctx := testOrganizerContext(unauthorizedID)

	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, unauthorizedID, tournament.ID).Return(false, nil)

	err := svc.RecordMatchResult(ctx, tournament.ID, uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTournamentService_RBAC_NilAuthorizationAllowsAll(t *testing.T) {
	repo := new(MockTournamentRepo)
	wallet := new(MockWalletCmd)
	billing := new(MockBillingHandler)
	bracketGen := tournament_usecases.NewGenerateBracketsUseCase(billing, repo)
	svc := tournament_services.NewTournamentService(repo, wallet, bracketGen, nil, nil)

	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	ctx := testOrganizerContext(uuid.New())

	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)

	err := svc.OpenRegistration(ctx, tournament.ID)
	assert.NoError(t, err)
}

// ============================================================================
// Event Publishing Tests
// ============================================================================

func TestTournamentService_EventPublishing_CreateTournament(t *testing.T) {
	svc, repo, _, _, events := createServiceWithMocks()
	organizerID := uuid.New()
	ctx := testOrganizerContext(organizerID)

	repo.On("Save", ctx, mock.Anything).Return(nil)
	events.On("PublishTournamentCreated", ctx, mock.Anything).Return(nil)

	cmd := tournament_in.CreateTournamentCommand{
		ResourceOwner:     shared.ResourceOwner{UserID: organizerID},
		Name:              "Test",
		Description:       "Test desc",
		GameID:            replay_common.GameIDKey("cs2"),
		GameMode:          "5v5",
		Region:            "NA",
		Format:            tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants:   8,
		MinParticipants:   2,
		EntryFee:          wallet_vo.NewAmountFromCents(0),
		Currency:          wallet_vo.Currency("USD"),
		StartTime:         time.Now().Add(24 * time.Hour),
		RegistrationOpen:  time.Now().Add(-time.Hour),
		RegistrationClose: time.Now().Add(12 * time.Hour),
		Rules:             tournament_entities.TournamentRules{BestOf: 3},
		OrganizerID:       organizerID,
	}

	result, err := svc.CreateTournament(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	events.AssertCalled(t, "PublishTournamentCreated", ctx, mock.Anything)
}

func TestTournamentService_EventPublishing_CancelTournament(t *testing.T) {
	svc, repo, _, auth, events := createServiceWithMocks()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	tournament.EntryFee = wallet_vo.NewAmountFromCents(0)
	ctx := testOrganizerContext(organizerID)

	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, organizerID, tournament.ID).Return(true, nil)
	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	events.On("PublishTournamentCancelled", ctx, mock.Anything).Return(nil)

	cmd := tournament_in.CancelTournamentCommand{TournamentID: tournament.ID, Reason: "insufficient participants"}
	err := svc.CancelTournament(ctx, cmd)
	assert.NoError(t, err)
	events.AssertCalled(t, "PublishTournamentCancelled", ctx, mock.Anything)
}

// ============================================================================
// Registration & Wallet Integration Tests
// ============================================================================

func TestTournamentService_RegisterPlayer_ChargesEntryFee(t *testing.T) {
	svc, repo, wallet, _, events := createServiceWithMocks()
	organizerID := uuid.New()
	playerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	tournament.Status = tournament_entities.TournamentStatusRegistration
	ctx := testOrganizerContext(playerID)

	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	wallet.On("DebitWallet", ctx, mock.MatchedBy(func(cmd wallet_in.DebitWalletCommand) bool {
		return cmd.UserID == playerID
	})).Return(&wallet_entities.WalletTransaction{}, nil)
	events.On("PublishPlayerRegistered", ctx, mock.Anything, playerID).Return(nil)

	cmd := tournament_in.RegisterPlayerCommand{TournamentID: tournament.ID, PlayerID: playerID, DisplayName: "TestPlayer"}
	err := svc.RegisterPlayer(ctx, cmd)
	assert.NoError(t, err)
	wallet.AssertCalled(t, "DebitWallet", ctx, mock.Anything)
	events.AssertCalled(t, "PublishPlayerRegistered", ctx, mock.Anything, playerID)
}

func TestTournamentService_RegisterPlayer_EntryFeeFailureRollsBack(t *testing.T) {
	svc, repo, wallet, _, _ := createServiceWithMocks()
	organizerID := uuid.New()
	playerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	tournament.Status = tournament_entities.TournamentStatusRegistration
	ctx := testOrganizerContext(playerID)

	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	wallet.On("DebitWallet", ctx, mock.Anything).Return(nil, assert.AnError)

	cmd := tournament_in.RegisterPlayerCommand{TournamentID: tournament.ID, PlayerID: playerID, DisplayName: "TestPlayer"}
	err := svc.RegisterPlayer(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entry fee")
	assert.False(t, tournament.IsPlayerRegistered(playerID))
}

func TestTournamentService_UnregisterPlayer_RefundsEntryFee(t *testing.T) {
	svc, repo, wallet, _, events := createServiceWithMocks()
	organizerID := uuid.New()
	playerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	tournament.Status = tournament_entities.TournamentStatusRegistration
	_ = tournament.RegisterPlayer(playerID, "TestPlayer")
	ctx := testOrganizerContext(playerID)

	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	wallet.On("CreditWallet", ctx, mock.MatchedBy(func(cmd wallet_in.CreditWalletCommand) bool {
		return cmd.UserID == playerID
	})).Return(&wallet_entities.WalletTransaction{}, nil)
	events.On("PublishPlayerUnregistered", ctx, mock.Anything, playerID).Return(nil)

	cmd := tournament_in.UnregisterPlayerCommand{TournamentID: tournament.ID, PlayerID: playerID}
	err := svc.UnregisterPlayer(ctx, cmd)
	assert.NoError(t, err)
	wallet.AssertCalled(t, "CreditWallet", ctx, mock.Anything)
	events.AssertCalled(t, "PublishPlayerUnregistered", ctx, mock.Anything, playerID)
}

func TestTournamentService_CancelTournament_RefundsAllParticipants(t *testing.T) {
	svc, repo, wallet, auth, events := createServiceWithMocks()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	tournament.Status = tournament_entities.TournamentStatusRegistration
	p1 := uuid.New()
	p2 := uuid.New()
	_ = tournament.RegisterPlayer(p1, "P1")
	_ = tournament.RegisterPlayer(p2, "P2")
	ctx := testOrganizerContext(organizerID)

	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, organizerID, tournament.ID).Return(true, nil)
	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	wallet.On("CreditWallet", ctx, mock.Anything).Return(&wallet_entities.WalletTransaction{}, nil)
	events.On("PublishTournamentCancelled", ctx, mock.Anything).Return(nil)

	cmd := tournament_in.CancelTournamentCommand{TournamentID: tournament.ID, Reason: "cancelled by organizer"}
	err := svc.CancelTournament(ctx, cmd)
	assert.NoError(t, err)
	wallet.AssertNumberOfCalls(t, "CreditWallet", 2)
}

// ============================================================================
// Match Result & Bracket Advancement Tests
// ============================================================================

func TestTournamentService_RecordMatchResult_Success(t *testing.T) {
	svc, repo, _, auth, events := createServiceWithMocks()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	tournament.Status = tournament_entities.TournamentStatusInProgress
	p1 := uuid.New()
	p2 := uuid.New()
	tournament.Participants = []tournament_entities.TournamentPlayer{
		{PlayerID: p1, DisplayName: "P1", Status: "registered"},
		{PlayerID: p2, DisplayName: "P2", Status: "registered"},
	}
	matchID := uuid.New()
	tournament.Matches = []tournament_entities.TournamentMatch{
		{MatchID: matchID, Round: 1, Player1ID: p1, Player2ID: p2, Status: tournament_entities.MatchStatusScheduled},
	}
	ctx := testOrganizerContext(organizerID)

	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, organizerID, tournament.ID).Return(true, nil)
	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	events.On("PublishMatchResultRecorded", ctx, mock.Anything, matchID, p1).Return(nil)

	err := svc.RecordMatchResult(ctx, tournament.ID, matchID, p1)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Update", ctx, mock.Anything)
	events.AssertCalled(t, "PublishMatchResultRecorded", ctx, mock.Anything, matchID, p1)
}

func TestTournamentService_AdvanceBracket_Success(t *testing.T) {
	svc, repo, _, auth, events := createServiceWithMocks()
	organizerID := uuid.New()
	tournament := createTestTournamentEntity(organizerID)
	tournament.Status = tournament_entities.TournamentStatusInProgress
	p1 := uuid.New()
	p2 := uuid.New()
	p3 := uuid.New()
	p4 := uuid.New()
	tournament.Participants = []tournament_entities.TournamentPlayer{
		{PlayerID: p1, DisplayName: "P1", Status: "registered"},
		{PlayerID: p2, DisplayName: "P2", Status: "registered"},
		{PlayerID: p3, DisplayName: "P3", Status: "registered"},
		{PlayerID: p4, DisplayName: "P4", Status: "registered"},
	}
	now := time.Now()
	tournament.Matches = []tournament_entities.TournamentMatch{
		{MatchID: uuid.New(), Round: 1, Player1ID: p1, Player2ID: p2, WinnerID: &p1, Status: tournament_entities.MatchStatusCompleted, CompletedAt: &now},
		{MatchID: uuid.New(), Round: 1, Player1ID: p3, Player2ID: p4, WinnerID: &p3, Status: tournament_entities.MatchStatusCompleted, CompletedAt: &now},
	}
	ctx := testOrganizerContext(organizerID)

	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, organizerID, tournament.ID).Return(true, nil)
	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	events.On("PublishBracketAdvanced", ctx, mock.Anything, mock.Anything).Return(nil)

	err := svc.AdvanceBracket(ctx, tournament.ID)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Update", ctx, mock.Anything)
	events.AssertCalled(t, "PublishBracketAdvanced", ctx, mock.Anything, mock.Anything)
}

// ============================================================================
// Full Lifecycle Test (Financial-Grade)
// ============================================================================

func TestTournamentService_FullLifecycle(t *testing.T) {
	svc, repo, wallet, auth, events := createServiceWithMocks()
	organizerID := uuid.New()
	ctx := testOrganizerContext(organizerID)

	// 1. Create tournament
	repo.On("Save", ctx, mock.Anything).Return(nil)
	events.On("PublishTournamentCreated", ctx, mock.Anything).Return(nil)

	createCmd := tournament_in.CreateTournamentCommand{
		ResourceOwner:     shared.ResourceOwner{UserID: organizerID},
		Name:              "Full Lifecycle Tournament",
		Description:       "Test",
		GameID:            replay_common.GameIDKey("cs2"),
		GameMode:          "5v5",
		Region:            "NA",
		Format:            tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants:   4,
		MinParticipants:   2,
		EntryFee:          wallet_vo.NewAmountFromCents(500),
		Currency:          wallet_vo.Currency("USD"),
		StartTime:         time.Now().Add(24 * time.Hour),
		RegistrationOpen:  time.Now().Add(-time.Hour),
		RegistrationClose: time.Now().Add(12 * time.Hour),
		Rules:             tournament_entities.TournamentRules{BestOf: 3},
		OrganizerID:       organizerID,
	}

	tournament, err := svc.CreateTournament(ctx, createCmd)
	require.NoError(t, err)
	require.NotNil(t, tournament)
	assert.Equal(t, tournament_entities.TournamentStatusDraft, tournament.Status)

	// 2. Open registration
	auth.On("IsPlatformAdmin", ctx).Return(false)
	auth.On("IsOrganizer", ctx, organizerID, tournament.ID).Return(true, nil)
	repo.On("FindByID", ctx, tournament.ID).Return(tournament, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	events.On("PublishRegistrationOpened", ctx, mock.Anything).Return(nil)

	err = svc.OpenRegistration(ctx, tournament.ID)
	require.NoError(t, err)
	assert.Equal(t, tournament_entities.TournamentStatusRegistration, tournament.Status)

	// 3. Register players
	player1 := uuid.New()
	player2 := uuid.New()
	wallet.On("DebitWallet", ctx, mock.Anything).Return(&wallet_entities.WalletTransaction{}, nil)
	events.On("PublishPlayerRegistered", ctx, mock.Anything, mock.Anything).Return(nil)

	err = svc.RegisterPlayer(ctx, tournament_in.RegisterPlayerCommand{
		TournamentID: tournament.ID, PlayerID: player1, DisplayName: "Player1",
	})
	require.NoError(t, err)

	err = svc.RegisterPlayer(ctx, tournament_in.RegisterPlayerCommand{
		TournamentID: tournament.ID, PlayerID: player2, DisplayName: "Player2",
	})
	require.NoError(t, err)

	assert.True(t, tournament.IsPlayerRegistered(player1))
	assert.True(t, tournament.IsPlayerRegistered(player2))

	// 4. Close registration
	events.On("PublishRegistrationClosed", ctx, mock.Anything).Return(nil)
	err = svc.CloseRegistration(ctx, tournament.ID)
	require.NoError(t, err)
	assert.Equal(t, tournament_entities.TournamentStatusReady, tournament.Status)

	wallet.AssertNumberOfCalls(t, "DebitWallet", 2)
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ tournament_out.TournamentRepository = (*MockTournamentRepo)(nil)
var _ wallet_in.WalletCommand = (*MockWalletCmd)(nil)
var _ tournament_out.TournamentAuthorization = (*MockTournamentAuth)(nil)
var _ tournament_out.TournamentEventPublisher = (*MockEventPublisher)(nil)
var _ billing_in.BillableOperationCommandHandler = (*MockBillingHandler)(nil)
