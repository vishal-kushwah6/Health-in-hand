package community

import (
	"context"
	"errors"
	"strings"
	"time"

	media "artGallery/internal/media/domain"
	"artGallery/internal/modules/user"
	"artGallery/internal/shared/database"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrNoProfile = errors.New("no profile")
	ErrBadImages = errors.New("images must be uploaded through the app")
)

// ProfileFinder is the one thing this module needs from the user module.
type ProfileFinder interface {
	FindByClerkID(ctx context.Context, clerkUserID string) (*user.UserProfile, error)
}

// PostService. Rules: anyone can READ posts; you must be a MEMBER of the
// community to post, like or comment. Counters (likes, comments) are updated
// together with their records inside a transaction.
type PostService struct {
	tx          database.TxManager
	posts       PostRepository
	communities CommunityRepository
	profiles    ProfileFinder

	// optional image support (see WithImages)
	store       media.Storage
	imagePrefix string
}

func NewPostService(tx database.TxManager, posts PostRepository, communities CommunityRepository, profiles ProfileFinder) *PostService {
	return &PostService{tx: tx, posts: posts, communities: communities, profiles: profiles}
}

// WithImages enables photo posts. imagePrefix is the only URL prefix accepted
// for post images (CloudinaryPrefix), so a client cannot attach arbitrary links.
// Without it, posts containing images are rejected.
func (s *PostService) WithImages(store media.Storage, imagePrefix string) *PostService {
	s.store = store
	s.imagePrefix = imagePrefix
	return s
}

func (s *PostService) validImages(urls []string) bool {
	if len(urls) == 0 {
		return true
	}
	if s.imagePrefix == "" {
		return false
	}
	for _, u := range urls {
		if !strings.HasPrefix(u, s.imagePrefix) {
			return false
		}
	}
	return true
}

// PostView is a post plus per-request extras.
type PostView struct {
	Post          Post
	LikedByMe     bool
	CommunityName string
}

func (s *PostService) requireMember(ctx context.Context, userID string, communityID bson.ObjectID) error {
	if _, err := s.communities.FindByID(ctx, communityID); err != nil {
		return err
	}
	ok, err := s.communities.IsMember(ctx, communityID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotMember
	}
	return nil
}

func (s *PostService) author(ctx context.Context, userID string) (*user.UserProfile, error) {
	p, err := s.profiles.FindByClerkID(ctx, userID)
	if errors.Is(err, user.ErrNotFound) {
		return nil, ErrNoProfile
	}
	return p, err
}

func (s *PostService) views(ctx context.Context, userID string, posts []Post) ([]PostView, error) {
	postIDs := make([]bson.ObjectID, 0, len(posts))
	communityIDs := make([]bson.ObjectID, 0, len(posts))
	seen := make(map[bson.ObjectID]struct{}, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
		if _, dup := seen[p.CommunityID]; !dup {
			seen[p.CommunityID] = struct{}{}
			communityIDs = append(communityIDs, p.CommunityID)
		}
	}
	liked, err := s.posts.LikedBy(ctx, userID, postIDs)
	if err != nil {
		return nil, err
	}
	comms, err := s.communities.FindByIDs(ctx, communityIDs)
	if err != nil {
		return nil, err
	}
	names := make(map[bson.ObjectID]string, len(comms))
	for _, c := range comms {
		names[c.ID] = c.Name
	}
	out := make([]PostView, 0, len(posts))
	for _, p := range posts {
		out = append(out, PostView{Post: p, LikedByMe: liked[p.ID], CommunityName: names[p.CommunityID]})
	}
	return out, nil
}

func (s *PostService) one(ctx context.Context, userID string, p *Post) (*PostView, error) {
	v, err := s.views(ctx, userID, []Post{*p})
	if err != nil {
		return nil, err
	}
	return &v[0], nil
}

func (s *PostService) Create(ctx context.Context, userID string, req CreatePostRequest) (*PostView, error) {
	if !s.validImages(req.Images) {
		return nil, ErrBadImages
	}
	cid, err := bson.ObjectIDFromHex(req.CommunityID)
	if err != nil {
		return nil, ErrCommunityNotFound
	}
	if err := s.requireMember(ctx, userID, cid); err != nil {
		return nil, err
	}
	profile, err := s.author(ctx, userID)
	if err != nil {
		return nil, err
	}
	post, err := req.ToModel(userID, profile.Username, profile.ProfilePicture)
	if err != nil {
		return nil, ErrCommunityNotFound
	}
	if err := s.posts.CreatePost(ctx, &post); err != nil {
		return nil, err
	}
	return s.one(ctx, userID, &post)
}

// List: one community when communityId is given, otherwise the feed of every
// community the caller has joined.
func (s *PostService) List(ctx context.Context, userID string, q PostQuery) ([]PostView, int64, error) {
	var ids []bson.ObjectID
	if q.CommunityID != "" {
		id, err := bson.ObjectIDFromHex(q.CommunityID)
		if err != nil {
			return nil, 0, ErrCommunityNotFound
		}
		if _, err := s.communities.FindByID(ctx, id); err != nil {
			return nil, 0, err
		}
		ids = []bson.ObjectID{id}
	} else {
		joined, err := s.communities.JoinedIDs(ctx, userID)
		if err != nil {
			return nil, 0, err
		}
		if len(joined) == 0 {
			return []PostView{}, 0, nil
		}
		ids = joined
	}
	posts, total, err := s.posts.ListPosts(ctx, ids, q.Skip(), int64(q.Limit))
	if err != nil {
		return nil, 0, err
	}
	v, err := s.views(ctx, userID, posts)
	return v, total, err
}

func (s *PostService) Get(ctx context.Context, userID string, id bson.ObjectID) (*PostView, error) {
	p, err := s.posts.FindPost(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.one(ctx, userID, p)
}

// Delete: author only. Removes the post together with its likes and comments.
func (s *PostService) Delete(ctx context.Context, userID string, id bson.ObjectID) error {
	p, err := s.posts.FindPost(ctx, id)
	if err != nil {
		return err
	}
	if p.AuthorID != userID {
		return ErrForbidden
	}
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.posts.DeletePost(ctx, userID, id); err != nil {
			return err
		}
		if err := s.posts.DeleteLikesOf(ctx, id); err != nil {
			return err
		}
		return s.posts.DeleteCommentsOf(ctx, id)
	})
	if err != nil {
		return err
	}

	// Best effort: remove the photos from Cloudinary too. A failure here must
	// not undo the deletion (at worst an orphan file stays in storage).
	if s.store != nil {
		for _, img := range p.Images {
			if key := publicIDFromURL(img); key != "" {
				_ = s.store.Delete(ctx, key)
			}
		}
	}
	return nil
}

// ToggleLike likes the post, or un-likes it if you already liked it.
func (s *PostService) ToggleLike(ctx context.Context, userID string, postID bson.ObjectID) (LikeResponse, error) {
	post, err := s.posts.FindPost(ctx, postID)
	if err != nil {
		return LikeResponse{}, err
	}
	if err := s.requireMember(ctx, userID, post.CommunityID); err != nil {
		return LikeResponse{}, err
	}

	liked := false
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		removed, err := s.posts.RemoveLike(ctx, postID, userID)
		if err != nil {
			return err
		}
		if removed {
			liked = false
			return s.posts.IncPost(ctx, postID, fieldLikeCount, -1)
		}
		if err := s.posts.AddLike(ctx, &PostLike{PostID: postID, UserID: userID, CreatedAt: time.Now().UTC()}); err != nil {
			return err
		}
		liked = true
		return s.posts.IncPost(ctx, postID, fieldLikeCount, 1)
	})
	if err != nil {
		return LikeResponse{}, err
	}
	updated, err := s.posts.FindPost(ctx, postID)
	if err != nil {
		return LikeResponse{}, err
	}
	return LikeResponse{Liked: liked, LikeCount: updated.LikeCount}, nil
}

func (s *PostService) ListComments(ctx context.Context, postID bson.ObjectID, skip, limit int64) ([]Comment, int64, error) {
	if _, err := s.posts.FindPost(ctx, postID); err != nil {
		return nil, 0, err
	}
	return s.posts.ListComments(ctx, postID, skip, limit)
}

func (s *PostService) AddComment(ctx context.Context, userID string, postID bson.ObjectID, req CreateCommentRequest) (*Comment, error) {
	post, err := s.posts.FindPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	if err := s.requireMember(ctx, userID, post.CommunityID); err != nil {
		return nil, err
	}
	profile, err := s.author(ctx, userID)
	if err != nil {
		return nil, err
	}
	c := req.ToModel(postID, userID, profile.Username, profile.ProfilePicture)
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.posts.AddComment(ctx, &c); err != nil {
			return err
		}
		return s.posts.IncPost(ctx, postID, fieldCommentCount, 1)
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}
