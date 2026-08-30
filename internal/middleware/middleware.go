package middleware

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leg100/surl/v2"
	"github.com/redis/go-redis/v9"
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
	Cache(ctx context.Context, data shared.UserCredential) error
}

type DatabaseUserQuery interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (generated.User, error)
}
type AuthMiddleware struct {
	sessionStore      RedisSessionStore
	userCacheStore    RedisUserCache
	userDatabaseQuery DatabaseUserQuery
}

func NewAuthMiddleware(authSession RedisSessionStore, userCache RedisUserCache, que DatabaseUserQuery) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore:      authSession,
		userCacheStore:    userCache,
		userDatabaseQuery: que,
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
			c.AbortWithStatusJSON(401, shared.NewErrorResponse(401, "you don't have permision to access this feature"))
			return
		}

		c.Next()
	}
}

func (mw *AuthMiddleware) RequireSignedUrl(signer *surl.Signer) gin.HandlerFunc {

	return func(c *gin.Context) {
		signedURL := "http://" + c.Request.Host + c.Request.URL.String()
		fmt.Println("URL THAT BEING VERIFIED : " + signedURL)
		err := signer.Verify(signedURL)
		if err != nil {

			fmt.Println("ERROR KEY VERIFY : " + err.Error())
			c.AbortWithStatusJSON(403, shared.NewErrorResponse(403, "access forbidden! you dont have permision to access this media!"))
			return
		}

		c.Next()
	}

}

func (mw *AuthMiddleware) RequireUsersStatus(status []generated.UserStatus) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, ex := GetUserIdFromContext(c)
		if !ex {
			c.AbortWithStatusJSON(403, shared.NewErrorResponse(403, "access forbidden, you need to login to access this feature"))
			return
		}

		data, err := mw.userCacheStore.GetCache(c.Request.Context(), userId)

		if err != nil {
			if errors.Is(err, redis.Nil) {
				fmt.Println("USER CACHE MISS!")

				usrData, dbErr := mw.userDatabaseQuery.GetUserByID(c.Request.Context(), userId)
				if dbErr != nil {
					if errors.Is(dbErr, pgx.ErrNoRows) {
						c.AbortWithStatusJSON(403, shared.NewErrorResponse(403, "access forbidden, your account doesn't exist! please contact the admin for more information"))
						return
					}
					c.AbortWithStatusJSON(500, shared.NewErrorResponse(500, "something wrong! please visit another time"))
					return
				}

				data = shared.UserCredential{
					UserId: usrData.ID,
					Status: string(usrData.Status),
					Role:   string(usrData.Role),
				}

				mw.userCacheStore.Cache(c.Request.Context(), data)

			} else {
				c.AbortWithStatusJSON(500, shared.NewErrorResponse(500, "something wrong! please visit another time"))
				return
			}
		}

		allowed := slices.Contains(status, generated.UserStatus(data.Status))
		if !allowed {
			c.AbortWithStatusJSON(403, shared.NewErrorResponse(403, "access forbidden"))
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
