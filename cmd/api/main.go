package main

import (
	"adv-bknd/internal/config"
	"adv-bknd/internal/infrastructure"
	"adv-bknd/internal/repository"
	"adv-bknd/internal/service"
	transport "adv-bknd/internal/transport/http"
	"adv-bknd/migrations"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
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

	slog.Info("starting server", "port", cfg.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.HTTPPort, wrappedRouter); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
