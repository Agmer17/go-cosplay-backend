package middleware

import (
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	AccessTokenKey  = "access_token"
	CurrentUserId   = "user_id"
	CurrentUserRole = "role"
)

type Middleware struct {
	rdb *redis.Client
}

func (mw *Middleware) HydrateUserContext() gin.HandlerFunc {

	return func(c *gin.Context) {
		token, err := c.Cookie(AccessTokenKey)
		if err != nil {
			return
		}

		key := "sessions:" + token

		data, err := utils.RedisGetStringObject[shared.SessionsData](c.Request.Context(), mw.rdb, key)

		if err != nil {
			return
		}

		c.Set(CurrentUserId, data.UserId)
		c.Next()
	}
}
