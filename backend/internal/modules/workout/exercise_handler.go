package workout

import (
	"artGallery/internal/common"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExerciseHandler struct {
	svc *ExerciseService
}

func NewExerciseHandler(svc *ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{svc: svc}
}

func (h *ExerciseHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/exercises")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Archive)
}

// GET /api/exercises?scope=all|mine|library&muscleGroup=chest&equipment=barbell&search=press&page=1
func (h *ExerciseHandler) List(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var q ExerciseQuery
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
	out := make([]ExerciseResponse, 0, len(items))
	for i := range items {
		out = append(out, NewExerciseResponse(&items[i]))
	}
	c.JSON(http.StatusOK, common.PaginatedResponse[ExerciseResponse]{
		Items: out, Page: q.Page, Limit: q.Limit, Total: total,
	})
}

// GET /api/exercises/:id
func (h *ExerciseHandler) Get(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	e, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, NewExerciseResponse(e))
}

// POST /api/exercises  (creates a custom exercise owned by the caller)
func (h *ExerciseHandler) Create(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var req CreateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, req) {
		return
	}
	e, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, NewExerciseResponse(e))
}

// PUT /api/exercises/:id
func (h *ExerciseHandler) Update(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req UpdateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, req) {
		return
	}
	e, err := h.svc.Update(c.Request.Context(), uid, id, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, NewExerciseResponse(e))
}

// DELETE /api/exercises/:id  (archives; 204 on success)
func (h *ExerciseHandler) Archive(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.svc.Archive(c.Request.Context(), uid, id); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ExerciseHandler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrExerciseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "exercise not found"})
	case errors.Is(err, ErrReadOnlyExercise):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrExerciseNameTaken):
		c.JSON(http.StatusConflict, gin.H{
			"error":  "validation failed",
			"fields": gin.H{"name": "you already have an exercise with this name"},
		})
	default:
		log.Printf("exercise error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
