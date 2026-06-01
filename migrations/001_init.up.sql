CREATE TABLE IF NOT EXISTS jobs (
    id              UUID PRIMARY KEY,
    job_type        TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    priority        SMALLINT NOT NULL DEFAULT 2,
    status          TEXT NOT NULL,
    run_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    max_attempts    INT NOT NULL DEFAULT 5,
    attempt_count   INT NOT NULL DEFAULT 0,
    timeout_seconds INT NOT NULL DEFAULT 300,
    idempotency_key TEXT,
    correlation_id  TEXT,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    CONSTRAINT jobs_status_check CHECK (status IN (
        'pending','queued','running','completed','failed','retrying','dead','cancelled'
    ))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_jobs_idempotency
    ON jobs (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs (status);
