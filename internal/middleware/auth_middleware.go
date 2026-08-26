package middleware

import (
	"context"

	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	AccessTokenKey    = "access_token"
	CurrentUserId     = "user_id"
	CurrentUserRole   = "role"
	CurrentUserStatus = "status"
)

type RedisSessionStore interface {
	GetSession(ctx context.Context, token string) (shared.SessionsData, error)
}

type RedisUserCache interface {
	GetCache(ctx context.Context, id uuid.UUID) (shared.UserCredential, error)
}
type AuthMiddleware struct {
	sessionStore   RedisSessionStore
	userCacheStore RedisUserCache
}

func NewAuthMiddleware(authSession RedisSessionStore, userCache RedisUserCache) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore:   authSession,
		userCacheStore: userCache,
	}
}

func (mw *AuthMiddleware) HydrateUserContext() gin.HandlerFunc {

	return func(c *gin.Context) {
		token, err := c.Cookie(AccessTokenKey)
		if err != nil {
			c.Next()
			return
		}

		data, err := mw.sessionStore.GetSession(c.Request.Context(), token)
		if err != nil {
			c.Next()
			return
		}

		c.Set(CurrentUserId, data.UserId)
		c.Next()
	}
}

func (mw *AuthMiddleware) AuthenticatedUserOnly() gin.HandlerFunc {

	return func(c *gin.Context) {

		// ini dapet dari hydrate context yg global!
		_, ex := c.Get(CurrentUserId)

		if !ex {
			c.AbortWithStatusJSON(403, shared.NewErrorResponse(403, "you don't have permision to access this feature"))
			return
		}

		c.Next()
	}
}

func GetUserIdFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(CurrentUserId)

	if !exists {
		return uuid.Nil, false
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return userID, true
}
