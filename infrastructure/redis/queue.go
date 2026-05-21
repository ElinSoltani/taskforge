package redis

import (
	"context"
	"encoding/json"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/taskforge/taskforge/config"
	"github.com/taskforge/taskforge/domain"
)

type Queue struct {
	client *goredis.Client
	cfg    config.RedisConfig
	worker config.WorkerConfig
}

func NewQueue(cfg *config.Config) *Queue {
	return &Queue{
		client: goredis.NewClient(&goredis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
		cfg:    cfg.Redis,
		worker: cfg.Worker,
	}
}

func (q *Queue) Client() *goredis.Client { return q.client }

func (q *Queue) EnsureGroup(ctx context.Context) error {
	err := q.client.XGroupCreateMkStream(ctx, q.cfg.Stream, q.cfg.ConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create group: %w", err)
	}
	return nil
}

func (q *Queue) Enqueue(ctx context.Context, msg domain.QueueMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = q.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: q.cfg.Stream,
		Values: map[string]interface{}{"message": string(data)},
	}).Result()
	return err
}

func (q *Queue) Dequeue(ctx context.Context) (*domain.ConsumedMessage, error) {
	streams, err := q.client.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    q.cfg.ConsumerGroup,
		Consumer: q.worker.ConsumerName,
		Streams:  []string{q.cfg.Stream, ">"},
		Count:    1,
		Block:    q.worker.BlockTimeout,
	}).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}
		return nil, err
	}
	for _, s := range streams {
		for _, m := range s.Messages {
			raw, _ := m.Values["message"].(string)
			var payload domain.QueueMessage
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				return nil, err
			}
			return &domain.ConsumedMessage{Stream: s.Stream, MessageID: m.ID, Payload: payload}, nil
		}
	}
	return nil, nil
}

func (q *Queue) Ack(ctx context.Context, stream, messageID string) error {
	return q.client.XAck(ctx, stream, q.cfg.ConsumerGroup, messageID).Err()
}

func (q *Queue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

var _ domain.Queue = (*Queue)(nil)
