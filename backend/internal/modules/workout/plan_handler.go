package workout

import (
	"artGallery/internal/common"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	svc *PlanService
}

func NewPlanHandler(svc *PlanService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

func (h *PlanHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/workouts")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

// GET /api/workouts?day=monday
func (h *PlanHandler) List(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var q PlanQuery
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
		h.fail(c, err)
		return
	}
	out := make([]PlanSummaryResponse, 0, len(items))
	for i := range items {
		out = append(out, NewPlanSummaryResponse(&items[i]))
	}
	c.JSON(http.StatusOK, common.PaginatedResponse[PlanSummaryResponse]{
		Items: out, Page: q.Page, Limit: q.Limit, Total: total,
	})
}

// POST /api/workouts
func (h *PlanHandler) Create(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, req) {
		return
	}
	d, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, NewPlanResponse(d.Plan, d.Names))
}

// GET /api/workouts/:id
func (h *PlanHandler) Get(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	d, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, NewPlanResponse(d.Plan, d.Names))
}

// PUT /api/workouts/:id
func (h *PlanHandler) Update(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, req) {
		return
	}
	d, err := h.svc.Update(c.Request.Context(), uid, id, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, NewPlanResponse(d.Plan, d.Names))
}

// DELETE /api/workouts/:id
func (h *PlanHandler) Delete(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uid, id); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PlanHandler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrPlanNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "workout plan not found"})
	case errors.Is(err, ErrInvalidExercises):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "validation failed",
			"fields": gin.H{"exercises": "contains an exercise that does not exist"},
		})
	default:
		log.Printf("plan error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
