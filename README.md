# TaskForge

Distributed job processing: REST API → PostgreSQL → Redis Streams → workers.

## Quick start

```bash
docker compose up --build
```

Create a job:

```bash
curl -s -X POST http://localhost:8080/v1/jobs \
  -H 'Content-Type: application/json' \
  -d '{"job_type":"ping","payload":{"message":"hello"}}'
```

Check status (use `id` from the response):

```bash
curl -s http://localhost:8080/v1/jobs/<job-id>
```

Worker logs will show the `ping` handler running.

## Local development (without Docker)

```bash
# Postgres on :5433, Redis on :6380 (see docker-compose ports)
export POSTGRES_HOST=localhost POSTGRES_PORT=5433
export POSTGRES_USER=taskforge POSTGRES_PASSWORD=taskforge POSTGRES_DB=taskforge
export POSTGRES_MIGRATION_PATH=file://migrations
export REDIS_ADDR=localhost:6380

go run ./cmd/api    # runs migrations on startup (go-pg + golang-migrate)
go run ./cmd/worker
```

## How REST talks to domain

REST never imports postgres/redis. The flow is:

```
HTTP JSON → rest/dto (bind + Validate) → mapper ToCreateJobInput → service → repository → infrastructure
domain model ← service ← repository
domain model → mapper JobResponseFromDomain → rest/dto → HTTP JSON
```

| Layer | Responsibility |
|-------|----------------|
| `interface/rest/dto` | Request/response shapes, field validation, domain mapping |
| `interface/rest/handler` | Gin binding, status codes, error mapping |
| `service` | Business rules (idempotency, enqueue workflow) |
| `domain/model` | Core types (`Job`, `CreateJobInput`) |

## Layout (aligned with other Backend services)

```
domain/model/          # Job aggregate
domain/repository/     # JobStore, JobQueue interface signatures
domain/handler/        # JobHandler interface
repository/            # Thin delegates to infrastructure
service/               # Use cases
infrastructure/postgres/
  main.go              # NewPostgres, Ping, migrations
  dto/ + mapper        # DB rows ↔ domain model
infrastructure/redis/
  main.go              # NewRedis, Ping
```

## Architecture (this slice)

1. **POST /v1/jobs** — persist job (`pending` → `queued`), **XADD** to Redis stream
2. **Worker** — **XREADGROUP**, claim job in Postgres (`running`), run handler, **XACK**, mark `completed`

Stream: `taskforge:queue:normal` · Consumer group: `taskforge-workers`

### Migrations

| Command | Action |
|---------|--------|
| `make migrate` | Apply `001_init.up.sql` |
| `make migrate-down` | Roll back `001_init.down.sql` (drops `jobs`) |
