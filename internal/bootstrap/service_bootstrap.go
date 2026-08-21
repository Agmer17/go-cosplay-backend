package bootstrap

import (
	"github.com/Agmer17/go-cosplay-backend/internal/auth"
	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/redis/go-redis/v9"
)

type serviceConfigs struct {
	AuthService *auth.AuthService
}

type serviceConfigsParams struct {
	GoogleOauthId     string
	GoogleOauthSecret string
	redisCli          *redis.Client
	Database          *db.Database
}

func NewServiceConfigs(env serviceConfigsParams) *serviceConfigs {

	authSvc := *auth.NewAuthService(&auth.CreateAuthServiceParams{
		GoogleAuthId: env.GoogleOauthId,
		GoogleSecret: env.GoogleOauthSecret,
		RedisClient:  env.redisCli,
		Database:     env.Database,
	})

	return &serviceConfigs{
		AuthService: &authSvc,
	}

}
