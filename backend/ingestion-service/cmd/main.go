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

	"github.com/digital-memory/ingestion-service/internal/database"
	"github.com/digital-memory/ingestion-service/internal/handlers"
	"github.com/digital-memory/ingestion-service/internal/middleware"
	"github.com/digital-memory/ingestion-service/internal/queue"
)

func init() {
	// Load .env file if it exists
	_ = godotenv.Load("../../.env")
}

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatal("DATABASE_URL environment variable not set")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	// Initialize database
	logger.Info("Initializing database connection", zap.String("url", dbURL))
	db, err := database.NewPostgresDB(dbURL)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis queue producer
	logger.Info("Initializing Redis queue", zap.String("url", redisURL))
	redisProducer, err := queue.NewRedisProducer(redisURL)
	if err != nil {
		logger.Fatal("Failed to initialize Redis queue", zap.Error(err))
	}
	defer redisProducer.Close()

	// Create Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.ErrorHandlingMiddleware())
	router.Use(middleware.RateLimitMiddleware())

	// Initialize handler and register routes
	handlerService := handlers.NewEventHandler(db, redisProducer, logger)
	registerRoutes(router, handlerService)

	// Start server
	logger.Info("Starting ingestion service", zap.String("port", port))

	go func() {
		if err := router.Run(":" + port); err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down ingestion service")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Graceful shutdown
	if err := shutdownGracefully(ctx, logger); err != nil {
		logger.Error("Shutdown error", zap.Error(err))
	}
}

func registerRoutes(router *gin.Engine, handler *handlers.EventHandler) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// Webhook endpoints
	api := router.Group("/webhook")
	{
		api.POST("/slack", handler.HandleSlackEvent)
		api.POST("/github", handler.HandleGitHubEvent)
	}

	// Metrics endpoint (Prometheus)
	router.GET("/metrics", handler.Metrics)

	// Status endpoint
	router.GET("/status", handler.Status)
}

func shutdownGracefully(ctx context.Context, logger *zap.Logger) error {
	select {
	case <-time.After(5 * time.Second):
		return fmt.Errorf("shutdown timeout exceeded")
	case <-ctx.Done():
		return ctx.Err()
	}
}
