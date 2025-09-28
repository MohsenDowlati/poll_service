package domain

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionPoll = "polls"

type pollType string

const (
	singleChoice pollType = "single_choice"
	multiChoice  pollType = "multi_choice"
	slide        pollType = "slide"
)

func ParsePollType(value string) (pollType, error) {
	switch value {
	case "", string(singleChoice):
		return singleChoice, nil
	case string(multiChoice):
		return multiChoice, nil
	case string(slide):
		return slide, nil
	default:
		return "", fmt.Errorf("invalid poll type: %s", value)
	}
}

type Poll struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	SheetID     primitive.ObjectID `bson:"sheetID"`
	Title       string             `bson:"title"`
	Options     []string           `bson:"options"`
	PollType    pollType           `bson:"pollType"`
	Participant int                `bson:"participant"`
	Votes       []int              `bson:"votes"`
	Description string             `bson:"description"`
	CreatedAt   time.Time          `bson:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"`
}

type PollRepository interface {
	Create(ctx context.Context, poll *Poll) error
	GetPollBySheetID(ctx context.Context, sheetID string, pagination PaginationQuery) ([]Poll, int64, error)
	EditPoll(ctx context.Context, poll *Poll) error
	SubmitVote(ctx context.Context, id string, votes []int) error
	Delete(ctx context.Context, id string) error
}
