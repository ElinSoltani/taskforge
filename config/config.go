package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	pgcfg "github.com/taskforge/taskforge/infrastructure/postgres"
	rediscfg "github.com/taskforge/taskforge/infrastructure/redis"
)

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Worker   WorkerConfig
}

type HTTPConfig struct {
	Addr    string
	BaseURL string
}

type PostgresConfig struct {
	Host          string
	Port          int
	Username      string
	Password      string
	Database      string
	MigrationPath string
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
		HTTP: HTTPConfig{
			Addr:    env("HTTP_ADDR", ":8080"),
			BaseURL: env("HTTP_BASE_URL", "http://localhost:8080"),
		},
		Postgres: PostgresConfig{
			Host:          env("POSTGRES_HOST", "postgres"),
			Port:          envInt("POSTGRES_PORT", 5432),
			Username:      env("POSTGRES_USER", "taskforge"),
			Password:      env("POSTGRES_PASSWORD", "taskforge"),
			Database:      env("POSTGRES_DB", "taskforge"),
			MigrationPath: env("POSTGRES_MIGRATION_PATH", "file://migrations"),
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
	if cfg.Postgres.Host == "" || cfg.Postgres.Database == "" {
		return nil, fmt.Errorf("postgres config incomplete")
	}
	applyInfrastructureConfig(cfg)
	return cfg, nil
}

func applyInfrastructureConfig(cfg *Config) {
	pgcfg.SetConfig(pgcfg.Config{
		Host:          cfg.Postgres.Host,
		Port:          cfg.Postgres.Port,
		Username:      cfg.Postgres.Username,
		Password:      cfg.Postgres.Password,
		Database:      cfg.Postgres.Database,
		MigrationPath: cfg.Postgres.MigrationPath,
	})
	rediscfg.SetConfig(rediscfg.Config{
		Addr:          cfg.Redis.Addr,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.DB,
		Stream:        cfg.Redis.Stream,
		ConsumerGroup: cfg.Redis.ConsumerGroup,
		ConsumerName:  cfg.Worker.ConsumerName,
		BlockTimeout:  int(cfg.Worker.BlockTimeout / time.Millisecond),
	})
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
