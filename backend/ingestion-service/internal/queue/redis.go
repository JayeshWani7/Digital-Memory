package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/digital-memory/ingestion-service/internal/models"
)

// RedisProducer handles publishing messages to Redis streams
type RedisProducer struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisProducer creates a new Redis producer
func NewRedisProducer(redisURL string) (*RedisProducer, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger, _ := zap.NewProduction()
	return &RedisProducer{client: client, logger: logger}, nil
}

// PublishEvent publishes an event to the appropriate Redis stream
func (rp *RedisProducer) PublishEvent(event *models.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Determine stream name based on event type
	streamName := rp.getStreamName(event.Source, event.EventType)

	// Create queue message
	queueMsg := &models.QueueMessage{
		EventID:   event.ID,
		Source:    event.Source,
		EventType: event.EventType,
		Timestamp: event.ReceivedAt,
		Data:      event.RawData,
	}

	// Marshal to JSON
	msgData, err := json.Marshal(queueMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal queue message: %w", err)
	}

	// Publish to stream
	_, err = rp.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"event": string(msgData),
		},
	}).Result()

	if err != nil {
		rp.logger.Error("Failed to publish event", zap.Error(err), zap.String("stream", streamName))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	rp.logger.Info("Event published", zap.String("event_id", event.ID), zap.String("stream", streamName))
	return nil
}

// GetStreamName returns the stream name for an event
func (rp *RedisProducer) getStreamName(source models.EventSource, eventType models.EventType) string {
	return fmt.Sprintf("events.%s.%s", source, eventType)
}

// Close closes the Redis connection
func (rp *RedisProducer) Close() error {
	return rp.client.Close()
}

// GetStreamStats returns statistics about streams
func (rp *RedisProducer) GetStreamStats(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	// Get all streams starting with "events."
	keys, err := rp.client.Keys(ctx, "events.*").Result()
	if err != nil {
		rp.logger.Error("Failed to get stream stats", zap.Error(err))
		return stats
	}

	for _, key := range keys {
		info, err := rp.client.XLen(ctx, key).Result()
		if err != nil {
			continue
		}
		stats[key] = info
	}

	return stats
}
