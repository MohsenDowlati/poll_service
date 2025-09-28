package repository

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type pollRepository struct {
	database   mongo.Database
	collection string
}

func (pr *pollRepository) Delete(ctx context.Context, id string) error {
	collection := pr.database.Collection(pr.collection)

	hexID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": hexID})
	return err
}

func (pr *pollRepository) Create(ctx context.Context, poll *domain.Poll) error {
	collection := pr.database.Collection(pr.collection)
	_, err := collection.InsertOne(ctx, poll)
	return err
}

func (pr *pollRepository) GetPollBySheetID(ctx context.Context, sheetID string, pagination domain.PaginationQuery) ([]domain.Poll, int64, error) {
	collection := pr.database.Collection(pr.collection)

	idHex, err := primitive.ObjectIDFromHex(sheetID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"sheetID": idHex}
	findOptions := options.Find()
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

	var polls []domain.Poll
	if err = cursor.All(ctx, &polls); err != nil {
		return nil, 0, err
	}
	if polls == nil {
		polls = []domain.Poll{}
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return polls, total, nil
}

func (pr *pollRepository) EditPoll(ctx context.Context, poll *domain.Poll) error {
	collection := pr.database.Collection(pr.collection)

	filter := bson.M{"_id": poll.ID}

	//TODO: config update
	update := bson.M{
		"$set": bson.M{
			"title":       poll.Title,
			"description": poll.Description,
			"options":     poll.Options,
			"updatedAt":   time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update poll: %v", err)
	}

	return nil

}

func (pr *pollRepository) SubmitVote(ctx context.Context, id string, votes []int) error {
	collection := pr.database.Collection(pr.collection)

	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": idHex}

	updateDoc := bson.M{}
	for i, voteCount := range votes {
		updateDoc[fmt.Sprintf("Votes.%d", i)] = voteCount
	}

	update := bson.M{
		"$inc": updateDoc,
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func NewPollRepository(database mongo.Database, collection string) domain.PollRepository {
	return &pollRepository{
		database:   database,
		collection: collection,
	}
}
