package workout

import (
	"artGallery/internal/common"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type SessionHandler struct {
	svc *SessionService
}

func NewSessionHandler(svc *SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/workout-sessions")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.DELETE("/:id", h.Delete)

	// kept outside the group so it can never clash with "/:id"
	rg.GET("/last-sets", h.LastSets)
}

// POST /api/workout-sessions  (sent once when the user taps "Finish")
func (h *SessionHandler) Create(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badJSON(c)
		return
	}
	if invalid(c, req) {
		return
	}
	s, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, NewSessionResponse(s))
}

// GET /api/workout-sessions?from=2026-10-01&to=2026-10-31&page=1
func (h *SessionHandler) List(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	var q SessionQuery
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
	out := make([]SessionSummaryResponse, 0, len(items))
	for i := range items {
		out = append(out, NewSessionSummaryResponse(&items[i]))
	}
	c.JSON(http.StatusOK, common.PaginatedResponse[SessionSummaryResponse]{
		Items: out, Page: q.Page, Limit: q.Limit, Total: total,
	})
}

// GET /api/workout-sessions/:id
func (h *SessionHandler) Get(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	s, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, NewSessionResponse(s))
}

// DELETE /api/workout-sessions/:id
func (h *SessionHandler) Delete(c *gin.Context) {
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

// GET /api/last-sets?exerciseIds=id1,id2,id3
func (h *SessionHandler) LastSets(c *gin.Context) {
	uid, ok := currentUser(c)
	if !ok {
		return
	}
	raw := strings.Split(c.Query("exerciseIds"), ",")
	ids := make([]bson.ObjectID, 0, len(raw))
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		id, err := bson.ObjectIDFromHex(r)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
			return
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 || len(ids) > 30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "send between 1 and 30 exerciseIds"})
		return
	}

	sets, err := h.svc.LastSets(c.Request.Context(), uid, ids)
	if err != nil {
		h.fail(c, err)
		return
	}
	out := LastSetsResponse{Items: make(map[string][]LastSetResponse, len(sets))}
	for id, list := range sets {
		resp := make([]LastSetResponse, 0, len(list))
		for _, st := range list {
			resp = append(resp, LastSetResponse{Reps: st.Reps, WeightKg: st.WeightKg})
		}
		out.Items[id] = resp
	}
	c.JSON(http.StatusOK, out)
}

func (h *SessionHandler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "workout session not found"})
	case errors.Is(err, ErrInvalidExercises):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": gin.H{"exercises": "contains an exercise that does not exist"}})
	case errors.Is(err, ErrInvalidPlan):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": gin.H{"workoutPlanId": "plan does not exist"}})
	case errors.Is(err, ErrInvalidDate):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": gin.H{"date": "cannot be in the future"}})
	default:
		log.Printf("session error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
