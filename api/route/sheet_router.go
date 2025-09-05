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

func NewSheetRouter(env *bootstrap.Env, db mongo.Database, contextTimeout time.Duration, group *gin.RouterGroup) {
	sr := repository.NewSheetRepository(db, domain.CollectionSheet)
	sc := controller.SheetController{
		SheetuseCase: usecase.NewSheetUseCase(sr, contextTimeout),
	}

	group.POST("/sheet/create", sc.Create)
	group.PUT("/sheet/delete?id={id}", sc.Delete)
	group.GET("/sheet/fetch", sc.Fetch)
	group.GET("/sheet/fetch?id={id}", sc.FetchByID)
}
