package user

import (
	"artGallery/internal/shared/database"
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrNotFound      = errors.New("profile not found")
	ErrAlreadyExists = errors.New("profile already exists")
	ErrUsernameTaken = errors.New("username already taken")
)

// Repository is the only place that knows about MongoDB for this module.
type Repository interface {
	FindByClerkID(ctx context.Context, clerkUserID string) (*UserProfile, error)
	Create(ctx context.Context, p *UserProfile) error
	Update(ctx context.Context, p *UserProfile) error
}

type mongoRepository struct {
	col *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{col: db.Collection(database.ColUserProfiles)}
}

func (r *mongoRepository) FindByClerkID(ctx context.Context, clerkUserID string) (*UserProfile, error) {
	var p UserProfile
	err := r.col.FindOne(ctx, bson.M{"clerkUserId": clerkUserID}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *mongoRepository) Create(ctx context.Context, p *UserProfile) error {
	res, err := r.col.InsertOne(ctx, p)
	if err != nil {
		return mapDuplicate(err)
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		p.ID = id
	}
	return nil
}

func (r *mongoRepository) Update(ctx context.Context, p *UserProfile) error {
	res, err := r.col.ReplaceOne(ctx, bson.M{"clerkUserId": p.ClerkUserID}, p)
	if err != nil {
		return mapDuplicate(err)
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDuplicate turns a MongoDB unique-index violation into a domain error.
func mapDuplicate(err error) error {
	if !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if strings.Contains(err.Error(), "clerkUserId") {
		return ErrAlreadyExists
	}
	return ErrUsernameTaken
}
