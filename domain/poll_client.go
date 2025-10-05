package domain

import (
	"context"
	"errors"
)

var (
	ErrNoVotesSubmitted   = errors.New("no votes submitted")
	ErrNoOpinionSubmitted = errors.New("no opinion submitted")
)

type PollClientRequest struct {
	ID     string   `json:"id" form:"id"`
	Votes  []int    `json:"votes" form:"votes"`
	Inputs []string `json:"inputs" form:"inputs"`
}

type PollClientResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Options     []string `json:"options"`
	PollType    PollType `json:"poll_type"`
	Description string   `json:"description"`
}

type PollClientSheetMeta struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	IsPhoneRequired bool   `json:"is_phone_required"`
}

type PollClientUsecase interface {
	GetBySheetID(c context.Context, sheetID string, pagination PaginationQuery) ([]Poll, int64, error)
	GetSheet(c context.Context, sheetID string) (Sheet, error)
	SubmitVote(c context.Context, payload PollClientRequest) error
}
