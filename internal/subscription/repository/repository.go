package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/subscription/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlanRepository defines persistence operations for plans
type PlanRepository interface {
	Create(ctx context.Context, plan *domain.Plan) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Plan, error)
	FindAll(ctx context.Context) ([]*domain.Plan, error)
}

// SubscriptionRepository defines persistence operations for subscriptions
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *domain.Subscription) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Subscription, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Subscription, error)
	FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) (*domain.Subscription, error)
	Update(ctx context.Context, subscription *domain.Subscription) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
