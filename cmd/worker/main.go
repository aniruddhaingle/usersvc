package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"usersvc/internal/config"
	"usersvc/internal/infrastructure"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	//rbmq only for wrkr
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
	msgs, err := rabbitClient.Consume()
	if err != nil {
		slog.Error("failed to start consumer", "error", err)
		os.Exit(1)
	}
	//chan to handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
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
			//simulating work bcz i am too lazy to implement actual logic for mail
			time.Sleep(500 * time.Millisecond)

			//a safe manula ack
			d.Ack(false)
		}
	}()
	slog.Info("worker started, waiting for messages")
	<-stop
	slog.Info("shutting down worker...")
	//close conn to stop cons
	rabbitClient.Close()
	//wait for proc to fin
	wg.Wait()
	slog.Info("worker stopped gracefully")
}
