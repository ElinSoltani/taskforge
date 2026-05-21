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
export POSTGRES_DSN="postgres://taskforge:taskforge@localhost:5433/taskforge?sslmode=disable"
export REDIS_ADDR="localhost:6380"
psql "$POSTGRES_DSN" -f migrations/001_init.up.sql
# Rollback: psql "$POSTGRES_DSN" -f migrations/001_init.down.sql

go run ./cmd/api
go run ./cmd/worker
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
