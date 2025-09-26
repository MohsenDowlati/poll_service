package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SheetController struct {
	SheetuseCase        domain.SheetUseCase
	NotificationUsecase domain.NotificationUsecase
}

// Create registers a new sheet.
// @Summary Create sheet
// @Description Create a new sheet (verified admin or super admin).
// @Tags Sheets
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param title formData string true "Sheet title"
// @Param venue formData string true "Sheet venue"
// @Param description formData string false "Sheet description"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/sheet/create [post]
func (sc *SheetController) Create(c *gin.Context) {
	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))

	if userID == "" || (userType != domain.VerifiedAdmin && userType != domain.SuperAdmin) {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	var sheet domain.Sheet

	if err := c.ShouldBind(&sheet); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "invalid user identifier"})
		return
	}

	now := time.Now()

	sheet.ID = primitive.NewObjectID()
	sheet.UserID = ownerID
	sheet.CreatedAt = now
	sheet.UpdatedAt = now

	if userType == domain.SuperAdmin {
		sheet.Status = domain.SheetStatusPublished
		sheet.ApprovedBy = ownerID
		sheet.ApprovedAt = now
	} else {
		sheet.Status = domain.SheetStatusPending
	}

	if err = sc.SheetuseCase.Create(c, sheet); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	if userType == domain.VerifiedAdmin && sc.NotificationUsecase != nil {
		if err = sc.NotificationUsecase.CreateForSheet(c, &sheet); err != nil {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
			return
		}
	}

	message := "sheet created!"
	if sheet.Status == domain.SheetStatusPending {
		message = "sheet submitted for approval"
	} else {
		message = "sheet published"
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse{Message: message})
}

// Fetch lists sheets for the authenticated admin.
// @Summary List sheets
// @Description Retrieve sheets depending on admin role.
// @Tags Sheets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Sheet
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/sheet/fetch [get]
func (sc *SheetController) Fetch(c *gin.Context) {
	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))

	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	var (
		sheets []domain.Sheet
		err    error
	)

	switch userType {
	case domain.SuperAdmin:
		sheets, err = sc.SheetuseCase.GetAll(c)
	case domain.VerifiedAdmin:
		sheets, err = sc.SheetuseCase.GetByUserID(c, userID)
	default:
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, sheets)
}

func (sc *SheetController) FetchByID(c *gin.Context) {
	identifier := c.Param("id")
	if identifier == "" {
		identifier = c.Query("id")
	}

	if identifier == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id is required"})
		return
	}

	sheet, err := sc.SheetuseCase.GetByID(c, identifier)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))

	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if userType != domain.SuperAdmin && sheet.UserID.Hex() != userID {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, sheet)
}

// Delete removes a sheet owned by the authenticated admin.
// @Summary Delete sheet
// @Description Delete a sheet by identifier.
// @Tags Sheets
// @Produce json
// @Security BearerAuth
// @Param id query string true "Sheet identifier"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/sheet/delete [put]
func (sc *SheetController) Delete(c *gin.Context) {
	identifier := c.Param("id")
	if identifier == "" {
		identifier = c.Query("id")
	}

	if identifier == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id is required"})
		return
	}

	sheet, err := sc.SheetuseCase.GetByID(c, identifier)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))

	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if userType != domain.SuperAdmin && sheet.UserID.Hex() != userID {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if err = sc.SheetuseCase.Delete(c, identifier); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "sheet deleted!"})
}
