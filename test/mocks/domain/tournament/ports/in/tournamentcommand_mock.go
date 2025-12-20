package tournament_in

import (
	"context"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	"github.com/stretchr/testify/mock"
)

// MockTournamentCommand is a mock implementation of TournamentCommand
type MockTournamentCommand struct {
	mock.Mock
}

// CreateTournament provides a mock function
func (_m *MockTournamentCommand) CreateTournament(ctx context.Context, cmd tournament_in.CreateTournamentCommand) (*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, tournament_in.CreateTournamentCommand) (*tournament_entities.Tournament, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, tournament_in.CreateTournamentCommand) *tournament_entities.Tournament); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// UpdateTournament provides a mock function
func (_m *MockTournamentCommand) UpdateTournament(ctx context.Context, cmd tournament_in.UpdateTournamentCommand) (*tournament_entities.Tournament, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *tournament_entities.Tournament
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, tournament_in.UpdateTournamentCommand) (*tournament_entities.Tournament, error)); ok {
		return rf(ctx, cmd)
	}

	if rf, ok := ret.Get(0).(func(context.Context, tournament_in.UpdateTournamentCommand) *tournament_entities.Tournament); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tournament_entities.Tournament)
		}
	}
	r1 = ret.Error(1)

	return r0, r1
}

// DeleteTournament provides a mock function
func (_m *MockTournamentCommand) DeleteTournament(ctx context.Context, tournamentID uuid.UUID) error {
	ret := _m.Called(ctx, tournamentID)

	return ret.Error(0)
}

// RegisterPlayer provides a mock function
func (_m *MockTournamentCommand) RegisterPlayer(ctx context.Context, cmd tournament_in.RegisterPlayerCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// UnregisterPlayer provides a mock function
func (_m *MockTournamentCommand) UnregisterPlayer(ctx context.Context, cmd tournament_in.UnregisterPlayerCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// OpenRegistration provides a mock function
func (_m *MockTournamentCommand) OpenRegistration(ctx context.Context, tournamentID uuid.UUID) error {
	ret := _m.Called(ctx, tournamentID)

	return ret.Error(0)
}

// CloseRegistration provides a mock function
func (_m *MockTournamentCommand) CloseRegistration(ctx context.Context, tournamentID uuid.UUID) error {
	ret := _m.Called(ctx, tournamentID)

	return ret.Error(0)
}

// StartTournament provides a mock function
func (_m *MockTournamentCommand) StartTournament(ctx context.Context, tournamentID uuid.UUID) error {
	ret := _m.Called(ctx, tournamentID)

	return ret.Error(0)
}

// GenerateBrackets provides a mock function
func (_m *MockTournamentCommand) GenerateBrackets(ctx context.Context, tournamentID uuid.UUID) error {
	ret := _m.Called(ctx, tournamentID)

	return ret.Error(0)
}

// ScheduleMatches provides a mock function
func (_m *MockTournamentCommand) ScheduleMatches(ctx context.Context, cmd tournament_in.ScheduleMatchesCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// RescheduleMatch provides a mock function
func (_m *MockTournamentCommand) RescheduleMatch(ctx context.Context, cmd tournament_in.RescheduleMatchCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// ReportMatchResult provides a mock function
func (_m *MockTournamentCommand) ReportMatchResult(ctx context.Context, cmd tournament_in.ReportMatchResultCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// CompleteTournament provides a mock function
func (_m *MockTournamentCommand) CompleteTournament(ctx context.Context, cmd tournament_in.CompleteTournamentCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// CancelTournament provides a mock function
func (_m *MockTournamentCommand) CancelTournament(ctx context.Context, cmd tournament_in.CancelTournamentCommand) error {
	ret := _m.Called(ctx, cmd)

	return ret.Error(0)
}

// NewMockTournamentCommand creates a new instance of MockTournamentCommand
func NewMockTournamentCommand(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTournamentCommand {
	mock := &MockTournamentCommand{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
