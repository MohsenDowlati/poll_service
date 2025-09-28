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

type sheetRepository struct {
	database   mongo.Database
	collection string
}

func (sr *sheetRepository) GetByUserID(ctx context.Context, userID string, pagination domain.PaginationQuery) ([]domain.Sheet, int64, error) {
	collection := sr.database.Collection(sr.collection)

	UID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"userID": UID}
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

	var result []domain.Sheet
	if err = cursor.All(ctx, &result); err != nil {
		return nil, 0, err
	}
	if result == nil {
		result = []domain.Sheet{}
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (sr *sheetRepository) GetByID(ctx context.Context, id string) (domain.Sheet, error) {
	collection := sr.database.Collection(sr.collection)

	UID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.Sheet{}, err
	}

	var result domain.Sheet
	err = collection.FindOne(ctx, bson.M{"_id": UID}).Decode(&result)
	return result, err
}

func (sr *sheetRepository) Create(ctx context.Context, sheet domain.Sheet) error {
	collection := sr.database.Collection(sr.collection)

	_, err := collection.InsertOne(ctx, sheet)
	return err
}

func (sr *sheetRepository) GetAll(ctx context.Context, pagination domain.PaginationQuery) ([]domain.Sheet, int64, error) {
	collection := sr.database.Collection(sr.collection)

	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if skip := pagination.Skip(); skip > 0 {
		findOptions.SetSkip(skip)
	}
	if limit := pagination.Limit(); limit > 0 {
		findOptions.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, 0, err
	}

	var sheets []domain.Sheet
	if err = cursor.All(ctx, &sheets); err != nil {
		return nil, 0, err
	}
	if sheets == nil {
		sheets = []domain.Sheet{}
	}

	total, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return nil, 0, err
	}

	return sheets, total, nil
}

func (sr *sheetRepository) Delete(ctx context.Context, id string) error {
	collection := sr.database.Collection(sr.collection)

	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: idHex}})
	return err
}

func (sr *sheetRepository) UpdateStatus(ctx context.Context, id string, status domain.SheetStatus, approvedBy primitive.ObjectID, approvedAt time.Time) error {
	collection := sr.database.Collection(sr.collection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"approvedBy": approvedBy,
			"approvedAt": approvedAt,
			"updatedAt":  approvedAt,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func NewSheetRepository(db mongo.Database, collection string) domain.SheetRepository {
	return &sheetRepository{
		database:   db,
		collection: collection,
	}
}
