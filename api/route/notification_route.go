package route

import (
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
	"github.com/gin-gonic/gin"
)

func NewNotificationRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	nr := repository.NewNotificationRepository(db, domain.CollectionNotification)
	ur := repository.NewUserRepository(db, domain.CollectionUser)

	nc := controller.NotificationController{
		NotificationUsecase: usecase.NewNotificationUsecase(nr, ur, timeout),
	}

	group.GET("/poll/notifications", nc.FetchPending)
	group.POST("/poll/notifications/:id/approve", nc.Approve)
	group.POST("/poll/notifications/:id/reject", nc.Reject)
}
