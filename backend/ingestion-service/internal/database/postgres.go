package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/digital-memory/ingestion-service/internal/models"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	conn   *sql.DB
	logger *zap.Logger
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(dbURL string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger, _ := zap.NewProduction()
	return &PostgresDB{conn: db, logger: logger}, nil
}

// StoreEvent stores a raw event in the database
func (db *PostgresDB) StoreEvent(event *models.Event) error {
	rawDataJSON, err := json.Marshal(event.RawData)
	if err != nil {
		return fmt.Errorf("failed to marshal raw data: %w", err)
	}

	query := `
		INSERT INTO events 
		(id, source, source_id, event_type, raw_data, author, channel, received_at, processing_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (source, source_id) DO UPDATE SET
		raw_data = $5,
		author = $6,
		channel = $7
		RETURNING id
	`

	var returnedID string
	err = db.conn.QueryRow(
		query,
		event.ID,
		event.Source,
		event.SourceID,
		event.EventType,
		rawDataJSON,
		event.Author,
		event.Channel,
		event.ReceivedAt,
		models.StatusPending,
	).Scan(&returnedID)

	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	event.ID = returnedID
	return nil
}

// GetEventBySourceID retrieves an event by source and source ID
func (db *PostgresDB) GetEventBySourceID(source models.EventSource, sourceID string) (*models.Event, error) {
	query := `
		SELECT id, source, source_id, event_type, raw_data, author, channel, 
		       received_at, processing_status, processed_at, error_message, error_count
		FROM events
		WHERE source = $1 AND source_id = $2
	`

	event := &models.Event{}
	var rawDataJSON []byte

	err := db.conn.QueryRow(query, source, sourceID).Scan(
		&event.ID, &event.Source, &event.SourceID, &event.EventType,
		&rawDataJSON, &event.Author, &event.Channel, &event.ReceivedAt,
		&event.ProcessingStatus, &event.ProcessedAt, &event.ErrorMessage, &event.ErrorCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if err := json.Unmarshal(rawDataJSON, &event.RawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw data: %w", err)
	}

	return event, nil
}

// UpdateEventStatus updates the processing status of an event
func (db *PostgresDB) UpdateEventStatus(eventID string, status models.ProcessingStatus) error {
	query := `UPDATE events SET processing_status = $1, processed_at = $2 WHERE id = $3`
	_, err := db.conn.Exec(query, status, time.Now(), eventID)
	return err
}

// RecordEventError records an error for an event
func (db *PostgresDB) RecordEventError(eventID string, errMsg string) error {
	query := `
		UPDATE events 
		SET error_message = $1, error_count = error_count + 1, last_error_at = $2
		WHERE id = $3 AND error_count < 5
	`
	_, err := db.conn.Exec(query, errMsg, time.Now(), eventID)
	return err
}

// GetPendingEvents retrieves events that are pending processing
func (db *PostgresDB) GetPendingEvents(limit int) ([]*models.Event, error) {
	query := `
		SELECT id, source, source_id, event_type, raw_data, author, channel, 
		       received_at, processing_status, processed_at, error_message, error_count
		FROM events
		WHERE processing_status = $1
		ORDER BY received_at ASC
		LIMIT $2
	`

	rows, err := db.conn.Query(query, models.StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		var rawDataJSON []byte

		err := rows.Scan(
			&event.ID, &event.Source, &event.SourceID, &event.EventType,
			&rawDataJSON, &event.Author, &event.Channel, &event.ReceivedAt,
			&event.ProcessingStatus, &event.ProcessedAt, &event.ErrorMessage, &event.ErrorCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending event: %w", err)
		}

		if err := json.Unmarshal(rawDataJSON, &event.RawData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw data: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// GetEventStats returns statistics about events
func (db *PostgresDB) GetEventStats() map[string]interface{} {
	query := `
		SELECT 
			source,
			event_type,
			processing_status,
			COUNT(*) as count
		FROM events
		GROUP BY source, event_type, processing_status
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		db.logger.Error("Failed to get event stats", zap.Error(err))
		return map[string]interface{}{}
	}
	defer rows.Close()

	stats := make(map[string]interface{})
	for rows.Next() {
		var source, eventType, status string
		var count int
		if err := rows.Scan(&source, &eventType, &status, &count); err != nil {
			db.logger.Error("Failed to scan stats", zap.Error(err))
			continue
		}
		key := fmt.Sprintf("%s_%s_%s", source, eventType, status)
		stats[key] = count
	}

	return stats
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}
