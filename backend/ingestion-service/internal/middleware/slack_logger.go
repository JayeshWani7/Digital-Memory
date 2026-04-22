package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SlackLoggerMiddleware logs all incoming Slack webhook requests
func SlackLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only log Slack webhook endpoints
		if c.Request.URL.Path != "/webhook/slack" {
			c.Next()
			return
		}

		start := time.Now()

		// Read the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("Failed to read request body", zap.Error(err))
			c.Next()
			return
		}

		// Restore the body so the handler can read it
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		// Parse as JSON for structured logging
		var slackEvent map[string]interface{}
		if err := json.Unmarshal(body, &slackEvent); err == nil {
			// Log with structured fields
			logger.Info("Slack webhook received",
				zap.String("event_type", getString(slackEvent, "type")),
				zap.String("team_id", getString(slackEvent, "team_id")),
				zap.String("event_id", getString(slackEvent, "event_id")),
				zap.Int64("event_time", getInt64(slackEvent, "event_time")),
				zap.Duration("latency", time.Since(start)),
			)
		} else {
			// Fallback: avoid logging raw payload to prevent leaking sensitive data
			logger.Info("Slack webhook received",
				zap.Bool("payload_parse_failed", true),
				zap.Int("payload_size_bytes", len(body)),
				zap.Duration("latency", time.Since(start)),
			)
		}

		c.Next()
	}
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key].(float64); ok {
		return int64(val)
	}
	return 0
}
