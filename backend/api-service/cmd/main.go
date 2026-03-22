package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/digital-memory/api-service/internal/database"
	"github.com/digital-memory/api-service/internal/handlers"
	"github.com/digital-memory/api-service/internal/middleware"
	"github.com/digital-memory/api-service/internal/vector_db"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Get configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatal("DATABASE_URL environment variable not set")
	}

	// Initialize database
	logger.Info("Initializing database connection", zap.String("url", dbURL))
	db, err := database.NewPostgresDB(dbURL)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize vector DB
	logger.Info("Initializing vector database")
	vectorDB, err := vector_db.NewPgVectorDB(db, logger)
	if err != nil {
		logger.Fatal("Failed to initialize vector DB", zap.Error(err))
	}

	// Create Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.ErrorHandlingMiddleware())
	router.Use(middleware.RateLimitMiddleware())

	// Initialize handlers
	handlerService := handlers.NewQueryHandler(db, vectorDB, logger)
	registerRoutes(router, handlerService)

	// Start server
	logger.Info("Starting API service", zap.String("port", port))

	go func() {
		if err := router.Run(":" + port); err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down API service")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := shutdownGracefully(ctx, logger); err != nil {
		logger.Error("Shutdown error", zap.Error(err))
	}
}

func registerRoutes(router *gin.Engine, handler *handlers.QueryHandler) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// Status
	router.GET("/status", handler.Status)

	// Metrics
	router.GET("/metrics", handler.Metrics)

	// API routes
	api := router.Group("/api/v1")
	{
		// Query endpoint
		api.POST("/query", handler.Query)

		// History endpoint
		api.GET("/history", handler.History)

		// Entities endpoint
		api.GET("/entities", handler.GetEntities)

		// Entity details
		api.GET("/entities/:name", handler.GetEntityDetails)
	}
}

func shutdownGracefully(ctx context.Context, logger *zap.Logger) error {
	select {
	case <-time.After(5 * time.Second):
		return fmt.Errorf("shutdown timeout exceeded")
	case <-ctx.Done():
		return ctx.Err()
	}
}
