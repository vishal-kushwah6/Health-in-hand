package community

import (
	"artGallery/internal/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	svc *PostService
}

func NewPostHandler(svc *PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

func (h *PostHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/posts")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/like", h.Like)
	g.GET("/:id/comments", h.Comments)
	g.POST("/:id/comments", h.AddComment)
}

func postResponse(v PostView) PostResponse {
	r := NewPostResponse(&v.Post, v.LikedByMe)
	r.CommunityName = v.CommunityName
	return r
}

// GET /api/posts?communityId=...   (no communityId = feed of my communities)
func (h *PostHandler) List(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var q PostQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, q) {
		return
	}
	q.Normalize()

	items, total, err := h.svc.List(c.Request.Context(), uid, q)
	if err != nil {
		respondError(c, err)
		return
	}
	out := make([]PostResponse, 0, len(items))
	for _, v := range items {
		out = append(out, postResponse(v))
	}
	c.JSON(http.StatusOK, common.PaginatedResponse[PostResponse]{Items: out, Page: q.Page, Limit: q.Limit, Total: total})
}

// POST /api/posts
func (h *PostHandler) Create(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, req) {
		return
	}
	v, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, postResponse(*v))
}

// GET /api/posts/:id
func (h *PostHandler) Get(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	v, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, postResponse(*v))
}

// DELETE /api/posts/:id
func (h *PostHandler) Delete(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uid, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /api/posts/:id/like  (toggle)
func (h *PostHandler) Like(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	resp, err := h.svc.ToggleLike(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GET /api/posts/:id/comments
func (h *PostHandler) Comments(c *gin.Context) {
	if _, ok := currentUser(c); !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	var q common.PaginationQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, q) {
		return
	}
	q.Normalize()

	items, total, err := h.svc.ListComments(c.Request.Context(), id, q.Skip(), int64(q.Limit))
	if err != nil {
		respondError(c, err)
		return
	}
	out := make([]CommentResponse, 0, len(items))
	for i := range items {
		out = append(out, NewCommentResponse(&items[i]))
	}
	c.JSON(http.StatusOK, common.PaginatedResponse[CommentResponse]{Items: out, Page: q.Page, Limit: q.Limit, Total: total})
}

// POST /api/posts/:id/comments
func (h *PostHandler) AddComment(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, req) {
		return
	}
	cm, err := h.svc.AddComment(c.Request.Context(), uid, id, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, NewCommentResponse(cm))
}
