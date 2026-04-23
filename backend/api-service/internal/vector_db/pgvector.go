package vector_db

import (
	"fmt"
	"sort"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/digital-memory/api-service/internal/database"
	"github.com/digital-memory/api-service/internal/models"
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
func (pvdb *PgVectorDB) SearchSimilar(embedding []float32, topK int) ([]models.SearchMatch, error) {
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

	results := make([]models.SearchMatch, 0, topK)

	for rows.Next() {
		var knowledgeID string
		var similarity float64

		if err := rows.Scan(&knowledgeID, &similarity); err != nil {
			pvdb.logger.Error("Failed to scan result", zap.Error(err))
			continue
		}

		results = append(results, models.SearchMatch{
			ID:    knowledgeID,
			Score: similarity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate results: %w", err)
	}

	sortSearchMatches(results)

	return results, nil
}

// sortSearchMatches keeps search results ordered by descending similarity score.
// Ties fall back to the knowledge ID so API responses stay deterministic.
func sortSearchMatches(results []models.SearchMatch) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].ID < results[j].ID
		}
		return results[i].Score > results[j].Score
	})
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
