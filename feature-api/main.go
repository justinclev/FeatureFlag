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
	"github.com/featureflags/feature-api/internal/handlers"
	"github.com/featureflags/feature-api/internal/middleware"
	"github.com/featureflags/feature-api/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid config", "error", err)
		os.Exit(1)
	}

	mongoClient, database, err := db.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer db.Disconnect(mongoClient)

	redisClient, err := cache.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer cache.Close(redisClient)

	repo := repository.NewMongoRedisRepository(database.Collection("flags"), redisClient)
	h := handlers.New(repo, logger)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.CORS(middleware.Logging(logger, mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("feature-api started", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}
