# TaskForge

Distributed job processing: REST API → PostgreSQL → Redis Streams → workers.

## Quick start

```bash
docker compose up --build
```

| Service | Port (host) |
|---------|-------------|
| API | `8080` |
| Postgres | `5433` |
| Redis | `6380` |

```bash
# Create
curl -s -X POST http://localhost:8080/v1/jobs \
  -H 'Content-Type: application/json' \
  -d '{"job_type":"ping","payload":{"message":"hello"}}'

# Status (replace <job-id> with id from response)
curl -s http://localhost:8080/v1/jobs/<job-id>

# Optional idempotency
curl -s -X POST http://localhost:8080/v1/jobs \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: my-unique-key-123' \
  -d '{"job_type":"ping","payload":{"message":"hello"}}'
```

Worker logs show the handler running. Poll `GET /v1/jobs/<id>` until `status` is `completed`.

## API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/jobs` | Create job (`202 Accepted`, or `200` if idempotent duplicate) |
| `GET` | `/v1/jobs/{id}` | Job status and metadata |
| `GET` | `/health` | Liveness |
| `GET` | `/ready` | Postgres + Redis readiness |

## Built-in job types

| `job_type` | Behavior |
|------------|----------|
| `ping` | Logs payload `message`; set `"fail": true` in payload to simulate a retryable error |
| `fail` | Always fails with a retryable error until `max_attempts` |

## Job statuses

`pending` → `queued` → `running` → `completed`

On failure (with attempts left): `running` → `retrying` → (scheduler) → `queued` → `running` …

Terminal: `dead` (max attempts, `TerminalError`, or unknown job type)

## Makefile

| Command | Action |
|---------|--------|
| `make up` | `docker compose up --build -d` |
| `make down` | Stop stack |
| `make logs` | Follow `api` + `worker` logs |
| `make build` | Build `bin/api` and `bin/worker` locally |
| `make migrate` | Apply SQL migrations via Compose (`001` + `002`) |
| `make migrate-down` | Roll back `002` then `001` (drops `jobs`) |
| `make test-job` | POST a sample `ping` job |

## Local development (without Docker)

```bash
export POSTGRES_HOST=localhost POSTGRES_PORT=5433
export POSTGRES_USER=taskforge POSTGRES_PASSWORD=taskforge POSTGRES_DB=taskforge
export POSTGRES_MIGRATION_PATH=file://migrations
export REDIS_ADDR=localhost:6380
export REDIS_STREAM=taskforge:queue:normal
export REDIS_CONSUMER_GROUP=taskforge-workers
export WORKER_CONSUMER_NAME=worker-1

# Optional
export BACKOFF_BASE=5s BACKOFF_MAX=15m RETRY_SCHEDULER_INTERVAL=5s
export HTTP_BASE_URL=http://localhost:8080
export SERVICE_MAX_ATTEMPTS=5 SERVICE_TIMEOUT_SECONDS=300

go run ./cmd/api     # applies migrations on startup
go run ./cmd/worker  # retry scheduler runs in background
```

Manual migration (if needed):

```bash
psql "postgres://taskforge:taskforge@localhost:5433/taskforge?sslmode=disable" \
  -f migrations/001_init.up.sql \
  -f migrations/002_jobs_retry_index.up.sql
```

## Migrations

Files in [`migrations/`](migrations/). Applied automatically when **api** or **worker** starts (`golang-migrate`, `POSTGRES_MIGRATION_PATH`, default `file://migrations`). Compose also runs the same SQL via a `migrate` service before api/worker.

| Version | Up | Down | Purpose |
|---------|----|------|---------|
| `001` | `001_init.up.sql` | `001_init.down.sql` | `jobs` table, idempotency index, status index |
| `002` | `002_jobs_retry_index.up.sql` | `002_jobs_retry_index.down.sql` | Index `(status, run_at)` for retry scheduler |

Docker images use `file:///app/migrations`; local runs use `file://migrations` from the project root.

## Configuration

### HTTP

| Env | Default | Description |
|-----|---------|-------------|
| `HTTP_ADDR` | `:8080` | API listen address |
| `HTTP_BASE_URL` | `http://localhost:8080` | Base URL in job response `links.self` |

### Service

| Env | Default | Description |
|-----|---------|-------------|
| `SERVICE_MAX_ATTEMPTS` | `5` | Max execution attempts per job |
| `SERVICE_TIMEOUT_SECONDS` | `300` | Per-attempt handler timeout (seconds) |

### Postgres

| Env | Default (Docker) | Description |
|-----|------------------|-------------|
| `POSTGRES_HOST` | `postgres` | Host |
| `POSTGRES_PORT` | `5432` | Port |
| `POSTGRES_USER` | `taskforge` | User |
| `POSTGRES_PASSWORD` | `taskforge` | Password |
| `POSTGRES_DB` | `taskforge` | Database |
| `POSTGRES_MIGRATION_PATH` | `file://migrations` | golang-migrate source |

### Redis

| Env | Default | Description |
|-----|---------|-------------|
| `REDIS_ADDR` | `redis:6379` | Address |
| `REDIS_PASSWORD` | _(empty)_ | Password |
| `REDIS_DB` | `0` | DB number |
| `REDIS_STREAM` | `taskforge:queue:normal` | Stream for job messages |
| `REDIS_CONSUMER_GROUP` | `taskforge-workers` | Consumer group |

### Worker

| Env | Default | Description |
|-----|---------|-------------|
| `WORKER_CONSUMER_NAME` | `worker-1` | Redis consumer name |
| `WORKER_BLOCK_TIMEOUT` | `2s` | `XREADGROUP` block timeout |

### Retries (exponential backoff)

| Env | Default | Description |
|-----|---------|-------------|
| `BACKOFF_BASE` | `5s` | Base delay (`base * 2^attempt`) |
| `BACKOFF_MAX` | `15m` | Cap on backoff delay |
| `RETRY_SCHEDULER_INTERVAL` | `5s` | How often due `retrying` jobs are re-enqueued |

On handler failure the worker sets `retrying` with `run_at = now + backoff` (full jitter), **acks** the Redis message, and the **scheduler** promotes due jobs to `queued` and **XADD**s again.

Handlers may return `domain/error.RetryableError` or `TerminalError` to control retry vs dead.

Test retries:

```bash
curl -s -X POST http://localhost:8080/v1/jobs \
  -H 'Content-Type: application/json' \
  -d '{"job_type":"ping","payload":{"fail":true}}'

curl -s -X POST http://localhost:8080/v1/jobs \
  -H 'Content-Type: application/json' \
  -d '{"job_type":"fail","payload":{}}'
```

## How REST talks to domain

REST never imports postgres/redis:

```
HTTP JSON → rest/dto (bind + Validate) → mapper → service → repository → infrastructure
domain model ← service ← repository
domain model → mapper → rest/dto → HTTP JSON
```

| Layer | Responsibility |
|-------|----------------|
| `interface/rest/dto` | Request/response shapes, validation, domain mapping |
| `interface/rest/handler` | Gin binding, status codes, error mapping |
| `service` | Business rules (idempotency, enqueue workflow) |
| `domain/model` | `Job`, `QueueMessage`, `ConsumedMessage` |
| `domain/enum` | `JobStatus` |
| `repository` | Delegates to `JobStore` + `JobQueue` |
| `infrastructure/postgres` | go-pg, DTOs, mappers |
| `infrastructure/redis` | Redis Streams queue |

## Project layout

```
cmd/api/               # HTTP server
cmd/worker/            # Consumer + retry scheduler
config/
domain/model/          # Job, QueueMessage, ConsumedMessage
domain/enum/           # JobStatus
domain/repository/     # JobStore, JobQueue interfaces
domain/handler/
domain/error/
repository/
service/
worker/                # Runner, backoff, scheduler, built-in handlers
interface/rest/dto/
infrastructure/postgres/
infrastructure/redis/
migrations/
```

## Architecture

1. **POST /v1/jobs** — persist (`pending`), **XADD** Redis, `queued`
2. **Worker** — **XREADGROUP**, claim `running`, run handler
3. **Success** — `completed`, **XACK**
4. **Failure** — `retrying` + backoff, **XACK**; scheduler re-enqueues when due

Stream: `taskforge:queue:normal` · Consumer group: `taskforge-workers`

See [Job statuses](#job-statuses) and [Migrations](#migrations) for lifecycle and schema detail.
