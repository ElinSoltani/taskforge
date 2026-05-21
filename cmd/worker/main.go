package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/taskforge/taskforge/config"
	"github.com/taskforge/taskforge/domain"
	"github.com/taskforge/taskforge/infrastructure/postgres"
	redisq "github.com/taskforge/taskforge/infrastructure/redis"
	"github.com/taskforge/taskforge/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	store, err := postgres.New(cfg.Postgres.DSN)
	if err != nil {
		slog.Error("postgres", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	queue := redisq.NewQueue(cfg)
	defer queue.Client().Close()

	handlers := map[string]domain.JobHandler{
		"ping": worker.PingHandler{},
	}

	runner := worker.NewRunner(store, queue, cfg.Worker.ConsumerName, handlers)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runner.Run(ctx); err != nil && ctx.Err() == nil {
		slog.Error("worker", "error", err)
		os.Exit(1)
	}
}
