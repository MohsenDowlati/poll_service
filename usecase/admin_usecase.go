package usecase

import (
	"context"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
)

type adminUsecase struct {
	userRepository domain.UserRepository
	contextTimeout time.Duration
}

func (au *adminUsecase) Delete(c context.Context, userID string) error {
	ctx, cancel := context.WithTimeout(c, au.contextTimeout)
	defer cancel()

	return au.userRepository.DeleteUser(ctx, userID)
}

func NewAdminUsecase(userRepository domain.UserRepository, timeout time.Duration) domain.AdminUsecase {
	return &adminUsecase{
		userRepository: userRepository,
		contextTimeout: timeout,
	}
}

func (au *adminUsecase) VerifyUser(c context.Context, userID string, isVerified bool) error {
	ctx, cancel := context.WithTimeout(c, au.contextTimeout)
	defer cancel()

	adminRole := domain.VerifiedAdmin
	if !isVerified {
		adminRole = domain.CanceledUser
	}

	return au.userRepository.UpdateAdminStatus(ctx, userID, adminRole, isVerified)
}

func (au *adminUsecase) Fetch(c context.Context, pagination domain.PaginationQuery) ([]domain.User, int64, error) {
	ctx, cancel := context.WithTimeout(c, au.contextTimeout)
	defer cancel()

	return au.userRepository.Fetch(ctx, pagination)
}
