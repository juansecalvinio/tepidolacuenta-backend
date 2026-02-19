package usecase

import (
	"context"
	"errors"

	"juansecalvinio/tepidolacuenta/internal/branch/domain"
	"juansecalvinio/tepidolacuenta/internal/branch/repository"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UseCase defines the interface for branch use cases
type UseCase interface {
	Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateBranchInput) (*domain.Branch, error)
	GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.Branch, error)
	GetByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.Branch, error)
	Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateBranchInput) (*domain.Branch, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

type branchUseCase struct {
	repo           repository.Repository
	restaurantRepo restaurantRepo.Repository
}

// NewBranchUseCase creates a new branch use case
func NewBranchUseCase(repo repository.Repository, restaurantRepo restaurantRepo.Repository) UseCase {
	return &branchUseCase{
		repo:           repo,
		restaurantRepo: restaurantRepo,
	}
}

// Create creates a new branch
func (uc *branchUseCase) Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateBranchInput) (*domain.Branch, error) {
	// Parse restaurant ID
	restaurantID, err := primitive.ObjectIDFromHex(input.RestaurantID)
	if err != nil {
		return nil, errors.New("invalid restaurant ID")
	}

	// Verify restaurant exists and belongs to user
	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	// Create branch
	branch := domain.NewBranch(restaurantID, input.Address, input.Description)

	// Save to database
	if err := uc.repo.Create(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

// GetByID retrieves a branch by ID
func (uc *branchUseCase) GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.Branch, error) {
	branch, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership via restaurant
	restaurant, err := uc.restaurantRepo.FindByID(ctx, branch.RestaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return branch, nil
}

// GetByRestaurantID retrieves all branches for a restaurant
func (uc *branchUseCase) GetByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.Branch, error) {
	// Verify ownership
	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return uc.repo.FindByRestaurantID(ctx, restaurantID)
}

// Update updates a branch
func (uc *branchUseCase) Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateBranchInput) (*domain.Branch, error) {
	// Find branch
	branch, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership via restaurant
	restaurant, err := uc.restaurantRepo.FindByID(ctx, branch.RestaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	if input.Address != "" {
		branch.Address = input.Address
	}

	if input.Description != "" {
		branch.Description = input.Description
	}

	if input.IsActive != nil {
		branch.IsActive = *input.IsActive
	}

	// Save changes
	if err := uc.repo.Update(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

// Delete deletes a branch
func (uc *branchUseCase) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	// Find branch
	branch, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify ownership via restaurant
	restaurant, err := uc.restaurantRepo.FindByID(ctx, branch.RestaurantID)
	if err != nil {
		return err
	}

	if restaurant.UserID != userID {
		return pkg.ErrUnauthorized
	}

	return uc.repo.Delete(ctx, id)
}
