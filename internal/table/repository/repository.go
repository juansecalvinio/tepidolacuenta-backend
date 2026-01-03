package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/table/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository defines the interface for table persistence operations
type Repository interface {
	Create(ctx context.Context, table *domain.Table) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Table, error)
	FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Table, error)
	FindByRestaurantAndNumber(ctx context.Context, restaurantID primitive.ObjectID, number int) (*domain.Table, error)
	Update(ctx context.Context, table *domain.Table) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
