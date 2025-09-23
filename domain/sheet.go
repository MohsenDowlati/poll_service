package domain

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const CollectionSheet = "sheets"

type Sheet struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"userID" json:"-"`
	Title       string             `bson:"title" form:"title" binding:"required" json:"title"`
	Venue       string             `bson:"venue" form:"venue" binding:"required" json:"venue"`
	Description string             `bson:"description" form:"description" json:"description"`
	CreatedAt   time.Time          `bson:"createdAt" json:"-"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"-"`
}

type SheetRepository interface {
	Create(ctx context.Context, sheet Sheet) error
	GetAll(ctx context.Context) ([]Sheet, error)
	GetByUserID(ctx context.Context, userID string) ([]Sheet, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (Sheet, error)
}

type SheetUseCase interface {
	Create(c context.Context, sheet Sheet) error
	GetAll(c context.Context) ([]Sheet, error)
	Delete(c context.Context, id string) error
	GetByUserID(c context.Context, userID string) ([]Sheet, error)
	GetByID(c context.Context, id string) (Sheet, error)
}
