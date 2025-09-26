package main

import (
	"time"

	route "github.com/amitshekhariitbhu/go-backend-clean-architecture/api/route"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	_ "github.com/amitshekhariitbhu/go-backend-clean-architecture/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Poll Service API
// @version         1.0
// @description     API documentation for the Poll service.
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {

	app := bootstrap.App()

	env := app.Env

	db := app.Mongo.Database(env.DBName)
	defer app.CloseDBConnection()

	timeout := time.Duration(env.ContextTimeout) * time.Second

	gin := gin.Default()

	gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	route.Setup(env, timeout, db, gin)

	gin.Run(env.ServerAddress)
}
