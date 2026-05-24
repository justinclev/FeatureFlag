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
	repo := repository.NewMongoRedisRepository(database.Collection(cfg.MongoCollectionName), redisClient, cfg.CacheTTL, cfg.RedisCachePrefix)
	h := handlers.New(repo, logger, eval, cfg.RequestTimeout)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Principal optimization: Reorder middleware for maximum safety and observability.
	// 1. Recovery is at the top to catch any panic.
	// 2. Auth is next to fail fast and log unauthorized attempts.
	// 3. CORS and Logging follow.
	handler := middleware.Recovery(logger,
		middleware.APIKeyAuth(cfg.APIKey, logger,
			middleware.CORS(cfg.CORSAllowedOrigin,
				middleware.Logging(logger, mux),
			),
		),
	)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
		// Principal Security: Limit max header size to 1MB to prevent resource exhaustion.
		MaxHeaderBytes: 1 << 20,
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		return err
	}
	logger.Info("server stopped")
	return nil
}
