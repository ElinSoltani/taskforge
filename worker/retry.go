package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	domainerror "github.com/taskforge/taskforge/domain/error"
	"github.com/taskforge/taskforge/domain/model"
	"github.com/taskforge/taskforge/pkg/backoff"
	"github.com/taskforge/taskforge/repository"
)

// RetryConfig controls exponential backoff between attempts.
type RetryConfig struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// HandleFailure classifies the error and schedules retry or dead status.
func HandleFailure(ctx context.Context, repo *repository.Repository, job *model.Job, execErr error, cfg RetryConfig) error {
	if execErr == nil {
		return nil
	}

	var terminal *domainerror.TerminalError
	if errors.As(execErr, &terminal) {
		return repo.MarkDead(ctx, job, execErr.Error())
	}

	if !job.HasAttemptsRemaining() {
		msg := fmt.Sprintf("max attempts (%d) exceeded: %v", job.MaxAttempts, execErr)
		return repo.MarkDead(ctx, job, msg)
	}

	var retryable *domainerror.RetryableError
	if !errors.As(execErr, &retryable) {
		// Default: unknown handler errors are retryable until max attempts.
		retryable = &domainerror.RetryableError{Message: execErr.Error()}
	}

	delay := backoff.Delay(job.AttemptCount, cfg.BaseDelay, cfg.MaxDelay)
	runAt := time.Now().UTC().Add(delay)

	slog.Info("scheduling job retry",
		"job_id", job.ID,
		"attempt", job.AttemptCount,
		"max_attempts", job.MaxAttempts,
		"retry_in", delay.String(),
		"error", retryable.Error(),
	)

	return repo.ScheduleRetry(ctx, job.ID, execErr.Error(), runAt)
}
