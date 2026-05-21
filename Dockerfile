FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api
RUN CGO_ENABLED=0 go build -o /worker ./cmd/worker

FROM alpine:3.20 AS api
RUN apk add --no-cache wget
COPY --from=builder /api /api
EXPOSE 8080
ENTRYPOINT ["/api"]

FROM alpine:3.20 AS worker
COPY --from=builder /worker /worker
ENTRYPOINT ["/worker"]
