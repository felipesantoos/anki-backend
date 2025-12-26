package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/infra/database"
	"github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/logger"
	// Uncomment to enable automatic migrations on startup
	"github.com/felipesantos/anki-backend/pkg/migrate"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		// If unable to load config, use default logger
		logger.InitLogger("INFO", "development")
		log := logger.GetLogger()
		log.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Initialize logger with configuration
	logger.InitLogger(cfg.Logger.Level, cfg.Logger.Environment)
	log := logger.GetLogger()

	log.Info("Application starting",
		"environment", cfg.Logger.Environment,
		"log_level", cfg.Logger.Level,
	)

	// 3. Initialize database connection
	db, err := database.NewDatabase(cfg.Database, log)
	if err != nil {
		log.Error("Failed to initialize database connection", "error", err)
		os.Exit(1)
	}

	// Optional: Run migrations on startup
	// Uncomment the lines below to run migrations automatically
	// In production, it's usually better to run migrations manually via CI/CD
	//
	if err := migrate.RunMigrations(cfg.Database, log); err != nil {
	    log.Error("Failed to run migrations", "error", err)
	    os.Exit(1)
	}

	// 4. Initialize Redis connection
	rdb, err := redis.NewRedis(cfg.Redis, log)
	if err != nil {
		log.Error("Failed to initialize Redis connection", "error", err)
		// Redis is optional for some applications, but we'll treat it as an error
		// If optional, you can comment out os.Exit(1)
		os.Exit(1)
	}

	// TODO: Initialize HTTP router
	// TODO: Initialize services
	// TODO: Register routes
	// TODO: Start HTTP server

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Store db and redis in context or make them available to handlers
	_ = db  // TODO: Pass db to services/repositories
	_ = rdb // TODO: Pass redis to services

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutdown signal received, shutting down gracefully...")
		cancel()
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Close Redis connection gracefully
	log.Info("Closing Redis connection...")
	if err := rdb.Close(); err != nil {
		log.Error("Error closing Redis connection", "error", err)
	} else {
		log.Info("Redis connection closed successfully")
	}

	// Close database connection gracefully
	log.Info("Closing database connection...")
	if err := db.Close(); err != nil {
		log.Error("Error closing database connection", "error", err)
	} else {
		log.Info("Database connection closed successfully")
	}

	time.Sleep(1 * time.Second)
	log.Info("Server stopped")
}
