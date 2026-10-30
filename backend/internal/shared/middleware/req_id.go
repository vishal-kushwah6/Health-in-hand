package middleware

import (
	logger "artGallery/internal/shared/logging"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqID := uuid.NewString()
		ctx := context.WithValue(
			c.Request.Context(),
			logger.RequestIDKey,
			reqID,
		)

		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set("X-Request-ID", reqID)
		l.Info(ctx, "http request come", map[string]any{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		})

		c.Next()
		l.Info(ctx, "http request completed", map[string]any{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"duration": time.Since(start).Milliseconds(),
		})

	}
}
