package domain

import "context"

type PollClientRequest struct {
	ID    string `form:"id"`
	Votes []int  `form:"votes"`
}

type PollClientResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Options     []string `json:"options"`
	PollType    pollType `json:"poll_type"`
	Description string   `json:"description"`
}

type PollClientUsecase interface {
	GetBySheetID(c context.Context, sheetID string) ([]Poll, error)
	SubmitVote(c context.Context, id string, votes []int) error
}
