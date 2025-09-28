package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type notificationUsecase struct {
	notificationRepository domain.NotificationRepository
	userRepository         domain.UserRepository
	sheetRepository        domain.SheetRepository
	contextTimeout         time.Duration
}

func NewNotificationUsecase(notificationRepository domain.NotificationRepository, userRepository domain.UserRepository, sheetRepository domain.SheetRepository, timeout time.Duration) domain.NotificationUsecase {
	return &notificationUsecase{
		notificationRepository: notificationRepository,
		userRepository:         userRepository,
		sheetRepository:        sheetRepository,
		contextTimeout:         timeout,
	}
}

func (nu *notificationUsecase) CreateForNewUser(c context.Context, user *domain.User) error {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	notification := domain.Notification{
		ID:               primitive.NewObjectID(),
		Type:             domain.NotificationTypeUserSignup,
		SubjectID:        user.ID,
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

func (nu *notificationUsecase) CreateForSheet(c context.Context, sheet *domain.Sheet) error {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	notification := domain.Notification{
		ID:         primitive.NewObjectID(),
		Type:       domain.NotificationTypeSheetApproval,
		SubjectID:  sheet.ID,
		UserID:     sheet.UserID,
		SheetID:    sheet.ID,
		SheetTitle: sheet.Title,
		SheetVenue: sheet.Venue,
		Status:     domain.NotificationPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return nu.notificationRepository.Create(ctx, &notification)
}

func (nu *notificationUsecase) FetchPending(c context.Context, pagination domain.PaginationQuery) ([]domain.Notification, int64, error) {
	ctx, cancel := context.WithTimeout(c, nu.contextTimeout)
	defer cancel()

	return nu.notificationRepository.FetchPending(ctx, pagination)
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

	resolverID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return domain.Notification{}, err
	}

	resolvedAt := time.Now()

	switch notification.Type {
	case domain.NotificationTypeUserSignup:
		if err = nu.userRepository.UpdateAdminStatus(ctx, notification.UserID.Hex(), domain.VerifiedAdmin, true); err != nil {
			return domain.Notification{}, err
		}
	case domain.NotificationTypeSheetApproval:
		if notification.SheetID.IsZero() {
			return domain.Notification{}, fmt.Errorf("sheet identifier missing for notification %s", notificationID)
		}
		if err = nu.sheetRepository.UpdateStatus(ctx, notification.SheetID.Hex(), domain.SheetStatusPublished, resolverID, resolvedAt); err != nil {
			return domain.Notification{}, err
		}
	default:
		return domain.Notification{}, fmt.Errorf("unsupported notification type %s", notification.Type)
	}

	if err = nu.notificationRepository.UpdateStatus(ctx, notificationID, domain.NotificationApproved, resolverID, resolvedAt); err != nil {
		return domain.Notification{}, err
	}

	if err = nu.notificationRepository.Delete(ctx, notificationID); err != nil {
		return domain.Notification{}, err
	}

	notification.Status = domain.NotificationApproved
	notification.ResolvedBy = resolverID
	notification.UpdatedAt = resolvedAt

	return notification, nil
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

	resolverID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return domain.Notification{}, err
	}

	resolvedAt := time.Now()

	switch notification.Type {
	case domain.NotificationTypeUserSignup:
		if err = nu.userRepository.UpdateAdminStatus(ctx, notification.UserID.Hex(), domain.CanceledUser, false); err != nil {
			return domain.Notification{}, err
		}
	case domain.NotificationTypeSheetApproval:
		if notification.SheetID.IsZero() {
			return domain.Notification{}, fmt.Errorf("sheet identifier missing for notification %s", notificationID)
		}
		if err = nu.sheetRepository.UpdateStatus(ctx, notification.SheetID.Hex(), domain.SheetStatusRejected, resolverID, resolvedAt); err != nil {
			return domain.Notification{}, err
		}
	default:
		return domain.Notification{}, fmt.Errorf("unsupported notification type %s", notification.Type)
	}

	if err = nu.notificationRepository.UpdateStatus(ctx, notificationID, domain.NotificationRejected, resolverID, resolvedAt); err != nil {
		return domain.Notification{}, err
	}

	if err = nu.notificationRepository.Delete(ctx, notificationID); err != nil {
		return domain.Notification{}, err
	}

	notification.Status = domain.NotificationRejected
	notification.ResolvedBy = resolverID
	notification.UpdatedAt = resolvedAt

	return notification, nil
}
