package repository

import (
	"context"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type sheetRepository struct {
	database   mongo.Database
	collection string
}

func (sr *sheetRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Sheet, error) {
	collection := sr.database.Collection(sr.collection)

	UID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(ctx, bson.M{"userID": UID})
	if err != nil {
		return nil, err
	}

	var result []domain.Sheet
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	if result == nil {
		return []domain.Sheet{}, nil
	}

	return result, nil
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

func (sr *sheetRepository) GetAll(ctx context.Context) ([]domain.Sheet, error) {
	collection := sr.database.Collection(sr.collection)

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var sheets []domain.Sheet

	if err = cursor.All(ctx, &sheets); err != nil {
		return nil, err
	}

	if sheets == nil {
		return []domain.Sheet{}, nil
	}

	return sheets, nil
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
