package repository

import (
	"context"
	"time"

	"juansecalvinio/tepidolacuenta/internal/auth/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository defines the interface for user persistence operations
type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
	LinkGoogleID(ctx context.Context, id primitive.ObjectID, googleID string) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SaveResetToken(ctx context.Context, id primitive.ObjectID, token string, expiry time.Time) error
	FindByResetToken(ctx context.Context, token string) (*domain.User, error)
	UpdatePasswordAndClearToken(ctx context.Context, id primitive.ObjectID, hashedPassword string) error
}
