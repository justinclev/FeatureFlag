package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/featureflags/feature-api/internal/cache"
	"github.com/featureflags/feature-api/internal/config"
	"github.com/featureflags/feature-api/internal/db"
	"github.com/featureflags/feature-api/internal/evaluator"
	"github.com/featureflags/feature-api/internal/handlers"
	"github.com/featureflags/feature-api/internal/middleware"
	"github.com/featureflags/feature-api/internal/repository"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	mongoClient, database, err := db.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect to MongoDB", "error", err)
		return err
	}
	defer db.Disconnect(mongoClient)

	redisClient, err := cache.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		return err
	}
	defer cache.Close(redisClient)

	eval := evaluator.New()
	repo := repository.NewMongoRedisRepository(database.Collection(cfg.MongoCollectionName), redisClient, logger, cfg.CacheTTL, cfg.RedisCachePrefix)
	h := handlers.New(repo, logger, eval, cfg.RequestTimeout)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Middleware chain: Recovery -> BodyLimit -> Auth -> CORS -> Logging
	handler := middleware.Chain(mux,
		middleware.Recovery(logger),
		middleware.BodyLimit(1<<20), // 1MB
		middleware.APIKeyAuth(cfg.APIKey, logger),
		middleware.CORS(cfg.CORSAllowedOrigin),
		middleware.Logging(logger),
	)

	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("feature-api started", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			stop <- syscall.SIGINT
		}
	}()

	<-stop
	logger.Info("shutting down gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server first to stop receiving new requests
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}

	// Now close dependencies using the same shutdown context (polite cleanup)
	db.Disconnect(mongoClient)
	cache.Close(redisClient)

	logger.Info("server stopped")
	return nil
}
