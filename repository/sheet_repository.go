package repository

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type sheetRepository struct {
	database   mongo.Database
	collection string
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

	err = cursor.All(ctx, &sheets)
	if sheets == nil {
		return []domain.Sheet{}, err
	}

	return sheets, err
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

func NewSheetRepository(db mongo.Database, collection string) domain.SheetRepository {
	return &sheetRepository{
		database:   db,
		collection: collection,
	}
}
