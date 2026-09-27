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

func NewAdminPollRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	apr := repository.NewPollRepository(db, domain.CollectionPoll)
	asr := repository.NewSheetRepository(db, domain.CollectionSheet)
	apc := &controller.PollAdminController{
		PollAdminUsecase: usecase.NewPollAdminUsecase(apr, timeout),
		SheetUsecase:     usecase.NewSheetUseCase(asr, nil, timeout),
	}

	adminGroup := group.Group("")
	adminGroup.Use(middleware.RequireRoles(domain.VerifiedAdmin, domain.SuperAdmin))
	adminGroup.POST("/create", apc.Create)
	adminGroup.POST("/edit", apc.Edit)
	adminGroup.GET("/admin/fetch", apc.GetBySheetID)
	adminGroup.PUT("/delete", apc.Delete)
}

func NewClientPollRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	cpr := repository.NewPollRepository(db, domain.CollectionPoll)
	sr := repository.NewSheetRepository(db, domain.CollectionSheet)
	cpc := &controller.PollClientController{
		PollClientUsecse: usecase.NewPollClientUsecase(cpr, sr, timeout),
	}

	group.POST("/submit", cpc.Submit)
	group.GET("/client/fetch", cpc.Fetch)
}
