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
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"userID" json:"-"`
	Title           string             `bson:"title" form:"title" binding:"required" json:"title"`
	Venue           string             `bson:"venue" form:"venue" binding:"required" json:"venue"`
	Description     string             `bson:"description" form:"description" json:"description"`
	Status          SheetStatus        `bson:"status" json:"status"`
	IsPhoneRequired bool               `bson:"isPhoneRequired" form:"is_phone_required" json:"is_phone_required"`
	ApprovedBy      primitive.ObjectID `bson:"approvedBy,omitempty" json:"approved_by,omitempty"`
	ApprovedAt      time.Time          `bson:"approvedAt,omitempty" json:"approved_at,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"-"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"-"`
}

type SheetCreatePoll struct {
	Title       string   `json:"title" form:"title"`
	Description string   `json:"description,omitempty" form:"description"`
	Options     []string `json:"options" form:"options"`
	PollType    string   `json:"poll_type" form:"poll_type"`
	Category    string   `json:"category" form:"category"`
}

type SheetCreateRequest struct {
	Title           string            `json:"title" form:"title"`
	Venue           string            `json:"venue" form:"venue"`
	IsPhoneRequired bool              `json:"is_phone_required" form:"is_phone_required"`
	Polls           []SheetCreatePoll `json:"polls" form:"polls"`
}

type SheetCreateResponse struct {
	Message string              `json:"message"`
	Sheet   Sheet               `json:"sheet"`
	Polls   []PollAdminResponse `json:"polls,omitempty"`
}

func (r SheetCreateRequest) EffectiveTitle() string {
	return r.Title
}

type SheetRepository interface {
	Create(ctx context.Context, sheet Sheet) error
	GetAll(ctx context.Context, pagination PaginationQuery) ([]Sheet, int64, error)
	GetByUserID(ctx context.Context, userID string, pagination PaginationQuery) ([]Sheet, int64, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (Sheet, error)
	UpdateStatus(ctx context.Context, id string, status SheetStatus, approvedBy primitive.ObjectID, approvedAt time.Time) error
}

type SheetUseCase interface {
	Create(c context.Context, sheet Sheet) error
	GetAll(c context.Context, pagination PaginationQuery) ([]SheetListItem, int64, error)
	Delete(c context.Context, id string) error
	GetByUserID(c context.Context, userID string, pagination PaginationQuery) ([]SheetListItem, int64, error)
	GetByID(c context.Context, id string) (Sheet, error)
	UpdateStatus(c context.Context, id string, status SheetStatus, approvedBy primitive.ObjectID, approvedAt time.Time) error
}
