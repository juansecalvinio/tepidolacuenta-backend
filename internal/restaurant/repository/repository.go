package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/restaurant/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository defines the interface for restaurant persistence operations
type Repository interface {
	Create(ctx context.Context, restaurant *domain.Restaurant) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Restaurant, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Restaurant, error)
	Update(ctx context.Context, restaurant *domain.Restaurant) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
