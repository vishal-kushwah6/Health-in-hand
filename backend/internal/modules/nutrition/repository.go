package nutrition

import (
	"context"
	"time"

	"artGallery/internal/shared/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	Save(ctx context.Context, msgs ...*Message) error
	// Recent returns the user's latest messages, OLDEST first.
	Recent(ctx context.Context, userID string, limit int64) ([]Message, error)
	CountUserMessagesSince(ctx context.Context, userID string, since time.Time) (int64, error)
	Clear(ctx context.Context, userID string) error
}

type mongoRepository struct {
	col *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{col: db.Collection(database.ColNutritionMessages)}
}

func (r *mongoRepository) Save(ctx context.Context, msgs ...*Message) error {
	docs := make([]any, len(msgs))
	for i, m := range msgs {
		docs[i] = m
	}
	res, err := r.col.InsertMany(ctx, docs)
	if err != nil {
		return err
	}
	for i, id := range res.InsertedIDs {
		if oid, ok := id.(bson.ObjectID); ok {
			msgs[i].ID = oid
		}
	}
	return nil
}

func (r *mongoRepository) Recent(ctx context.Context, userID string, limit int64) ([]Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(limit)
	cur, err := r.col.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, err
	}
	var out []Message
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 { // newest-first -> oldest-first
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (r *mongoRepository) CountUserMessagesSince(ctx context.Context, userID string, since time.Time) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{"userId": userID, "role": RoleUser, "createdAt": bson.M{"$gte": since}})
}

func (r *mongoRepository) Clear(ctx context.Context, userID string) error {
	_, err := r.col.DeleteMany(ctx, bson.M{"userId": userID})
	return err
}
