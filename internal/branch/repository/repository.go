package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/branch/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	Create(ctx context.Context, branch *domain.Branch) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Branch, error)
	FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Branch, error)
	Update(ctx context.Context, branch *domain.Branch) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
