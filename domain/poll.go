package domain

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const CollectionPoll = "polls"

type pollType string

const (
	singleChoice pollType = "single_choice"
	multiChoice  pollType = "multi_choice"
	slide        pollType = "slide"
)

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
	GetPollBySheetID(ctx context.Context, sheetID string) (poll []Poll, err error)
	EditPoll(ctx context.Context, poll *Poll) error
	SubmitVote(ctx context.Context, id string, votes []int) error
}
