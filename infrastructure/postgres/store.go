package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/taskforge/taskforge/domain"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(dsn string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

func (s *Store) Create(ctx context.Context, job *domain.Job) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO jobs (id, job_type, payload, priority, status, run_at, max_attempts, attempt_count, timeout_seconds, idempotency_key, correlation_id, created_at, updated_at)
		VALUES ($1,$2,$3,2,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		job.ID, job.JobType, job.Payload, job.Status, job.RunAt,
		job.MaxAttempts, job.AttemptCount, job.TimeoutSeconds,
		job.IdempotencyKey, job.CorrelationID, job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (s *Store) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	return scanJob(s.pool.QueryRow(ctx, `
		SELECT id, job_type, payload, status, run_at, max_attempts, attempt_count, timeout_seconds, idempotency_key, correlation_id, created_at, updated_at
		FROM jobs WHERE id = $1`, id))
}

func (s *Store) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Job, error) {
	return scanJob(s.pool.QueryRow(ctx, `
		SELECT id, job_type, payload, status, run_at, max_attempts, attempt_count, timeout_seconds, idempotency_key, correlation_id, created_at, updated_at
		FROM jobs WHERE idempotency_key = $1`, key))
}

func (s *Store) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	_, err := s.pool.Exec(ctx, `UPDATE jobs SET status = $2, updated_at = $3 WHERE id = $1`, id, status, time.Now().UTC())
	return err
}

func (s *Store) Claim(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	job, err := scanJob(tx.QueryRow(ctx, `
		SELECT id, job_type, payload, status, run_at, max_attempts, attempt_count, timeout_seconds, idempotency_key, correlation_id, created_at, updated_at
		FROM jobs WHERE id = $1 FOR UPDATE`, id))
	if err != nil {
		return nil, err
	}
	if job.Status != domain.StatusQueued {
		return nil, domain.ErrInvalidTransition
	}

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `UPDATE jobs SET status = $2, attempt_count = attempt_count + 1, started_at = $3, updated_at = $3 WHERE id = $1`,
		id, domain.StatusRunning, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	job.Status = domain.StatusRunning
	job.AttemptCount++
	return job, nil
}

func (s *Store) Complete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `UPDATE jobs SET status = $2, finished_at = $3, updated_at = $3 WHERE id = $1`, id, domain.StatusCompleted, now)
	return err
}

func scanJob(row pgx.Row) (*domain.Job, error) {
	var j domain.Job
	var status string
	err := row.Scan(&j.ID, &j.JobType, &j.Payload, &status, &j.RunAt, &j.MaxAttempts, &j.AttemptCount, &j.TimeoutSeconds, &j.IdempotencyKey, &j.CorrelationID, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJobNotFound
		}
		return nil, err
	}
	j.Status = domain.Status(status)
	return &j, nil
}

var _ domain.JobRepository = (*Store)(nil)
