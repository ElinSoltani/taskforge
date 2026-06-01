FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api
RUN CGO_ENABLED=0 go build -o /worker ./cmd/worker

FROM alpine:3.20 AS api
WORKDIR /app
RUN apk add --no-cache wget
COPY --from=builder /api /app/api
COPY --from=builder /app/migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/api"]

FROM alpine:3.20 AS worker
WORKDIR /app
COPY --from=builder /worker /app/worker
COPY --from=builder /app/migrations /app/migrations
ENTRYPOINT ["/app/worker"]
