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
		ID:        notification.ID.Hex(),
		Type:      notification.Type,
		SubjectID: notification.SubjectID.Hex(),
		Status:    notification.Status,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}

	if !notification.UserID.IsZero() {
		response.UserID = notification.UserID.Hex()
	}
	if notification.UserName != "" {
		response.UserName = notification.UserName
	}
	if notification.UserPhone != "" {
		response.UserPhone = notification.UserPhone
	}
	if notification.UserOrganization != "" {
		response.UserOrganization = notification.UserOrganization
	}
	if !notification.SheetID.IsZero() {
		response.SheetID = notification.SheetID.Hex()
	}
	if notification.SheetTitle != "" {
		response.SheetTitle = notification.SheetTitle
	}
	if notification.SheetVenue != "" {
		response.SheetVenue = notification.SheetVenue
	}
	if !notification.ResolvedBy.IsZero() {
		response.ResolvedBy = notification.ResolvedBy.Hex()
	}

	return response
}
