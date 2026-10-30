package community

import (
	"strings"
	"time"

	"artGallery/internal/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// =====================================================================
// Community DTOs
// =====================================================================

// CreateCommunityRequest: POST /api/communities
type CreateCommunityRequest struct {
	Name        string   `json:"name"        validate:"required,notblank,min=3,max=60"`
	Description string   `json:"description" validate:"omitempty,max=500"`
	Category    Category `json:"category"    validate:"required,oneof=running gym weight_loss muscle_building calisthenics yoga powerlifting nutrition other"`
	Image       string   `json:"image"       validate:"omitempty,url,max=500"`
}

func (r CreateCommunityRequest) ToModel(creatorID string) Community {
	now := time.Now().UTC()
	return Community{
		Name: strings.TrimSpace(r.Name), Description: strings.TrimSpace(r.Description),
		Category: r.Category, Image: r.Image, CreatedBy: creatorID,
		MemberCount: 0, // service inserts the creator as first member and sets this to 1
		CreatedAt:   now, UpdatedAt: now,
	}
}

// UpdateCommunityRequest: PUT /api/communities/:id (creator only; check in service).
type UpdateCommunityRequest struct {
	Name        *string   `json:"name"        validate:"omitempty,notblank,min=3,max=60"`
	Description *string   `json:"description" validate:"omitempty,max=500"`
	Category    *Category `json:"category"    validate:"omitempty,oneof=running gym weight_loss muscle_building calisthenics yoga powerlifting nutrition other"`
	Image       *string   `json:"image"       validate:"omitempty,url,max=500"`
}

func (r UpdateCommunityRequest) Apply(c *Community) {
	if r.Name != nil {
		c.Name = strings.TrimSpace(*r.Name)
	}
	if r.Description != nil {
		c.Description = strings.TrimSpace(*r.Description)
	}
	if r.Category != nil {
		c.Category = *r.Category
	}
	if r.Image != nil {
		c.Image = *r.Image
	}
	c.UpdatedAt = time.Now().UTC()
}

// CommunityQuery: GET /api/communities?category=gym&search=lift&joined=true
type CommunityQuery struct {
	common.PaginationQuery
	Category Category `json:"category" form:"category" validate:"omitempty,oneof=running gym weight_loss muscle_building calisthenics yoga powerlifting nutrition other"`
	Search   string   `json:"search"   form:"search"   validate:"omitempty,max=50"`
	Joined   bool     `json:"joined"   form:"joined"` // true = only communities I belong to
}

type CommunityResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Category    Category  `json:"category"`
	Image       string    `json:"image,omitempty"`
	MemberCount int       `json:"memberCount"`
	IsMember    bool      `json:"isMember"` // computed per request, not stored
	CreatedAt   time.Time `json:"createdAt"`
}

func NewCommunityResponse(c *Community, isMember bool) CommunityResponse {
	return CommunityResponse{
		ID: c.ID.Hex(), Name: c.Name, Description: c.Description, Category: c.Category,
		Image: c.Image, MemberCount: c.MemberCount, IsMember: isMember, CreatedAt: c.CreatedAt,
	}
}

// =====================================================================
// Post DTOs
// =====================================================================

// CreatePostRequest: POST /api/posts
type CreatePostRequest struct {
	CommunityID string   `json:"communityId" validate:"required,objectid"`
	Content     string   `json:"content"     validate:"required,notblank,max=2000"`
	Images      []string `json:"images"      validate:"omitempty,max=4,dive,url,max=500"`
}

// ToModel: author info comes from the profile of the authenticated user.
func (r CreatePostRequest) ToModel(authorID, username, picture string) (Post, error) {
	cid, err := bson.ObjectIDFromHex(r.CommunityID)
	if err != nil {
		return Post{}, err
	}
	now := time.Now().UTC()
	return Post{
		AuthorID: authorID, AuthorUsername: username, AuthorPicture: picture,
		CommunityID: cid, Content: strings.TrimSpace(r.Content), Images: r.Images,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

// UpdatePostRequest: PUT /api/posts/:id (author only; check in service).
type UpdatePostRequest struct {
	Content string `json:"content" validate:"required,notblank,max=2000"`
}

// PostQuery: GET /api/posts?communityId=...  (omit communityId = feed of my communities)
type PostQuery struct {
	common.PaginationQuery
	CommunityID string `json:"communityId" form:"communityId" validate:"omitempty,objectid"`
}

type PostResponse struct {
	ID             string    `json:"id"`
	AuthorID       string    `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	AuthorPicture  string    `json:"authorPicture,omitempty"`
	CommunityID    string    `json:"communityId"`
	CommunityName  string    `json:"communityName,omitempty"`
	Content        string    `json:"content"`
	Images         []string  `json:"images,omitempty"`
	LikeCount      int       `json:"likeCount"`
	CommentCount   int       `json:"commentCount"`
	LikedByMe      bool      `json:"likedByMe"` // computed per request, not stored
	CreatedAt      time.Time `json:"createdAt"`
}

func NewPostResponse(p *Post, likedByMe bool) PostResponse {
	return PostResponse{
		ID: p.ID.Hex(), AuthorID: p.AuthorID, AuthorUsername: p.AuthorUsername,
		AuthorPicture: p.AuthorPicture, CommunityID: p.CommunityID.Hex(), Content: p.Content,
		Images: p.Images, LikeCount: p.LikeCount, CommentCount: p.CommentCount,
		LikedByMe: likedByMe, CreatedAt: p.CreatedAt,
	}
}

// LikeResponse: returned by POST /api/posts/:id/like (acts as a toggle).
type LikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"likeCount"`
}

// =====================================================================
// Comment DTOs
// =====================================================================

// CreateCommentRequest: POST /api/posts/:id/comments
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,notblank,max=500"`
}

func (r CreateCommentRequest) ToModel(postID bson.ObjectID, authorID, username, picture string) Comment {
	return Comment{
		PostID: postID, AuthorID: authorID, AuthorUsername: username, AuthorPicture: picture,
		Content: strings.TrimSpace(r.Content), CreatedAt: time.Now().UTC(),
	}
}

type CommentResponse struct {
	ID             string    `json:"id"`
	PostID         string    `json:"postId"`
	AuthorID       string    `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	AuthorPicture  string    `json:"authorPicture,omitempty"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
}

func NewCommentResponse(c *Comment) CommentResponse {
	return CommentResponse{
		ID: c.ID.Hex(), PostID: c.PostID.Hex(), AuthorID: c.AuthorID,
		AuthorUsername: c.AuthorUsername, AuthorPicture: c.AuthorPicture,
		Content: c.Content, CreatedAt: c.CreatedAt,
	}
}
