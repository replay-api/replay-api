package squad_usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_usecases "github.com/replay-api/replay-api/pkg/domain/squad/usecases"
	squad_value_objects "github.com/replay-api/replay-api/pkg/domain/squad/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPlayerProfileReader struct {
	mock.Mock
}

func (m *MockPlayerProfileReader) Compile(ctx context.Context, aggregations []common.SearchAggregation, options common.SearchResultOptions) (*common.Search, error) {
	args := m.Called(ctx, aggregations, options)
	return args.Get(0).(*common.Search), args.Error(1)
}

func (m *MockPlayerProfileReader) Search(ctx context.Context, search common.Search) ([]squad_entities.PlayerProfile, error) {
	args := m.Called(ctx, search)
	return args.Get(0).([]squad_entities.PlayerProfile), args.Error(1)
}

type MockSquadHistoryWriter struct {
	mock.Mock
}

func (m *MockPlayerProfileReader) GetByID(ctx context.Context, id uuid.UUID) (*squad_entities.PlayerProfile, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*squad_entities.PlayerProfile), args.Error(1)
}

func (m *MockSquadHistoryWriter) CreateMany(ctx context.Context, histories []*squad_entities.SquadHistory) error {
	args := m.Called(ctx, histories)
	return args.Error(0)
}

func (m *MockSquadHistoryWriter) Create(ctx context.Context, history *squad_entities.SquadHistory) (*squad_entities.SquadHistory, error) {
	args := m.Called(ctx, history)
	return args.Get(0).(*squad_entities.SquadHistory), args.Error(1)
}

func TestProcessMemberships(t *testing.T) {
	mockPlayerProfileReader := new(MockPlayerProfileReader)
	mockSquadHistoryWriter := new(MockSquadHistoryWriter)

	updateSquadUseCase := &squad_usecases.UpdateSquadUseCase{
		PlayerProfileReader: mockPlayerProfileReader,
		SquadHistoryWriter:  mockSquadHistoryWriter,
	}

	playerProfileID := uuid.New()
	userID := uuid.New()
	squadID := uuid.New()

	playerProfile := squad_entities.PlayerProfile{
		BaseEntity: common.BaseEntity{
			ID: playerProfileID,
			ResourceOwner: common.ResourceOwner{
				UserID:   userID,
				TenantID: uuid.New(),
			},
		},
	}

	mockPlayerProfileReader.On("Search", mock.Anything, mock.MatchedBy(func(search common.Search) bool {
		for _, filter := range search.SearchParams {
			for _, v := range filter.Params {
				for _, p := range v.ValueParams {
					if p.Field == "ID" {
						for _, id := range p.Values {
							if id == playerProfileID.String() {
								return true
							}
						}
					}
				}
			}
		}
		return true
	})).Return([]squad_entities.PlayerProfile{playerProfile}, nil)

	mockPlayerProfileReader.On("Search", mock.Anything, mock.MatchedBy(func(search common.Search) bool {
		for _, filter := range search.SearchParams {
			for _, v := range filter.Params {
				for _, p := range v.ValueParams {
					if p.Field == "ID" {
						for _, id := range p.Values {
							if id == playerProfileID.String() {
								return true
							}
						}
					}
				}
			}
		}
		return true
	})).Return([]squad_entities.PlayerProfile{}, common.NewErrNotFound(common.ResourceTypePlayerProfile, "ID", uuid.New().String()))
	mockSquadHistoryWriter.On("Create", mock.Anything, mock.Anything).Return(&squad_entities.SquadHistory{}, nil)

	tests := []struct {
		name          string
		squad         *squad_entities.Squad
		members       map[string]squad_in.CreateSquadMembershipInput
		expectedError error
	}{
		{
			name: "Valid Membership",
			squad: &squad_entities.Squad{
				BaseEntity: common.BaseEntity{
					ID: squadID,
					ResourceOwner: common.ResourceOwner{
						UserID:   userID,
						TenantID: uuid.New(),
					},
				},
			},
			members: map[string]squad_in.CreateSquadMembershipInput{
				playerProfileID.String(): {
					Type:   squad_value_objects.SquadMembershipTypeMember,
					Roles:  []string{"role1"},
					Status: squad_value_objects.SquadMembershipStatusActive,
				},
			},
			expectedError: nil,
		},
		{
			name: "Demote Membership",
			squad: &squad_entities.Squad{
				BaseEntity: common.BaseEntity{
					ID: squadID,
					ResourceOwner: common.ResourceOwner{
						UserID:   userID,
						TenantID: uuid.New(),
					},
				},
				Membership: []squad_value_objects.SquadMembership{
					{
						UserID: userID,
						Type:   squad_value_objects.SquadMembershipTypeAdmin,
					},
				},
			},
			members: map[string]squad_in.CreateSquadMembershipInput{
				playerProfileID.String(): {
					Type:   squad_value_objects.SquadMembershipTypeMember,
					Roles:  []string{"role1"},
					Status: squad_value_objects.SquadMembershipStatusActive,
				},
			},
			expectedError: nil,
		},
		{
			name: "Promote Membership",
			squad: &squad_entities.Squad{
				BaseEntity: common.BaseEntity{
					ID: squadID,
					ResourceOwner: common.ResourceOwner{
						UserID:   userID,
						TenantID: uuid.New(),
					},
				},
				Membership: []squad_value_objects.SquadMembership{
					{
						UserID: userID,
						Type:   squad_value_objects.SquadMembershipTypeMember,
					},
				},
			},
			members: map[string]squad_in.CreateSquadMembershipInput{
				playerProfileID.String(): {
					Type:   squad_value_objects.SquadMembershipTypeAdmin,
					Roles:  []string{"role1"},
					Status: squad_value_objects.SquadMembershipStatusActive,
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
			ctx = context.WithValue(ctx, common.UserIDKey, userID)
			ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
			memberships, err := updateSquadUseCase.ProcessMemberships(ctx, tt.squad, tt.members)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, memberships)
			}
		})
	}
}
