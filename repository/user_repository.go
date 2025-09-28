package repository

import (
	"context"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepository struct {
	database   mongo.Database
	collection string
}

func (ur *userRepository) DeleteUser(c context.Context, id string) error {
	collection := ur.database.Collection(ur.collection)
	_, err := collection.DeleteOne(c, bson.M{"_id": id})
	return err
}

func (ur *userRepository) GetByPhone(c context.Context, phone string) (domain.User, error) {
	collection := ur.database.Collection(ur.collection)
	var user domain.User
	err := collection.FindOne(c, bson.M{"phone": phone}).Decode(&user)
	return user, err

}

func (ur *userRepository) VerifyUser(c context.Context, id string) error {
	collection := ur.database.Collection(ur.collection)

	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": idHex}
	update := bson.M{"$set": bson.M{"isVerified": true}}

	_, err = collection.UpdateOne(c, filter, update, options.Update().SetUpsert(true))
	return err
}

func NewUserRepository(db mongo.Database, collection string) domain.UserRepository {
	return &userRepository{
		database:   db,
		collection: collection,
	}
}

func (ur *userRepository) Create(c context.Context, user *domain.User) error {
	collection := ur.database.Collection(ur.collection)

	_, err := collection.InsertOne(c, user)

	return err
}

func (ur *userRepository) Fetch(c context.Context, pagination domain.PaginationQuery) ([]domain.User, int64, error) {
	collection := ur.database.Collection(ur.collection)

	findOptions := options.Find().SetProjection(bson.D{{Key: "password", Value: 0}})
	if skip := pagination.Skip(); skip > 0 {
		findOptions.SetSkip(skip)
	}
	if limit := pagination.Limit(); limit > 0 {
		findOptions.SetLimit(limit)
	}

	cursor, err := collection.Find(c, bson.D{}, findOptions)
	if err != nil {
		return nil, 0, err
	}

	var users []domain.User
	if err = cursor.All(c, &users); err != nil {
		return nil, 0, err
	}
	if users == nil {
		users = []domain.User{}
	}

	total, err := collection.CountDocuments(c, bson.D{})
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (ur *userRepository) GetByEmail(c context.Context, email string) (domain.User, error) {
	collection := ur.database.Collection(ur.collection)
	var user domain.User
	err := collection.FindOne(c, bson.M{"email": email}).Decode(&user)
	return user, err
}

func (ur *userRepository) GetByID(c context.Context, id string) (domain.User, error) {
	collection := ur.database.Collection(ur.collection)

	var user domain.User

	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return user, err
	}

	err = collection.FindOne(c, bson.M{"_id": idHex}).Decode(&user)
	return user, err
}
func (ur *userRepository) UpdateAdminStatus(c context.Context, id string, admin domain.UserType, isVerified bool) error {
	collection := ur.database.Collection(ur.collection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"admin":      admin,
			"isVerified": isVerified,
		},
	}

	_, err = collection.UpdateOne(c, bson.M{"_id": objectID}, update)
	return err
}
