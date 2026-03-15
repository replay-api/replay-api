package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	messaging_entities "github.com/replay-api/replay-api/pkg/domain/messaging/entities"
	messaging_out "github.com/replay-api/replay-api/pkg/domain/messaging/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionMatchComments  = "match_comments"
	CollectionDirectMessages = "direct_messages"
	CollectionTeamMessages   = "team_messages"
)

// --- Comment Repository ---

type CommentMongoRepository struct {
	collection *mongo.Collection
}

func NewCommentMongoRepository(db *mongo.Database) messaging_out.CommentRepository {
	return &CommentMongoRepository{
		collection: db.Collection(CollectionMatchComments),
	}
}

func (r *CommentMongoRepository) Save(ctx context.Context, comment *messaging_entities.Comment) error {
	_, err := r.collection.InsertOne(ctx, comment)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to insert comment", "error", err)
		return fmt.Errorf("failed to save comment: %w", err)
	}
	return nil
}

func (r *CommentMongoRepository) FindByID(ctx context.Context, id uuid.UUID) (*messaging_entities.Comment, error) {
	var comment messaging_entities.Comment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("comment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find comment: %w", err)
	}
	return &comment, nil
}

func (r *CommentMongoRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID, limit, offset int, sort string) ([]*messaging_entities.Comment, int64, error) {
	filter := bson.M{
		"match_id": matchID,
		"status":   bson.M{"$ne": messaging_entities.CommentStatusDeleted},
		"parent_id": nil, // Only top-level comments
	}

	// Sort order
	sortField := bson.D{{Key: "created_at", Value: -1}} // newest first
	if sort == "oldest" {
		sortField = bson.D{{Key: "created_at", Value: 1}}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	opts := options.Find().
		SetSort(sortField).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find comments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []*messaging_entities.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, fmt.Errorf("failed to decode comments: %w", err)
	}

	return comments, total, nil
}

func (r *CommentMongoRepository) FindReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*messaging_entities.Comment, int64, error) {
	filter := bson.M{
		"parent_id": parentID,
		"status":    bson.M{"$ne": messaging_entities.CommentStatusDeleted},
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count replies: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find replies: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []*messaging_entities.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, fmt.Errorf("failed to decode replies: %w", err)
	}

	return comments, total, nil
}

func (r *CommentMongoRepository) Update(ctx context.Context, comment *messaging_entities.Comment) error {
	comment.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": comment.ID}, comment)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}
	return nil
}

func (r *CommentMongoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}

func (r *CommentMongoRepository) IncrementReplyCount(ctx context.Context, parentID uuid.UUID, delta int) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": parentID}, bson.M{
		"$inc": bson.M{"reply_count": delta},
		"$set": bson.M{"updated_at": time.Now()},
	})
	if err != nil {
		return fmt.Errorf("failed to increment reply count: %w", err)
	}
	return nil
}

// --- Direct Message Repository ---

type DirectMessageMongoRepository struct {
	collection *mongo.Collection
}

func NewDirectMessageMongoRepository(db *mongo.Database) messaging_out.DirectMessageRepository {
	return &DirectMessageMongoRepository{
		collection: db.Collection(CollectionDirectMessages),
	}
}

func (r *DirectMessageMongoRepository) Save(ctx context.Context, message *messaging_entities.DirectMessage) error {
	_, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to insert direct message", "error", err)
		return fmt.Errorf("failed to save direct message: %w", err)
	}
	return nil
}

func (r *DirectMessageMongoRepository) FindByID(ctx context.Context, id uuid.UUID) (*messaging_entities.DirectMessage, error) {
	var dm messaging_entities.DirectMessage
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&dm)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("direct message not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find direct message: %w", err)
	}
	return &dm, nil
}

func (r *DirectMessageMongoRepository) FindByConversation(ctx context.Context, conversationID string, limit, offset int) ([]*messaging_entities.DirectMessage, int64, error) {
	filter := bson.M{
		"conversation_id": conversationID,
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []*messaging_entities.DirectMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, 0, fmt.Errorf("failed to decode messages: %w", err)
	}

	return messages, total, nil
}

func (r *DirectMessageMongoRepository) ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*messaging_entities.Conversation, error) {
	// Aggregate to get the latest message per conversation for this user
	pipeline := mongo.Pipeline{
		// Match messages involving this user
		{{Key: "$match", Value: bson.M{
			"$or": bson.A{
				bson.M{"sender_id": userID},
				bson.M{"recipient_id": userID},
			},
		}}},
		// Sort by created_at desc to get latest messages first
		{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		// Group by conversation_id
		{{Key: "$group", Value: bson.M{
			"_id":          "$conversation_id",
			"last_message": bson.M{"$first": "$content"},
			"last_at":      bson.M{"$first": "$created_at"},
			"sender_id":    bson.M{"$first": "$sender_id"},
			"recipient_id": bson.M{"$first": "$recipient_id"},
		}}},
		// Sort by last message time
		{{Key: "$sort", Value: bson.D{{Key: "last_at", Value: -1}}}},
		{{Key: "$skip", Value: int64(offset)}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		ConversationID string    `bson:"_id"`
		LastMessage    string    `bson:"last_message"`
		LastAt         time.Time `bson:"last_at"`
		SenderID       uuid.UUID `bson:"sender_id"`
		RecipientID    uuid.UUID `bson:"recipient_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode conversations: %w", err)
	}

	conversations := make([]*messaging_entities.Conversation, 0, len(results))
	for _, res := range results {
		// Determine the other participant
		otherUserID := res.RecipientID
		if res.SenderID != userID {
			otherUserID = res.SenderID
		}

		// Count unread messages in this conversation
		unreadCount, _ := r.collection.CountDocuments(ctx, bson.M{
			"conversation_id": res.ConversationID,
			"recipient_id":    userID,
			"read_at":         nil,
		})

		conversations = append(conversations, &messaging_entities.Conversation{
			ConversationID: res.ConversationID,
			Participant: messaging_entities.AuthorSummary{
				ID: otherUserID,
				// DisplayName and AvatarURL will be populated by the controller/service layer
			},
			LastMessage:   res.LastMessage,
			LastMessageAt: res.LastAt,
			UnreadCount:   int(unreadCount),
		})
	}

	return conversations, nil
}

func (r *DirectMessageMongoRepository) MarkConversationRead(ctx context.Context, conversationID string, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.collection.UpdateMany(ctx,
		bson.M{
			"conversation_id": conversationID,
			"recipient_id":    userID,
			"read_at":         nil,
		},
		bson.M{
			"$set": bson.M{"read_at": now, "updated_at": now},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to mark conversation read: %w", err)
	}
	return nil
}

func (r *DirectMessageMongoRepository) Update(ctx context.Context, message *messaging_entities.DirectMessage) error {
	message.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": message.ID}, message)
	if err != nil {
		return fmt.Errorf("failed to update direct message: %w", err)
	}
	return nil
}

func (r *DirectMessageMongoRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"recipient_id": userID,
		"read_at":      nil,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}
	return count, nil
}

// --- Team Message Repository ---

type TeamMessageMongoRepository struct {
	collection *mongo.Collection
}

func NewTeamMessageMongoRepository(db *mongo.Database) messaging_out.TeamMessageRepository {
	return &TeamMessageMongoRepository{
		collection: db.Collection(CollectionTeamMessages),
	}
}

func (r *TeamMessageMongoRepository) Save(ctx context.Context, message *messaging_entities.TeamMessage) error {
	_, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to insert team message", "error", err)
		return fmt.Errorf("failed to save team message: %w", err)
	}
	return nil
}

func (r *TeamMessageMongoRepository) FindByID(ctx context.Context, id uuid.UUID) (*messaging_entities.TeamMessage, error) {
	var msg messaging_entities.TeamMessage
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&msg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("team message not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find team message: %w", err)
	}
	return &msg, nil
}

func (r *TeamMessageMongoRepository) FindByTeamAndChannel(ctx context.Context, teamID uuid.UUID, channel messaging_entities.ChannelType, limit, offset int) ([]*messaging_entities.TeamMessage, int64, error) {
	filter := bson.M{"team_id": teamID}
	if channel != "" {
		filter["channel"] = channel
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count team messages: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find team messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []*messaging_entities.TeamMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, 0, fmt.Errorf("failed to decode team messages: %w", err)
	}

	return messages, total, nil
}

func (r *TeamMessageMongoRepository) ListTeamChannels(ctx context.Context, teamID uuid.UUID) ([]*messaging_entities.TeamChannelSummary, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"team_id": teamID}}},
		{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		{{Key: "$group", Value: bson.M{
			"_id":          "$channel",
			"last_message": bson.M{"$first": "$content"},
			"last_at":      bson.M{"$first": "$created_at"},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "last_at", Value: -1}}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to list team channels: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		Channel     string    `bson:"_id"`
		LastMessage string    `bson:"last_message"`
		LastAt      time.Time `bson:"last_at"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode team channels: %w", err)
	}

	summaries := make([]*messaging_entities.TeamChannelSummary, 0, len(results))
	for _, res := range results {
		summaries = append(summaries, &messaging_entities.TeamChannelSummary{
			TeamID:      teamID,
			Channel:     messaging_entities.ChannelType(res.Channel),
			LastMessage: res.LastMessage,
			LastAt:      res.LastAt,
		})
	}

	return summaries, nil
}
