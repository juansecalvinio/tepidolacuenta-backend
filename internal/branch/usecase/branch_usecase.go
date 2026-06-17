package usecase

import (
	"context"
	"errors"

	"juansecalvinio/tepidolacuenta/internal/branch/domain"
	"juansecalvinio/tepidolacuenta/internal/branch/repository"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	subscriptionRepo "juansecalvinio/tepidolacuenta/internal/subscription/repository"
	subscriptionDomain "juansecalvinio/tepidolacuenta/internal/subscription/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UseCase defines the interface for branch use cases
type UseCase interface {
	Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateBranchInput) (*domain.Branch, error)
	GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, restaurantIDHint *primitive.ObjectID) (*domain.Branch, error)
	GetByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID, restaurantIDHint *primitive.ObjectID) ([]*domain.Branch, error)
	Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateBranchInput) (*domain.Branch, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

type branchUseCase struct {
	repo             repository.Repository
	restaurantRepo   restaurantRepo.Repository
	subscriptionRepo subscriptionRepo.SubscriptionRepository
	planRepo         subscriptionRepo.PlanRepository
}

// NewBranchUseCase creates a new branch use case
func NewBranchUseCase(
	repo repository.Repository,
	restaurantRepo restaurantRepo.Repository,
	subscriptionRepo subscriptionRepo.SubscriptionRepository,
	planRepo subscriptionRepo.PlanRepository,
) UseCase {
	return &branchUseCase{
		repo:             repo,
		restaurantRepo:   restaurantRepo,
		subscriptionRepo: subscriptionRepo,
		planRepo:         planRepo,
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

	// Validate plan branch limit
	if err := uc.checkBranchPlanLimit(ctx, restaurantID); err != nil {
		return nil, err
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
func (uc *branchUseCase) GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, restaurantIDHint *primitive.ObjectID) (*domain.Branch, error) {
	branch, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	restaurant, err := uc.restaurantRepo.FindByID(ctx, branch.RestaurantID)
	if err != nil {
		return nil, err
	}

	if err := authorizeRestaurantAccess(restaurant.ID, restaurant.UserID, userID, restaurantIDHint); err != nil {
		return nil, err
	}

	return branch, nil
}

// GetByRestaurantID retrieves all branches for a restaurant
func (uc *branchUseCase) GetByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID, restaurantIDHint *primitive.ObjectID) ([]*domain.Branch, error) {
	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if err := authorizeRestaurantAccess(restaurant.ID, restaurant.UserID, userID, restaurantIDHint); err != nil {
		return nil, err
	}

	return uc.repo.FindByRestaurantID(ctx, restaurantID)
}

// authorizeRestaurantAccess checks whether the caller can access the given restaurant.
// Employees carry their restaurantID in the JWT (hint); owners are matched by UserID.
func authorizeRestaurantAccess(restaurantID, ownerUserID, callerUserID primitive.ObjectID, hint *primitive.ObjectID) error {
	if hint != nil {
		if restaurantID != *hint {
			return pkg.ErrForbidden
		}
		return nil
	}
	if ownerUserID != callerUserID {
		return pkg.ErrUnauthorized
	}
	return nil
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

// checkBranchPlanLimit verifies the restaurant's plan allows creating another branch
func (uc *branchUseCase) checkBranchPlanLimit(ctx context.Context, restaurantID primitive.ObjectID) error {
	subscription, err := uc.subscriptionRepo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		// No subscription found — block creation
		return pkg.ErrPlanLimitReached
	}

	plan, err := uc.planRepo.FindByID(ctx, subscription.PlanID)
	if err != nil {
		return err
	}

	// Unlimited branches
	if plan.MaxBranches == subscriptionDomain.Unlimited {
		return nil
	}

	existing, err := uc.repo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		return err
	}

	if len(existing) >= plan.MaxBranches {
		return pkg.ErrPlanLimitReached
	}

	return nil
}
