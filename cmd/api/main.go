package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"usersvc/internal/config"
	"usersvc/internal/infrastructure"
	"usersvc/internal/repository"
	"usersvc/internal/service"
	transport "usersvc/internal/transport/http"
	"usersvc/migrations"
)

func main() {
	//setup generic json logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	//db
	db, err := repository.NewDB(cfg.DBURL, migrations.FS)
	if err != nil {
		slog.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	//redis
	redisClient, err := infrastructure.NewRedisClient(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	//rbmq retry loop
	var rabbitClient *infrastructure.RabbitMQClient
	for i := 0; i < 10; i++ {
		rabbitClient, err = infrastructure.NewRabbitMQClient(cfg.RabbitMQURL)
		if err == nil {
			break
		}
		slog.Info("waiting for rabbitmq", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		slog.Error("failed to connect to rabbitmq after retries", "error", err)
		os.Exit(1)
	}
	defer rabbitClient.Close()

	//run the worker inside the api when the host has no separate worker process (eg render free tier)
	if os.Getenv("EMBED_WORKER") == "true" {
		msgs, err := rabbitClient.Consume()
		if err != nil {
			slog.Error("failed to start embedded consumer", "error", err)
			os.Exit(1)
		}
		go func() {
			for d := range msgs {
				var event map[string]interface{}
				if err := json.Unmarshal(d.Body, &event); err != nil {
					slog.Error("failed to unmarshal message", "error", err, "body", string(d.Body))
					d.Nack(false, false)
					continue
				}
				slog.Info("Event processed",
					"user_id", event["user_id"],
					"queue_name", "user_events",
				)
				d.Ack(false)
			}
		}()
		slog.Info("embedded worker started")
	}

	//wiringg
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, redisClient, rabbitClient)
	handler := transport.NewHandler(userService)

	//router
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	//mdlwr chain
	wrappedRouter := transport.LoggingMiddleware(
		transport.RecoveryMiddleware(mux),
	)

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: wrappedRouter}

	//chan to handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("starting server", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}
	slog.Info("server stopped gracefully")
}
