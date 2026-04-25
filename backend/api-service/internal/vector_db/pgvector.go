package vector_db

import (
	"fmt"
	"sort"

	"go.uber.org/zap"

	"github.com/digital-memory/api-service/internal/database"
	"github.com/digital-memory/api-service/internal/models"
)

// PgVectorDB handles vector similarity search using PostgreSQL.
// DEMO MODE: Uses pre-computed similarity_score column instead of pgvector
// so the API runs without the pgvector extension installed.
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

// SearchSimilar performs similarity search.
//
// DEMO MODE: Reads the pre-computed `similarity_score` column from the
// knowledge table and returns results sorted highest → lowest.
// This exercises the EXACT same sort.Slice ranking fix that was implemented
// to resolve the unordered-map bug, without requiring pgvector.
//
// In production this query would be replaced with a cosine-distance vector
// search: `ORDER BY embedding <=> $1::vector LIMIT $2`
func (pvdb *PgVectorDB) SearchSimilar(embedding []float32, topK int) ([]models.SearchMatch, error) {
	pvdb.logger.Info("[DEMO MODE] Running ranked search via similarity_score column",
		zap.Int("topK", topK))

	// Query reads pre-seeded scores — same data that flows into sort.Slice
	query := `
		SELECT id, similarity_score
		FROM knowledge
		ORDER BY similarity_score DESC
		LIMIT $1
	`

	rows, err := pvdb.db.Query(query, topK)
	if err != nil {
		pvdb.logger.Error("Failed to search knowledge", zap.Error(err))
		return nil, fmt.Errorf("failed to search knowledge: %w", err)
	}
	defer rows.Close()

	// -----------------------------------------------------------------------
	// THE FIX: results are collected into a []SearchMatch (slice), then
	// sorted with sort.Slice to guarantee descending order.
	//
	// BEFORE the fix: results were stored in a Go map[string]float64, which
	// has random iteration order — so the API response was non-deterministic.
	// AFTER the fix: sort.Slice ensures highest similarity is always first.
	// -----------------------------------------------------------------------
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

	// Apply sort — this is the core of the fix
	sortSearchMatches(results)

	pvdb.logger.Info("[DEMO MODE] Search complete",
		zap.Int("results", len(results)),
		zap.Any("ranking", func() []float64 {
			scores := make([]float64, len(results))
			for i, r := range results {
				scores[i] = r.Score
			}
			return scores
		}()))

	return results, nil
}

// sortSearchMatches sorts results by descending similarity score.
// Ties are broken by knowledge ID for deterministic ordering.
//
// This is the KEY FIX: replaces the old Go map (random iteration order)
// with an explicit sort that guarantees highest-similarity results first.
func sortSearchMatches(results []models.SearchMatch) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].ID < results[j].ID
		}
		return results[i].Score > results[j].Score
	})
}

// GetEmbeddingStats returns statistics about knowledge items
func (pvdb *PgVectorDB) GetEmbeddingStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalCount int
	if err := pvdb.db.QueryRow("SELECT COUNT(*) FROM knowledge WHERE similarity_score > 0").Scan(&totalCount); err != nil {
		return nil, err
	}
	stats["total"] = totalCount

	var pendingCount int
	if err := pvdb.db.QueryRow("SELECT COUNT(*) FROM knowledge WHERE similarity_score = 0").Scan(&pendingCount); err != nil {
		return nil, err
	}
	stats["pending"] = pendingCount

	var allCount int
	if err := pvdb.db.QueryRow("SELECT COUNT(*) FROM knowledge").Scan(&allCount); err != nil {
		return nil, err
	}

	if allCount > 0 {
		coverage := float64(totalCount) / float64(allCount) * 100
		stats["coverage_percent"] = coverage
	}

	stats["demo_mode"] = true

	return stats, nil
}

// embeddingToFloat64 converts float32 slice to float64 slice (kept for future use)
func embeddingToFloat64(embedding []float32) []float64 {
	result := make([]float64, len(embedding))
	for i, v := range embedding {
		result[i] = float64(v)
	}
	return result
}
