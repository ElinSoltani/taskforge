package redis

import (
	"context"
	"encoding/json"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/taskforge/taskforge/domain/model"
)

func (r *redisClient) EnsureGroup(ctx context.Context) error {
	if err := ensureStreamGroup(ctx, r.db, cfg.Stream, cfg.ConsumerGroup); err != nil {
		return fmt.Errorf("main queue: %w", err)
	}
	dlqGroup := cfg.DLQConsumerGroup
	if dlqGroup == "" {
		dlqGroup = "taskforge-dlq"
	}
	if err := ensureStreamGroup(ctx, r.db, cfg.DLQStream, dlqGroup); err != nil {
		return fmt.Errorf("dlq stream: %w", err)
	}
	return nil
}

func ensureStreamGroup(ctx context.Context, db *goredis.Client, stream, group string) error {
	err := db.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

func (r *redisClient) Enqueue(ctx context.Context, msg model.QueueMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = r.db.XAdd(ctx, &goredis.XAddArgs{
		Stream: cfg.Stream,
		Values: map[string]interface{}{"message": string(data)},
	}).Result()
	return err
}

func (r *redisClient) Dequeue(ctx context.Context) (*model.ConsumedMessage, error) {
	streams, err := r.db.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    cfg.ConsumerGroup,
		Consumer: cfg.ConsumerName,
		Streams:  []string{cfg.Stream, ">"},
		Count:    1,
		Block:    blockDuration(),
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
			var payload model.QueueMessage
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				return nil, err
			}
			return &model.ConsumedMessage{
				Stream:    s.Stream,
				MessageID: m.ID,
				Payload:   payload,
			}, nil
		}
	}
	return nil, nil
}

func (r *redisClient) Ack(ctx context.Context, stream, messageID string) error {
	return r.db.XAck(ctx, stream, cfg.ConsumerGroup, messageID).Err()
}

func (r *redisClient) Ping(ctx context.Context) error {
	return r.db.Ping(ctx).Err()
}
