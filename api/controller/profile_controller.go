package controller

import (
	"errors"
	"net/http"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

type ProfileController struct {
	ProfileUsecase domain.ProfileUsecase
}

// Fetch returns the authenticated user profile.
// @Summary Get current user profile
// @Description Retrieve the profile for the authenticated user.
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Profile
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/profile [get]
func (pc *ProfileController) Fetch(c *gin.Context) {
	userID := c.GetString("x-user-id")
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "not authorized"})
		return
	}

	profile, err := pc.ProfileUsecase.GetProfileByID(c, userID)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) || errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "unable to load profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
