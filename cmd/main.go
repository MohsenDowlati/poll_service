package main

import (
	"log"
	"strings"
	"time"

	route "github.com/amitshekhariitbhu/go-backend-clean-architecture/api/route"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	_ "github.com/amitshekhariitbhu/go-backend-clean-architecture/docs"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/internal/validation"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
			return validation.Phone(fl.Field().String())
		}); err != nil {
			log.Fatalf("failed to register phone validator: %v", err)
		}
	}

	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	allowedOrigins := strings.Split(env.CORSAllowedOrigins, ",")
	formattedOrigins := make([]string, 0, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			formattedOrigins = append(formattedOrigins, trimmed)
		}
	}

	if len(formattedOrigins) == 0 {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false
	} else {
		corsConfig.AllowOrigins = formattedOrigins
	}

	gin.Use(cors.New(corsConfig))

	gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	route.Setup(env, timeout, db, gin)

	gin.Run(env.ServerAddress)
}
