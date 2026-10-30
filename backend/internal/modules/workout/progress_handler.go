package workout

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProgressHandler struct {
	svc *ProgressService
}

func NewProgressHandler(svc *ProgressService) *ProgressHandler {
	return &ProgressHandler{svc: svc}
}

func (h *ProgressHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/progress")
	g.GET("", h.Summary)
	g.GET("/exercise", h.Exercise)
	g.GET("/records", h.Records)
}

// GET /api/progress?tz=330
func (h *ProgressHandler) Summary(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var q ProgressQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, q) {
		return
	}
	resp, err := h.svc.Summary(c.Request.Context(), uid, q.TZ)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GET /api/progress/exercise?exerciseId=...&range=three_months
func (h *ProgressHandler) Exercise(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var q ExerciseProgressQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, q) {
		return
	}
	id, err := bson.ObjectIDFromHex(q.ExerciseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.svc.Exercise(c.Request.Context(), uid, id, q.Range)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GET /api/progress/records
func (h *ProgressHandler) Records(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	items, err := h.svc.Records(c.Request.Context(), uid)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *ProgressHandler) fail(c *gin.Context, err error) {
	log.Printf("progress error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
