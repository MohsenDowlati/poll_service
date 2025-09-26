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

// FetchPending lists pending notifications for approval.
// @Summary List pending notifications
// @Description Retrieve pending notifications (super admin only).
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.NotificationResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/poll/notifications [get]
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

// Approve marks a notification as approved.
// @Summary Approve notification
// @Description Approve a pending notification (super admin only).
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification identifier"
// @Success 200 {object} domain.NotificationResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/poll/notifications/{id}/approve [post]
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

// Reject marks a notification as rejected.
// @Summary Reject notification
// @Description Reject a pending notification (super admin only).
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification identifier"
// @Success 200 {object} domain.NotificationResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/poll/notifications/{id}/reject [post]
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
