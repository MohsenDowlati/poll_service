package usecase

import (
	"context"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/internal/tokenutil"
)

type loginUsecase struct {
	userRepository domain.UserRepository
	contextTimeout time.Duration
}

func (lu *loginUsecase) GetUserByPhone(c context.Context, phone string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	return lu.userRepository.GetByPhone(ctx, phone)
}

func (lu *loginUsecase) GetUserByEmail(c context.Context, email string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()

	return lu.userRepository.GetByEmail(ctx, email)
}

func NewLoginUsecase(userRepository domain.UserRepository, timeout time.Duration) domain.LoginUsecase {
	return &loginUsecase{
		userRepository: userRepository,
		contextTimeout: timeout,
	}
}

func (lu *loginUsecase) CreateAccessToken(user *domain.User, secret string, expiry int) (accessToken string, err error) {
	return tokenutil.CreateAccessToken(user, secret, expiry)
}

func (lu *loginUsecase) CreateRefreshToken(user *domain.User, secret string, expiry int) (refreshToken string, err error) {
	return tokenutil.CreateRefreshToken(user, secret, expiry)
}
