package repository

import (
	"context"
	"errors"
	"time"

	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/subscription/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// --- Plan MongoDB Repository ---

type mongoPlanRepository struct {
	collection *mongo.Collection
}

func NewMongoPlanRepository(db *mongo.Database) PlanRepository {
	return &mongoPlanRepository{
		collection: db.Collection("plans"),
	}
}

func (r *mongoPlanRepository) Create(ctx context.Context, plan *domain.Plan) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, plan)
	if err != nil {
		return err
	}

	plan.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoPlanRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var plan domain.Plan
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&plan)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &plan, nil
}

func (r *mongoPlanRepository) FindAll(ctx context.Context) ([]*domain.Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	plans := make([]*domain.Plan, 0)
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, err
	}

	return plans, nil
}

// --- Subscription MongoDB Repository ---

type mongoSubscriptionRepository struct {
	collection *mongo.Collection
}

func NewMongoSubscriptionRepository(db *mongo.Database) SubscriptionRepository {
	return &mongoSubscriptionRepository{
		collection: db.Collection("subscriptions"),
	}
}

func (r *mongoSubscriptionRepository) Create(ctx context.Context, subscription *domain.Subscription) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, subscription)
	if err != nil {
		return err
	}

	subscription.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoSubscriptionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var subscription domain.Subscription
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&subscription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *mongoSubscriptionRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	subscriptions := make([]*domain.Subscription, 0)
	if err := cursor.All(ctx, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *mongoSubscriptionRepository) FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) (*domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var subscription domain.Subscription
	err := r.collection.FindOne(ctx, bson.M{"restaurant_id": restaurantID}).Decode(&subscription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *mongoSubscriptionRepository) Update(ctx context.Context, subscription *domain.Subscription) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	subscription.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"plan_id":                 subscription.PlanID,
			"status":                  subscription.Status,
			"trial_started_at":        subscription.TrialStartedAt,
			"trial_ends_at":           subscription.TrialEndsAt,
			"payment_subscription_id": subscription.PaymentSubscriptionID,
			"updated_at":              subscription.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": subscription.ID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return pkg.ErrNotFound
	}

	return nil
}

func (r *mongoSubscriptionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
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
