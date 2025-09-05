package controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

type SheetController struct {
	SheetuseCase domain.SheetUseCase
}

func (sc *SheetController) Create(c *gin.Context) {
	var sheet domain.Sheet

	err := c.ShouldBind(&sheet)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	userID := c.GetString("x-user-id")
	sheet.ID = primitive.NewObjectID()

	sheet.UserID, err = primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	sheet.CreatedAt = time.Now()
	sheet.UpdatedAt = time.Now()

	err = sc.SheetuseCase.Create(c, sheet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse{Message: "sheet created!"})

}

func (sc *SheetController) Fetch(c *gin.Context) {
	tasks, err := sc.SheetuseCase.GetAll(c)

	userID := c.GetString("x-user-id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "user id is empty"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (sc *SheetController) Delete(c *gin.Context) {
	id := c.Param("id")

	userID := c.GetString("x-user-id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "user id is empty"})
		return
	}

	err := sc.SheetuseCase.Delete(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "sheet deleted!"})
}
