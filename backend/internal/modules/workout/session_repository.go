package workout

import (
	"context"
	"errors"
	"time"

	"artGallery/internal/shared/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrSessionNotFound = errors.New("workout session not found")
	ErrInvalidPlan     = errors.New("workout plan does not exist")
	ErrInvalidDate     = errors.New("date cannot be in the future")
)

type SessionRepository interface {
	Create(ctx context.Context, s *WorkoutSession) error
	FindByID(ctx context.Context, userID string, id bson.ObjectID) (*WorkoutSession, error)
	List(ctx context.Context, userID string, q SessionQuery) ([]WorkoutSession, int64, error)
	Delete(ctx context.Context, userID string, id bson.ObjectID) error
	// AllCompleted returns the user's completed sessions, oldest first (capped at 3000 as a safety limit).
	AllCompleted(ctx context.Context, userID string) ([]WorkoutSession, error)
	// LastSets returns, per exercise id (hex), the completed sets of its most recent session.
	LastSets(ctx context.Context, userID string, ids []bson.ObjectID) (map[string][]SetLog, error)
}

type mongoSessionRepository struct {
	col *mongo.Collection
}

func NewMongoSessionRepository(db *mongo.Database) SessionRepository {
	return &mongoSessionRepository{col: db.Collection(database.ColWorkoutSessions)}
}

func (r *mongoSessionRepository) Create(ctx context.Context, s *WorkoutSession) error {
	res, err := r.col.InsertOne(ctx, s)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		s.ID = id
	}
	return nil
}

func (r *mongoSessionRepository) FindByID(ctx context.Context, userID string, id bson.ObjectID) (*WorkoutSession, error) {
	var s WorkoutSession
	err := r.col.FindOne(ctx, bson.M{"_id": id, "userId": userID}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *mongoSessionRepository) List(ctx context.Context, userID string, q SessionQuery) ([]WorkoutSession, int64, error) {
	filter := bson.M{"userId": userID}

	// from / to are YYYY-MM-DD (already validated); "to" is inclusive of that whole day.
	dateRange := bson.M{}
	if q.From != "" {
		if t, err := time.Parse("2006-01-02", q.From); err == nil {
			dateRange["$gte"] = t
		}
	}
	if q.To != "" {
		if t, err := time.Parse("2006-01-02", q.To); err == nil {
			dateRange["$lt"] = t.AddDate(0, 0, 1)
		}
	}
	if len(dateRange) > 0 {
		filter["date"] = dateRange
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(q.Skip()).
		SetLimit(int64(q.Limit))
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var items []WorkoutSession
	if err := cur.All(ctx, &items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *mongoSessionRepository) Delete(ctx context.Context, userID string, id bson.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "userId": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *mongoSessionRepository) LastSets(ctx context.Context, userID string, ids []bson.ObjectID) (map[string][]SetLog, error) {
	out := make(map[string][]SetLog, len(ids))
	for _, id := range ids {
		var s WorkoutSession
		err := r.col.FindOne(ctx,
			bson.M{"userId": userID, "exercises.exerciseId": id},
			options.FindOne().SetSort(bson.D{{Key: "date", Value: -1}}),
		).Decode(&s)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, ex := range s.Exercises {
			if ex.ExerciseID != id {
				continue
			}
			var done []SetLog
			for _, st := range ex.Sets {
				if st.Completed {
					done = append(done, st)
				}
			}
			if len(done) > 0 {
				out[id.Hex()] = done
			}
			break
		}
	}
	return out, nil
}

func (r *mongoSessionRepository) AllCompleted(ctx context.Context, userID string) ([]WorkoutSession, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: 1}, {Key: "_id", Value: 1}}).
		SetLimit(3000)
	cur, err := r.col.Find(ctx, bson.M{"userId": userID, "completed": true}, opts)
	if err != nil {
		return nil, err
	}
	var items []WorkoutSession
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
