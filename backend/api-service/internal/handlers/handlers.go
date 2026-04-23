package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/digital-memory/api-service/internal/database"
	"github.com/digital-memory/api-service/internal/models"
	"github.com/digital-memory/api-service/internal/vector_db"
)

// QueryHandler handles query-related endpoints
type QueryHandler struct {
	db       *database.PostgresDB
	vectorDB *vector_db.PgVectorDB
	logger   *zap.Logger
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(db *database.PostgresDB, vectorDB *vector_db.PgVectorDB, logger *zap.Logger) *QueryHandler {
	return &QueryHandler{
		db:       db,
		vectorDB: vectorDB,
		logger:   logger,
	}
}

// HealthCheck returns service health
func (qh *QueryHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}

// Status returns service status
func (qh *QueryHandler) Status(c *gin.Context) {
	dbStats := qh.db.GetDBStats()
	_, _ = qh.vectorDB.GetEmbeddingStats()

	response := models.StatusResponse{
		Status: "operational",
		Uptime: time.Now().Unix(),
		DB:     dbStats,
	}

	c.JSON(http.StatusOK, response)
}

// Metrics returns service metrics
func (qh *QueryHandler) Metrics(c *gin.Context) {
	stats := qh.db.GetDBStats()

	c.JSON(http.StatusOK, gin.H{
		"total_events":       stats.TotalEvents,
		"processed_events":   stats.ProcessedEvents,
		"total_knowledge":    stats.TotalKnowledge,
		"with_embeddings":    stats.WithEmbeddings,
		"embedding_coverage": float64(stats.WithEmbeddings) / float64(stats.TotalKnowledge) * 100,
	})
}

// Query performs a semantic search query
func (qh *QueryHandler) Query(c *gin.Context) {
	startTime := time.Now()

	var req models.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		qh.logger.Warn("Invalid query request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Default topK
	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.TopK > 50 {
		req.TopK = 50
	}

	// For MVP: Placeholder for actual embedding generation
	// In production, call embedding API or use local embeddings
	embedding, err := qh.generateDummyEmbedding(req.Query)
	if err != nil {
		qh.logger.Error("Failed to generate embedding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "embedding generation failed"})
		return
	}

	// Search similar embeddings
	scores, err := qh.vectorDB.SearchSimilar(embedding, req.TopK)
	if err != nil {
		qh.logger.Error("Failed to search embeddings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	// Hydrate the ranked matches without changing the response schema.
	results, err := qh.db.SearchByIDs(scores)
	if err != nil {
		qh.logger.Error("Failed to fetch results", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed"})
		return
	}

	// Record query for analytics
	duration := time.Since(startTime)
	go func() {
		_ = qh.db.RecordQuery(req.Query, req.TopK, len(results), int(duration.Milliseconds()))
	}()

	response := models.QueryResponse{
		Query:    req.Query,
		Results:  results,
		Count:    len(results),
		Duration: duration.String(),
	}

	c.JSON(http.StatusOK, response)
}

// History returns event history
func (qh *QueryHandler) History(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	events, total, err := qh.db.GetEventHistory(limit, offset)
	if err != nil {
		qh.logger.Error("Failed to get history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
		return
	}

	response := models.HistoryResponse{
		Events: events,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}

	c.JSON(http.StatusOK, response)
}

// GetEntities returns all entities
func (qh *QueryHandler) GetEntities(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 200 {
		limit = 200
	}

	entities, total, err := qh.db.GetEntities(limit, offset)
	if err != nil {
		qh.logger.Error("Failed to get entities", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entities": entities,
		"total":    total,
		"offset":   offset,
		"limit":    limit,
	})
}

// GetEntityDetails returns details about a specific entity
func (qh *QueryHandler) GetEntityDetails(c *gin.Context) {
	name := c.Param("name")

	entity, err := qh.db.GetEntityDetails(name)
	if err != nil {
		qh.logger.Error("Failed to get entity details", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entity"})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// Helper function to generate dummy embedding
// In production, call OpenAI or use local embedding model
func (qh *QueryHandler) generateDummyEmbedding(text string) ([]float32, error) {
	// For MVP, return a fixed-size random embedding
	// In production, call OpenAI embedding API or use sentence-transformers
	embedding := make([]float32, 1536) // OpenAI embedding size

	for i := range embedding {
		// Simple hash-based embedding (not cryptographically secure)
		h := 0
		for _, c := range text {
			h = ((h << 5) - h) + int(c)
		}
		seed := h + i
		// Simple pseudo-random number
		embedding[i] = float32((seed*1103515245+12345)%2147483648) / 2147483648.0
	}

	return embedding, nil
}
