package community

import (
	"context"
	"errors"

	"artGallery/internal/shared/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrPostNotFound = errors.New("post not found")

// Field names of the counters kept on a post.
const (
	fieldLikeCount    = "likeCount"
	fieldCommentCount = "commentCount"
)

type PostRepository interface {
	CreatePost(ctx context.Context, p *Post) error
	FindPost(ctx context.Context, id bson.ObjectID) (*Post, error)
	ListPosts(ctx context.Context, communityIDs []bson.ObjectID, skip, limit int64) ([]Post, int64, error)
	DeletePost(ctx context.Context, authorID string, id bson.ObjectID) error
	IncPost(ctx context.Context, id bson.ObjectID, field string, delta int) error

	AddLike(ctx context.Context, l *PostLike) error
	RemoveLike(ctx context.Context, postID bson.ObjectID, userID string) (bool, error)
	LikedBy(ctx context.Context, userID string, postIDs []bson.ObjectID) (map[bson.ObjectID]bool, error)
	DeleteLikesOf(ctx context.Context, postID bson.ObjectID) error

	AddComment(ctx context.Context, c *Comment) error
	ListComments(ctx context.Context, postID bson.ObjectID, skip, limit int64) ([]Comment, int64, error)
	DeleteCommentsOf(ctx context.Context, postID bson.ObjectID) error
}

type mongoPostRepository struct {
	posts    *mongo.Collection
	likes    *mongo.Collection
	comments *mongo.Collection
}

func NewMongoPostRepository(db *mongo.Database) PostRepository {
	return &mongoPostRepository{
		posts:    db.Collection(database.ColPosts),
		likes:    db.Collection(database.ColPostLikes),
		comments: db.Collection(database.ColComments),
	}
}

func (r *mongoPostRepository) CreatePost(ctx context.Context, p *Post) error {
	res, err := r.posts.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		p.ID = id
	}
	return nil
}

func (r *mongoPostRepository) FindPost(ctx context.Context, id bson.ObjectID) (*Post, error) {
	var p Post
	err := r.posts.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *mongoPostRepository) ListPosts(ctx context.Context, communityIDs []bson.ObjectID, skip, limit int64) ([]Post, int64, error) {
	filter := bson.M{"communityId": bson.M{"$in": communityIDs}}
	total, err := r.posts.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)
	cur, err := r.posts.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var out []Post
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *mongoPostRepository) DeletePost(ctx context.Context, authorID string, id bson.ObjectID) error {
	res, err := r.posts.DeleteOne(ctx, bson.M{"_id": id, "authorId": authorID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *mongoPostRepository) IncPost(ctx context.Context, id bson.ObjectID, field string, delta int) error {
	res, err := r.posts.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{field: delta}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *mongoPostRepository) AddLike(ctx context.Context, l *PostLike) error {
	_, err := r.likes.InsertOne(ctx, l)
	return err
}

func (r *mongoPostRepository) RemoveLike(ctx context.Context, postID bson.ObjectID, userID string) (bool, error) {
	res, err := r.likes.DeleteOne(ctx, bson.M{"postId": postID, "userId": userID})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

func (r *mongoPostRepository) LikedBy(ctx context.Context, userID string, postIDs []bson.ObjectID) (map[bson.ObjectID]bool, error) {
	cur, err := r.likes.Find(ctx, bson.M{"userId": userID, "postId": bson.M{"$in": postIDs}})
	if err != nil {
		return nil, err
	}
	var rows []PostLike
	if err := cur.All(ctx, &rows); err != nil {
		return nil, err
	}
	out := make(map[bson.ObjectID]bool, len(rows))
	for _, l := range rows {
		out[l.PostID] = true
	}
	return out, nil
}

func (r *mongoPostRepository) DeleteLikesOf(ctx context.Context, postID bson.ObjectID) error {
	_, err := r.likes.DeleteMany(ctx, bson.M{"postId": postID})
	return err
}

func (r *mongoPostRepository) AddComment(ctx context.Context, c *Comment) error {
	res, err := r.comments.InsertOne(ctx, c)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		c.ID = id
	}
	return nil
}

// Comments read oldest first, like a conversation.
func (r *mongoPostRepository) ListComments(ctx context.Context, postID bson.ObjectID, skip, limit int64) ([]Comment, int64, error) {
	filter := bson.M{"postId": postID}
	total, err := r.comments.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: 1}, {Key: "_id", Value: 1}}).
		SetSkip(skip).
		SetLimit(limit)
	cur, err := r.comments.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var out []Comment
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *mongoPostRepository) DeleteCommentsOf(ctx context.Context, postID bson.ObjectID) error {
	_, err := r.comments.DeleteMany(ctx, bson.M{"postId": postID})
	return err
}
