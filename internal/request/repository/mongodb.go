package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/request/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("requests"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, request *domain.Request) error {
	_, err := r.collection.InsertOne(ctx, request)
	return err
}

func (r *mongoRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Request, error) {
	var request domain.Request
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}
	return &request, nil
}

func (r *mongoRepository) FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Request, error) {
	filter := bson.M{"restaurantId": restaurantID}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []*domain.Request
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *mongoRepository) FindPendingByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Request, error) {
	filter := bson.M{
		"restaurantId": restaurantID,
		"status":       domain.StatusPending,
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []*domain.Request
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *mongoRepository) Update(ctx context.Context, request *domain.Request) error {
	filter := bson.M{"_id": request.ID}
	update := bson.M{"$set": request}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return pkg.ErrNotFound
	}

	return nil
}

func (r *mongoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return pkg.ErrNotFound
	}

	return nil
}
