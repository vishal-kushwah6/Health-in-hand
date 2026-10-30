package community

import (
	"artGallery/internal/shared/middleware"
	validation "artGallery/internal/shared/validator"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Helpers shared by the handlers in this module.

func currentUser(c *gin.Context) (string, bool) {
	uid, ok := middleware.ClerkUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	return uid, ok
}

// invalid writes a 400 and returns true when the struct fails validation.
func invalid(c *gin.Context, v any) bool {
	if errs := validation.Validate(v); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "fields": errs})
		return true
	}
	return false
}

func badJSON(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
}

func idParam(c *gin.Context) (bson.ObjectID, bool) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return bson.ObjectID{}, false
	}
	return id, true
}

// respondError maps domain errors to HTTP status codes in one place.
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrCommunityNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "community not found"})
	case errors.Is(err, ErrPostNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
	case errors.Is(err, ErrCommunityNameTaken):
		c.JSON(http.StatusConflict, gin.H{
			"error":  "validation failed",
			"fields": gin.H{"name": "a community with this name already exists"},
		})
	case errors.Is(err, ErrNotMember):
		c.JSON(http.StatusForbidden, gin.H{"error": "join the community first"})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only do this to your own content"})
	case errors.Is(err, ErrNoProfile):
		c.JSON(http.StatusBadRequest, gin.H{"error": "create your profile first"})
	case errors.Is(err, ErrBadImages):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "validation failed",
			"fields": gin.H{"images": "photos must be uploaded through the app first"},
		})
	default:
		log.Printf("community error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
