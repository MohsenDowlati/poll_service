package repository

import (
	"context"
	"strings"
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

func (sr *sheetRepository) GetByUserID(ctx context.Context, userID string, pagination domain.PaginationQuery, filter domain.SheetListFilter) ([]domain.Sheet, int64, error) {
	collection := sr.database.Collection(sr.collection)

	UID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	query := bson.M{"userID": UID}
	applySheetFilters(query, filter)
	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if skip := pagination.Skip(); skip > 0 {
		findOptions.SetSkip(skip)
	}
	if limit := pagination.Limit(); limit > 0 {
		findOptions.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var result []domain.Sheet
	if err = cursor.All(ctx, &result); err != nil {
		return nil, 0, err
	}
	if result == nil {
		result = []domain.Sheet{}
	}

	total, err := collection.CountDocuments(ctx, query)
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

func (sr *sheetRepository) GetAll(ctx context.Context, pagination domain.PaginationQuery, filter domain.SheetListFilter) ([]domain.Sheet, int64, error) {
	collection := sr.database.Collection(sr.collection)

	query := bson.M{}
	if len(filter.OwnerIDs) > 0 {
		query["userID"] = bson.M{"$in": filter.OwnerIDs}
	}
	applySheetFilters(query, filter)

	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if skip := pagination.Skip(); skip > 0 {
		findOptions.SetSkip(skip)
	}
	if limit := pagination.Limit(); limit > 0 {
		findOptions.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var sheets []domain.Sheet
	if err = cursor.All(ctx, &sheets); err != nil {
		return nil, 0, err
	}
	if sheets == nil {
		sheets = []domain.Sheet{}
	}

	total, err := collection.CountDocuments(ctx, query)
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
	deleted, err := collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: idHex}})
	if err != nil {
		return err
	}
	if deleted == 0 {
		return domain.ErrSheetNotFound
	}
	return nil
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

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}
	if result == nil || result.MatchedCount == 0 {
		return domain.ErrSheetNotFound
	}
	return nil
}

func applySheetFilters(criteria bson.M, filter domain.SheetListFilter) {
	if criteria == nil {
		return
	}

	if len(filter.Statuses) > 0 {
		statuses := make([]domain.SheetStatus, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			if status == "" {
				continue
			}
			statuses = append(statuses, status)
		}

		if len(statuses) > 0 {
			criteria["status"] = bson.M{"$in": statuses}
		}
	}

	if venue := strings.TrimSpace(filter.Venue); venue != "" {
		criteria["venue"] = primitive.Regex{Pattern: venue, Options: "i"}
	}

	dateRange := bson.M{}
	if filter.DateFrom != nil {
		dateRange["$gte"] = filter.DateFrom.UTC()
	}
	if filter.DateTo != nil {
		dateRange["$lte"] = filter.DateTo.UTC()
	}
	if len(dateRange) > 0 {
		criteria["createdAt"] = dateRange
	}
}

func NewSheetRepository(db mongo.Database, collection string) domain.SheetRepository {
	return &sheetRepository{
		database:   db,
		collection: collection,
	}
}
