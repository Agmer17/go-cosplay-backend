package utils

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

func RedisSetStringObject[T any](
	ctx context.Context,
	rdb *redis.Client,
	key string,
	value T,
	ttl time.Duration,
) error {

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, data, ttl).Err()
}

func RedisGetStringObject[T any](
	ctx context.Context,
	rdb *redis.Client,
	key string,
) (*T, error) {

	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var obj T

	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}

	return &obj, nil
}
