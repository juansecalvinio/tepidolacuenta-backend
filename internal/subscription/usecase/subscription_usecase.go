package usecase

import (
	"context"
	"errors"
	"time"

	"juansecalvinio/tepidolacuenta/internal/pkg"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	"juansecalvinio/tepidolacuenta/internal/subscription/domain"
	"juansecalvinio/tepidolacuenta/internal/subscription/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UseCase defines the interface for subscription use cases
type UseCase interface {
	// Plan operations
	GetAllPlans(ctx context.Context) ([]*domain.Plan, error)
	GetPlanByID(ctx context.Context, id primitive.ObjectID) (*domain.Plan, error)

	// Subscription operations
	Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateSubscriptionInput) (*domain.Subscription, error)
	GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.SubscriptionWithPlan, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Subscription, error)
	GetByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID) (*domain.SubscriptionWithPlan, error)
	Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateSubscriptionInput) (*domain.Subscription, error)
	Cancel(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error

	// EnsureComped garantiza que un user cortesía tenga suscripción activa (ver método).
	EnsureComped(ctx context.Context, userID, restaurantID primitive.ObjectID) error
}

type subscriptionUseCase struct {
	planRepo         repository.PlanRepository
	subscriptionRepo repository.SubscriptionRepository
	restaurantRepo   restaurantRepo.Repository
}

// NewSubscriptionUseCase creates a new subscription use case
func NewSubscriptionUseCase(
	planRepo repository.PlanRepository,
	subscriptionRepo repository.SubscriptionRepository,
	restaurantRepo restaurantRepo.Repository,
) UseCase {
	return &subscriptionUseCase{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		restaurantRepo:   restaurantRepo,
	}
}

// GetAllPlans retrieves all available plans
func (uc *subscriptionUseCase) GetAllPlans(ctx context.Context) ([]*domain.Plan, error) {
	return uc.planRepo.FindAll(ctx)
}

// GetPlanByID retrieves a plan by its ID
func (uc *subscriptionUseCase) GetPlanByID(ctx context.Context, id primitive.ObjectID) (*domain.Plan, error) {
	return uc.planRepo.FindByID(ctx, id)
}

// Create creates a new subscription for a restaurant
func (uc *subscriptionUseCase) Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateSubscriptionInput) (*domain.Subscription, error) {
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

	// Verify restaurant doesn't already have an active subscription
	existing, err := uc.subscriptionRepo.FindByRestaurantID(ctx, restaurantID)
	if err != nil && !errors.Is(err, pkg.ErrNotFound) {
		return nil, err
	}
	if existing != nil && existing.Status != domain.SubscriptionStatusCanceled && existing.Status != domain.SubscriptionStatusExpired {
		return nil, pkg.ErrSubscriptionAlreadyExists
	}

	// Parse plan ID
	planID, err := primitive.ObjectIDFromHex(input.PlanID)
	if err != nil {
		return nil, errors.New("invalid plan ID")
	}

	// Verify plan exists
	plan, err := uc.planRepo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Determine initial status
	status := domain.SubscriptionStatusActive
	if input.StartTrial {
		status = domain.SubscriptionStatusTrialing
	}

	subscription := domain.NewSubscription(userID, restaurantID, planID, status)
	subscription.PaymentSubscriptionID = input.PaymentSubscriptionID
	subscription.PurchasedBranches = plan.IncludedBranches

	// Set trial period using the plan's configured trial days
	if input.StartTrial && plan.TrialDays > 0 {
		now := time.Now()
		trialEnds := now.AddDate(0, 0, plan.TrialDays)
		subscription.TrialStartedAt = &now
		subscription.TrialEndsAt = &trialEnds
	}

	if err := uc.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetByID retrieves a subscription by ID with its associated plan embedded
func (uc *subscriptionUseCase) GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.SubscriptionWithPlan, error) {
	subscription, err := uc.subscriptionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if subscription.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	plan, err := uc.planRepo.FindByID(ctx, subscription.PlanID)
	if err != nil {
		return nil, err
	}

	return &domain.SubscriptionWithPlan{
		ID:                    subscription.ID,
		UserID:                subscription.UserID,
		RestaurantID:          subscription.RestaurantID,
		PlanID:                subscription.PlanID,
		Plan:                  plan,
		Status:                subscription.Status,
		TrialStartedAt:        subscription.TrialStartedAt,
		TrialEndsAt:           subscription.TrialEndsAt,
		PaymentSubscriptionID: subscription.PaymentSubscriptionID,
		CreatedAt:             subscription.CreatedAt,
		UpdatedAt:             subscription.UpdatedAt,
	}, nil
}

// GetByUserID retrieves all subscriptions for a user
func (uc *subscriptionUseCase) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Subscription, error) {
	return uc.subscriptionRepo.FindByUserID(ctx, userID)
}

// GetByRestaurantID retrieves the active subscription for a restaurant
func (uc *subscriptionUseCase) GetByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID) (*domain.SubscriptionWithPlan, error) {
	// Verify restaurant exists and belongs to user
	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	subscription, err := uc.subscriptionRepo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	plan, err := uc.planRepo.FindByID(ctx, subscription.PlanID)
	if err != nil {
		return nil, err
	}

	return &domain.SubscriptionWithPlan{
		ID:                    subscription.ID,
		UserID:                subscription.UserID,
		RestaurantID:          subscription.RestaurantID,
		PlanID:                subscription.PlanID,
		Plan:                  plan,
		Status:                subscription.Status,
		TrialStartedAt:        subscription.TrialStartedAt,
		TrialEndsAt:           subscription.TrialEndsAt,
		PaymentSubscriptionID: subscription.PaymentSubscriptionID,
		CreatedAt:             subscription.CreatedAt,
		UpdatedAt:             subscription.UpdatedAt,
	}, nil
}

// Update updates a subscription's plan, status, or payment details
func (uc *subscriptionUseCase) Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateSubscriptionInput) (*domain.Subscription, error) {
	subscription, err := uc.subscriptionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if subscription.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	if input.PlanID != "" {
		planID, err := primitive.ObjectIDFromHex(input.PlanID)
		if err != nil {
			return nil, errors.New("invalid plan ID")
		}
		if _, err := uc.planRepo.FindByID(ctx, planID); err != nil {
			return nil, err
		}
		subscription.PlanID = planID
	}

	if input.Status != "" {
		subscription.Status = input.Status
	}

	if input.PaymentSubscriptionID != "" {
		subscription.PaymentSubscriptionID = input.PaymentSubscriptionID
	}

	if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

// Cancel cancels a subscription
func (uc *subscriptionUseCase) Cancel(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	subscription, err := uc.subscriptionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if subscription.UserID != userID {
		return pkg.ErrUnauthorized
	}

	subscription.Status = domain.SubscriptionStatusCanceled

	return uc.subscriptionRepo.Update(ctx, subscription)
}

// EnsureComped garantiza que un usuario cortesía tenga una suscripción
// perpetuamente activa en el plan Business para su restaurante, de modo que el
// paywall (que vive sólo en el frontend) lo deje pasar sin pasar por selección
// de plan ni pago. Es idempotente: si ya está activa/Business/holgada, no hace
// nada. Está pensada para correr en cada login de un usuario marcado como
// comped en la DB. Si no existe suscripción para el restaurante, crea una.
func (uc *subscriptionUseCase) EnsureComped(ctx context.Context, userID, restaurantID primitive.ObjectID) error {
	businessPlan, err := uc.findBusinessPlan(ctx)
	if err != nil {
		return err
	}

	sub, err := uc.subscriptionRepo.FindByRestaurantID(ctx, restaurantID)
	if err != nil {
		if !errors.Is(err, pkg.ErrNotFound) {
			return err
		}
		// Sin suscripción previa: creamos una cortesía activa.
		sub = domain.NewSubscription(userID, restaurantID, businessPlan.ID, domain.SubscriptionStatusActive)
		sub.ApplyComped(businessPlan.ID)
		return uc.subscriptionRepo.Create(ctx, sub)
	}

	if sub.IsCompedActive(businessPlan.ID) {
		return nil
	}

	sub.ApplyComped(businessPlan.ID)
	return uc.subscriptionRepo.Update(ctx, sub)
}

// findBusinessPlan devuelve el plan Business (el más alto), que es el que se le
// asigna a las cuentas cortesía.
func (uc *subscriptionUseCase) findBusinessPlan(ctx context.Context) (*domain.Plan, error) {
	plans, err := uc.planRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if p.Name == domain.PlanNameBusiness {
			return p, nil
		}
	}
	return nil, pkg.ErrNotFound
}
