package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const squadInvitationCollectionName = "squad_invitations"

type SquadInvitationMongoDBRepository struct {
	mongodb.MongoDBRepository[squad_entities.SquadInvitation]
}

func NewSquadInvitationMongoDBRepository(client *mongo.Client, dbName string) squad_out.SquadInvitationWriter {
	repo := mongodb.NewMongoDBRepository[squad_entities.SquadInvitation](client, dbName, squad_entities.SquadInvitation{}, squadInvitationCollectionName, "SquadInvitation")

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"SquadID":         true,
		"SquadName":       true,
		"PlayerProfileID": true,
		"PlayerName":      true,
		"InviterID":       true,
		"InviterName":     true,
		"InvitationType":  true,
		"Status":          true,
		"Role":            true,
		"Message":         true,
		"ExpiresAt":       true,
		"RespondedAt":     true,
		"ResourceOwner":   true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}, map[string]string{
		"ID":              "_id",
		"SquadID":         "squad_id",
		"SquadName":       "squad_name",
		"PlayerProfileID": "player_profile_id",
		"PlayerName":      "player_name",
		"InviterID":       "inviter_id",
		"InviterName":     "inviter_name",
		"InvitationType":  "invitation_type",
		"Status":          "status",
		"Role":            "role",
		"Message":         "message",
		"ExpiresAt":       "expires_at",
		"RespondedAt":     "responded_at",
		"ResourceOwner":   "resource_owner",
		"CreatedAt":       "created_at",
		"UpdatedAt":       "updated_at",
	})

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "squad_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "player_profile_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "squad_id", Value: 1}, {Key: "player_profile_id", Value: 1}, {Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0), // TTL index
		},
	}

	_, err := repo.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Warn("Failed to create squad invitation indexes", "error", err)
	}

	return &SquadInvitationMongoDBRepository{
		MongoDBRepository: *repo,
	}
}

// Create creates a new invitation
func (r *SquadInvitationMongoDBRepository) Create(ctx context.Context, invitation *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error) {
	if invitation.ID == uuid.Nil {
		invitation.ID = uuid.New()
	}
	invitation.CreatedAt = time.Now()
	invitation.UpdatedAt = time.Now()

	_, err := r.MongoDBRepository.Create(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create squad invitation", "error", err, "squad_id", invitation.SquadID, "player_id", invitation.PlayerProfileID)
		return nil, err
	}

	slog.InfoContext(ctx, "Squad invitation created", "id", invitation.ID, "squad_id", invitation.SquadID, "player_id", invitation.PlayerProfileID)
	return invitation, nil
}

// Update updates an existing invitation
func (r *SquadInvitationMongoDBRepository) Update(ctx context.Context, invitation *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error) {
	invitation.UpdatedAt = time.Now()

	_, err := r.MongoDBRepository.Update(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update squad invitation", "error", err, "id", invitation.ID)
		return nil, err
	}

	slog.InfoContext(ctx, "Squad invitation updated", "id", invitation.ID, "status", invitation.Status)
	return invitation, nil
}

// Delete deletes an invitation
func (r *SquadInvitationMongoDBRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.MongoDBRepository.DeleteOneWithRLS(ctx, bson.M{"_id": id})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete squad invitation", "error", err, "id", id)
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	slog.InfoContext(ctx, "Squad invitation deleted", "id", id)
	return nil
}

// SquadInvitationReaderRepository provides read operations
type SquadInvitationReaderRepository struct {
	mongodb.MongoDBRepository[squad_entities.SquadInvitation]
}

func NewSquadInvitationReaderRepository(client *mongo.Client, dbName string) squad_out.SquadInvitationReader {
	repo := mongodb.NewMongoDBRepository[squad_entities.SquadInvitation](client, dbName, squad_entities.SquadInvitation{}, squadInvitationCollectionName, "SquadInvitation")
	return &SquadInvitationReaderRepository{
		MongoDBRepository: *repo,
	}
}

// GetByID gets an invitation by ID
func (r *SquadInvitationReaderRepository) GetByID(ctx context.Context, id uuid.UUID) (*squad_entities.SquadInvitation, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

// GetBySquadID gets all invitations for a squad
func (r *SquadInvitationReaderRepository) GetBySquadID(ctx context.Context, squadID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
	cursor, err := r.MongoDBRepository.FindWithRLS(ctx, bson.M{
		"squad_id": squadID,
		"status":   squad_entities.InvitationStatusPending,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get invitations by squad ID", "error", err, "squad_id", squadID)
		return nil, err
	}
	defer cursor.Close(ctx)

	var invitations []squad_entities.SquadInvitation
	if err := cursor.All(ctx, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

// GetByPlayerID gets all invitations for a player
func (r *SquadInvitationReaderRepository) GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
	cursor, err := r.MongoDBRepository.FindWithRLS(ctx, bson.M{
		"player_profile_id": playerID,
		"status":            squad_entities.InvitationStatusPending,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get invitations by player ID", "error", err, "player_id", playerID)
		return nil, err
	}
	defer cursor.Close(ctx)

	var invitations []squad_entities.SquadInvitation
	if err := cursor.All(ctx, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

// GetPendingBySquadAndPlayer gets a pending invitation between a squad and player
func (r *SquadInvitationReaderRepository) GetPendingBySquadAndPlayer(ctx context.Context, squadID, playerID uuid.UUID) (*squad_entities.SquadInvitation, error) {
	var invitation squad_entities.SquadInvitation
	err := r.MongoDBRepository.FindOneWithRLS(ctx, bson.M{
		"squad_id":          squadID,
		"player_profile_id": playerID,
		"status":            squad_entities.InvitationStatusPending,
		"expires_at":        bson.M{"$gt": time.Now()},
	}).Decode(&invitation)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &invitation, nil
}

