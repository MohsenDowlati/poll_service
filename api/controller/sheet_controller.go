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

	//TODO: check if user is valid
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

	//TODO: Send a validation for super admin
	c.JSON(http.StatusCreated, domain.SuccessResponse{Message: "sheet created!"})

}

func (sc *SheetController) Fetch(c *gin.Context) {
	userID := c.GetString("x-user-id")
	userType := c.GetString("x-user-type")

	var sheets []domain.Sheet
	var err error

	if domain.UserType(userType) == domain.SuperAdmin {
		sheets, err = sc.SheetuseCase.GetAll(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
			return
		}
	} else if domain.UserType(userType) == domain.VerifiedAdmin {
		sheets, err = sc.SheetuseCase.GetByUserID(c, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "invalid user type"})
		return
	}

	c.JSON(http.StatusOK, sheets)
}

func (sc *SheetController) FetchByID(c *gin.Context) {

	//TODO: check if id is okay for admin

	id := c.Param("id")

	t := c.GetString("x-user-type")

	if domain.UserType(t) != domain.SuperAdmin {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "user id is empty"})
		return
	}

	var sheet domain.Sheet

	sheet, err := sc.SheetuseCase.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, sheet)
}

func (sc *SheetController) Delete(c *gin.Context) {
	//TODO: check auth
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
