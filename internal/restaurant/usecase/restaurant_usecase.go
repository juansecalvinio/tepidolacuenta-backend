package usecase

import (
	"context"
	"errors"

	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/restaurant/domain"
	"juansecalvinio/tepidolacuenta/internal/restaurant/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UseCase defines the interface for restaurant use cases
type UseCase interface {
	Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateRestaurantInput) (*domain.Restaurant, error)
	GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.Restaurant, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Restaurant, error)
	Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateRestaurantInput) (*domain.Restaurant, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

type restaurantUseCase struct {
	repo repository.Repository
}

// NewRestaurantUseCase creates a new restaurant use case
func NewRestaurantUseCase(repo repository.Repository) UseCase {
	return &restaurantUseCase{
		repo: repo,
	}
}

// Create creates a new restaurant
func (uc *restaurantUseCase) Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateRestaurantInput) (*domain.Restaurant, error) {
	// Validate input
	if !pkg.IsValidRestaurantName(input.Name) {
		return nil, errors.New("restaurant name must be between 3 and 100 characters")
	}

	// Create restaurant
	restaurant := domain.NewRestaurant(userID, input.Name, input.Address, input.Phone, input.Description)

	// Save to database
	if err := uc.repo.Create(ctx, restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

// GetByID retrieves a restaurant by ID
func (uc *restaurantUseCase) GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.Restaurant, error) {
	restaurant, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return restaurant, nil
}

// GetByUserID retrieves all restaurants for a user
func (uc *restaurantUseCase) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Restaurant, error) {
	return uc.repo.FindByUserID(ctx, userID)
}

// Update updates a restaurant
func (uc *restaurantUseCase) Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateRestaurantInput) (*domain.Restaurant, error) {
	// Find restaurant
	restaurant, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	// Update fields if provided
	if input.Name != "" {
		if !pkg.IsValidRestaurantName(input.Name) {
			return nil, errors.New("restaurant name must be between 3 and 100 characters")
		}
		restaurant.Name = input.Name
	}

	if input.Address != "" {
		restaurant.Address = input.Address
	}

	if input.Phone != "" {
		restaurant.Phone = input.Phone
	}

	if input.Description != "" {
		restaurant.Description = input.Description
	}

	// Save changes
	if err := uc.repo.Update(ctx, restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

// Delete deletes a restaurant
func (uc *restaurantUseCase) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	// Find restaurant
	restaurant, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify ownership
	if restaurant.UserID != userID {
		return pkg.ErrUnauthorized
	}

	return uc.repo.Delete(ctx, id)
}
