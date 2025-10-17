package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionPoll = "polls"

type PollType string

const (
	singleChoice PollType = "single_choice"
	multiChoice  PollType = "multi_choice"
	slide        PollType = "slide"
	opinion      PollType = "opinion"
)

var (
	PollTypeSingleChoice = singleChoice
	PollTypeMultiChoice  = multiChoice
	PollTypeSlide        = slide
	PollTypeOpinion      = opinion
)

func ParsePollType(value string) (PollType, error) {
	switch value {
	case "", string(singleChoice):
		return singleChoice, nil
	case string(multiChoice):
		return multiChoice, nil
	case string(slide):
		return slide, nil
	case string(opinion):
		return opinion, nil
	default:
		return "", fmt.Errorf("invalid poll type: %s", value)
	}
}

func (t PollType) MinOptions() int {
	switch t {
	case slide, opinion:
		return 1
	default:
		return 2
	}
}

func (t PollType) VoteSlots(optionCount int) int {
	if optionCount < 0 {
		optionCount = 0
	}

	if t == opinion {
		return 1
	}

	return optionCount
}

type Poll struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	SheetID     primitive.ObjectID `bson:"sheetID"`
	Title       string             `bson:"title"`
	Category    []string           `bson:"category"`
	Options     []string           `bson:"options"`
	PollType    PollType           `bson:"pollType"`
	Participant int                `bson:"participant"`
	Votes       []int              `bson:"votes"`
	Responses   []string           `bson:"responses,omitempty"`
	Description string             `bson:"description"`
	CreatedAt   time.Time          `bson:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"`
}

type PollRepository interface {
	Create(ctx context.Context, poll *Poll) error
	GetPollBySheetID(ctx context.Context, sheetID string, pagination PaginationQuery) ([]Poll, int64, error)
	GetByID(ctx context.Context, id string) (Poll, error)
	EditPoll(ctx context.Context, poll *Poll) error
	SubmitVote(ctx context.Context, id string, votes []int) error
	AppendOpinionResponse(ctx context.Context, id string, responses []string) error
	Delete(ctx context.Context, id string) error
	DeleteBySheetID(ctx context.Context, sheetID string) error
}

var ErrPollNotFound = errors.New("poll not found")
