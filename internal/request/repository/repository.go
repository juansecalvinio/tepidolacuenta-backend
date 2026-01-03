package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/request/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository defines the interface for request persistence
type Repository interface {
	Create(ctx context.Context, request *domain.Request) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Request, error)
	FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Request, error)
	FindPendingByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Request, error)
	Update(ctx context.Context, request *domain.Request) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
