package metadata_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/out"
	replay_services_metadata "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/services/metadata"
)

type mockReplayFileMetadataReader struct {
	replayFiles []replay_entity.ReplayFile
}

func (m *mockReplayFileMetadataReader) Search(ctx context.Context, s common.Search) ([]replay_entity.ReplayFile, error) {
	return m.replayFiles, nil
}

func (m *mockReplayFileMetadataReader) Compile(ctx context.Context, searchParams []common.SearchAggregation, resultOptions common.SearchResultOptions) (*common.Search, error) {
	return &common.Search{
		SearchParams:  searchParams,
		ResultOptions: resultOptions,
	}, nil
}

func (m *mockReplayFileMetadataReader) GetByID(ctx context.Context, id uuid.UUID) (*replay_entity.ReplayFile, error) {
	for _, file := range m.replayFiles {
		if file.ID == id {
			return &file, nil
		}
	}
	return nil, fmt.Errorf("replay file not found")
}

func TestReplayFileQueryService_Filter(t *testing.T) {
	tenantID := uuid.New()
	clientID := uuid.New()
	userID := uuid.New()

	entity := common.NewEntity(common.ResourceOwner{TenantID: tenantID, ClientID: clientID, UserID: userID})

	sampleReplayFiles := []replay_entity.ReplayFile{
		{
			ID:            entity.ID,
			GameID:        common.CS2_GAME_ID,
			NetworkID:     common.SteamNetworkIDKey,
			Header:        struct{ Filestamp string }{Filestamp: "HLTV-1.0.0"},
			ResourceOwner: entity.ResourceOwner,
			CreatedAt:     entity.CreatedAt,
			UpdatedAt:     entity.UpdatedAt,
		},
	}

	tests := []struct {
		name           string
		searchParams   []common.SearchAggregation
		resultOptions  common.SearchResultOptions
		mockReader     replay_out.ReplayFileMetadataReader
		expectedOutput []replay_entity.ReplayFile
		expectedError  error
		contextValues  map[interface{}]interface{}
	}{
		{
			name: "Valid Query - GameID",
			searchParams: []common.SearchAggregation{
				{
					Params: []common.SearchParameter{
						{
							ValueParams: []common.SearchableValue{
								{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}},
							},
						},
					},
				},
			},
			resultOptions: common.SearchResultOptions{Limit: 10},
			mockReader:    &mockReplayFileMetadataReader{replayFiles: sampleReplayFiles},
			expectedOutput: []replay_entity.ReplayFile{
				sampleReplayFiles[0],
			},
			contextValues: map[interface{}]interface{}{
				common.TenantIDKey: tenantID,
				common.ClientIDKey: clientID,
				common.UserIDKey:   userID,
			},
		},
		{
			name:           "Invalid Query - Missing TenantID in Context",
			searchParams:   []common.SearchAggregation{}, // Empty search params
			resultOptions:  common.SearchResultOptions{Limit: 10},
			mockReader:     &mockReplayFileMetadataReader{replayFiles: sampleReplayFiles},
			expectedOutput: nil,
			expectedError:  fmt.Errorf("GetResourceOwner.IsMissingTenant: tenant_id missing in context context.Background"),
			contextValues:  map[interface{}]interface{}{}, // Empty context to trigger the error
		},
		{
			name:           "Invalid Query - Mismatched TenantID in Context",
			searchParams:   []common.SearchAggregation{}, // Empty search params
			resultOptions:  common.SearchResultOptions{Limit: 10},
			mockReader:     &mockReplayFileMetadataReader{replayFiles: sampleReplayFiles},
			expectedOutput: nil,
			expectedError:  fmt.Errorf("GetResourceOwner.IsMissingTenant: tenant_id missing in context context.Background.WithValue(common.ContextKey, 00000000-0000-0000-0000-000000000000)"),
			contextValues:  map[interface{}]interface{}{common.TenantIDKey: uuid.Nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			for key, value := range tt.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}
			service := replay_services_metadata.NewReplayFileQueryService(tt.mockReader)

			defer func() {
				if r := recover(); r != nil {
					if !reflect.DeepEqual(r, tt.expectedError) {
						t.Errorf("Expected error %v, but got %v", tt.expectedError, r)
					}
				}
			}()

			s, err := service.Compile(ctx, tt.searchParams, tt.resultOptions)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v, but got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error %v, but got %v", tt.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			output, err := service.Search(ctx, *s)

			if tt.expectedError == nil {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(output, tt.expectedOutput) {
					t.Errorf("Expected output %v, but got %v", tt.expectedOutput, output)
				}
			}
		})
	}
}
