package route

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/middleware"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/gin-gonic/gin"
)

func Setup(env *bootstrap.Env, timeout time.Duration, db mongo.Database, gin *gin.Engine) {
	publicRouter := gin.Group("/api/v1")
	// All Public APIs
	NewSignupRouter(env, timeout, db, publicRouter)
	NewLoginRouter(env, timeout, db, publicRouter)
	NewRefreshTokenRouter(env, timeout, db, publicRouter)
	NewClientPollRouter(env, timeout, db, publicRouter)

	protectedRouter := gin.Group("/api/v1")
	// Middleware to verify AccessToken
	protectedRouter.Use(middleware.JwtAuthMiddleware(env.AccessTokenSecret))
	// All Private APIs
	NewProfileRouter(env, timeout, db, protectedRouter)
	NewAdminPollRouter(env, timeout, db, protectedRouter)
	NewNotificationRouter(env, timeout, db, protectedRouter)
	NewSheetRouter(env, db, timeout, protectedRouter)
	NewAdminRouter(env, timeout, db, protectedRouter)
}
