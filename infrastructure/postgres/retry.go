package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	domainerror "github.com/taskforge/taskforge/domain/error"
	"github.com/taskforge/taskforge/domain/model"
	"github.com/taskforge/taskforge/infrastructure/postgres/dto"
)

func truncateError(msg string, max int) string {
	if len(msg) <= max {
		return msg
	}
	return msg[:max]
}

func (p *postgres) ScheduleRetry(ctx context.Context, id uuid.UUID, lastError string, runAt time.Time) error {
	now := time.Now().UTC()
	errMsg := truncateError(lastError, 2000)
	res, err := p.db.WithContext(ctx).Model(&dto.Job{}).
		Set("status = ?", string(model.JobStatusRetrying)).
		Set("run_at = ?", runAt.UTC()).
		Set("last_error = ?", errMsg).
		Set("started_at = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("status = ?", string(model.JobStatusRunning)).
		Update()
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domainerror.ErrInvalidTransition
	}
	return nil
}

func (p *postgres) MarkDead(ctx context.Context, id uuid.UUID, lastError string) error {
	now := time.Now().UTC()
	errMsg := truncateError(lastError, 2000)
	res, err := p.db.WithContext(ctx).Model(&dto.Job{}).
		Set("status = ?", string(model.JobStatusDead)).
		Set("last_error = ?", errMsg).
		Set("finished_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("status = ?", string(model.JobStatusRunning)).
		Update()
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domainerror.ErrInvalidTransition
	}
	return nil
}

// ListDueForRetry returns jobs ready to be moved back to the queue.
func (p *postgres) ListDueForRetry(ctx context.Context, limit int) ([]*model.Job, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []dto.Job
	err := p.db.WithContext(ctx).Model(&rows).
		Where("status = ?", string(model.JobStatusRetrying)).
		Where("run_at <= ?", time.Now().UTC()).
		Order("run_at ASC").
		Limit(limit).
		Select()
	if err != nil {
		return nil, err
	}
	jobs := make([]*model.Job, 0, len(rows))
	for i := range rows {
		jobs = append(jobs, rows[i].ToDomain())
	}
	return jobs, nil
}

// MarkQueuedIfDue atomically promotes a due retrying job to queued.
func (p *postgres) MarkQueuedIfDue(ctx context.Context, id uuid.UUID) (bool, error) {
	now := time.Now().UTC()
	res, err := p.db.WithContext(ctx).Model(&dto.Job{}).
		Set("status = ?", string(model.JobStatusQueued)).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("status = ?", string(model.JobStatusRetrying)).
		Where("run_at <= ?", now).
		Update()
	if err != nil {
		return false, err
	}
	return res.RowsAffected() > 0, nil
}
