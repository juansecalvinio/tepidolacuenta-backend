package repository

import (
	"context"
	"errors"
	"time"

	"juansecalvinio/tepidolacuenta/internal/branch/domain"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("branches"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, branch *domain.Branch) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, branch)
	if err != nil {
		return err
	}

	branch.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var branch domain.Branch
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&branch)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &branch, nil
}

func (r *mongoRepository) FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"restaurant_id": restaurantID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	branches := make([]*domain.Branch, 0)
	if err := cursor.All(ctx, &branches); err != nil {
		return nil, err
	}

	return branches, nil
}

func (r *mongoRepository) Update(ctx context.Context, branch *domain.Branch) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	branch.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"address":     branch.Address,
			"description": branch.Description,
			"is_active":   branch.IsActive,
			"updated_at":  branch.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": branch.ID}, update)
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
