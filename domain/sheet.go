package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionSheet = "sheets"

type SheetStatus string

const (
	SheetStatusPending   SheetStatus = "pending"
	SheetStatusPublished SheetStatus = "published"
	SheetStatusRejected  SheetStatus = "rejected"
)

type Sheet struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"userID" json:"-"`
	Title       string             `bson:"title" form:"title" binding:"required" json:"title"`
	Venue       string             `bson:"venue" form:"venue" binding:"required" json:"venue"`
	Description string             `bson:"description" form:"description" json:"description"`
	Status      SheetStatus        `bson:"status" json:"status"`
	ApprovedBy  primitive.ObjectID `bson:"approvedBy,omitempty" json:"approved_by,omitempty"`
	ApprovedAt  time.Time          `bson:"approvedAt,omitempty" json:"approved_at,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"-"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"-"`
}

type SheetRepository interface {
	Create(ctx context.Context, sheet Sheet) error
	GetAll(ctx context.Context) ([]Sheet, error)
	GetByUserID(ctx context.Context, userID string) ([]Sheet, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (Sheet, error)
	UpdateStatus(ctx context.Context, id string, status SheetStatus, approvedBy primitive.ObjectID, approvedAt time.Time) error
}

type SheetUseCase interface {
	Create(c context.Context, sheet Sheet) error
	GetAll(c context.Context) ([]Sheet, error)
	Delete(c context.Context, id string) error
	GetByUserID(c context.Context, userID string) ([]Sheet, error)
	GetByID(c context.Context, id string) (Sheet, error)
	UpdateStatus(c context.Context, id string, status SheetStatus, approvedBy primitive.ObjectID, approvedAt time.Time) error
}
