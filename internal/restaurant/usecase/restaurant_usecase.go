package usecase

import (
	"context"
	"log"
	"time"

	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/restaurant/domain"
	"juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	subscriptionDomain "juansecalvinio/tepidolacuenta/internal/subscription/domain"
	subscriptionRepo "juansecalvinio/tepidolacuenta/internal/subscription/repository"

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
	repo             repository.Repository
	planRepo         subscriptionRepo.PlanRepository
	subscriptionRepo subscriptionRepo.SubscriptionRepository
}

// NewRestaurantUseCase creates a new restaurant use case
func NewRestaurantUseCase(
	repo repository.Repository,
	planRepo subscriptionRepo.PlanRepository,
	subscriptionRepo subscriptionRepo.SubscriptionRepository,
) UseCase {
	return &restaurantUseCase{
		repo:             repo,
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// Create creates a new restaurant and automatically starts a trial subscription
func (uc *restaurantUseCase) Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateRestaurantInput) (*domain.Restaurant, error) {
	restaurant := domain.NewRestaurant(userID, input.Name, input.CUIT)

	if err := uc.repo.Create(ctx, restaurant); err != nil {
		return nil, err
	}

	if err := uc.startTrial(ctx, userID, restaurant.ID); err != nil {
		log.Printf("Warning: failed to start trial for restaurant %s: %v", restaurant.ID.Hex(), err)
	}

	return restaurant, nil
}

// startTrial creates a trialing subscription using the Basic plan
func (uc *restaurantUseCase) startTrial(ctx context.Context, userID, restaurantID primitive.ObjectID) error {
	plans, err := uc.planRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	var basicPlan *subscriptionDomain.Plan
	for _, p := range plans {
		if p.Name == subscriptionDomain.PlanNameBasico {
			basicPlan = p
			break
		}
	}
	if basicPlan == nil {
		return pkg.ErrNotFound
	}

	subscription := subscriptionDomain.NewSubscription(userID, restaurantID, basicPlan.ID, subscriptionDomain.SubscriptionStatusTrialing)

	now := time.Now()
	trialEnds := now.AddDate(0, 0, basicPlan.TrialDays)
	subscription.TrialStartedAt = &now
	subscription.TrialEndsAt = &trialEnds

	return uc.subscriptionRepo.Create(ctx, subscription)
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
		restaurant.Name = input.Name
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
