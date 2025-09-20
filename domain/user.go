package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CollectionUser = "users"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id"`
	Name         string             `bson:"name"`
	Email        string             `bson:"email"`
	Phone        string             `bson:"phone"`
	Password     string             `bson:"password"`
	IsVerified   bool               `bson:"isVerified"`
	IsSuperAdmin bool               `bson:"isSuperAdmin"`
}

type UserRepository interface {
	Create(c context.Context, user *User) error
	Fetch(c context.Context) ([]User, error)
	GetByEmail(c context.Context, email string) (User, error)
	GetByID(c context.Context, id string) (User, error)
	GetByPhone(c context.Context, phone string) (User, error)
	VerifyUser(c context.Context, id string) error
}
