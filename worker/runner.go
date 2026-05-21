package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/taskforge/taskforge/domain"
)

type Runner struct {
	repo     domain.JobRepository
	queue    domain.Queue
	registry map[string]domain.JobHandler
	workerID string
}

func NewRunner(repo domain.JobRepository, queue domain.Queue, workerID string, handlers map[string]domain.JobHandler) *Runner {
	return &Runner{repo: repo, queue: queue, registry: handlers, workerID: workerID}
}

func (r *Runner) Run(ctx context.Context) error {
	if err := r.queue.EnsureGroup(ctx); err != nil {
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

		msg, err := r.queue.Dequeue(ctx)
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

func (r *Runner) process(ctx context.Context, msg *domain.ConsumedMessage) error {
	job, err := r.repo.Claim(ctx, msg.Payload.JobID)
	if err != nil {
		_ = r.queue.Ack(ctx, msg.Stream, msg.MessageID)
		return err
	}

	handler, ok := r.registry[job.JobType]
	if !ok {
		slog.Error("unknown job type", "job_type", job.JobType)
		_ = r.queue.Ack(ctx, msg.Stream, msg.MessageID)
		return domain.ErrInvalidInput
	}

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(job.TimeoutSeconds)*time.Second)
	defer cancel()

	start := time.Now()
	err = handler.Execute(runCtx, job)
	if err != nil {
		slog.Error("handler failed", "job_id", job.ID, "error", err)
		return err
	}

	if err := r.repo.Complete(ctx, job.ID); err != nil {
		return err
	}
	if err := r.queue.Ack(ctx, msg.Stream, msg.MessageID); err != nil {
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

// PingHandler is the default job handler for smoke tests.
type PingHandler struct{}

func (PingHandler) Execute(ctx context.Context, job *domain.Job) error {
	var p struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(job.Payload, &p)
	slog.InfoContext(ctx, "ping", "job_id", job.ID, "message", p.Message)
	return nil
}
