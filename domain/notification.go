package domain

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNotification = "notifications"

var ErrNotificationResolved = errors.New("notification already resolved")

type NotificationStatus string

const (
	NotificationPending  NotificationStatus = "pending"
	NotificationApproved NotificationStatus = "approved"
	NotificationRejected NotificationStatus = "rejected"
)

type Notification struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	UserID           primitive.ObjectID `bson:"userID"`
	UserName         string             `bson:"userName"`
	UserPhone        string             `bson:"userPhone"`
	UserOrganization string             `bson:"userOrganization"`
	Status           NotificationStatus `bson:"status"`
	CreatedAt        time.Time          `bson:"createdAt"`
	UpdatedAt        time.Time          `bson:"updatedAt"`
	ResolvedBy       primitive.ObjectID `bson:"resolvedBy,omitempty"`
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	FetchPending(ctx context.Context) ([]Notification, error)
	GetByID(ctx context.Context, id string) (Notification, error)
	UpdateStatus(ctx context.Context, id string, status NotificationStatus, resolvedBy primitive.ObjectID, updatedAt time.Time) error
}

type NotificationUsecase interface {
	CreateForNewUser(ctx context.Context, user *User) error
	FetchPending(ctx context.Context) ([]Notification, error)
	Approve(ctx context.Context, notificationID string, adminID string) (Notification, error)
	Reject(ctx context.Context, notificationID string, adminID string) (Notification, error)
}

type NotificationResponse struct {
	ID               string             `json:"id"`
	UserID           string             `json:"user_id"`
	UserName         string             `json:"user_name"`
	UserPhone        string             `json:"user_phone"`
	UserOrganization string             `json:"user_organization"`
	Status           NotificationStatus `json:"status"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	ResolvedBy       string             `json:"resolved_by,omitempty"`
}
