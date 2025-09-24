package usecase

import (
	"context"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type notificationUsecase struct {
	notificationRepository domain.NotificationRepository
	userRepository         domain.UserRepository
	contextTimeout         time.Duration
}

func NewNotificationUsecase(notificationRepository domain.NotificationRepository, userRepository domain.UserRepository, timeout time.Duration) domain.NotificationUsecase {
	return &notificationUsecase{
		notificationRepository: notificationRepository,
		userRepository:         userRepository,
		contextTimeout:         timeout,
	}
}

func (nu *notificationUsecase) CreateForNewUser(c context.Context, user *domain.User) error {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	notification := domain.Notification{
		ID:               primitive.NewObjectID(),
		UserID:           user.ID,
		UserName:         user.Name,
		UserPhone:        user.Phone,
		UserOrganization: user.Organization,
		Status:           domain.NotificationPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return nu.notificationRepository.Create(ctx, &notification)
}

func (nu *notificationUsecase) FetchPending(c context.Context) ([]domain.Notification, error) {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	return nu.notificationRepository.FetchPending(ctx)
}

func (nu *notificationUsecase) Approve(c context.Context, notificationID string, adminID string) (domain.Notification, error) {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	notification, err := nu.notificationRepository.GetByID(ctx, notificationID)
	if err != nil {
		return domain.Notification{}, err
	}

	if notification.Status != domain.NotificationPending {
		return domain.Notification{}, domain.ErrNotificationResolved
	}

	objectID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return domain.Notification{}, err
	}

	err = nu.userRepository.UpdateAdminStatus(ctx, notification.UserID.Hex(), domain.VerifiedAdmin, true)
	if err != nil {
		return domain.Notification{}, err
	}

	updatedAt := time.Now()
	err = nu.notificationRepository.UpdateStatus(ctx, notificationID, domain.NotificationApproved, objectID, updatedAt)
	if err != nil {
		return domain.Notification{}, err
	}

	return nu.notificationRepository.GetByID(ctx, notificationID)
}

func (nu *notificationUsecase) Reject(c context.Context, notificationID string, adminID string) (domain.Notification, error) {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	notification, err := nu.notificationRepository.GetByID(ctx, notificationID)
	if err != nil {
		return domain.Notification{}, err
	}

	if notification.Status != domain.NotificationPending {
		return domain.Notification{}, domain.ErrNotificationResolved
	}

	objectID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return domain.Notification{}, err
	}

	err = nu.userRepository.UpdateAdminStatus(ctx, notification.UserID.Hex(), domain.CanceledUser, false)
	if err != nil {
		return domain.Notification{}, err
	}

	updatedAt := time.Now()
	err = nu.notificationRepository.UpdateStatus(ctx, notificationID, domain.NotificationRejected, objectID, updatedAt)
	if err != nil {
		return domain.Notification{}, err
	}

	return nu.notificationRepository.GetByID(ctx, notificationID)
}
