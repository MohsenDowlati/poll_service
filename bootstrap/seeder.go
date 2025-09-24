package bootstrap

import (
	"context"
	"errors"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func SeedSuperAdmin(env *Env, db mongo.Database, timeout time.Duration) error {
	if env.SuperAdminPhone == "" || env.SuperAdminPassword == "" {
		return nil
	}

	userRepository := repository.NewUserRepository(db, domain.CollectionUser)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	user, err := userRepository.GetByPhone(ctx, env.SuperAdminPhone)
	if err == nil {
		if user.Admin != domain.SuperAdmin || !user.IsVerified {
			if err := userRepository.UpdateAdminStatus(ctx, user.ID.Hex(), domain.SuperAdmin, true); err != nil {
				return err
			}
		}
		return nil
	}

	if !errors.Is(err, mongodriver.ErrNoDocuments) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(env.SuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	superAdmin := domain.User{
		ID:           primitive.NewObjectID(),
		Name:         env.SuperAdminName,
		Email:        env.SuperAdminEmail,
		Phone:        env.SuperAdminPhone,
		Password:     string(hashedPassword),
		Organization: env.SuperAdminOrganization,
		Admin:        domain.SuperAdmin,
		IsVerified:   true,
	}

	return userRepository.Create(ctx, &superAdmin)
}
