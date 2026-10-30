package community

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Category string

const (
	CategoryRunning        Category = "running"
	CategoryGym            Category = "gym"
	CategoryWeightLoss     Category = "weight_loss"
	CategoryMuscleBuilding Category = "muscle_building"
	CategoryCalisthenics   Category = "calisthenics"
	CategoryYoga           Category = "yoga"
	CategoryPowerlifting   Category = "powerlifting"
	CategoryNutrition      Category = "nutrition"
	CategoryOther          Category = "other"
)

// All user ids below are Clerk user ids.

// Community. Collection: "communities".
// MemberCount is a counter kept in sync with community_members ($inc on join/leave).
type Community struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Description string        `bson:"description,omitempty"`
	Category    Category      `bson:"category"`
	Image       string        `bson:"image,omitempty"`
	CreatedBy   string        `bson:"createdBy"`
	MemberCount int           `bson:"memberCount"`
	CreatedAt   time.Time     `bson:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt"`
}

// CommunityMember: one document per (community, user). Collection: "community_members".
type CommunityMember struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	CommunityID bson.ObjectID `bson:"communityId"`
	UserID      string        `bson:"userId"`
	JoinedAt    time.Time     `bson:"joinedAt"`
}

// Post. Collection: "posts".
// AuthorUsername/AuthorPicture are copied at write time so feeds need no join.
type Post struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	AuthorID       string        `bson:"authorId"`
	AuthorUsername string        `bson:"authorUsername"`
	AuthorPicture  string        `bson:"authorPicture,omitempty"`
	CommunityID    bson.ObjectID `bson:"communityId"`
	Content        string        `bson:"content"`
	Images         []string      `bson:"images,omitempty"`
	LikeCount      int           `bson:"likeCount"`
	CommentCount   int           `bson:"commentCount"`
	CreatedAt      time.Time     `bson:"createdAt"`
	UpdatedAt      time.Time     `bson:"updatedAt"`
}

// PostLike: one document per (post, user). Collection: "post_likes".
type PostLike struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	PostID    bson.ObjectID `bson:"postId"`
	UserID    string        `bson:"userId"`
	CreatedAt time.Time     `bson:"createdAt"`
}

// Comment. Collection: "comments".
type Comment struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	PostID         bson.ObjectID `bson:"postId"`
	AuthorID       string        `bson:"authorId"`
	AuthorUsername string        `bson:"authorUsername"`
	AuthorPicture  string        `bson:"authorPicture,omitempty"`
	Content        string        `bson:"content"`
	CreatedAt      time.Time     `bson:"createdAt"`
}

func CommunityIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "name", Value: 1}},
			Options: options.Index().
				SetName("name_ci_unique").
				SetUnique(true).
				SetCollation(&options.Collation{Locale: "en", Strength: 2}),
		},
		{Keys: bson.D{{Key: "category", Value: 1}}},
	}
}

// MemberIndexes: the unique pair makes "join twice" fail at the DB level.
func MemberIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "communityId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "userId", Value: 1}}},
	}
}

func PostIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "communityId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "authorId", Value: 1}, {Key: "createdAt", Value: -1}}},
	}
}

// LikeIndexes: the unique pair makes "like twice" fail at the DB level.
func LikeIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "postId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)},
	}
}

func CommentIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "postId", Value: 1}, {Key: "createdAt", Value: 1}}},
	}
}
