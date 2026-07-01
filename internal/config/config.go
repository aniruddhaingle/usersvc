package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBURL       string
	RedisURL    string
	RabbitMQURL string
	HTTPPort    string
}

// Why did i use "method muation" vs "factory functions"
// - ensure that the cfg now doesn't have any half upated config \n
// - its cleanr to init a var as a result of func calling
// -
func Load() (*Config, error) {
	cfg := &Config{
		DBURL:       os.Getenv("DB_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		RabbitMQURL: os.Getenv("RABBITMQ_URL"),
		HTTPPort:    os.Getenv("HTTP_PORT"),
	}

	if cfg.DBURL == "" {
		return nil, fmt.Errorf("dburl is needed")
	}
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("redisurl needed")
	}
	if cfg.RabbitMQURL == "" {
		return nil, fmt.Errorf("rabbitmq url is needed")
	}
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8080"
	}
	return cfg, nil
}
