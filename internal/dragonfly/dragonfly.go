package dragonfly

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Dragonfly struct {
	Client *redis.Client
}

func NewDragonflyClient(ctx context.Context, host string, port int) (*Dragonfly, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Dragonfly{Client: client}, nil
}

func (d *Dragonfly) Close() {
	d.Client.Close()
}

// should probably be a lua script
func (d *Dragonfly) UpsertBatchCounter(ctx context.Context, key string, batchSize int64) (int64, int64, error) {
	val, err := d.Client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			err = d.Client.Set(ctx, key, batchSize, 0).Err()
			if err != nil {
				return 0, 0, fmt.Errorf("set key %s: %w", key, err)
			}
			return 0, batchSize, nil
		}
		return 0, 0, fmt.Errorf("get key %s: %w", key, err)
	}

	err = d.Client.Set(ctx, key, val+batchSize, 0).Err()
	if err != nil {
		return 0, 0, fmt.Errorf("increment key %s: %w", key, err)
	}

	return val + 1, val + batchSize, nil
}
