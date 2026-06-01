.PHONY: build up down logs migrate migrate-down test-job

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f api worker

migrate:
	docker compose run --rm migrate

migrate-down:
	docker compose --profile tools run --rm migrate-down

test-job:
	curl -s -X POST http://localhost:8080/v1/jobs \
		-H 'Content-Type: application/json' \
		-d '{"job_type":"ping","payload":{"message":"hello"}}' | jq .
