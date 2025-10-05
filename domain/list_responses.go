package domain

import "time"

type SheetListItem struct {
	ID        string      `json:"id"`
	UserName  string      `json:"user_name"`
	Title     string      `json:"title"`
	Venue     string      `json:"venue"`
	Status    SheetStatus `json:"status"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type SheetListResponse struct {
	Data       []SheetListItem  `json:"data"`
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
	Sheet      PollClientSheetMeta  `json:"sheet"`
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
