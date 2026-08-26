package caching

import (
	"context"
	"time"

	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const userCahceExpiration = 7 * 24 * time.Hour

type UserCache struct {
	rdb *redis.Client
}

func NewUserDataCache(r *redis.Client) *UserCache {
	return &UserCache{
		rdb: r,
	}
}

func (uc *UserCache) Cache(ctx context.Context, data shared.UserCredential) error {
	key := "users:" + data.UserId.String()
	return utils.RedisSetStringObject(ctx, uc.rdb, key, data, userCahceExpiration)
}

func (uc *UserCache) EvictCache(ctx context.Context, id uuid.UUID) error {
	key := "users:" + id.String()
	return uc.rdb.Del(ctx, key).Err()
}

func (uc *UserCache) GetCache(ctx context.Context, id uuid.UUID) (shared.UserCredential, error) {

	key := "users:" + id.String()

	data, err := utils.RedisGetStringObject[shared.UserCredential](ctx, uc.rdb, key)
	if err != nil {
		return shared.UserCredential{}, err
	}

	return *data, nil
}
