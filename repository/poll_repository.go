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

	count, err := collection.DeleteOne(ctx, bson.M{"_id": hexID})
	if err != nil {
		return err
	}

	if count == 0 {
		return domain.ErrPollNotFound
	}

	return nil
}

func (pr *pollRepository) DeleteBySheetID(ctx context.Context, sheetID string) error {
	collection := pr.database.Collection(pr.collection)

	hexID, err := primitive.ObjectIDFromHex(sheetID)
	if err != nil {
		return err
	}

	for {
		deleted, err := collection.DeleteOne(ctx, bson.M{"sheetID": hexID})
		if err != nil {
			return err
		}

		if deleted == 0 {
			return nil
		}
	}
}

func (pr *pollRepository) Create(ctx context.Context, poll *domain.Poll) error {
	collection := pr.database.Collection(pr.collection)
	_, err := collection.InsertOne(ctx, poll)
	return err
}

func (pr *pollRepository) GetByID(ctx context.Context, id string) (domain.Poll, error) {
	collection := pr.database.Collection(pr.collection)

	var poll domain.Poll
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return poll, err
	}

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&poll)
	return poll, err
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

	update := bson.M{
		"$set": bson.M{
			"title":       poll.Title,
			"description": poll.Description,
			"options":     poll.Options,
			"pollType":    poll.PollType,
			"category":    poll.Category,
			"updatedAt":   time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update poll: %v", err)
	}

	return nil

}

func (pr *pollRepository) AppendOpinionResponse(ctx context.Context, id string, responses []string) error {
	collection := pr.database.Collection(pr.collection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	if len(responses) == 0 {
		return domain.ErrNoOpinionSubmitted
	}

	update := bson.M{
		"$push": bson.M{
			"responses": bson.M{
				"$each": responses,
			},
		},
		"$inc": bson.M{
			"participant": 1,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
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
		if voteCount == 0 {
			continue
		}
		updateDoc[fmt.Sprintf("votes.%d", i)] = voteCount
	}

	if len(updateDoc) == 0 {
		return domain.ErrNoVotesSubmitted
	}

	updateDoc["participant"] = 1

	update := bson.M{
		"$inc": updateDoc,
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
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
