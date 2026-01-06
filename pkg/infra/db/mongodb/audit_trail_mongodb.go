package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

// AuditTrailMongoDBRepository implements bank-grade audit trail persistence
type AuditTrailMongoDBRepository struct {
	mongodb.MongoDBRepository[*billing_entities.AuditTrailEntry]
}

// NewAuditTrailMongoDBRepository creates a new audit trail repository
func NewAuditTrailMongoDBRepository(client *mongo.Client, dbName string) *AuditTrailMongoDBRepository {
	repo := mongodb.NewMongoDBRepository[*billing_entities.AuditTrailEntry](client, dbName, &billing_entities.AuditTrailEntry{}, "audit_trail", "AuditTrailEntry")

	repo.InitQueryableFields(map[string]bool{
		"ID":             true,
		"EventType":      true,
		"Severity":       true,
		"Timestamp":      true,
		"ActorUserID":    true,
		"ActorType":      true,
		"ActorIP":        true,
		"ActorUserAgent": true,
		"ActorSessionID": true,
		"TargetUserID":   true,
		"TargetType":     true,
		"TargetID":       true,
		"Amount":         true,
		"Currency":       true,
		"BalanceBefore":  true,
		"BalanceAfter":   true,
		"TransactionID":  true,
		"ExternalRef":    true,
		"ProviderRef":    true,
		"Description":    true,
		"PreviousEntryID": true,
		"Hash":           true,
		"RetentionUntil": true,
		"Exportable":     true,
	}, map[string]string{
		"ID":             "_id",
		"EventType":      "event_type",
		"Severity":       "severity",
		"Timestamp":      "timestamp",
		"ActorUserID":    "actor_user_id",
		"ActorType":      "actor_type",
		"ActorIP":        "actor_ip",
		"ActorUserAgent": "actor_user_agent",
		"ActorSessionID": "actor_session_id",
		"TargetUserID":   "target_user_id",
		"TargetType":     "target_type",
		"TargetID":       "target_id",
		"Amount":         "amount",
		"Currency":       "currency",
		"BalanceBefore":  "balance_before",
		"BalanceAfter":   "balance_after",
		"TransactionID":  "transaction_id",
		"ExternalRef":    "external_ref",
		"ProviderRef":    "provider_ref",
		"Description":    "description",
		"PreviousEntryID": "previous_entry_id",
		"Hash":           "hash",
		"RetentionUntil": "retention_until",
		"Exportable":     "exportable",
	})

	auditRepo := &AuditTrailMongoDBRepository{
		MongoDBRepository: *repo,
	}

	// Create indexes for compliance queries
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		// Primary lookup by ID
		{
			Keys:    bson.D{{Key: "_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// User-based queries (most common for audit history)
		{
			Keys:    bson.D{{Key: "actor_user_id", Value: 1}, {Key: "timestamp", Value: -1}},
			Options: options.Index().SetName("idx_actor_user_timestamp"),
		},
		// Target-based queries (for transaction trails)
		{
			Keys:    bson.D{{Key: "target_type", Value: 1}, {Key: "target_id", Value: 1}, {Key: "timestamp", Value: -1}},
			Options: options.Index().SetName("idx_target_timestamp"),
		},
		// Event type queries (for compliance reports)
		{
			Keys:    bson.D{{Key: "event_type", Value: 1}, {Key: "timestamp", Value: -1}},
			Options: options.Index().SetName("idx_event_type_timestamp"),
		},
		// Severity queries (for alerting)
		{
			Keys:    bson.D{{Key: "severity", Value: 1}, {Key: "timestamp", Value: -1}},
			Options: options.Index().SetName("idx_severity_timestamp"),
		},
		// Transaction linking
		{
			Keys:    bson.D{{Key: "transaction_id", Value: 1}},
			Options: options.Index().SetName("idx_transaction_id").SetSparse(true),
		},
		// Time-based queries (for reports and archival)
		{
			Keys:    bson.D{{Key: "timestamp", Value: -1}},
			Options: options.Index().SetName("idx_timestamp"),
		},
		// Tenant isolation
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "timestamp", Value: -1}},
			Options: options.Index().SetName("idx_tenant_timestamp"),
		},
		// Chain verification (hash chain integrity)
		{
			Keys:    bson.D{{Key: "target_type", Value: 1}, {Key: "target_id", Value: 1}, {Key: "_id", Value: 1}},
			Options: options.Index().SetName("idx_chain_verification"),
		},
		// Retention management
		{
			Keys:    bson.D{{Key: "retention_until", Value: 1}},
			Options: options.Index().SetName("idx_retention").SetExpireAfterSeconds(0),
		},
	}

	_, err := auditRepo.MongoDBRepository.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Warn("Failed to create audit trail indexes", "error", err)
	}

	return auditRepo
}

// Create persists a new audit entry (append-only)
func (r *AuditTrailMongoDBRepository) Create(ctx context.Context, entry *billing_entities.AuditTrailEntry) error {
	_, err := r.MongoDBRepository.Collection().InsertOne(ctx, entry)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create audit entry",
			"error", err,
			"audit_id", entry.ID,
			"event_type", entry.EventType,
		)
		return err
	}

	slog.DebugContext(ctx, "Audit entry created",
		"audit_id", entry.ID,
		"event_type", entry.EventType,
	)

	return nil
}

// GetByID retrieves a specific audit entry
func (r *AuditTrailMongoDBRepository) GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.AuditTrailEntry, error) {
	filter := bson.M{"_id": id}
	var entry billing_entities.AuditTrailEntry

	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &entry, nil
}

// GetLatestForTarget gets the most recent audit entry for a target
func (r *AuditTrailMongoDBRepository) GetLatestForTarget(ctx context.Context, targetType string, targetID uuid.UUID) (*billing_entities.AuditTrailEntry, error) {
	filter := bson.M{
		"target_type": targetType,
		"target_id":   targetID,
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var entry billing_entities.AuditTrailEntry
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter, opts).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &entry, nil
}

// GetByUser retrieves audit entries for a specific user
func (r *AuditTrailMongoDBRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]billing_entities.AuditTrailEntry, error) {
	filter := bson.M{"actor_user_id": userID}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	return r.findMany(ctx, filter, opts)
}

// GetByTarget retrieves audit entries for a specific target entity
func (r *AuditTrailMongoDBRepository) GetByTarget(ctx context.Context, targetType string, targetID uuid.UUID, limit, offset int) ([]billing_entities.AuditTrailEntry, error) {
	filter := bson.M{
		"target_type": targetType,
		"target_id":   targetID,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	return r.findMany(ctx, filter, opts)
}

// GetByEventType retrieves entries of a specific event type
func (r *AuditTrailMongoDBRepository) GetByEventType(ctx context.Context, eventType billing_entities.AuditEventType, from, to time.Time, limit, offset int) ([]billing_entities.AuditTrailEntry, error) {
	filter := bson.M{
		"event_type": eventType,
		"timestamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	return r.findMany(ctx, filter, opts)
}

// GetBySeverity retrieves entries at or above a severity level
func (r *AuditTrailMongoDBRepository) GetBySeverity(ctx context.Context, severity billing_entities.AuditSeverity, from, to time.Time, limit, offset int) ([]billing_entities.AuditTrailEntry, error) {
	// Order severities by importance
	severityOrder := map[billing_entities.AuditSeverity]int{
		billing_entities.AuditSeverityInfo:     1,
		billing_entities.AuditSeverityWarning:  2,
		billing_entities.AuditSeverityCritical: 3,
		billing_entities.AuditSeverityAlert:    4,
	}

	minSeverity := severityOrder[severity]
	var severities []billing_entities.AuditSeverity
	for sev, order := range severityOrder {
		if order >= minSeverity {
			severities = append(severities, sev)
		}
	}

	filter := bson.M{
		"severity": bson.M{"$in": severities},
		"timestamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	return r.findMany(ctx, filter, opts)
}

// GetChainForVerification retrieves entries for hash chain verification
func (r *AuditTrailMongoDBRepository) GetChainForVerification(ctx context.Context, targetType string, targetID uuid.UUID, from, to time.Time) ([]billing_entities.AuditTrailEntry, error) {
	filter := bson.M{
		"target_type": targetType,
		"target_id":   targetID,
		"timestamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}) // Ordered by creation for chain verification

	return r.findMany(ctx, filter, opts)
}

// GetForComplianceReport retrieves entries for compliance reporting
func (r *AuditTrailMongoDBRepository) GetForComplianceReport(ctx context.Context, from, to time.Time) ([]billing_entities.AuditTrailEntry, error) {
	filter := bson.M{
		"timestamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
		"exportable": true,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: 1}})

	return r.findMany(ctx, filter, opts)
}

// CountByType counts entries by event type within a period
func (r *AuditTrailMongoDBRepository) CountByType(ctx context.Context, eventType billing_entities.AuditEventType, from, to time.Time) (int64, error) {
	filter := bson.M{
		"event_type": eventType,
		"timestamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	return r.MongoDBRepository.Collection().CountDocuments(ctx, filter)
}

// GetFinancialSummary retrieves aggregated financial audit data
func (r *AuditTrailMongoDBRepository) GetFinancialSummary(ctx context.Context, userID uuid.UUID, from, to time.Time) (*billing_entities.AuditSummary, error) {
	pipeline := mongo.Pipeline{
		// Match user and time range
		{{Key: "$match", Value: bson.M{
			"actor_user_id": userID,
			"timestamp": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}}},
		// Group and aggregate
		{{Key: "$facet", Value: bson.M{
			"totals": []bson.M{
				{"$group": bson.M{
					"_id":         nil,
					"total":       bson.M{"$sum": 1},
					"deposits":    bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$event_type", "DEPOSIT"}}, "$amount", 0}}},
					"withdrawals": bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$event_type", "WITHDRAWAL"}}, "$amount", 0}}},
					"unique_ips":  bson.M{"$addToSet": "$actor_ip"},
				}},
			},
			"by_type": []bson.M{
				{"$group": bson.M{
					"_id":   "$event_type",
					"count": bson.M{"$sum": 1},
				}},
			},
			"by_severity": []bson.M{
				{"$group": bson.M{
					"_id":   "$severity",
					"count": bson.M{"$sum": 1},
				}},
			},
			"failed_logins": []bson.M{
				{"$match": bson.M{"event_type": "LOGIN_FAILED"}},
				{"$count": "count"},
			},
		}}},
	}

	cursor, err := r.MongoDBRepository.Collection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Parse results
	summary := &billing_entities.AuditSummary{
		UserID:           userID,
		StartDate:        from,
		EndDate:          to,
		GeneratedAt:      time.Now().UTC(),
		EventsByType:     make(map[billing_entities.AuditEventType]int),
		EventsBySeverity: make(map[billing_entities.AuditSeverity]int),
	}

	if cursor.Next(ctx) {
		var result struct {
			Totals []struct {
				Total       int      `bson:"total"`
				Deposits    float64  `bson:"deposits"`
				Withdrawals float64  `bson:"withdrawals"`
				UniqueIPs   []string `bson:"unique_ips"`
			} `bson:"totals"`
			ByType []struct {
				ID    billing_entities.AuditEventType `bson:"_id"`
				Count int                              `bson:"count"`
			} `bson:"by_type"`
			BySeverity []struct {
				ID    billing_entities.AuditSeverity `bson:"_id"`
				Count int                             `bson:"count"`
			} `bson:"by_severity"`
			FailedLogins []struct {
				Count int `bson:"count"`
			} `bson:"failed_logins"`
		}

		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		if len(result.Totals) > 0 {
			summary.TotalEvents = result.Totals[0].Total
			summary.TotalDeposits = result.Totals[0].Deposits
			summary.TotalWithdrawals = result.Totals[0].Withdrawals
			summary.UniqueIPs = len(result.Totals[0].UniqueIPs)
			summary.NetChange = result.Totals[0].Deposits - result.Totals[0].Withdrawals
		}

		for _, t := range result.ByType {
			summary.EventsByType[t.ID] = t.Count
		}

		for _, s := range result.BySeverity {
			summary.EventsBySeverity[s.ID] = s.Count
		}

		if len(result.FailedLogins) > 0 {
			summary.FailedLogins = result.FailedLogins[0].Count
		}
	}

	return summary, nil
}

// ArchiveOldEntries moves entries past retention to cold storage
func (r *AuditTrailMongoDBRepository) ArchiveOldEntries(ctx context.Context, before time.Time) (int64, error) {
	// In production, this would move to cold storage (S3, etc.)
	// For now, we mark as anonymized for GDPR compliance
	filter := bson.M{
		"retention_until": bson.M{"$lt": before},
		"anonymized":      false,
	}

	update := bson.M{
		"$set": bson.M{
			"anonymized":      true,
			"actor_ip":        "ARCHIVED",
			"actor_user_agent": "ARCHIVED",
			"metadata":        nil,
		},
	}

	result, err := r.MongoDBRepository.Collection().UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	slog.Info("Archived old audit entries",
		"count", result.ModifiedCount,
		"before", before,
	)

	return result.ModifiedCount, nil
}

// findMany is a helper for finding multiple entries
func (r *AuditTrailMongoDBRepository) findMany(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]billing_entities.AuditTrailEntry, error) {
	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []billing_entities.AuditTrailEntry
	if err = cursor.All(ctx, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

