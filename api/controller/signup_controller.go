package controller

import (
	"net/http"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type SignupController struct {
	SignupUsecase       domain.SignupUsecase
	NotificationUsecase domain.NotificationUsecase
	Env                 *bootstrap.Env
}

// Signup registers a new user and issues auth tokens.
// @Summary Register a new user
// @Description Register a new user and receive access and refresh tokens.
// @Tags Auth
// @Accept mpfd
// @Produce json
// @Param name formData string true "Full name"
// @Param phone formData string true "Phone number"
// @Param organization formData string true "Organization name"
// @Param password formData string true "Password"
// @Success 200 {object} domain.SignupResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Router /api/v1/signup [post]
func (sc *SignupController) Signup(c *gin.Context) {
	var request domain.SignupRequest

	err := c.ShouldBind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	_, err = sc.SignupUsecase.GetUserByPhone(c, request.Phone)
	if err == nil {
		c.JSON(http.StatusConflict, domain.ErrorResponse{Message: "User already exists with the given phone"})
		return
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	user := domain.User{
		ID:           primitive.NewObjectID(),
		Name:         request.Name,
		Email:        request.Phone,
		Phone:        request.Phone,
		Organization: request.Organization,
		Password:     string(encryptedPassword),
		Admin:        domain.NewUser,
		IsVerified:   false,
		CreatedAt:    time.Now(),
	}

	err = sc.SignupUsecase.Create(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	if sc.NotificationUsecase != nil {
		if err = sc.NotificationUsecase.CreateForNewUser(c, &user); err != nil {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
			return
		}
	}

	accessToken, err := sc.SignupUsecase.CreateAccessToken(&user, sc.Env.AccessTokenSecret, sc.Env.AccessTokenExpiryHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	refreshToken, err := sc.SignupUsecase.CreateRefreshToken(&user, sc.Env.RefreshTokenSecret, sc.Env.RefreshTokenExpiryHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	setAuthCookies(c, sc.Env, accessToken, refreshToken)

	signupResponse := domain.SignupResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(http.StatusOK, signupResponse)
}
