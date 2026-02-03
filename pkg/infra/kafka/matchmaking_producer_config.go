package kafka

import (
	"os"
	"strconv"
	"time"
)

// MatchmakingProducerConfig holds configuration for the matchmaking Kafka producers.
// All settings can be externalised via environment variables for non-prod/test clusters.
type MatchmakingProducerConfig struct {
	// TopicNames
	TopicCommands string // PlayerQueued (e.g. matchmaking.commands)
	TopicMatches  string // MatchCompleted (e.g. matchmaking.matches)

	// Producer behaviour
	Enabled       bool          // Can disable producer for non-prod or testing
	MaxRetries    int           // Max retries on transient failures
	InitialBackoff time.Duration // Initial backoff between retries
	MaxBackoff    time.Duration // Max backoff cap
	ProduceTimeout time.Duration // Timeout for each produce call

	// Source identifier for event envelope (e.g. "replay-api")
	Source string
}

// DefaultMatchmakingProducerConfig returns sensible defaults.
func DefaultMatchmakingProducerConfig() *MatchmakingProducerConfig {
	return &MatchmakingProducerConfig{
		TopicCommands:   getEnvOrDefault("KAFKA_TOPIC_MATCHMAKING_COMMANDS", "matchmaking.commands"),
		TopicMatches:    getEnvOrDefault("KAFKA_TOPIC_MATCHMAKING_MATCHES", "matchmaking.matches"),
		Enabled:         getEnvBoolOrDefault("KAFKA_MATCHMAKING_PRODUCER_ENABLED", true),
		MaxRetries:      getEnvIntOrDefault("KAFKA_MATCHMAKING_PRODUCER_MAX_RETRIES", 3),
		InitialBackoff:  getEnvDurationOrDefault("KAFKA_MATCHMAKING_PRODUCER_INITIAL_BACKOFF", 100*time.Millisecond),
		MaxBackoff:      getEnvDurationOrDefault("KAFKA_MATCHMAKING_PRODUCER_MAX_BACKOFF", 2*time.Second),
		ProduceTimeout:  getEnvDurationOrDefault("KAFKA_MATCHMAKING_PRODUCER_TIMEOUT", 10*time.Second),
		Source:          getEnvOrDefault("KAFKA_MATCHMAKING_SOURCE", "replay-api"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		b, _ := strconv.ParseBool(v)
		return b
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		i, _ := strconv.Atoi(v)
		return i
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
