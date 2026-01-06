package metadata_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
	replay_services_metadata "github.com/replay-api/replay-api/pkg/domain/replay/services/metadata"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type mockReplayFileMetadataReader struct {
	replayFiles []replay_entity.ReplayFile
}

func (m *mockReplayFileMetadataReader) Search(ctx context.Context, s shared.Search) ([]replay_entity.ReplayFile, error) {
	return m.replayFiles, nil
}

func (m *mockReplayFileMetadataReader) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	return &shared.Search{
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

	entity := shared.NewEntity(shared.ResourceOwner{TenantID: tenantID, ClientID: clientID, UserID: userID})

	sampleReplayFiles := []replay_entity.ReplayFile{
		{
			ID:            entity.ID,
			GameID:        replay_common.CS2_GAME_ID,
			NetworkID:     replay_common.SteamNetworkIDKey,
			Header:        struct{ Filestamp string }{Filestamp: "HLTV-1.0.0"},
			ResourceOwner: entity.ResourceOwner,
			CreatedAt:     entity.CreatedAt,
			UpdatedAt:     entity.UpdatedAt,
		},
	}

	tests := []struct {
		name           string
		searchParams   []shared.SearchAggregation
		resultOptions  shared.SearchResultOptions
		mockReader     replay_out.ReplayFileMetadataReader
		expectedOutput []replay_entity.ReplayFile
		expectedError  error
		contextValues  map[interface{}]interface{}
	}{
		{
			name: "Valid Query - GameID",
			searchParams: []shared.SearchAggregation{
				{
					Params: []shared.SearchParameter{
						{
							ValueParams: []shared.SearchableValue{
								{Field: "GameID", Values: []interface{}{replay_common.CS2_GAME_ID}},
							},
						},
					},
				},
			},
			resultOptions: shared.SearchResultOptions{Limit: 10},
			mockReader:    &mockReplayFileMetadataReader{replayFiles: sampleReplayFiles},
			expectedOutput: []replay_entity.ReplayFile{
				sampleReplayFiles[0],
			},
			contextValues: map[interface{}]interface{}{
				shared.TenantIDKey: tenantID,
				shared.ClientIDKey: clientID,
				shared.UserIDKey:   userID,
			},
		},
		{
			name:           "Invalid Query - Missing TenantID in Context",
			searchParams:   []shared.SearchAggregation{}, // Empty search params
			resultOptions:  shared.SearchResultOptions{Limit: 10},
			mockReader:     &mockReplayFileMetadataReader{replayFiles: sampleReplayFiles},
			expectedOutput: nil,
			expectedError:  fmt.Errorf("GetResourceOwner.IsMissingTenant: tenant_id missing in context context.Background"),
			contextValues:  map[interface{}]interface{}{}, // Empty context to trigger the error
		},
		{
			name:           "Invalid Query - Mismatched TenantID in Context",
			searchParams:   []shared.SearchAggregation{}, // Empty search params
			resultOptions:  shared.SearchResultOptions{Limit: 10},
			mockReader:     &mockReplayFileMetadataReader{replayFiles: sampleReplayFiles},
			expectedOutput: nil,
			expectedError:  fmt.Errorf("GetResourceOwner.IsMissingTenant: tenant_id missing in context context.Background.WithValue(shared.ContextKey, 00000000-0000-0000-0000-000000000000)"),
			contextValues:  map[interface{}]interface{}{shared.TenantIDKey: uuid.Nil},
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
