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
	"github.com/taskforge/taskforge/infrastructure/redis"
	"github.com/taskforge/taskforge/interface/rest"
	"github.com/taskforge/taskforge/repository"
	"github.com/taskforge/taskforge/service"
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

	if err := rdb.EnsureGroup(context.Background()); err != nil {
		slog.Error("redis group", "error", err)
		os.Exit(1)
	}

	repo := repository.NewRepository(pg, rdb)
	jobs := service.NewJobService(repo, cfg.Service)
	handler := rest.NewHandler(jobs, cfg.HTTP.BaseURL)
	router := rest.NewRouter(handler, pgReadiness{}, rdbReadiness{})

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

type pgReadiness struct{}

func (pgReadiness) Ping(ctx context.Context) error { return postgres.Ping(ctx) }

type rdbReadiness struct{}

func (rdbReadiness) Ping(ctx context.Context) error { return redis.Ping(ctx) }
