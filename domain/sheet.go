package domain

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const CollectionSheet = "sheets"

type Sheet struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      primitive.ObjectID `bson:"userID"`
	Description string             `bson:"description"`
	CreatedAt   time.Time          `bson:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"`
}

type SheetRepository interface {
	Create(ctx context.Context, sheet Sheet) error
	GetAll(ctx context.Context) ([]Sheet, error)
	Delete(ctx context.Context, id string) error
}
