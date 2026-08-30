package bootstrap

import (
	"github.com/Agmer17/go-cosplay-backend/internal/auth"
	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/likes"
	"github.com/Agmer17/go-cosplay-backend/internal/posts"
	"github.com/Agmer17/go-cosplay-backend/internal/storage"
	"github.com/Agmer17/go-cosplay-backend/internal/users"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/leg100/surl/v2"
	"github.com/redis/go-redis/v9"
)

type serviceConfigs struct {
	AuthService  *auth.AuthService
	UserService  *users.UsersService
	PostsService *posts.PostsService
	LikesService *likes.PostLikesService

	UrlSigner *surl.Signer
}

type serviceConfigsParams struct {
	GoogleOauthId     string
	GoogleOauthSecret string
	redisCli          *redis.Client
	Database          *db.Database
	cacheConfigs      *CacheConfigs
}

func NewServiceConfigs(env serviceConfigsParams) *serviceConfigs {

	storage := storage.NewFileStorage(10)

	authSvc := *auth.NewAuthService(&auth.CreateAuthServiceParams{
		GoogleAuthId: env.GoogleOauthId,
		GoogleSecret: env.GoogleOauthSecret,
		Database:     env.Database,
		AuthSession:  env.cacheConfigs.sessionStore,
		UserCache:    env.cacheConfigs.usersDataCache,
	})

	usersService := users.NewUsersService(env.Database, storage, env.cacheConfigs.usersDataCache)

	likeSvc := likes.NewPostLikesService(env.Database)
	cryptoKey, _ := utils.GenerateSecureString(48)

	signer := surl.New([]byte(cryptoKey))
	postsService := posts.NewPostsService(env.Database, storage, signer)

	return &serviceConfigs{
		AuthService:  &authSvc,
		UserService:  usersService,
		PostsService: postsService,
		LikesService: likeSvc,
		UrlSigner:    signer,
	}

}
