package bootstrap

import (
	"context"
	"net/http"
	"os"

	"github.com/Agmer17/go-cosplay-backend/internal/auth"
	"github.com/Agmer17/go-cosplay-backend/internal/db"
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
}

func NewApp(configs Configs, rtr *gin.Engine) *App {

	// setting up the redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     configs.RedisHost,
		Password: "",
		DB:       0,
	})

	database := db.NewDatabase(configs.DatabaseUrl, configs.AppContext)

	serviceConfigs := NewServiceConfigs(serviceConfigsParams{
		GoogleOauthId:     configs.GoogleOauthClientId,
		GoogleOauthSecret: configs.GoogleOauthSecret,
		redisCli:          redisClient,
		Database:          database,
	})

	return &App{
		Router:         rtr,
		Database:       database,
		ServiceConfigs: serviceConfigs,
		RedisClient:    redisClient,
	}

}

type BootstrapHandler interface {
	RegisterRoutes(r gin.IRouter)
}

func (app *App) SetupRoutes() {

	// create something in here
	authHandler := auth.NewAuthHandler(app.ServiceConfigs.AuthService)

	var hs []BootstrapHandler = []BootstrapHandler{authHandler}

	api := app.Router.Group("/api")
	for _, h := range hs {
		h.RegisterRoutes(api)
	}

	app.Router.StaticFS("/uploads", http.Dir("./tmp/uploads"))
}

func (a *App) Run() {
	a.Router.Run("0.0.0.0:80")
}
