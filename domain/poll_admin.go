package domain

import "context"

type PollAdminRequest struct {
	SheetID     string   `form:"sheet_id"`
	Title       string   `form:"title"`
	Options     []string `form:"options"`
	PollType    pollType `form:"poll_type"`
	Description string   `form:"description"`
}

type PollAdminResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Options     []string `json:"options"`
	PollType    pollType `json:"poll_type"`
	Participant int      `json:"participant"`
	Votes       []int    `json:"votes"`
	Description string   `json:"description"`
}

type PollAdminUsecase interface {
	CreatePoll(c context.Context, poll *Poll) error
	GetBySheetID(c context.Context, sheetID string) ([]Poll, error)
	EditPoll(c context.Context, poll *Poll) error
}
