package route

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/middleware"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
	"github.com/gin-gonic/gin"
)

func NewAdminRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	ur := repository.NewUserRepository(db, domain.CollectionUser)

	ac := controller.AdminController{
		AdminUsecase: usecase.NewAdminUsecase(ur, timeout),
	}

	adminGroup := group.Group("")
	adminGroup.Use(middleware.RequireRoles(domain.SuperAdmin))
	adminGroup.GET("/admin/users", ac.Fetch)
	adminGroup.POST("/admin/users/status", ac.UpdateStatus)
	adminGroup.DELETE("/admin/users/:id", ac.Delete)
}
