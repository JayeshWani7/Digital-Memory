package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/digital-memory/ingestion-service/internal/database"
	"github.com/digital-memory/ingestion-service/internal/models"
	"github.com/digital-memory/ingestion-service/internal/queue"
)

// EventHandler handles incoming webhook events
type EventHandler struct {
	db           *database.PostgresDB
	queue        *queue.RedisProducer
	logger       *zap.Logger
	requestCount int64
	successCount int64
}

// NewEventHandler creates a new event handler
func NewEventHandler(db *database.PostgresDB, q *queue.RedisProducer, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		db:     db,
		queue:  q,
		logger: logger,
	}
}

// HealthCheck returns the health status of the service
func (h *EventHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}

// Status returns the current status of the service
func (h *EventHandler) Status(c *gin.Context) {
	stats := h.db.GetEventStats()
	streamStats := h.queue.GetStreamStats(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"service":        "ingestion-service",
		"uptime_seconds": time.Now().Unix(),
		"request_count":  h.requestCount,
		"success_count":  h.successCount,
		"event_stats":    stats,
		"queue_stats":    streamStats,
	})
}

// HandleSlackEvent handles incoming Slack webhook events
func (h *EventHandler) HandleSlackEvent(c *gin.Context) {
	h.requestCount++

	// Verify Slack signature
	if !h.verifySlackSignature(c) {
		h.logger.Warn("Invalid Slack signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Parse Slack event
	var slackEvent models.SlackEvent
	if err := json.Unmarshal(body, &slackEvent); err != nil {
		h.logger.Error("Failed to parse Slack event", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Handle URL verification challenge (Slack setup)
	if slackEvent.Type == "url_verification" {
		var challenge struct {
			Challenge string `json:"challenge"`
		}
		if err := json.Unmarshal(body, &challenge); err == nil {
			c.JSON(http.StatusOK, gin.H{"challenge": challenge.Challenge})
			return
		}
	}

	// Process event
	if slackEvent.Type == "event_callback" {
		if err := h.processSlackEvent(&slackEvent); err != nil {
			h.logger.Error("Failed to process Slack event", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "processing failed"})
			return
		}
		h.successCount++
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleGitHubEvent handles incoming GitHub webhook events
func (h *EventHandler) HandleGitHubEvent(c *gin.Context) {
	h.requestCount++

	// Verify GitHub signature
	if !h.verifyGitHubSignature(c) {
		h.logger.Warn("Invalid GitHub signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Parse GitHub event
	var githubEvent models.GitHubEvent
	if err := json.Unmarshal(body, &githubEvent); err != nil {
		h.logger.Error("Failed to parse GitHub event", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Process event
	if err := h.processGitHubEvent(&githubEvent, string(body)); err != nil {
		h.logger.Error("Failed to process GitHub event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "processing failed"})
		return
	}
	h.successCount++

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Metrics endpoint
func (h *EventHandler) Metrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"requests_total":      h.requestCount,
		"requests_successful": h.successCount,
		"timestamp":           time.Now().Unix(),
	})
}

// Helper functions

func (h *EventHandler) processSlackEvent(slackEvent *models.SlackEvent) error {
	if slackEvent.Event.Type != "message" {
		return nil
	}

	event := &models.Event{
		ID:         uuid.New().String(),
		Source:     models.SourceSlack,
		SourceID:   slackEvent.Event.Timestamp,
		EventType:  models.EventMessage,
		RawData:    h.slackEventToMap(slackEvent),
		Author:     slackEvent.Event.User,
		Channel:    slackEvent.Event.Channel,
		ReceivedAt: time.Now(),
	}

	// Store event
	if err := h.db.StoreEvent(event); err != nil {
		return err
	}

	// Publish to queue
	return h.queue.PublishEvent(event)
}

func (h *EventHandler) processGitHubEvent(githubEvent *models.GitHubEvent, rawBody string) error {
	var event *models.Event
	var sourceID string

	switch {
	case githubEvent.PullRequest != nil:
		sourceID = fmt.Sprintf("pr-%d", githubEvent.PullRequest.Number)
		if githubEvent.Action == "opened" {
			event = &models.Event{
				ID:         uuid.New().String(),
				Source:     models.SourceGitHub,
				SourceID:   sourceID,
				EventType:  models.EventPRCreated,
				Author:     githubEvent.PullRequest.User.Login,
				Channel:    githubEvent.Repository.FullName,
				ReceivedAt: time.Now(),
			}
		} else if githubEvent.Action == "synchronize" || githubEvent.Action == "edited" {
			event = &models.Event{
				ID:         uuid.New().String(),
				Source:     models.SourceGitHub,
				SourceID:   sourceID,
				EventType:  models.EventPRUpdated,
				Author:     githubEvent.PullRequest.User.Login,
				Channel:    githubEvent.Repository.FullName,
				ReceivedAt: time.Now(),
			}
		}
	}

	if event == nil {
		return nil // Event type not processed
	}

	// Parse raw data
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(rawBody), &rawData); err != nil {
		return err
	}
	event.RawData = rawData

	// Store event
	if err := h.db.StoreEvent(event); err != nil {
		return err
	}

	// Publish to queue
	return h.queue.PublishEvent(event)
}

func (h *EventHandler) slackEventToMap(slackEvent *models.SlackEvent) map[string]interface{} {
	return map[string]interface{}{
		"token":      slackEvent.Token,
		"team_id":    slackEvent.TeamID,
		"event_id":   slackEvent.EventID,
		"event_type": slackEvent.Event.Type,
		"user":       slackEvent.Event.User,
		"text":       slackEvent.Event.Text,
		"channel":    slackEvent.Event.Channel,
		"timestamp":  slackEvent.Event.Timestamp,
	}
}

func (h *EventHandler) verifySlackSignature(c *gin.Context) bool {
	timestamp := c.GetHeader("X-Slack-Request-Timestamp")
	signature := c.GetHeader("X-Slack-Signature")
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	environment := os.Getenv("ENV")

	if signingSecret == "" {
		h.logger.Warn("SLACK_SIGNING_SECRET not configured")
		return environment == "development"
	}

	if timestamp == "" || signature == "" {
		return false
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	if ts < now-300 || ts > now+300 {
		return false
	}

	// Read body (we'll need to re-read it)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(baseString))
	expectedSignature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (h *EventHandler) verifyGitHubSignature(c *gin.Context) bool {
	signature := c.GetHeader("X-Hub-Signature-256")
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	environment := os.Getenv("ENV")

	if secret == "" {
		h.logger.Warn("GITHUB_WEBHOOK_SECRET not configured")
		return environment == "development"
	}

	if signature == "" {
		return false
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(body)
	expectedSig := "sha256=" + hex.EncodeToString(hash.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}
