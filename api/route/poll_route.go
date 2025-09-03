package route

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/api/controller"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
	"github.com/gin-gonic/gin"
	"time"
)

func NewAdminPollRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	apr := repository.NewPollRepository(db, domain.CollectionPoll)
	apc := &controller.PollAdminController{
		PollAdminUsecase: usecase.NewPollAdminUsecase(apr, timeout),
	}

	group.POST("/create", apc.Create)
	group.POST("/edit?id={id}", apc.Edit)
	group.GET("/admin/fetch?id={id}", apc.GetBySheetID)
	group.PUT("/delete?id={id}", apc.Delete)
}

func NewClientPollRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	cpr := repository.NewPollRepository(db, domain.CollectionPoll)
	cpc := &controller.PollClientController{
		PollClientUsecse: usecase.NewPollClientUsecase(cpr, timeout),
	}

	group.POST("/submit", cpc.Submit)
	group.GET("/client/fetch?id={id}", cpc.Fetch)
}
