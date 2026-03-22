package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		logger.Info(
			"HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.RequestURI),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	}
}

// ErrorHandlingMiddleware handles panics
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

// RateLimitMiddleware provides basic rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic rate limiting can be added here
		c.Next()
	}
}
