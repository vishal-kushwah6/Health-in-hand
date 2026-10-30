package nutrition

import (
	"artGallery/internal/shared/middleware"
	validation "artGallery/internal/shared/validator"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/nutrition")
	g.GET("/messages", h.List)
	g.POST("/messages", h.Send)
	g.DELETE("/messages", h.Clear)
}

// callerID returns the signed-in user id (named so it cannot clash with the imported "user" package).
func callerID(c *gin.Context) (string, bool) {
	uid, ok := middleware.ClerkUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	return uid, ok
}

// GET /api/nutrition/messages  (latest 50, oldest first)
func (h *Handler) List(c *gin.Context) {
	uid, ok := callerID(c)
	if !ok {
		return
	}
	msgs, err := h.svc.List(c.Request.Context(), uid)
	if err != nil {
		h.fail(c, err)
		return
	}
	out := make([]MessageResponse, 0, len(msgs))
	for i := range msgs {
		out = append(out, NewMessageResponse(&msgs[i]))
	}
	c.JSON(http.StatusOK, ListResponse{Items: out})
}

// POST /api/nutrition/messages
func (h *Handler) Send(c *gin.Context) {
	uid, ok := callerID(c)
	if !ok {
		return
	}
	if !h.svc.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "The AI nutritionist is not configured on this server."})
		return
	}

	// The model call takes a few seconds: lift the server-wide 5s write limit for THIS request.
	rc := http.NewResponseController(c.Writer)
	_ = rc.SetWriteDeadline(time.Now().Add(90 * time.Second))

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if errs := validation.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": errs})
		return
	}

	userMsg, botMsg, err := h.svc.Send(c.Request.Context(), uid, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, SendResponse{UserMessage: NewMessageResponse(userMsg), AssistantMessage: NewMessageResponse(botMsg)})
}

// DELETE /api/nutrition/messages  (clear my chat)
func (h *Handler) Clear(c *gin.Context) {
	uid, ok := callerID(c)
	if !ok {
		return
	}
	if err := h.svc.Clear(c.Request.Context(), uid); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "The AI nutritionist is not configured on this server."})
	case errors.Is(err, ErrRateLimited):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "You've reached the hourly limit of messages. Please try again a bit later."})
	case errors.Is(err, ErrNoProfile):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Create your profile first so I can tailor my advice."})
	case errors.Is(err, ErrAIUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": "The nutritionist is unavailable right now. Please try again in a moment."})
	default:
		log.Printf("nutrition error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
