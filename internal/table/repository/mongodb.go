package repository

import (
	"context"
	"errors"
	"time"

	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/table/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository for tables
func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("tables"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, table *domain.Table) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, table)
	if err != nil {
		return err
	}

	table.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Table, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var table domain.Table
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&table)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &table, nil
}

func (r *mongoRepository) FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Table, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"restaurant_id": restaurantID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Initialize with empty slice instead of nil
	tables := make([]*domain.Table, 0)
	if err := cursor.All(ctx, &tables); err != nil {
		return nil, err
	}

	return tables, nil
}

func (r *mongoRepository) FindByRestaurantAndNumber(ctx context.Context, restaurantID primitive.ObjectID, number int) (*domain.Table, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var table domain.Table
	err := r.collection.FindOne(ctx, bson.M{
		"restaurant_id": restaurantID,
		"number":        number,
	}).Decode(&table)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No error, just not found
		}
		return nil, err
	}

	return &table, nil
}

func (r *mongoRepository) Update(ctx context.Context, table *domain.Table) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	table.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"number":     table.Number,
			"capacity":   table.Capacity,
			"qr_code":    table.QRCode,
			"is_active":  table.IsActive,
			"updated_at": table.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": table.ID}, update)
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
