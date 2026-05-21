package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/taskforge/taskforge/config"
	domainhandler "github.com/taskforge/taskforge/domain/handler"
	"github.com/taskforge/taskforge/infrastructure/postgres"
	"github.com/taskforge/taskforge/infrastructure/redis"
	"github.com/taskforge/taskforge/repository"
	"github.com/taskforge/taskforge/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	pg, err := postgres.NewPostgres()
	if err != nil {
		slog.Error("postgres", "error", err)
		os.Exit(1)
	}
	defer pg.Close()

	rdb, err := redis.NewRedis()
	if err != nil {
		slog.Error("redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	repo := repository.NewRepository(pg, rdb)
	handlers := map[string]domainhandler.JobHandler{
		"ping": worker.PingHandler{},
	}

	runner := worker.NewRunner(repo, cfg.Worker.ConsumerName, handlers)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runner.Run(ctx); err != nil && ctx.Err() == nil {
		slog.Error("worker", "error", err)
		os.Exit(1)
	}
}
