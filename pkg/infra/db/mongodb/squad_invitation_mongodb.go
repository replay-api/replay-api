package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const squadInvitationCollectionName = "squad_invitations"

type SquadInvitationMongoDBRepository struct {
	collection *mongo.Collection
}

func NewSquadInvitationMongoDBRepository(db *mongo.Database) squad_out.SquadInvitationWriter {
	collection := db.Collection(squadInvitationCollectionName)

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

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Warn("Failed to create squad invitation indexes", "error", err)
	}

	return &SquadInvitationMongoDBRepository{
		collection: collection,
	}
}

// Create creates a new invitation
func (r *SquadInvitationMongoDBRepository) Create(ctx context.Context, invitation *squad_entities.SquadInvitation) (*squad_entities.SquadInvitation, error) {
	if invitation.ID == uuid.Nil {
		invitation.ID = uuid.New()
	}
	invitation.CreatedAt = time.Now()
	invitation.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, invitation)
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

	filter := bson.M{"_id": invitation.ID}
	update := bson.M{"$set": invitation}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update squad invitation", "error", err, "id", invitation.ID)
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, mongo.ErrNoDocuments
	}

	slog.InfoContext(ctx, "Squad invitation updated", "id", invitation.ID, "status", invitation.Status)
	return invitation, nil
}

// Delete deletes an invitation
func (r *SquadInvitationMongoDBRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
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
	collection *mongo.Collection
}

func NewSquadInvitationReaderRepository(db *mongo.Database) squad_out.SquadInvitationReader {
	return &SquadInvitationReaderRepository{
		collection: db.Collection(squadInvitationCollectionName),
	}
}

// GetByID gets an invitation by ID
func (r *SquadInvitationReaderRepository) GetByID(ctx context.Context, id uuid.UUID) (*squad_entities.SquadInvitation, error) {
	var invitation squad_entities.SquadInvitation
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&invitation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		slog.ErrorContext(ctx, "Failed to get invitation by ID", "error", err, "id", id)
		return nil, err
	}
	return &invitation, nil
}

// GetBySquadID gets all invitations for a squad
func (r *SquadInvitationReaderRepository) GetBySquadID(ctx context.Context, squadID uuid.UUID) ([]squad_entities.SquadInvitation, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
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
	cursor, err := r.collection.Find(ctx, bson.M{
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
	err := r.collection.FindOne(ctx, bson.M{
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

