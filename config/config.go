package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Worker   WorkerConfig
}

type HTTPConfig struct {
	Addr string
}

type PostgresConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr          string
	Password      string
	DB            int
	Stream        string
	ConsumerGroup string
}

type WorkerConfig struct {
	ConsumerName string
	BlockTimeout time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTP: HTTPConfig{Addr: env("HTTP_ADDR", ":8080")},
		Postgres: PostgresConfig{
			DSN: env("POSTGRES_DSN", "postgres://taskforge:taskforge@postgres:5432/taskforge?sslmode=disable"),
		},
		Redis: RedisConfig{
			Addr:          env("REDIS_ADDR", "redis:6379"),
			Password:      env("REDIS_PASSWORD", ""),
			DB:            envInt("REDIS_DB", 0),
			Stream:        env("REDIS_STREAM", "taskforge:queue:normal"),
			ConsumerGroup: env("REDIS_CONSUMER_GROUP", "taskforge-workers"),
		},
		Worker: WorkerConfig{
			ConsumerName: env("WORKER_CONSUMER_NAME", "worker-1"),
			BlockTimeout: envDuration("WORKER_BLOCK_TIMEOUT", 2*time.Second),
		},
	}
	if cfg.Postgres.DSN == "" {
		return nil, fmt.Errorf("POSTGRES_DSN required")
	}
	return cfg, nil
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}

func envDuration(k string, d time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if dur, err := time.ParseDuration(v); err == nil {
			return dur
		}
	}
	return d
}
