package repository

import (
	"context"
	"fmt"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type pollRepository struct {
	database   mongo.Database
	collection string
}

func (pr *pollRepository) Create(ctx context.Context, poll *domain.Poll) error {
	collection := pr.database.Collection(pr.collection)
	_, err := collection.InsertOne(ctx, poll)
	return err
}

func (pr *pollRepository) GetPollBySheetID(ctx context.Context, sheetID string) (poll []domain.Poll, err error) {
	collection := pr.database.Collection(pr.collection)
	var polls []domain.Poll

	idHex, err := primitive.ObjectIDFromHex(sheetID)
	if err != nil {
		return polls, err
	}

	cursor, err := collection.Find(ctx, bson.D{{"sheetID", idHex}})

	err = cursor.All(ctx, &polls)

	if polls == nil {
		return []domain.Poll{}, nil
	}
	return polls, err
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
