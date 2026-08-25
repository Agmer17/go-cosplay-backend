package caching

import (
	"context"
	"time"

	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	rdb *redis.Client
}

func NewSessionStore(r *redis.Client) *SessionStore {
	return &SessionStore{
		rdb: r,
	}
}

func (ss *SessionStore) IssueSession(ctx context.Context, sess *shared.SessionsData, ttl time.Duration) error {
	key := "sessions:" + sess.SessionId
	sess.ExpiresAt = time.Now().Add(ttl)
	return utils.RedisSetStringObject(ctx, ss.rdb, key, sess, ttl)
}

func (ss *SessionStore) RevokeSession(ctx context.Context, key string) error {

	sessionKey := "sessions:" + key
	err := ss.rdb.Del(ctx, sessionKey).Err()
	if err != nil {
		return err
	}

	return nil
}
