package workout

import (
	"context"
	"errors"

	"artGallery/internal/shared/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrPlanNotFound     = errors.New("workout plan not found")
	ErrInvalidExercises = errors.New("plan contains an exercise that does not exist")
)

type PlanRepository interface {
	Create(ctx context.Context, p *WorkoutPlan) error
	FindByID(ctx context.Context, userID string, id bson.ObjectID) (*WorkoutPlan, error)
	List(ctx context.Context, userID string, q PlanQuery) ([]WorkoutPlan, int64, error)
	Update(ctx context.Context, p *WorkoutPlan) error
	Delete(ctx context.Context, userID string, id bson.ObjectID) error
}

type mongoPlanRepository struct {
	col *mongo.Collection
}

func NewMongoPlanRepository(db *mongo.Database) PlanRepository {
	return &mongoPlanRepository{col: db.Collection(database.ColWorkoutPlans)}
}

func (r *mongoPlanRepository) Create(ctx context.Context, p *WorkoutPlan) error {
	res, err := r.col.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		p.ID = id
	}
	return nil
}

// Every query includes userId, so one user can never read or change another's plan.
func (r *mongoPlanRepository) FindByID(ctx context.Context, userID string, id bson.ObjectID) (*WorkoutPlan, error) {
	var p WorkoutPlan
	err := r.col.FindOne(ctx, bson.M{"_id": id, "userId": userID}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrPlanNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *mongoPlanRepository) List(ctx context.Context, userID string, q PlanQuery) ([]WorkoutPlan, int64, error) {
	filter := bson.M{"userId": userID}
	if q.Day != "" {
		filter["scheduledDays"] = q.Day // matches if the array contains the day
	}
	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(q.Skip()).
		SetLimit(int64(q.Limit))
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var items []WorkoutPlan
	if err := cur.All(ctx, &items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *mongoPlanRepository) Update(ctx context.Context, p *WorkoutPlan) error {
	res, err := r.col.ReplaceOne(ctx, bson.M{"_id": p.ID, "userId": p.UserID}, p)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrPlanNotFound
	}
	return nil
}

func (r *mongoPlanRepository) Delete(ctx context.Context, userID string, id bson.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "userId": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrPlanNotFound
	}
	return nil
}
