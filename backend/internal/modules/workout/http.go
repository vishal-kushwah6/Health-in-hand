package workout

import (
	"artGallery/internal/shared/middleware"
	validation "artGallery/internal/shared/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Small helpers shared by every handler in this module.

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
