package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type redisClient struct {
	db *goredis.Client
}

var instance *redisClient

func NewRedis() (*redisClient, error) {
	if err := cfg.Validation(); err != nil {
		return nil, err
	}
	r := &redisClient{
		db: goredis.NewClient(&goredis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
	}
	if err := r.db.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	instance = r
	return r, nil
}

func (r *redisClient) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func GetClient() *goredis.Client {
	if instance == nil {
		return nil
	}
	return instance.db
}

func Ping(ctx context.Context) error {
	if instance == nil || instance.db == nil {
		return ErrRedisNotReady
	}
	if err := instance.db.Ping(ctx).Err(); err != nil {
		return ErrRedisNotReady
	}
	return nil
}

func blockDuration() time.Duration {
	if cfg.BlockTimeout <= 0 {
		return 2 * time.Second
	}
	return time.Duration(cfg.BlockTimeout) * time.Millisecond
}
