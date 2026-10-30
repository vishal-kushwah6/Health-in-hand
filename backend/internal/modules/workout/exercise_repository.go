package workout

import (
	"context"
	"errors"
	"regexp"

	"artGallery/internal/shared/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrExerciseNotFound  = errors.New("exercise not found")
	ErrExerciseNameTaken = errors.New("you already have an exercise with this name")
	ErrReadOnlyExercise  = errors.New("built-in exercises cannot be changed")
)

type ExerciseRepository interface {
	List(ctx context.Context, userID string, q ExerciseQuery) ([]Exercise, int64, error)
	// FindVisible returns an exercise only if it is built-in or owned by userID, and not archived.
	FindVisible(ctx context.Context, userID string, id bson.ObjectID) (*Exercise, error)
	// FindVisibleByIDs: only exercises the user may use right now (built-in or own, not archived).
	FindVisibleByIDs(ctx context.Context, userID string, ids []bson.ObjectID) ([]Exercise, error)
	// FindByIDs: no visibility or archive filter. Used only to show names of exercises already in a plan.
	FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]Exercise, error)
	Create(ctx context.Context, e *Exercise) error
	Update(ctx context.Context, e *Exercise) error
	Archive(ctx context.Context, userID string, id bson.ObjectID) error
	CountGlobal(ctx context.Context) (int64, error)
	InsertMany(ctx context.Context, items []Exercise) error
}

type mongoExerciseRepository struct {
	col *mongo.Collection
}

func NewMongoExerciseRepository(db *mongo.Database) ExerciseRepository {
	return &mongoExerciseRepository{col: db.Collection(database.ColExercises)}
}

// builtIn matches library exercises (the ownerId field is absent).
var builtIn = bson.M{"ownerId": bson.M{"$exists": false}}

func visibleTo(userID string) bson.A {
	return bson.A{builtIn, bson.M{"ownerId": userID}}
}

func (r *mongoExerciseRepository) List(ctx context.Context, userID string, q ExerciseQuery) ([]Exercise, int64, error) {
	filter := bson.M{"archived": false}
	switch q.Scope {
	case "mine":
		filter["ownerId"] = userID
	case "library":
		filter["ownerId"] = bson.M{"$exists": false}
	default:
		filter["$or"] = visibleTo(userID)
	}
	if q.MuscleGroup != "" {
		filter["muscleGroups"] = q.MuscleGroup // matches if the array contains it
	}
	if q.Equipment != "" {
		filter["equipment"] = q.Equipment
	}
	if q.Difficulty != "" {
		filter["difficulty"] = q.Difficulty
	}
	if q.Search != "" {
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(q.Search), "$options": "i"}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "name", Value: 1}}).
		SetCollation(&options.Collation{Locale: "en"}).
		SetSkip(q.Skip()).
		SetLimit(int64(q.Limit))
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var items []Exercise
	if err := cur.All(ctx, &items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *mongoExerciseRepository) FindVisible(ctx context.Context, userID string, id bson.ObjectID) (*Exercise, error) {
	filter := bson.M{"_id": id, "archived": false, "$or": visibleTo(userID)}
	var e Exercise
	err := r.col.FindOne(ctx, filter).Decode(&e)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrExerciseNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *mongoExerciseRepository) Create(ctx context.Context, e *Exercise) error {
	res, err := r.col.InsertOne(ctx, e)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrExerciseNameTaken
		}
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		e.ID = id
	}
	return nil
}

func (r *mongoExerciseRepository) Update(ctx context.Context, e *Exercise) error {
	filter := bson.M{"_id": e.ID, "ownerId": e.OwnerID, "archived": false}
	res, err := r.col.ReplaceOne(ctx, filter, e)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrExerciseNameTaken
		}
		return err
	}
	if res.MatchedCount == 0 {
		return ErrExerciseNotFound
	}
	return nil
}

func (r *mongoExerciseRepository) Archive(ctx context.Context, userID string, id bson.ObjectID) error {
	filter := bson.M{"_id": id, "ownerId": userID, "archived": false}
	res, err := r.col.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"archived": true}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrExerciseNotFound
	}
	return nil
}

func (r *mongoExerciseRepository) CountGlobal(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, builtIn)
}

func (r *mongoExerciseRepository) InsertMany(ctx context.Context, items []Exercise) error {
	docs := make([]any, len(items))
	for i := range items {
		docs[i] = items[i]
	}
	_, err := r.col.InsertMany(ctx, docs)
	return err
}

func (r *mongoExerciseRepository) FindVisibleByIDs(ctx context.Context, userID string, ids []bson.ObjectID) ([]Exercise, error) {
	filter := bson.M{"_id": bson.M{"$in": ids}, "archived": false, "$or": visibleTo(userID)}
	return r.findMany(ctx, filter)
}

func (r *mongoExerciseRepository) FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]Exercise, error) {
	return r.findMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
}

func (r *mongoExerciseRepository) findMany(ctx context.Context, filter bson.M) ([]Exercise, error) {
	cur, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var out []Exercise
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
