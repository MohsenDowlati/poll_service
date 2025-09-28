package domain

type SheetListResponse struct {
	Data       []Sheet          `json:"data"`
	Pagination PaginationResult `json:"pagination"`
}

type NotificationListResponse struct {
	Data       []NotificationResponse `json:"data"`
	Pagination PaginationResult       `json:"pagination"`
}

type PollAdminListResponse struct {
	Data       []PollAdminResponse `json:"data"`
	Pagination PaginationResult    `json:"pagination"`
}

type PollClientListResponse struct {
	Data       []PollClientResponse `json:"data"`
	Pagination PaginationResult     `json:"pagination"`
}

type UserListItem struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Phone        string   `json:"phone"`
	Organization string   `json:"organization"`
	Admin        UserType `json:"admin"`
	IsVerified   bool     `json:"is_verified"`
}

type UserListResponse struct {
	Data       []UserListItem   `json:"data"`
	Pagination PaginationResult `json:"pagination"`
}
