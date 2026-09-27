package domain

import "context"

type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   string `json:"age"`
	Admin UserType `json:"admin"`
	IsVerified bool `json:"is_verified"`
}

type ProfileUsecase interface {
	GetProfileByID(c context.Context, userID string) (*Profile, error)
}
