package repository

import (
	"context"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type notificationRepository struct {
	database   mongo.Database
	collection string
}

func NewNotificationRepository(db mongo.Database, collection string) domain.NotificationRepository {
	return &notificationRepository{
		database:   db,
		collection: collection,
	}
}

func (nr *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	collection := nr.database.Collection(nr.collection)
	_, err := collection.InsertOne(ctx, notification)
	return err
}

func (nr *notificationRepository) FetchPending(ctx context.Context, pagination domain.PaginationQuery) ([]domain.Notification, int64, error) {
	collection := nr.database.Collection(nr.collection)

	filter := bson.M{"status": domain.NotificationPending}
	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if skip := pagination.Skip(); skip > 0 {
		findOptions.SetSkip(skip)
	}
	if limit := pagination.Limit(); limit > 0 {
		findOptions.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}

	var notifications []domain.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}
	if notifications == nil {
		notifications = []domain.Notification{}
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (nr *notificationRepository) GetByID(ctx context.Context, id string) (domain.Notification, error) {
	collection := nr.database.Collection(nr.collection)

	var notification domain.Notification
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return notification, err
	}

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&notification)
	return notification, err
}

func (nr *notificationRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, resolvedBy primitive.ObjectID, updatedAt time.Time) error {
	collection := nr.database.Collection(nr.collection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updatedAt":  updatedAt,
			"resolvedBy": resolvedBy,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (nr *notificationRepository) Delete(ctx context.Context, id string) error {
	collection := nr.database.Collection(nr.collection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
