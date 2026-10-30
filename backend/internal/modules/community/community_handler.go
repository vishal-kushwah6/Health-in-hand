package community

import (
	"artGallery/internal/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommunityHandler struct {
	svc *CommunityService
}

func NewCommunityHandler(svc *CommunityService) *CommunityHandler {
	return &CommunityHandler{svc: svc}
}

func (h *CommunityHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/communities")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.POST("/:id/join", h.Join)
	g.DELETE("/:id/leave", h.Leave)
}

func view(v CommunityView) CommunityResponse {
	return NewCommunityResponse(&v.Community, v.IsMember)
}

// GET /api/communities?category=gym&search=lift&joined=true
func (h *CommunityHandler) List(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var q CommunityQuery
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
	out := make([]CommunityResponse, 0, len(items))
	for _, v := range items {
		out = append(out, view(v))
	}
	c.JSON(http.StatusOK, common.PaginatedResponse[CommunityResponse]{Items: out, Page: q.Page, Limit: q.Limit, Total: total})
}

// POST /api/communities
func (h *CommunityHandler) Create(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var req CreateCommunityRequest
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
	c.JSON(http.StatusCreated, view(*v))
}

// GET /api/communities/:id
func (h *CommunityHandler) Get(c *gin.Context) {
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
	c.JSON(http.StatusOK, view(*v))
}

// POST /api/communities/:id/join
func (h *CommunityHandler) Join(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	v, err := h.svc.Join(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, view(*v))
}

// DELETE /api/communities/:id/leave
func (h *CommunityHandler) Leave(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	v, err := h.svc.Leave(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, view(*v))
}
