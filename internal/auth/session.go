package auth

import (
	"context"
	"time"

	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
)

const session_expire = 7 * 24 * time.Hour

func (as *AuthService) CreateSessions(
	ctx context.Context,
	session shared.SessionsData,
) error {

	const sessionExpire = 7 * 24 * time.Hour

	session.ExpiresAt = time.Now().Add(sessionExpire)

	key := "sessions:" + session.SessionId

	return utils.RedisSetStringObject(
		ctx,
		as.rdb,
		key,
		session,
		sessionExpire,
	)
}

func (as *AuthService) ExpireSessions(ctx context.Context, key string) *shared.ErrorResponse {

	sessionKey := "sessions:" + key

	err := as.rdb.Del(ctx, sessionKey).Err()
	if err != nil {
		return shared.NewErrorResponse(500, "something wrong with the server, trying to logout another time")
	}

	return nil

}
