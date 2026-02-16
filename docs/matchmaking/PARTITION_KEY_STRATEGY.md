# Matchmaking Kafka Partition Key Strategy

## Overview

This document describes the partition key strategy used by replay-api when publishing **PlayerQueued** and **MatchCompleted** events to Kafka. The strategy aligns with Epic §10 (Topic Design for Scalability) and ensures ordering and consumer parallelism.

## Partition Keys

### PlayerQueued (topic: `matchmaking.commands`)

**Partition key:** `game_id` + `region` (e.g. `"uuid-game-uuid-us-east"`)

**Rationale:**
- Commands for the same game+region are ordered together (FIFO per partition)
- Consumers (match-making-api) can process queue joins for a given game/region in order
- Parallelism: different game/region combinations are distributed across partitions
- Enables scaling: more partitions = more consumers per game/region

### MatchCompleted (topic: `matchmaking.matches`)

**Partition key:** `match_id`

**Rationale:**
- All events for a given match are ordered together
- Downstream consumers (ratings, prizes, analytics) process match results in order
- Parallelism: different matches are distributed across partitions
- Enables scaling: more partitions = more consumers per match

## Behaviour on Retry Exhaustion

When all retries fail (after `MaxRetries` attempts with exponential backoff):

1. **Error is returned to the caller** — the caller can decide whether to retry, log, or fail the request
2. **Logging:** `slog.ErrorContext` with topic, key, and error
3. **DLQ:** Optional — full DLQ implementation is in scope of #35; for now, callers should handle the error (e.g. persist to DLQ manually)

## Configuration

See `MatchmakingProducerConfig` and environment variables in `.env.example`:

- `KAFKA_TOPIC_MATCHMAKING_COMMANDS` — topic for PlayerQueued (default: `matchmaking.commands`)
- `KAFKA_TOPIC_MATCHMAKING_MATCHES` — topic for MatchCompleted (default: `matchmaking.matches`)
- `KAFKA_MATCHMAKING_PRODUCER_ENABLED` — set to `false` to disable (e.g. non-prod, tests)
- `KAFKA_MATCHMAKING_PRODUCER_MAX_RETRIES` — retry count (default: 3)
- `KAFKA_MATCHMAKING_PRODUCER_INITIAL_BACKOFF` — e.g. `100ms`
- `KAFKA_MATCHMAKING_PRODUCER_MAX_BACKOFF` — e.g. `2s`

## References

- Epic §9 — Event Payloads (PlayerQueued, MatchCompleted)
- Epic §10 — Partitioning Strategy, Strimzi Kafka Integration Design
- Issue #16 — Event Schemas
- Issue #35 — DLQ and monitoring
