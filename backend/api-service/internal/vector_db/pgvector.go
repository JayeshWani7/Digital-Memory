package vector_db

import (
	"fmt"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/digital-memory/api-service/internal/database"
)

// PgVectorDB handles vector similarity search using PostgreSQL pgvector extension
type PgVectorDB struct {
	db     *database.PostgresDB
	logger *zap.Logger
}

// NewPgVectorDB creates a new PgVector database handler
func NewPgVectorDB(db *database.PostgresDB, logger *zap.Logger) (*PgVectorDB, error) {
	return &PgVectorDB{
		db:     db,
		logger: logger,
	}, nil
}

// SearchSimilar performs similarity search on embeddings
func (pvdb *PgVectorDB) SearchSimilar(embedding []float32, topK int) (map[string]float64, error) {
	// Convert embedding to PostgreSQL format
	embeddingStr := pq.Float64Array(embeddingToFloat64(embedding))

	query := `
		SELECT k.id, 1 - (k.embedding <=> $1::vector) as similarity_score
		FROM knowledge k
		WHERE k.embedding IS NOT NULL
		ORDER BY k.embedding <=> $1::vector
		LIMIT $2
	`

	rows, err := pvdb.db.Query(query, embeddingStr, topK)
	if err != nil {
		pvdb.logger.Error("Failed to search similar embeddings", zap.Error(err))
		return nil, fmt.Errorf("failed to search embeddings: %w", err)
	}
	defer rows.Close()

	results := make(map[string]float64)

	for rows.Next() {
		var knowledgeID string
		var similarity float64

		if err := rows.Scan(&knowledgeID, &similarity); err != nil {
			pvdb.logger.Error("Failed to scan result", zap.Error(err))
			continue
		}

		results[knowledgeID] = similarity
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate results: %w", err)
	}

	return results, nil
}

// GetEmbeddingStats returns statistics about embeddings
func (pvdb *PgVectorDB) GetEmbeddingStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count total embeddings
	var totalCount int
	if err := pvdb.db.QueryRow("SELECT COUNT(*) FROM knowledge WHERE embedding IS NOT NULL").Scan(&totalCount); err != nil {
		return nil, err
	}
	stats["total"] = totalCount

	// Count pending embeddings
	var pendingCount int
	if err := pvdb.db.QueryRow("SELECT COUNT(*) FROM knowledge WHERE embedding IS NULL").Scan(&pendingCount); err != nil {
		return nil, err
	}
	stats["pending"] = pendingCount

	// Get coverage percentage
	var allCount int
	if err := pvdb.db.QueryRow("SELECT COUNT(*) FROM knowledge").Scan(&allCount); err != nil {
		return nil, err
	}

	if allCount > 0 {
		coverage := float64(totalCount) / float64(allCount) * 100
		stats["coverage_percent"] = coverage
	}

	return stats, nil
}

// Helper to convert float32 slice to float64 slice
func embeddingToFloat64(embedding []float32) []float64 {
	result := make([]float64, len(embedding))
	for i, v := range embedding {
		result[i] = float64(v)
	}
	return result
}
