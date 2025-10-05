package domain

import "context"

type PollAdminRequest struct {
	SheetID     string   `form:"sheet_id"`
	Title       string   `form:"title"`
	Options     []string `form:"options"`
	PollType    PollType `form:"poll_type"`
	Category    string   `form:"category"`
	Description string   `form:"description"`
}

type PollAdminResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Options     []string `json:"options"`
	PollType    PollType `json:"poll_type"`
	Category    string   `json:"category"`
	Participant int      `json:"participant"`
	Votes       []int    `json:"votes"`
	Responses   []string `json:"responses,omitempty"`
	Description string   `json:"description"`
}

type PollAdminUsecase interface {
	CreatePoll(c context.Context, poll *Poll) error
	GetBySheetID(c context.Context, sheetID string, pagination PaginationQuery) ([]Poll, int64, error)
	EditPoll(c context.Context, poll *Poll) error
	Delete(c context.Context, id string) error
}
