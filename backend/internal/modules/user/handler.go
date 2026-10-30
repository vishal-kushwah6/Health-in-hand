package user

import (
	"artGallery/internal/shared/middleware"
	validation "artGallery/internal/shared/validator"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes expects a group that already uses the Clerk middleware.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/profile", h.Get)
	rg.POST("/profile", h.Create)
	rg.PUT("/profile", h.Update)
}

// GET /api/profile
// 404 means "signed in with Clerk but no profile yet" -> app shows Create Profile screen.
func (h *Handler) Get(c *gin.Context) {
	uid, ok := middleware.ClerkUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	p, err := h.svc.Get(c.Request.Context(), uid)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, NewProfileResponse(p))
}

// POST /api/profile
func (h *Handler) Create(c *gin.Context) {
	uid, ok := middleware.ClerkUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if errs := validation.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": errs})
		return
	}
	p, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, NewProfileResponse(p))
}

// PUT /api/profile (partial update: send only what changed)
func (h *Handler) Update(c *gin.Context) {
	uid, ok := middleware.ClerkUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if errs := validation.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": errs})
		return
	}
	p, err := h.svc.Update(c.Request.Context(), uid, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, NewProfileResponse(p))
}

// fail maps domain errors to HTTP status codes in one place.
func (h *Handler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
	case errors.Is(err, ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "profile already exists"})
	case errors.Is(err, ErrUsernameTaken):
		c.JSON(http.StatusConflict, gin.H{
			"error":  "validation failed",
			"fields": gin.H{"username": "is already taken"},
		})
	default:
		log.Printf("profile error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
