package repository

import (
	"context"
	"errors"
	"time"

	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/restaurant/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository for restaurants
func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("restaurants"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, restaurant)
	if err != nil {
		return err
	}

	restaurant.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Restaurant, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var restaurant domain.Restaurant
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&restaurant)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &restaurant, nil
}

func (r *mongoRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Restaurant, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Initialize with empty slice instead of nil
	restaurants := make([]*domain.Restaurant, 0)
	if err := cursor.All(ctx, &restaurants); err != nil {
		return nil, err
	}

	return restaurants, nil
}

func (r *mongoRepository) Update(ctx context.Context, restaurant *domain.Restaurant) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	restaurant.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":        restaurant.Name,
			"address":     restaurant.Address,
			"phone":       restaurant.Phone,
			"description": restaurant.Description,
			"updated_at":  restaurant.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": restaurant.ID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return pkg.ErrNotFound
	}

	return nil
}

func (r *mongoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return pkg.ErrNotFound
	}

	return nil
}
