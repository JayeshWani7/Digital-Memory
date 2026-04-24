package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/digital-memory/ingestion-service/internal/models"
)

// KafkaProducer handles publishing messages to Kafka topics
type KafkaProducer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

// NewKafkaProducer creates a new Kafka producer with retry support
func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("at least one Kafka broker is required")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		MaxAttempts:  1, // retries handled by publishWithRetry to avoid compounded attempts
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		BatchTimeout: 10 * time.Millisecond,
		Async:        false,
	}

	logger, _ := zap.NewProduction()

	producer := &KafkaProducer{
		writer: writer,
		logger: logger,
	}

	// Test connection against all brokers until one responds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastErr error
	connectedBroker := ""
	for _, broker := range brokers {
		if err := producer.ping(ctx, broker); err == nil {
			connectedBroker = broker
			break
		} else {
			lastErr = err
		}
	}
	if connectedBroker == "" {
		return nil, fmt.Errorf("failed to connect to any Kafka broker: %w", lastErr)
	}

	logger.Info("Connected to Kafka",
		zap.String("connected_broker", connectedBroker),
		zap.Strings("brokers", brokers),
	)
	return producer, nil
}

// ping checks Kafka connectivity
func (kp *KafkaProducer) ping(ctx context.Context, broker string) error {
	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

// PublishEvent publishes an event to the appropriate Kafka topic
func (kp *KafkaProducer) PublishEvent(event *models.Event) error {
	return kp.publishWithRetry(event, 3)
}

// publishWithRetry attempts to publish with exponential backoff
func (kp *KafkaProducer) publishWithRetry(event *models.Event, maxRetries int) error {
	topic := kp.getTopicName(event.Source, event.EventType)

	queueMsg := &models.KafkaMessage{
		EventID:       event.ID,
		Source:        event.Source,
		EventType:     event.EventType,
		Timestamp:     event.ReceivedAt,
		Data:          event.RawData,
		CorrelationID: event.ID,
		Version:       "1.0",
	}

	msgData, err := json.Marshal(queueMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal kafka message: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			kp.logger.Warn("Retrying Kafka publish",
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", backoff),
				zap.String("event_id", event.ID),
			)
			time.Sleep(backoff)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = kp.writer.WriteMessages(ctx, kafka.Message{
			Topic: topic,
			Key:   []byte(event.ID),
			Value: msgData,
		})
		cancel()

		if err == nil {
			kp.logger.Info("Event published to Kafka",
				zap.String("event_id", event.ID),
				zap.String("topic", topic),
			)
			return nil
		}
		lastErr = err
	}

	kp.logger.Error("Failed to publish event after retries",
		zap.Error(lastErr),
		zap.String("event_id", event.ID),
		zap.String("topic", topic),
	)
	return fmt.Errorf("failed to publish event after %d attempts: %w", maxRetries, lastErr)
}

// getTopicName returns the Kafka topic name for an event
func (kp *KafkaProducer) getTopicName(source models.EventSource, eventType models.EventType) string {
	return fmt.Sprintf("events.%s.%s", source, eventType)
}

// GetTopicStats returns the list of known topics for monitoring
func (kp *KafkaProducer) GetTopicStats(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})
	sources := []models.EventSource{models.SourceSlack, models.SourceGitHub}
	eventTypes := []models.EventType{
		models.EventMessage,
		models.EventPRCreated,
		models.EventPRUpdated,
		models.EventCommit,
	}

	for _, source := range sources {
		for _, eventType := range eventTypes {
			topic := kp.getTopicName(source, eventType)
			stats[topic] = "active"
		}
	}
	return stats
}

// Close closes the Kafka writer
func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}