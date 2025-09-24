package controller

import (
	"errors"
	"net/http"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	NotificationUsecase domain.NotificationUsecase
}

func (nc *NotificationController) FetchPending(c *gin.Context) {
	if domain.UserType(c.GetString("x-user-type")) != domain.SuperAdmin {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	notifications, err := nc.NotificationUsecase.FetchPending(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	var response []domain.NotificationResponse
	for _, notification := range notifications {
		response = append(response, mapNotificationToResponse(notification))
	}

	c.JSON(http.StatusOK, response)
}

func (nc *NotificationController) Approve(c *gin.Context) {
	if domain.UserType(c.GetString("x-user-type")) != domain.SuperAdmin {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	adminID := c.GetString("x-user-id")
	notificationID := c.Param("id")

	notification, err := nc.NotificationUsecase.Approve(c, notificationID, adminID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrNotificationResolved) {
			status = http.StatusConflict
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapNotificationToResponse(notification))
}

func (nc *NotificationController) Reject(c *gin.Context) {
	if domain.UserType(c.GetString("x-user-type")) != domain.SuperAdmin {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	adminID := c.GetString("x-user-id")
	notificationID := c.Param("id")

	notification, err := nc.NotificationUsecase.Reject(c, notificationID, adminID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrNotificationResolved) {
			status = http.StatusConflict
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapNotificationToResponse(notification))
}

func mapNotificationToResponse(notification domain.Notification) domain.NotificationResponse {
	response := domain.NotificationResponse{
		ID:               notification.ID.Hex(),
		UserID:           notification.UserID.Hex(),
		UserName:         notification.UserName,
		UserPhone:        notification.UserPhone,
		UserOrganization: notification.UserOrganization,
		Status:           notification.Status,
		CreatedAt:        notification.CreatedAt,
		UpdatedAt:        notification.UpdatedAt,
	}

	if !notification.ResolvedBy.IsZero() {
		response.ResolvedBy = notification.ResolvedBy.Hex()
	}

	return response
}
