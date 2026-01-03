package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/auth/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository defines the interface for user persistence operations
type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
