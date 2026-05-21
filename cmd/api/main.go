package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/taskforge/taskforge/config"
	"github.com/taskforge/taskforge/infrastructure/postgres"
	redisq "github.com/taskforge/taskforge/infrastructure/redis"
	"github.com/taskforge/taskforge/interface/rest"
	"github.com/taskforge/taskforge/service"
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

	if err := queue.EnsureGroup(context.Background()); err != nil {
		slog.Error("redis group", "error", err)
		os.Exit(1)
	}

	jobs := service.NewJobService(store, queue)
	handler := rest.NewHandler(jobs)
	router := rest.NewRouter(handler, store, queue)

	srv := &http.Server{Addr: cfg.HTTP.Addr, Handler: router}
	go func() {
		slog.Info("api listening", "addr", cfg.HTTP.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("api", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
