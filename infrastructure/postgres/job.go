package postgres

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	domainerror "github.com/taskforge/taskforge/domain/error"
	"github.com/taskforge/taskforge/domain/enum"
	"github.com/taskforge/taskforge/domain/model"
	"github.com/taskforge/taskforge/infrastructure/postgres/dto"
)

func (p *postgres) Create(ctx context.Context, job *model.Job) error {
	row := &dto.Job{}
	row.FromDomain(job)
	_, err := p.db.WithContext(ctx).Model(row).Insert()
	return err
}

func (p *postgres) GetByID(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	row := &dto.Job{}
	err := p.db.WithContext(ctx).Model(row).Where("id = ?", id).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, domainerror.ErrJobNotFound
		}
		return nil, err
	}
	return row.ToDomain(), nil
}

func (p *postgres) GetByIdempotencyKey(ctx context.Context, key string) (*model.Job, error) {
	row := &dto.Job{}
	err := p.db.WithContext(ctx).Model(row).Where("idempotency_key = ?", key).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, domainerror.ErrJobNotFound
		}
		return nil, err
	}
	return row.ToDomain(), nil
}

func (p *postgres) UpdateStatus(ctx context.Context, id uuid.UUID, status enum.JobStatus) error {
	_, err := p.db.WithContext(ctx).Model(&dto.Job{}).
		Set("status = ?", string(status)).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Update()
	return err
}

func (p *postgres) Claim(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	var result *model.Job
	err := p.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		row := &dto.Job{}
		err := tx.Model(row).Where("id = ?", id).For("UPDATE").Select()
		if err != nil {
			if err == pg.ErrNoRows {
				return domainerror.ErrJobNotFound
			}
			return err
		}
		if row.Status != string(enum.JobStatusQueued) {
			return domainerror.ErrInvalidTransition
		}

		now := time.Now().UTC()
		row.Status = string(enum.JobStatusRunning)
		row.AttemptCount++
		row.StartedAt = &now
		row.UpdatedAt = now

		_, err = tx.Model(row).
			Set("status = ?", row.Status).
			Set("attempt_count = ?", row.AttemptCount).
			Set("started_at = ?", row.StartedAt).
			Set("updated_at = ?", row.UpdatedAt).
			Where("id = ?", id).
			Update()
		if err != nil {
			return err
		}
		result = row.ToDomain()
		return nil
	})
	return result, err
}

func (p *postgres) Complete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := p.db.WithContext(ctx).Model(&dto.Job{}).
		Set("status = ?", string(enum.JobStatusCompleted)).
		Set("finished_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Update()
	return err
}
