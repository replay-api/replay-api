package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlayerQueuedInput_ValidateResourceOwnership(t *testing.T) {
	valid := &PlayerQueuedInput{
		PlayerID:        "player-1",
		GameID:          "game-1",
		Region:          "us-east",
		TenantID:        "tenant-1",
		ClientID:        "client-1",
		ResourceOwnerID: "rid-1",
	}
	require.NoError(t, valid.ValidateResourceOwnership())

	tests := []struct {
		name    string
		input   *PlayerQueuedInput
		wantErr bool
	}{
		{"missing tenant", &PlayerQueuedInput{TenantID: "", ClientID: "c", ResourceOwnerID: "r"}, true},
		{"missing client", &PlayerQueuedInput{TenantID: "t", ClientID: "", ResourceOwnerID: "r"}, true},
		{"missing resource_owner", &PlayerQueuedInput{TenantID: "t", ClientID: "c", ResourceOwnerID: ""}, true},
		{"whitespace tenant", &PlayerQueuedInput{TenantID: "  ", ClientID: "c", ResourceOwnerID: "r"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.ValidateResourceOwnership()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrResourceOwnershipInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMatchCompletedInput_ValidateResourceOwnership(t *testing.T) {
	valid := &MatchCompletedInput{
		MatchID:         "match-1",
		TenantID:        "tenant-1",
		ClientID:        "client-1",
		ResourceOwnerID: "rid-1",
	}
	require.NoError(t, valid.ValidateResourceOwnership())

	tests := []struct {
		name    string
		input   *MatchCompletedInput
		wantErr bool
	}{
		{"missing tenant", &MatchCompletedInput{TenantID: "", ClientID: "c", ResourceOwnerID: "r"}, true},
		{"missing client", &MatchCompletedInput{TenantID: "t", ClientID: "", ResourceOwnerID: "r"}, true},
		{"missing resource_owner", &MatchCompletedInput{TenantID: "t", ClientID: "c", ResourceOwnerID: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.ValidateResourceOwnership()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrResourceOwnershipInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMatchmakingProducer_PublishPlayerQueued_Disabled(t *testing.T) {
	cfg := DefaultMatchmakingProducerConfig()
	cfg.Enabled = false
	client, err := NewClient(&Config{BootstrapServers: "localhost:9092", Region: "test"})
	require.NoError(t, err)
	defer client.Close()

	mp := NewMatchmakingProducer(client, cfg)
	err = mp.PublishPlayerQueued(context.Background(), &PlayerQueuedInput{
		PlayerID:        "p1",
		GameID:          "g1",
		Region:          "us-east",
		TenantID:        "t1",
		ClientID:        "c1",
		ResourceOwnerID: "r1",
	})
	assert.ErrorIs(t, err, ErrProducerDisabled)
}

func TestMatchmakingProducer_PublishMatchCompleted_Disabled(t *testing.T) {
	cfg := DefaultMatchmakingProducerConfig()
	cfg.Enabled = false
	client, err := NewClient(&Config{BootstrapServers: "localhost:9092", Region: "test"})
	require.NoError(t, err)
	defer client.Close()

	mp := NewMatchmakingProducer(client, cfg)
	err = mp.PublishMatchCompleted(context.Background(), &MatchCompletedInput{
		MatchID:         "m1",
		TenantID:        "t1",
		ClientID:        "c1",
		ResourceOwnerID: "r1",
	})
	assert.ErrorIs(t, err, ErrProducerDisabled)
}
