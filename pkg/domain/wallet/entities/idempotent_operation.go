package wallet_entities

import (
	"time"

	"github.com/google/uuid"
)

// IdempotentOperation tracks operations to prevent duplicate executions
// Uses idempotency key as primary identifier to guarantee exactly-once semantics
// Automatically expires after 24 hours via MongoDB TTL index
type IdempotentOperation struct {
	Key             string          `bson:"_id" json:"key"` // Idempotency key (primary key)
	OperationType   string          `bson:"operation_type" json:"operation_type"`
	Status          OperationStatus `bson:"status" json:"status"`
	RequestPayload  interface{}     `bson:"request_payload" json:"request_payload"`
	ResponsePayload interface{}     `bson:"response_payload,omitempty" json:"response_payload,omitempty"`
	ResultID        *uuid.UUID      `bson:"result_id,omitempty" json:"result_id,omitempty"` // Transaction or Entry ID
	ErrorMessage    string          `bson:"error_message,omitempty" json:"error_message,omitempty"`
	CreatedAt       time.Time       `bson:"created_at" json:"created_at"`
	CompletedAt     *time.Time      `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	ExpiresAt       time.Time       `bson:"expires_at" json:"expires_at"` // For TTL index
	AttemptCount    int             `bson:"attempt_count" json:"attempt_count"`
	LastAttemptAt   *time.Time      `bson:"last_attempt_at,omitempty" json:"last_attempt_at,omitempty"`
}

// OperationStatus indicates the current state of the operation
type OperationStatus string

const (
	OperationStatusProcessing OperationStatus = "Processing" // Currently executing
	OperationStatusCompleted  OperationStatus = "Completed"  // Successfully completed
	OperationStatusFailed     OperationStatus = "Failed"     // Failed (can be retried)
)

// NewIdempotentOperation creates a new idempotent operation tracker
func NewIdempotentOperation(
	idempotencyKey string,
	operationType string,
	requestPayload interface{},
) *IdempotentOperation {
	now := time.Now().UTC()

	return &IdempotentOperation{
		Key:            idempotencyKey,
		OperationType:  operationType,
		Status:         OperationStatusProcessing,
		RequestPayload: requestPayload,
		CreatedAt:      now,
		ExpiresAt:      now.Add(24 * time.Hour), // Auto-cleanup after 24 hours
		AttemptCount:   1,
		LastAttemptAt:  &now,
	}
}

// MarkCompleted marks the operation as successfully completed
func (o *IdempotentOperation) MarkCompleted(resultID uuid.UUID, responsePayload interface{}) {
	now := time.Now().UTC()
	o.Status = OperationStatusCompleted
	o.ResultID = &resultID
	o.ResponsePayload = responsePayload
	o.CompletedAt = &now
}

// MarkFailed marks the operation as failed
func (o *IdempotentOperation) MarkFailed(errorMessage string) {
	o.Status = OperationStatusFailed
	o.ErrorMessage = errorMessage
}

// IncrementAttempt increments the attempt counter
func (o *IdempotentOperation) IncrementAttempt() {
	now := time.Now().UTC()
	o.AttemptCount++
	o.LastAttemptAt = &now
}

// IsProcessing returns true if the operation is currently processing
func (o *IdempotentOperation) IsProcessing() bool {
	return o.Status == OperationStatusProcessing
}

// IsCompleted returns true if the operation completed successfully
func (o *IdempotentOperation) IsCompleted() bool {
	return o.Status == OperationStatusCompleted
}

// IsFailed returns true if the operation failed
func (o *IdempotentOperation) IsFailed() bool {
	return o.Status == OperationStatusFailed
}

// CanRetry determines if a failed operation can be retried
func (o *IdempotentOperation) CanRetry(maxAttempts int) bool {
	return o.IsFailed() && o.AttemptCount < maxAttempts
}

// IsStale checks if the operation has been processing for too long
// Returns true if started > threshold ago and still processing
func (o *IdempotentOperation) IsStale(threshold time.Duration) bool {
	if !o.IsProcessing() {
		return false
	}

	if o.LastAttemptAt == nil {
		return time.Since(o.CreatedAt) > threshold
	}

	return time.Since(*o.LastAttemptAt) > threshold
}

// GetElapsedTime returns how long the operation has been running
func (o *IdempotentOperation) GetElapsedTime() time.Duration {
	if o.CompletedAt != nil {
		return o.CompletedAt.Sub(o.CreatedAt)
	}
	return time.Since(o.CreatedAt)
}
