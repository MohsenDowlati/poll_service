package domain

import "errors"

var ErrCannotModifySuperAdmin = errors.New("cannot modify super admin user")

type ErrorResponse struct {
	Message string `json:"message"`
}
