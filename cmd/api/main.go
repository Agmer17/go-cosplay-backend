package main

import (
	"context"

	"github.com/Agmer17/go-cosplay-backend/internal/bootstrap"
	"github.com/gin-gonic/gin"
)

func main() {

	appConfigs := bootstrap.NewConfigs(context.Background())

	mainRouter := gin.New()
	gin.SetMode(gin.ReleaseMode)
	mainRouter.Use(gin.Logger())
	mainRouter.Use(gin.Recovery())

	app := bootstrap.NewApp(appConfigs, mainRouter)

	app.SetupRoutes()

	defer app.Database.Pool.Close()
	defer app.RedisClient.Close()

	app.Run()

	println("success init")

}
