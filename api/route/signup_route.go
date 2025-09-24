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

func NewSignupRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	ur := repository.NewUserRepository(db, domain.CollectionUser)
	sr := repository.NewSheetRepository(db, domain.CollectionSheet)
	nr := repository.NewNotificationRepository(db, domain.CollectionNotification)

	sc := controller.SignupController{
		SignupUsecase:       usecase.NewSignupUsecase(ur, timeout),
		NotificationUsecase: usecase.NewNotificationUsecase(nr, ur, sr, timeout),
		Env:                 env,
	}
	group.POST("/signup", sc.Signup)
}
