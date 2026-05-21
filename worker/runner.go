package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	domainerror "github.com/taskforge/taskforge/domain/error"
	domainhandler "github.com/taskforge/taskforge/domain/handler"
	"github.com/taskforge/taskforge/domain/model"
	"github.com/taskforge/taskforge/repository"
)

type Runner struct {
	repo     *repository.Repository
	registry map[string]domainhandler.JobHandler
	workerID string
}

func NewRunner(repo *repository.Repository, workerID string, handlers map[string]domainhandler.JobHandler) *Runner {
	return &Runner{repo: repo, registry: handlers, workerID: workerID}
}

func (r *Runner) Run(ctx context.Context) error {
	if err := r.repo.EnsureGroup(ctx); err != nil {
		return err
	}
	slog.Info("worker started", "worker_id", r.workerID)

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopping", "worker_id", r.workerID)
			return ctx.Err()
		default:
		}

		msg, err := r.repo.Dequeue(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Warn("dequeue error", "error", err)
			continue
		}
		if msg == nil {
			continue
		}

		if err := r.process(ctx, msg); err != nil {
			slog.Error("process failed", "job_id", msg.Payload.JobID, "error", err)
		}
	}
}

func (r *Runner) process(ctx context.Context, msg *model.ConsumedMessage) error {
	job, err := r.repo.Claim(ctx, msg.Payload.JobID)
	if err != nil {
		_ = r.repo.Ack(ctx, msg.Stream, msg.MessageID)
		return err
	}

	handler, ok := r.registry[job.JobType]
	if !ok {
		slog.Error("unknown job type", "job_type", job.JobType)
		_ = r.repo.Ack(ctx, msg.Stream, msg.MessageID)
		return domainerror.ErrInvalidInput
	}

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(job.TimeoutSeconds)*time.Second)
	defer cancel()

	start := time.Now()
	if err := handler.Execute(runCtx, job); err != nil {
		slog.Error("handler failed", "job_id", job.ID, "error", err)
		return err
	}

	if err := r.repo.Complete(ctx, job.ID); err != nil {
		return err
	}
	if err := r.repo.Ack(ctx, msg.Stream, msg.MessageID); err != nil {
		return err
	}

	slog.Info("job completed",
		"job_id", job.ID,
		"job_type", job.JobType,
		"worker_id", r.workerID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}

type PingHandler struct{}

func (PingHandler) Execute(ctx context.Context, job *model.Job) error {
	var p struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(job.Payload, &p)
	slog.InfoContext(ctx, "ping", "job_id", job.ID, "message", p.Message)
	return nil
}
