package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/payment/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Payment, error)
	FindByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID) ([]*domain.Payment, error)
	FindByMPPreferenceID(ctx context.Context, preferenceID string) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
}
