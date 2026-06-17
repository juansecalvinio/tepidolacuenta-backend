package repository

import (
	"context"
	"errors"
	"time"

	"juansecalvinio/tepidolacuenta/internal/payment/domain"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("payments"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, payment *domain.Payment) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, payment)
	if err != nil {
		return err
	}

	payment.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var payment domain.Payment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&payment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &payment, nil
}

func (r *mongoRepository) FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"restaurant_id": restaurantID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	payments := make([]*domain.Payment, 0)
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *mongoRepository) FindByMPPreferenceID(ctx context.Context, preferenceID string) (*domain.Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var payment domain.Payment
	err := r.collection.FindOne(ctx, bson.M{"mp_preference_id": preferenceID}).Decode(&payment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &payment, nil
}

func (r *mongoRepository) Update(ctx context.Context, payment *domain.Payment) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	payment.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"mp_preference_id": payment.MPPreferenceID,
			"mp_payment_id":    payment.MPPaymentID,
			"status":           payment.Status,
			"updated_at":       payment.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": payment.ID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return pkg.ErrNotFound
	}

	return nil
}
