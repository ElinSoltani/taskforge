package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	domainerror "github.com/taskforge/taskforge/domain/error"
	"github.com/taskforge/taskforge/domain/enum"
	domainhandler "github.com/taskforge/taskforge/domain/handler"
	"github.com/taskforge/taskforge/domain/model"
	"github.com/taskforge/taskforge/repository"
)

type Runner struct {
	repo       *repository.Repository
	registry   map[string]domainhandler.JobHandler
	workerID   string
	retryCfg   RetryConfig
}

func NewRunner(repo *repository.Repository, workerID string, handlers map[string]domainhandler.JobHandler, retryCfg RetryConfig) *Runner {
	return &Runner{repo: repo, registry: handlers, workerID: workerID, retryCfg: retryCfg}
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
		if shouldAckAfterClaimFailure(ctx, r.repo, msg.Payload.JobID, err) {
			_ = r.repo.Ack(ctx, msg.Stream, msg.MessageID)
		} else {
			slog.Warn("claim failed, leaving message for redelivery", "job_id", msg.Payload.JobID, "error", err)
		}
		return err
	}

	handler, ok := r.registry[job.JobType]
	if !ok {
		slog.Error("unknown job type", "job_type", job.JobType)
		_ = HandleFailure(ctx, r.repo, job, &domainerror.TerminalError{
			Code:    "unknown_job_type",
			Message: "no handler registered for job type: " + job.JobType,
		}, r.retryCfg)
		_ = r.repo.Ack(ctx, msg.Stream, msg.MessageID)
		return domainerror.ErrInvalidInput
	}

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(job.TimeoutSeconds)*time.Second)
	defer cancel()

	start := time.Now()
	execErr := handler.Execute(runCtx, job)

	if execErr != nil {
		if err := HandleFailure(ctx, r.repo, job, execErr, r.retryCfg); err != nil {
			slog.Error("handle failure failed", "job_id", job.ID, "error", err)
		}
		_ = r.repo.Ack(ctx, msg.Stream, msg.MessageID)
		return execErr
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

// PingHandler processes ping jobs.
type PingHandler struct{}

func (PingHandler) Execute(ctx context.Context, job *model.Job) error {
	var p struct {
		Message string `json:"message"`
		Fail    bool   `json:"fail"`
	}
	_ = json.Unmarshal(job.Payload, &p)
	if p.Fail {
		return &domainerror.RetryableError{Code: "simulated_failure", Message: "ping fail requested"}
	}
	slog.InfoContext(ctx, "ping", "job_id", job.ID, "message", p.Message)
	return nil
}

// FailHandler always returns a retryable error (for backoff testing).
type FailHandler struct{}

func (FailHandler) Execute(ctx context.Context, job *model.Job) error {
	return &domainerror.RetryableError{Code: "intentional_fail", Message: "fail job type always retries"}
}

// DeadHandler fails terminally (immediate dead + DLQ, no retries).
type DeadHandler struct{}

func (DeadHandler) Execute(ctx context.Context, job *model.Job) error {
	return &domainerror.TerminalError{
		Code:    "intentional_dead",
		Message: "dead job type moves straight to DLQ",
	}
}

// shouldAckAfterClaimFailure drops stale queue messages; pending/queued jobs stay in the PEL for retry.
func shouldAckAfterClaimFailure(ctx context.Context, repo *repository.Repository, jobID uuid.UUID, err error) bool {
	if !errors.Is(err, domainerror.ErrInvalidTransition) {
		return true
	}
	job, getErr := repo.GetByID(ctx, jobID)
	if getErr != nil {
		return true
	}
	switch job.Status {
	case enum.JobStatusPending, enum.JobStatusQueued:
		return false
	default:
		return true
	}
}
