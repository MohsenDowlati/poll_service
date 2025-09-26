package controller

import (
	"net/http"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
)

type RefreshTokenController struct {
	RefreshTokenUsecase domain.RefreshTokenUsecase
	Env                 *bootstrap.Env
}

// RefreshToken issues a new access/refresh token pair.
// @Summary Refresh authentication tokens
// @Description Exchange an existing refresh token for a new access/refresh token pair.
// @Tags Auth
// @Accept mpfd
// @Produce json
// @Param refreshToken formData string false "Refresh token when cookie is unavailable"
// @Success 200 {object} domain.RefreshTokenResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /api/v1/refresh [post]
func (rtc *RefreshTokenController) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(domain.RefreshTokenCookieName)
	if err != nil || refreshToken == "" {
		var request domain.RefreshTokenRequest

		if bindErr := c.ShouldBind(&request); bindErr != nil {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: bindErr.Error()})
			return
		}
		refreshToken = request.RefreshToken
	}

	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "User not found"})
		return
	}

	id, err := rtc.RefreshTokenUsecase.ExtractIDFromToken(refreshToken, rtc.Env.RefreshTokenSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "User not found"})
		return
	}

	user, err := rtc.RefreshTokenUsecase.GetUserByID(c, id)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "User not found"})
		return
	}

	newAccessToken, err := rtc.RefreshTokenUsecase.CreateAccessToken(&user, rtc.Env.AccessTokenSecret, rtc.Env.AccessTokenExpiryHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	newRefreshToken, err := rtc.RefreshTokenUsecase.CreateRefreshToken(&user, rtc.Env.RefreshTokenSecret, rtc.Env.RefreshTokenExpiryHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	setAuthCookies(c, rtc.Env, newAccessToken, newRefreshToken)

	refreshTokenResponse := domain.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	c.JSON(http.StatusOK, refreshTokenResponse)
}
