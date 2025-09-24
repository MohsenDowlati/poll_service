package repository

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type notificationRepository struct {
	database   mongo.Database
	collection string
}

func (n notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	collection := n.database.Collection(n.collection)
	_, err := collection.InsertOne(ctx, notification)
	return err
}

func (n notificationRepository) FetchPending(ctx context.Context) ([]domain.Notification, error) {
	collection := n.database.Collection(n.collection)
	cursor, err := collection.Find(ctx, bson.D{})

	if err != nil {
		return nil, err
	}

	var notifications []domain.Notification

	err = cursor.All(ctx, &notifications)
	if notifications == nil {
		return []domain.Notification{}, err
	}

	return notifications, err
}

func (n notificationRepository) GetByID(ctx context.Context, id string) (domain.Notification, error) {
	collection := n.database.Collection(n.collection)
	var notification domain.Notification

	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return notification, err
	}

	err = collection.FindOne(ctx, bson.D{{Key: "id", Value: ID}}).Decode(&notification)
	return notification, err
}

func (n notificationRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, resolvedBy primitive.ObjectID, updatedAt time.Time) error {
	//TODO implement me
	panic("implement me")
}

func NewNotificationRepository(database mongo.Database, collection string) domain.NotificationRepository {
	return &notificationRepository{
		database:   database,
		collection: collection,
	}
}
