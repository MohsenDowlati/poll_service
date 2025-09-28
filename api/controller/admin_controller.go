package controller

import (
	"net/http"

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
