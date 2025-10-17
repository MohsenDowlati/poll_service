package controller

import (
	"net/http"
	"strings"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
	AdminUsecase domain.AdminUsecase
}

// Fetch lists users with pagination.
// @Summary List users
// @Description Retrieve users with pagination (super admin only).
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} domain.UserListResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/admin/users [get]
func (ac *AdminController) Fetch(c *gin.Context) {
	if domain.UserType(c.GetString("x-user-type")) != domain.SuperAdmin {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	pagination := extractPagination(c)

	users, total, err := ac.AdminUsecase.Fetch(c, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	items := make([]domain.UserListItem, 0, len(users))
	for _, user := range users {
		items = append(items, domain.UserListItem{
			ID:           user.ID.Hex(),
			Name:         user.Name,
			Phone:        user.Phone,
			Organization: user.Organization,
			Admin:        user.Admin,
			IsVerified:   user.IsVerified,
		})
	}

	response := domain.UserListResponse{
		Data:       items,
		Pagination: domain.NewPaginationResult(pagination, total),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateStatus updates a user's verification status (super admin only).
// @Summary Update user status
// @Description Activate or deactivate a user by toggling their verification status.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User identifier"
// @Param payload body domain.AdminRequest true "Status payload"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/admin/users/{id}/status [put]
func (ac *AdminController) UpdateStatus(c *gin.Context) {
	if domain.UserType(c.GetString("x-user-type")) != domain.SuperAdmin {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	var payload domain.AdminRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	bodyID := strings.TrimSpace(payload.UserID)

	if bodyID == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "user id is required"})
		return
	}

	if payload.IsVerified == nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "is_verified is required"})
		return
	}

	if err := ac.AdminUsecase.VerifyUser(c, bodyID, *payload.IsVerified); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	message := "user deactivated successfully"
	if *payload.IsVerified {
		message = "user activated successfully"
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: message})
}
