package bootstrap

import (
	"context"
	"net/http"
	"os"

	"github.com/Agmer17/go-cosplay-backend/internal/auth"
	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/middleware"
	"github.com/Agmer17/go-cosplay-backend/internal/posts"
	"github.com/Agmer17/go-cosplay-backend/internal/users"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Configs struct {
	GoogleOauthClientId string
	GoogleOauthSecret   string
	DatabaseUrl         string
	AppContext          context.Context
	RedisHost           string
}

func NewConfigs(ctx context.Context) Configs {
	return Configs{
		GoogleOauthClientId: os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleOauthSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		DatabaseUrl:         os.Getenv("DATABASE_URL"),
		RedisHost:           os.Getenv("REDIS_HOST"),
		AppContext:          ctx,
	}
}

type App struct {
	Router         *gin.Engine
	Database       *db.Database
	RedisClient    *redis.Client
	ServiceConfigs *serviceConfigs

	Middleware *middleware.AuthMiddleware
}

func NewApp(configs Configs, rtr *gin.Engine) *App {

	opt, _ := redis.ParseURL(configs.RedisHost)
	redisClient := redis.NewClient(opt)

	database := db.NewDatabase(configs.DatabaseUrl, configs.AppContext)
	cacheConf := NewCacheConfigs(redisClient)

	serviceConfigs := NewServiceConfigs(serviceConfigsParams{
		GoogleOauthId:     configs.GoogleOauthClientId,
		GoogleOauthSecret: configs.GoogleOauthSecret,
		redisCli:          redisClient,
		Database:          database,
		cacheConfigs:      cacheConf,
	})

	mid := middleware.NewAuthMiddleware(
		cacheConf.sessionStore,
		cacheConf.usersDataCache,
		database.Query,
	)

	return &App{
		Router:         rtr,
		Database:       database,
		ServiceConfigs: serviceConfigs,
		RedisClient:    redisClient,
		Middleware:     mid,
	}

}

type BootstrapHandler interface {
	RegisterRoutes(r gin.IRouter)
}

func (app *App) SetupRoutes() {
	// create something in here
	authHandler := auth.NewAuthHandler(app.ServiceConfigs.AuthService)
	userHandler := users.NewUserHandler(app.ServiceConfigs.UserService, app.Middleware)
	postsHandler := posts.NewPostsHandler(app.ServiceConfigs.PostsService, app.Middleware, app.ServiceConfigs.UrlSigner)
	var hs []BootstrapHandler = []BootstrapHandler{
		authHandler,
		userHandler,
		postsHandler,
	}

	api := app.Router.Group("/api")
	api.Use(app.Middleware.HydrateUserContext())
	for _, h := range hs {
		h.RegisterRoutes(api)
	}

	app.Router.StaticFS("/uploads", http.Dir("./tmp/uploads"))
}

func (a *App) Run() {
	a.Router.Run("0.0.0.0:80")
}
