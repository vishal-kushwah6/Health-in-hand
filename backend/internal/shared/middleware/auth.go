package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

// REFERENCE VERSION. If you already have a Clerk middleware, keep yours and
// just make sure it does c.Set("clerkUserId", <Clerk user id>).
//
//	go get github.com/clerk/clerk-sdk-go/v2

const ctxClerkUserID = "clerkUserId"

// ClerkAuth verifies the "Authorization: Bearer <session token>" header sent
// by the Expo app and stores the Clerk user id (the token's subject) in the
// Gin context.
func ClerkAuth(secretKey string) gin.HandlerFunc {
	clerk.SetKey(secretKey)

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if header == "" || token == header {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{Token: token})
		if err != nil {
			log.Println("clerk verify error:", err) // <-- add this line

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ctxClerkUserID, claims.Subject)
		c.Next()
	}
}

// ClerkUserID is what handlers call to know who is making the request.
func ClerkUserID(c *gin.Context) (string, bool) {
	id := c.GetString(ctxClerkUserID)
	return id, id != ""
}
