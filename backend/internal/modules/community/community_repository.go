package community

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
	ErrCommunityNotFound  = errors.New("community not found")
	ErrCommunityNameTaken = errors.New("community name taken")
	ErrNotMember          = errors.New("not a member")
	ErrAlreadyMember      = errors.New("already a member")
)

type CommunityRepository interface {
	Create(ctx context.Context, c *Community) error
	FindByID(ctx context.Context, id bson.ObjectID) (*Community, error)
	FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]Community, error)
	List(ctx context.Context, userID string, q CommunityQuery) ([]Community, int64, error)
	CountAll(ctx context.Context) (int64, error)
	InsertMany(ctx context.Context, items []Community) error
	IncMembers(ctx context.Context, id bson.ObjectID, delta int) error

	AddMember(ctx context.Context, m *CommunityMember) error
	RemoveMember(ctx context.Context, communityID bson.ObjectID, userID string) (bool, error)
	IsMember(ctx context.Context, communityID bson.ObjectID, userID string) (bool, error)
	// MemberOf returns which of the given communities the user belongs to.
	MemberOf(ctx context.Context, userID string, ids []bson.ObjectID) (map[bson.ObjectID]bool, error)
	JoinedIDs(ctx context.Context, userID string) ([]bson.ObjectID, error)
}

type mongoCommunityRepository struct {
	communities *mongo.Collection
	members     *mongo.Collection
}

func NewMongoCommunityRepository(db *mongo.Database) CommunityRepository {
	return &mongoCommunityRepository{
		communities: db.Collection(database.ColCommunities),
		members:     db.Collection(database.ColCommunityMembers),
	}
}

func (r *mongoCommunityRepository) Create(ctx context.Context, c *Community) error {
	res, err := r.communities.InsertOne(ctx, c)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrCommunityNameTaken
		}
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		c.ID = id
	}
	return nil
}

func (r *mongoCommunityRepository) FindByID(ctx context.Context, id bson.ObjectID) (*Community, error) {
	var c Community
	err := r.communities.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrCommunityNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *mongoCommunityRepository) FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]Community, error) {
	cur, err := r.communities.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	var out []Community
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *mongoCommunityRepository) List(ctx context.Context, userID string, q CommunityQuery) ([]Community, int64, error) {
	filter := bson.M{}
	if q.Category != "" {
		filter["category"] = q.Category
	}
	if q.Search != "" {
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(q.Search), "$options": "i"}
	}
	if q.Joined {
		ids, err := r.JoinedIDs(ctx, userID)
		if err != nil {
			return nil, 0, err
		}
		filter["_id"] = bson.M{"$in": ids}
	}

	total, err := r.communities.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "memberCount", Value: -1}, {Key: "name", Value: 1}}).
		SetCollation(&options.Collation{Locale: "en"}).
		SetSkip(q.Skip()).
		SetLimit(int64(q.Limit))
	cur, err := r.communities.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var out []Community
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *mongoCommunityRepository) CountAll(ctx context.Context) (int64, error) {
	return r.communities.CountDocuments(ctx, bson.M{})
}

func (r *mongoCommunityRepository) InsertMany(ctx context.Context, items []Community) error {
	docs := make([]any, len(items))
	for i := range items {
		docs[i] = items[i]
	}
	_, err := r.communities.InsertMany(ctx, docs)
	return err
}

func (r *mongoCommunityRepository) IncMembers(ctx context.Context, id bson.ObjectID, delta int) error {
	res, err := r.communities.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"memberCount": delta}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrCommunityNotFound
	}
	return nil
}

func (r *mongoCommunityRepository) AddMember(ctx context.Context, m *CommunityMember) error {
	if _, err := r.members.InsertOne(ctx, m); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAlreadyMember
		}
		return err
	}
	return nil
}

func (r *mongoCommunityRepository) RemoveMember(ctx context.Context, communityID bson.ObjectID, userID string) (bool, error) {
	res, err := r.members.DeleteOne(ctx, bson.M{"communityId": communityID, "userId": userID})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

func (r *mongoCommunityRepository) IsMember(ctx context.Context, communityID bson.ObjectID, userID string) (bool, error) {
	n, err := r.members.CountDocuments(ctx, bson.M{"communityId": communityID, "userId": userID})
	return n > 0, err
}

func (r *mongoCommunityRepository) MemberOf(ctx context.Context, userID string, ids []bson.ObjectID) (map[bson.ObjectID]bool, error) {
	cur, err := r.members.Find(ctx, bson.M{"userId": userID, "communityId": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	var rows []CommunityMember
	if err := cur.All(ctx, &rows); err != nil {
		return nil, err
	}
	out := make(map[bson.ObjectID]bool, len(rows))
	for _, m := range rows {
		out[m.CommunityID] = true
	}
	return out, nil
}

func (r *mongoCommunityRepository) JoinedIDs(ctx context.Context, userID string) ([]bson.ObjectID, error) {
	cur, err := r.members.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	var rows []CommunityMember
	if err := cur.All(ctx, &rows); err != nil {
		return nil, err
	}
	ids := make([]bson.ObjectID, 0, len(rows))
	for _, m := range rows {
		ids = append(ids, m.CommunityID)
	}
	return ids, nil
}
