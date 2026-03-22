package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/digital-memory/api-service/internal/models"
)

// PostgresDB represents PostgreSQL connection
type PostgresDB struct {
	conn   *sql.DB
	logger *zap.Logger
}

// NewPostgresDB creates new PostgreSQL connection
func NewPostgresDB(dbURL string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger, _ := zap.NewProduction()
	return &PostgresDB{conn: db, logger: logger}, nil
}

// GetKnowledgeByID retrieves knowledge by ID
func (db *PostgresDB) GetKnowledgeByID(knowledgeID string) (*models.QueryResult, error) {
	query := `
		SELECT k.id, k.summary, k.raw_text, k.tags, k.decisions,
		       e.source, e.author, e.channel, k.created_at
		FROM knowledge k
		JOIN events e ON k.event_id = e.id
		WHERE k.id = $1
	`

	result := &models.QueryResult{}
	var tagsJSON, decisionsJSON []byte

	err := db.conn.QueryRow(query, knowledgeID).Scan(
		&result.ID, &result.Summary, &result.RawText, &tagsJSON, &decisionsJSON,
		&result.Source, &result.Author, &result.Channel, &result.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge: %w", err)
	}

	// Unmarshal JSON fields
	if tagsJSON != nil {
		json.Unmarshal(tagsJSON, &result.Tags)
	}
	if decisionsJSON != nil {
		json.Unmarshal(decisionsJSON, &result.Decisions)
	}

	return result, nil
}

// SearchByIDs retrieves multiple knowledge items and sorts by similarity
func (db *PostgresDB) SearchByIDs(knowledgeIDs []string, scores map[string]float64) ([]models.QueryResult, error) {
	if len(knowledgeIDs) == 0 {
		return []models.QueryResult{}, nil
	}

	query := `
		SELECT k.id, k.summary, k.raw_text, k.tags, k.decisions,
		       e.source, e.author, e.channel, k.created_at
		FROM knowledge k
		JOIN events e ON k.event_id = e.id
		WHERE k.id = ANY($1)
	`

	rows, err := db.conn.Query(query, pq.Array(knowledgeIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge: %w", err)
	}
	defer rows.Close()

	results := make([]models.QueryResult, 0)

	for rows.Next() {
		result := models.QueryResult{}
		var tagsJSON, decisionsJSON []byte

		err := rows.Scan(
			&result.ID, &result.Summary, &result.RawText, &tagsJSON, &decisionsJSON,
			&result.Source, &result.Author, &result.Channel, &result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan knowledge: %w", err)
		}

		if tagsJSON != nil {
			json.Unmarshal(tagsJSON, &result.Tags)
		}
		if decisionsJSON != nil {
			json.Unmarshal(decisionsJSON, &result.Decisions)
		}

		// Add similarity score
		result.SimilarityScore = scores[result.ID]
		results = append(results, result)
	}

	return results, rows.Err()
}

// GetEventHistory returns recent events
func (db *PostgresDB) GetEventHistory(limit, offset int) ([]models.HistoryEvent, int, error) {
	countQuery := `SELECT COUNT(*) FROM events`
	var total int
	if err := db.conn.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, source, event_type, author, channel, processing_status, received_at, processed_at
		FROM events
		ORDER BY received_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	events := make([]models.HistoryEvent, 0)
	for rows.Next() {
		event := models.HistoryEvent{}
		err := rows.Scan(
			&event.ID, &event.Source, &event.EventType, &event.Author, &event.Channel,
			&event.ProcessingStatus, &event.ReceivedAt, &event.ProcessedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, total, rows.Err()
}

// GetEntities retrieves all entities
func (db *PostgresDB) GetEntities(limit, offset int) ([]models.Entity, int, error) {
	countQuery := `SELECT COUNT(*) FROM entities`
	var total int
	if err := db.conn.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, name, entity_type, description, first_mentioned_at, last_mentioned_at, mention_count
		FROM entities
		ORDER BY mention_count DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query entities: %w", err)
	}
	defer rows.Close()

	entities := make([]models.Entity, 0)
	for rows.Next() {
		entity := models.Entity{}
		err := rows.Scan(
			&entity.ID, &entity.Name, &entity.Type, &entity.Description,
			&entity.FirstMentionedAt, &entity.LastMentionedAt, &entity.MentionCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan entity: %w", err)
		}
		entities = append(entities, entity)
	}

	return entities, total, rows.Err()
}

// GetEntityDetails retrieves details about a specific entity
func (db *PostgresDB) GetEntityDetails(name string) (*models.Entity, error) {
	query := `
		SELECT id, name, entity_type, description, first_mentioned_at, last_mentioned_at, mention_count
		FROM entities
		WHERE name = $1
	`

	entity := &models.Entity{}
	err := db.conn.QueryRow(query, name).Scan(
		&entity.ID, &entity.Name, &entity.Type, &entity.Description,
		&entity.FirstMentionedAt, &entity.LastMentionedAt, &entity.MentionCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return entity, nil
}

// GetDBStats returns database statistics
func (db *PostgresDB) GetDBStats() models.DBStats {
	stats := models.DBStats{}

	// Total events
	db.conn.QueryRow("SELECT COUNT(*) FROM events").Scan(&stats.TotalEvents)

	// Processed events
	db.conn.QueryRow("SELECT COUNT(*) FROM events WHERE processing_status = 'completed'").Scan(&stats.ProcessedEvents)

	// Total knowledge
	db.conn.QueryRow("SELECT COUNT(*) FROM knowledge").Scan(&stats.TotalKnowledge)

	// With embeddings
	db.conn.QueryRow("SELECT COUNT(*) FROM knowledge WHERE embedding IS NOT NULL").Scan(&stats.WithEmbeddings)

	return stats
}

// RecordQuery logs a query for analytics
func (db *PostgresDB) RecordQuery(queryText string, topK, resultsCount, responseTimeMs int) error {
	query := `
		INSERT INTO queries (query_text, top_k, results_count, response_time_ms)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.conn.Exec(query, queryText, topK, resultsCount, responseTimeMs)
	return err
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Query executes a query and returns rows
func (db *PostgresDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow executes a query that returns a single row
func (db *PostgresDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}
