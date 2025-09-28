package domain

import "context"

type AdminRequest struct {
	UserID     string `form:"user_id"`
	IsVerified bool   `form:"is_verified"`
}

type AdminResponse struct {
	UserID       string   `json:"user_id"`
	Name         string   `json:"name"`
	Organization string   `json:"organization"`
	Admin        UserType `json:"admin"`
}

type AdminUsecase interface {
	VerifyUser(c context.Context, userID string, isVerified bool) error
	Fetch(c context.Context, pagination PaginationQuery) ([]User, int64, error)
	Delete(c context.Context, userID string) error
}
