package bootstrap

import (
	"github.com/Agmer17/go-cosplay-backend/internal/caching"
	"github.com/redis/go-redis/v9"
)

type CacheConfigs struct {
	sessionStore   *caching.SessionStore
	usersDataCache *caching.UserCache
}

func NewCacheConfigs(rdb *redis.Client) *CacheConfigs {
	return &CacheConfigs{
		sessionStore:   caching.NewSessionStore(rdb),
		usersDataCache: caching.NewUserDataCache(rdb),
	}
}
