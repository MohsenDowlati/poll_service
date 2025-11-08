package server

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	route "github.com/amitshekhariitbhu/go-backend-clean-architecture/api/route"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/internal/validation"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Run bootstraps the application and starts the HTTP server.
func Run() error {
	app := bootstrap.App()

	env := app.Env

	db := app.Mongo.Database(env.DBName)
	defer app.CloseDBConnection()

	timeout := time.Duration(env.ContextTimeout) * time.Second

	if err := bootstrap.SeedSuperAdmin(env, db, timeout); err != nil {
		return fmt.Errorf("seed super admin: %w", err)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
			return validation.Phone(fl.Field().String())
		}); err != nil {
			return fmt.Errorf("register phone validator: %w", err)
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
	formattedOrigins := expandLocalhostOrigins(allowedOrigins)

	if len(formattedOrigins) == 0 {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false
	} else {
		corsConfig.AllowOrigins = formattedOrigins
	}

	setGinMode(env.AppEnv)
	router := gin.Default()
	router.Use(cors.New(corsConfig))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	route.Setup(env, timeout, db, router)

	if err := router.Run(env.ServerAddress); err != nil {
		return fmt.Errorf("run gin server: %w", err)
	}

	return nil
}

func setGinMode(appEnv string) {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "production", "prod", "release":
		gin.SetMode(gin.ReleaseMode)
	case "test", "testing":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

func expandLocalhostOrigins(origins []string) []string {
	formatted := make([]string, 0, len(origins))
	seen := make(map[string]struct{})

	appendUnique := func(origin string) {
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		formatted = append(formatted, origin)
	}

	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}

		appendUnique(trimmed)

		parsed, err := url.Parse(trimmed)
		if err != nil || parsed.Host == "" {
			continue
		}

		if strings.Contains(parsed.Host, "localhost") {
			clone := *parsed
			clone.Host = strings.Replace(parsed.Host, "localhost", "127.0.0.1", 1)
			appendUnique(clone.String())
		}

		if strings.Contains(parsed.Host, "127.0.0.1") {
			clone := *parsed
			clone.Host = strings.Replace(parsed.Host, "127.0.0.1", "localhost", 1)
			appendUnique(clone.String())
		}
	}

	return formatted
}
