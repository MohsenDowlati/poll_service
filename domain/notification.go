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

type NotificationType string

const (
	NotificationPending  NotificationStatus = "pending"
	NotificationApproved NotificationStatus = "approved"
	NotificationRejected NotificationStatus = "rejected"
)

const (
	NotificationTypeUserSignup    NotificationType = "user_signup"
	NotificationTypeSheetApproval NotificationType = "sheet_approval"
)

type Notification struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Type             NotificationType   `bson:"type"`
	SubjectID        primitive.ObjectID `bson:"subjectID"`
	UserID           primitive.ObjectID `bson:"userID,omitempty"`
	UserName         string             `bson:"userName,omitempty"`
	UserPhone        string             `bson:"userPhone,omitempty"`
	UserOrganization string             `bson:"userOrganization,omitempty"`
	SheetID          primitive.ObjectID `bson:"sheetID,omitempty"`
	SheetTitle       string             `bson:"sheetTitle,omitempty"`
	SheetVenue       string             `bson:"sheetVenue,omitempty"`
	Status           NotificationStatus `bson:"status"`
	CreatedAt        time.Time          `bson:"createdAt"`
	UpdatedAt        time.Time          `bson:"updatedAt"`
	ResolvedBy       primitive.ObjectID `bson:"resolvedBy,omitempty"`
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	FetchPending(ctx context.Context, pagination PaginationQuery) ([]Notification, int64, error)
	GetByID(ctx context.Context, id string) (Notification, error)
	UpdateStatus(ctx context.Context, id string, status NotificationStatus, resolvedBy primitive.ObjectID, updatedAt time.Time) error
	Delete(ctx context.Context, id string) error
}

type NotificationUsecase interface {
	CreateForNewUser(ctx context.Context, user *User) error
	CreateForSheet(ctx context.Context, sheet *Sheet) error
	FetchPending(ctx context.Context, pagination PaginationQuery) ([]Notification, int64, error)
	Approve(ctx context.Context, notificationID string, adminID string) (Notification, error)
	Reject(ctx context.Context, notificationID string, adminID string) (Notification, error)
}

type NotificationResponse struct {
	ID               string             `json:"id"`
	Type             NotificationType   `json:"type"`
	SubjectID        string             `json:"subject_id"`
	UserID           string             `json:"user_id,omitempty"`
	UserName         string             `json:"user_name,omitempty"`
	UserPhone        string             `json:"user_phone,omitempty"`
	UserOrganization string             `json:"user_organization,omitempty"`
	SheetID          string             `json:"sheet_id,omitempty"`
	SheetTitle       string             `json:"sheet_title,omitempty"`
	SheetVenue       string             `json:"sheet_venue,omitempty"`
	Status           NotificationStatus `json:"status"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	ResolvedBy       string             `json:"resolved_by,omitempty"`
}
