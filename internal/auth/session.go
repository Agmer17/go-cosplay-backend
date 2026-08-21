package auth

import (
	"context"
	"time"

	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/google/uuid"
)

const session_expire = 7 * 24 * time.Hour

type SessionsData struct {
	SessionId string    `json:"session_id"`
	UserId    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (as *AuthService) CreateSessions(
	ctx context.Context,
	session SessionsData,
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
