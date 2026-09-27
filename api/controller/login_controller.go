package controller

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
)

type LoginController struct {
	LoginUsecase domain.LoginUsecase
	Env          *bootstrap.Env
}

// Login authenticates a user and issues new tokens.
// @Summary Login user
// @Description Authenticate a user using email and password and receive tokens.
// @Tags Auth
// @Accept mpfd
// @Produce json
// @Param email formData string false "Registered email address"
// @Param phone formData string false "Registered phone number"
// @Param password formData string true "Password"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /api/v1/login [post]
func (lc *LoginController) Login(c *gin.Context) {
	var request domain.LoginRequest

	err := c.ShouldBind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	var user domain.User
	identifier := strings.TrimSpace(request.Phone)
	switch {
	case identifier != "":
		user, err = lc.LoginUsecase.GetUserByPhone(c, identifier)
		if err != nil {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "invalid credentials"})
			return
		}
	case strings.TrimSpace(request.Email) != "":
		identifier = strings.TrimSpace(request.Email)
		user, err = lc.LoginUsecase.GetUserByEmail(c, identifier)
		if err != nil {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "invalid credentials"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "either phone or email is required"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)) != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Invalid credentials"})
		return
	}

	if !user.IsVerified && user.Admin != domain.SuperAdmin {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{Message: "account is pending approval"})
		return
	}
	if user.Admin == domain.CanceledUser {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{Message: "account is disabled"})
		return
	}

	accessToken, err := lc.LoginUsecase.CreateAccessToken(&user, lc.Env.AccessTokenSecret, lc.Env.AccessTokenExpiryHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	refreshToken, err := lc.LoginUsecase.CreateRefreshToken(&user, lc.Env.RefreshTokenSecret, lc.Env.RefreshTokenExpiryHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	setAuthCookies(c, lc.Env, accessToken, refreshToken)
	setSessionRoleCookie(c, lc.Env, user.Admin)

	loginResponse := domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(http.StatusOK, loginResponse)
}
